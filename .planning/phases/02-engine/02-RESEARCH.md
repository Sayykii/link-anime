# Phase 2: Engine - Research

**Researched:** 2026-02-23
**Domain:** FFmpeg video processing, libplacebo upscaling, Go process management
**Confidence:** HIGH

## Summary

Phase 2 implements the FFmpeg upscaling engine using libplacebo with Anime4K shaders via Vulkan GPU acceleration. The system needs to execute FFmpeg as a subprocess, parse progress from stderr, detect input duration via ffprobe for percentage calculation, and support context-based cancellation.

FFmpeg 8.0.1 is available on the system with libplacebo and Vulkan support compiled in. Anime4K shaders are already present at `/home/desktop/.config/mpv/shaders/`. The existing codebase uses `os/exec` for FFmpeg commands (see `probe.go`), which provides the foundation for this work.

**Primary recommendation:** Build an `Engine` struct with `Run(ctx, job, progressCb)` that wraps `exec.CommandContext`, parses FFmpeg stderr for progress updates, and calls the progress callback at ~1s intervals.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| ENG-01 | FFmpeg runner with libplacebo 2x upscaling via Vulkan | Verified: FFmpeg 8.0.1 has libplacebo with `custom_shader_path` option for Anime4K shaders. Vulkan hardware acceleration available. Command pattern documented below. |
| ENG-02 | Progress parsing from FFmpeg stderr (frame, fps, time) | Verified: FFmpeg outputs `frame=N fps=X time=HH:MM:SS.ms` to stderr. Regex pattern documented. Alternative: `-progress pipe:1` for structured output. |
| ENG-03 | Duration detection via ffprobe for percentage calculation | Verified: `ffprobe -v quiet -show_entries format=duration -of default=noprint_wrappers=1:nokey=1` returns duration in seconds as float. |
| ENG-04 | Context cancellation support to kill FFmpeg process | Go stdlib: `exec.CommandContext(ctx, ...)` automatically kills process on context cancellation. Cleanup of partial output file needed. |
</phase_requirements>

## Standard Stack

### Core

| Component | Version | Purpose | Why Standard |
|-----------|---------|---------|--------------|
| FFmpeg | 8.0.1 | Video transcoding with libplacebo filter | Already installed, has libplacebo+vulkan compiled in |
| ffprobe | 8.0.1 | Video metadata extraction (duration) | Bundled with FFmpeg |
| os/exec | Go stdlib | Process execution with context | Standard Go pattern, already used in probe.go |
| context | Go stdlib | Cancellation propagation | Idiomatic Go cancellation pattern |
| regexp | Go stdlib | Stderr progress parsing | Simple, no dependencies |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| bufio | Go stdlib | Line-by-line stderr reading | For parsing progress output |
| strconv | Go stdlib | Parsing numeric values | Converting fps/frame strings |
| time | Go stdlib | Duration parsing | Converting time strings to duration |
| path/filepath | Go stdlib | Output path construction | Building `_4k.mkv` output names |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Stderr parsing | `-progress pipe:1` | Structured output but requires separate stdout handling; stderr parsing is simpler |
| regexp | strings.Split | Less flexible; progress format can vary slightly |
| exec.CommandContext | Manual signal handling | More code, less idiomatic |

## Architecture Patterns

### Recommended Project Structure

```
internal/upscale/
├── probe.go          # [EXISTS] FFmpeg/libplacebo capability detection
├── engine.go         # [NEW] FFmpeg runner with progress
├── progress.go       # [NEW] Stderr parser for progress extraction
└── ffprobe.go        # [NEW] Duration detection
```

### Pattern 1: Engine with Progress Callback

**What:** Engine struct that runs FFmpeg and reports progress via callback
**When to use:** Any long-running subprocess that needs progress reporting

```go
// Source: Go idiom for subprocess with progress
type ProgressCallback func(progress models.UpscaleProgress)

type Engine struct {
    shaderDir string
}

func NewEngine(shaderDir string) *Engine {
    return &Engine{shaderDir: shaderDir}
}

func (e *Engine) Run(ctx context.Context, job *models.UpscaleJob, cb ProgressCallback) error {
    // 1. Get duration via ffprobe
    duration, err := e.probeDuration(ctx, job.InputPath)
    if err != nil {
        return fmt.Errorf("probe duration: %w", err)
    }
    
    // 2. Build FFmpeg command
    cmd := e.buildCommand(ctx, job)
    
    // 3. Capture stderr for progress
    stderr, _ := cmd.StderrPipe()
    
    // 4. Start process
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("start ffmpeg: %w", err)
    }
    
    // 5. Parse progress in goroutine
    go e.parseProgress(stderr, duration, job.ID, cb)
    
    // 6. Wait for completion
    if err := cmd.Wait(); err != nil {
        // Clean up partial output on error
        os.Remove(job.OutputPath)
        return fmt.Errorf("ffmpeg: %w", err)
    }
    
    return nil
}
```

### Pattern 2: FFmpeg Command Construction

**What:** Build FFmpeg command with libplacebo filter and Anime4K shader
**When to use:** Constructing the upscale command based on preset

```go
func (e *Engine) buildCommand(ctx context.Context, job *models.UpscaleJob) *exec.Cmd {
    shaderPath := e.getShaderPath(job.Preset)
    
    args := []string{
        "-y",                           // Overwrite output
        "-init_hw_device", "vulkan",    // Initialize Vulkan
        "-i", job.InputPath,            // Input file
        "-vf", fmt.Sprintf("libplacebo=w=iw*2:h=ih*2:custom_shader_path=%s", shaderPath),
        "-c:v", "libx264",              // Video codec
        "-crf", "18",                   // Quality (lower = better)
        "-preset", "medium",            // Encoding speed/quality tradeoff
        "-c:a", "copy",                 // Copy audio without re-encoding
        job.OutputPath,
    }
    
    return exec.CommandContext(ctx, "ffmpeg", args...)
}
```

### Pattern 3: Progress Parsing

**What:** Parse FFmpeg stderr for progress updates
**When to use:** Real-time progress extraction from FFmpeg output

```go
// FFmpeg stderr format: frame=  123 fps=24.5 ... time=00:05:23.45 ...
var progressRegex = regexp.MustCompile(`frame=\s*(\d+)\s+fps=\s*([\d.]+).*time=(\d+:\d+:\d+\.\d+)`)

func (e *Engine) parseProgress(r io.Reader, totalDuration float64, jobID int64, cb ProgressCallback) {
    scanner := bufio.NewScanner(r)
    scanner.Split(scanFFmpegLines) // Custom split for \r-terminated lines
    
    var lastUpdate time.Time
    for scanner.Scan() {
        line := scanner.Text()
        matches := progressRegex.FindStringSubmatch(line)
        if len(matches) < 4 {
            continue
        }
        
        // Throttle updates to ~1 per second
        if time.Since(lastUpdate) < time.Second {
            continue
        }
        lastUpdate = time.Now()
        
        frame, _ := strconv.Atoi(matches[1])
        fps, _ := strconv.ParseFloat(matches[2], 64)
        timeStr := matches[3]
        currentSecs := parseTimeToSeconds(timeStr)
        percent := (currentSecs / totalDuration) * 100
        
        cb(models.UpscaleProgress{
            JobID:   jobID,
            Frame:   frame,
            FPS:     fps,
            Time:    timeStr,
            Percent: percent,
        })
    }
}

// Custom scanner for FFmpeg's \r-terminated progress lines
func scanFFmpegLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
    if atEOF && len(data) == 0 {
        return 0, nil, nil
    }
    // Look for \r (carriage return) which FFmpeg uses for in-place updates
    if i := bytes.IndexByte(data, '\r'); i >= 0 {
        return i + 1, data[0:i], nil
    }
    // Also handle \n for final output
    if i := bytes.IndexByte(data, '\n'); i >= 0 {
        return i + 1, data[0:i], nil
    }
    if atEOF {
        return len(data), data, nil
    }
    return 0, nil, nil
}
```

### Pattern 4: Duration Detection

**What:** Use ffprobe to get video duration for percentage calculation
**When to use:** Before starting FFmpeg to know total length

```go
func (e *Engine) probeDuration(ctx context.Context, inputPath string) (float64, error) {
    cmd := exec.CommandContext(ctx, "ffprobe",
        "-v", "quiet",
        "-show_entries", "format=duration",
        "-of", "default=noprint_wrappers=1:nokey=1",
        inputPath,
    )
    
    out, err := cmd.Output()
    if err != nil {
        return 0, fmt.Errorf("ffprobe: %w", err)
    }
    
    duration, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
    if err != nil {
        return 0, fmt.Errorf("parse duration: %w", err)
    }
    
    return duration, nil
}
```

### Anti-Patterns to Avoid

- **Blocking on stderr read:** Always read stderr in a goroutine to prevent deadlock
- **Not cleaning up partial output:** If FFmpeg fails or is cancelled, delete the incomplete output file
- **Hardcoding shader paths:** Use configurable shader directory (match existing Presets map in probe.go)
- **Ignoring context:** Always use `exec.CommandContext` not `exec.Command`
- **Parsing stdout instead of stderr:** FFmpeg progress goes to stderr, stdout is for piping video data

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Process execution | Custom fork/exec | `os/exec` package | Handles platform differences, signal propagation |
| Context cancellation | Manual signal sending | `exec.CommandContext` | Automatic cleanup on cancel |
| Progress regex | Character-by-character parsing | `regexp` package | More maintainable, handles edge cases |
| Duration parsing | Manual string splitting | `strconv.ParseFloat` | Handles decimal formats correctly |

**Key insight:** Go's `os/exec` with `CommandContext` handles all the complexity of subprocess management, signal handling, and cleanup. The only custom code needed is progress parsing.

## Common Pitfalls

### Pitfall 1: FFmpeg stderr buffering

**What goes wrong:** Progress updates don't arrive in real-time
**Why it happens:** FFmpeg buffers stderr when output isn't a TTY
**How to avoid:** FFmpeg auto-detects TTY; use `-stats_period 1` if needed, or accept slight delay
**Warning signs:** Progress updates arrive in bursts rather than smoothly

### Pitfall 2: Carriage return handling

**What goes wrong:** Scanner doesn't split progress lines correctly
**Why it happens:** FFmpeg uses `\r` (carriage return) for in-place progress updates, not `\n`
**How to avoid:** Custom scanner split function that handles both `\r` and `\n`
**Warning signs:** Empty lines or concatenated progress output

### Pitfall 3: Vulkan device initialization failure

**What goes wrong:** libplacebo fails to initialize Vulkan
**Why it happens:** No GPU available, driver issues, or Vulkan not supported
**How to avoid:** The existing `Probe()` function checks for Vulkan availability; check before running jobs
**Warning signs:** FFmpeg exits with "Failed to initialize Vulkan" error

### Pitfall 4: Partial output file left behind

**What goes wrong:** Failed/cancelled job leaves incomplete video file
**Why it happens:** FFmpeg creates output file immediately on start
**How to avoid:** Always `os.Remove(job.OutputPath)` in error paths
**Warning signs:** Disk space consumed by incomplete files

### Pitfall 5: Shader file not found

**What goes wrong:** libplacebo can't load Anime4K shader
**Why it happens:** Path mismatch between configured and actual shader location
**How to avoid:** Validate shader path exists before starting job; use absolute paths
**Warning signs:** FFmpeg error mentioning shader file

## Code Examples

### Complete Engine Implementation Skeleton

```go
// Source: Project-specific implementation based on Go idioms
package upscale

import (
    "bufio"
    "bytes"
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "time"
    
    "link-anime/internal/models"
)

// ProgressCallback receives progress updates during upscaling.
type ProgressCallback func(models.UpscaleProgress)

// Engine runs FFmpeg upscaling jobs.
type Engine struct {
    shaderDir string
}

// NewEngine creates an upscaling engine.
func NewEngine(shaderDir string) *Engine {
    return &Engine{shaderDir: shaderDir}
}

// Run executes an upscaling job with progress reporting.
// It blocks until the job completes or the context is cancelled.
func (e *Engine) Run(ctx context.Context, job *models.UpscaleJob, cb ProgressCallback) error {
    // Get duration for percentage calculation
    duration, err := ProbeDuration(ctx, job.InputPath)
    if err != nil {
        return fmt.Errorf("probe duration: %w", err)
    }
    
    // Build and start FFmpeg
    cmd := e.buildCommand(ctx, job)
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return fmt.Errorf("stderr pipe: %w", err)
    }
    
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("start ffmpeg: %w", err)
    }
    
    // Parse progress in background
    done := make(chan struct{})
    go func() {
        e.parseProgress(stderr, duration, job.ID, cb)
        close(done)
    }()
    
    // Wait for completion
    err = cmd.Wait()
    <-done // Ensure progress goroutine finishes
    
    if err != nil {
        // Clean up partial output
        os.Remove(job.OutputPath)
        
        // Check if cancelled
        if ctx.Err() != nil {
            return ctx.Err()
        }
        return fmt.Errorf("ffmpeg: %w", err)
    }
    
    return nil
}
```

### Preset to Shader Mapping

```go
// Source: Based on existing probe.go pattern and Anime4K documentation
func (e *Engine) getShaderPath(preset string) string {
    shaders := map[string]string{
        // Mode A: Restore -> Upscale (best for most 1080p anime)
        "fast":     "Anime4K_Upscale_CNN_x2_S.glsl",
        "balanced": "Anime4K_Upscale_CNN_x2_M.glsl",
        "quality":  "Anime4K_Upscale_CNN_x2_L.glsl",
    }
    
    shader, ok := shaders[preset]
    if !ok {
        shader = shaders["balanced"]
    }
    
    return filepath.Join(e.shaderDir, shader)
}
```

### Output Path Generation

```go
// Source: Project requirement - "_4k.mkv" suffix
func GenerateOutputPath(inputPath string) string {
    ext := filepath.Ext(inputPath)
    base := strings.TrimSuffix(inputPath, ext)
    return base + "_4k.mkv"
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| waifu2x for upscaling | Anime4K via libplacebo | 2019+ | Real-time vs minutes per frame |
| CPU-based filtering | GPU via Vulkan | FFmpeg 5.0+ | 10-100x faster |
| Fixed shader pipelines | Modular Anime4K v4 | 2021 | Customizable quality/speed |
| Manual process management | exec.CommandContext | Go 1.7+ | Automatic cleanup |

**Deprecated/outdated:**
- Anime4K v1-v3 shader syntax: v4 uses new modular format
- `-progress_url` option: Use `-progress pipe:1` or stderr parsing instead

## Open Questions

1. **Shader chaining for higher quality**
   - What we know: Anime4K supports multiple shaders (e.g., Restore -> Upscale)
   - What's unclear: How to chain multiple shaders in libplacebo `custom_shader_path`
   - Recommendation: Start with single upscale shader; investigate chaining if quality insufficient

2. **Optimal encoder settings**
   - What we know: libx264 with CRF 18 provides good quality
   - What's unclear: Whether libx265 (HEVC) provides better quality/size tradeoff
   - Recommendation: Start with libx264; make encoder configurable in v2 (ADV-01)

3. **Hardware acceleration for encoding**
   - What we know: NVENC/VAAPI can accelerate encoding
   - What's unclear: Quality vs speed tradeoff, hardware availability detection
   - Recommendation: Use software encoding for v1; investigate HW encoding for v2

## Sources

### Primary (HIGH confidence)
- FFmpeg 8.0.1 built-in help (`ffmpeg -h filter=libplacebo`, `ffmpeg -h full`)
- Verified on local system with libplacebo and Vulkan support
- Go `os/exec` package documentation

### Secondary (MEDIUM confidence)
- Anime4K GitHub repository (https://github.com/bloc97/Anime4K) - shader documentation
- Anime4K GLSL_Instructions_Advanced.md - mode explanations
- FFmpeg official documentation (https://ffmpeg.org/ffmpeg.html) - progress options

### Tertiary (LOW confidence)
- None - all claims verified against official sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - FFmpeg/libplacebo verified on system, Go stdlib well-documented
- Architecture: HIGH - Follows established Go patterns, similar to existing probe.go
- Pitfalls: HIGH - Based on FFmpeg documentation and known behavior

**Research date:** 2026-02-23
**Valid until:** 2026-04-23 (60 days - stable domain, no expected changes)
