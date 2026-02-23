package upscale

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"link-anime/internal/models"
)

// ProgressCallback receives progress updates during upscaling.
type ProgressCallback func(models.UpscaleProgress)

// Engine runs FFmpeg upscaling jobs with libplacebo and Anime4K shaders.
type Engine struct {
	shaderDir string
}

// NewEngine creates an upscaling engine with the given shader directory.
func NewEngine(shaderDir string) *Engine {
	return &Engine{shaderDir: shaderDir}
}

// getShaderPath returns the full path to the shader file for a given preset.
// Falls back to "balanced" if the preset is unknown.
func (e *Engine) getShaderPath(preset string) string {
	shader, ok := Presets[preset]
	if !ok {
		shader = Presets["balanced"]
	}
	// Presets already contain full paths with ShaderDir prefix
	// but we want to use our configured shaderDir
	return filepath.Join(e.shaderDir, filepath.Base(shader))
}

// buildCommand constructs the FFmpeg command for upscaling.
// Uses Vulkan for libplacebo upscaling with Anime4K shaders.
// Attempts QSV hardware encoding first, falls back to x265 CPU encoding.
func (e *Engine) buildCommand(ctx context.Context, job *models.UpscaleJob, useQSV bool) *exec.Cmd {
	shaderPath := e.getShaderPath(job.Preset)

	// Pipeline: CPU decode → Vulkan (libplacebo with Anime4K) → QSV/CPU encode (HEVC)
	//
	// We use Vulkan for GPU-accelerated upscaling with Anime4K shaders.
	// For encoding, we try QSV (Intel Quick Sync) first for 3-5x speedup,
	// falling back to x265 CPU encoding if QSV isn't available.
	//
	// Note: QSV and Vulkan use separate hardware contexts. The frames go:
	// CPU → Vulkan GPU (upscale) → CPU → QSV GPU (encode)
	// This avoids the VAAPI interop issues we encountered.

	args := []string{
		"-y", // Overwrite output
		// Initialize Vulkan device for libplacebo (auto-detect GPU)
		"-init_hw_device", "vulkan=vk",
		// Set Vulkan as filter device for libplacebo
		"-filter_hw_device", "vk",
	}

	// Add QSV device if using hardware encoding
	if useQSV {
		args = append(args, "-init_hw_device", "qsv=qs")
	}

	args = append(args, "-i", job.InputPath)

	// Video filter chain for upscaling
	// 1. format=yuv420p: Normalize input format for Vulkan
	// 2. hwupload: Upload to Vulkan GPU
	// 3. libplacebo: Apply Anime4K shader and 2x upscale (frame_mixer=none skips unneeded temporal processing)
	// 4. hwdownload: Download from Vulkan back to CPU
	// 5. format=yuv420p: Receive frames from Vulkan (required format)
	// 6. format=nv12: Convert to NV12 for QSV (x265 works with both)
	vf := fmt.Sprintf("format=yuv420p,hwupload,libplacebo=w=iw*2:h=ih*2:custom_shader_path=%s:frame_mixer=none,hwdownload,format=yuv420p,format=nv12", shaderPath)

	if useQSV {
		// QSV encoder takes NV12 software frames directly (no hwupload needed)
		// The encoder handles the upload to QSV hardware internally
		args = append(args,
			"-vf", vf,
			"-c:v", "hevc_qsv",
			"-global_quality", "22", // Similar to CRF 22
			"-preset", "medium", // QSV presets: veryfast, faster, fast, medium, slow, veryslow
		)
	} else {
		args = append(args,
			"-vf", vf,
			"-c:v", "libx265",
			"-preset", "veryfast",
			"-crf", "22",
		)
	}

	args = append(args,
		"-c:a", "copy", // Copy audio without re-encoding
		"-c:s", "copy", // Copy subtitles
		job.OutputPath,
	)

	return exec.CommandContext(ctx, "ffmpeg", args...)
}

// Run executes an upscaling job with progress reporting.
// It blocks until the job completes or the context is cancelled.
// Tries QSV hardware encoding first, falls back to x265 if QSV fails.
func (e *Engine) Run(ctx context.Context, job *models.UpscaleJob, cb ProgressCallback) error {
	// Get duration for percentage calculation
	duration, err := ProbeDuration(ctx, job.InputPath)
	if err != nil {
		return fmt.Errorf("probe duration: %w", err)
	}

	// Try QSV first, fall back to x265 if it fails
	useQSV := true
	for {
		err := e.runFFmpeg(ctx, job, duration, useQSV, cb)
		if err == nil {
			return nil
		}

		// If QSV failed and we haven't tried x265 yet, fall back
		if useQSV && !isContextError(ctx, err) {
			// Log QSV failure and retry with x265
			useQSV = false
			continue
		}

		return err
	}
}

// isContextError checks if the error is due to context cancellation
func isContextError(ctx context.Context, err error) bool {
	return ctx.Err() != nil
}

// runFFmpeg executes a single FFmpeg encoding attempt
func (e *Engine) runFFmpeg(ctx context.Context, job *models.UpscaleJob, duration float64, useQSV bool, cb ProgressCallback) error {
	cmd := e.buildCommand(ctx, job, useQSV)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}

	// Parse progress from stderr in background, capturing last lines for error context
	done := make(chan struct{})
	var lastLines []string
	go func() {
		lastLines = parseProgressWithCapture(stderr, duration, job.ID, cb)
		close(done)
	}()

	// Wait for completion
	err = cmd.Wait()
	<-done // Ensure progress parsing completes

	if err != nil {
		// Clean up partial output on error or cancel
		os.Remove(job.OutputPath)

		// Check if cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// Include last stderr lines in error for debugging
		encoder := "x265"
		if useQSV {
			encoder = "QSV"
		}
		if len(lastLines) > 0 {
			return fmt.Errorf("ffmpeg (%s): %w\nstderr: %s", encoder, err, strings.Join(lastLines, "\n"))
		}
		return fmt.Errorf("ffmpeg (%s): %w", encoder, err)
	}

	return nil
}

// GenerateOutputPath creates the output path for an upscaled video.
// It takes an input path and returns a path with "_4k.mkv" suffix.
// Example: "/downloads/anime/episode.mkv" -> "/downloads/anime/episode_4k.mkv"
// Example: "/downloads/movie.mp4" -> "/downloads/movie_4k.mkv"
func GenerateOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + "_4k.mkv"
}
