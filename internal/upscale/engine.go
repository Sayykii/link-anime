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
// Uses Vulkan for libplacebo upscaling and VAAPI for hardware-accelerated encoding.
func (e *Engine) buildCommand(ctx context.Context, job *models.UpscaleJob) *exec.Cmd {
	shaderPath := e.getShaderPath(job.Preset)

	// Pipeline:
	// 1. Decode on CPU (most compatible)
	// 2. Upload to Vulkan GPU for libplacebo upscaling with Anime4K shaders
	// 3. Download from Vulkan to system memory
	// 4. Upload to VAAPI and encode with HEVC
	//
	// We use format=yuv420p before hwupload to VAAPI because VAAPI encoders
	// typically expect this format. The explicit @device syntax ensures each
	// hwupload goes to the correct hardware context.
	args := []string{
		"-y", // Overwrite output
		// Initialize Vulkan device for libplacebo
		"-init_hw_device", "vulkan=vk",
		// Initialize VAAPI device for encoding
		"-init_hw_device", "vaapi=va:/dev/dri/renderD128",
		// Set Vulkan as default filter device (for libplacebo)
		"-filter_hw_device", "vk",
		"-i", job.InputPath,
		// Video filter chain with explicit device binding:
		// 1. format=yuv420p - ensure compatible pixel format
		// 2. hwupload - upload to Vulkan (uses filter_hw_device=vk)
		// 3. libplacebo - upscale 2x with Anime4K shader
		// 4. hwdownload - download from Vulkan to CPU
		// 5. format=nv12 - convert to VAAPI-compatible format
		// 6. hwupload=extra_hw_frames=64 - upload to VAAPI with frame buffer
		"-vf", fmt.Sprintf("format=yuv420p,hwupload,libplacebo=w=iw*2:h=ih*2:custom_shader_path=%s,hwdownload,format=nv12,hwupload=extra_hw_frames=64:derive_device=va", shaderPath),
		// HEVC VAAPI encoder settings
		"-c:v", "hevc_vaapi",
		"-qp", "22", // Quality parameter (lower = better, 18-28 is good range)
		"-c:a", "copy", // Copy audio without re-encoding
		"-c:s", "copy", // Copy subtitles
		job.OutputPath,
	}

	return exec.CommandContext(ctx, "ffmpeg", args...)
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
		if len(lastLines) > 0 {
			return fmt.Errorf("ffmpeg: %w\nstderr: %s", err, strings.Join(lastLines, "\n"))
		}
		return fmt.Errorf("ffmpeg: %w", err)
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
