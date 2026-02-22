package upscale

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

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
func (e *Engine) buildCommand(ctx context.Context, job *models.UpscaleJob) *exec.Cmd {
	shaderPath := e.getShaderPath(job.Preset)

	args := []string{
		"-y",                        // Overwrite output
		"-init_hw_device", "vulkan", // Initialize Vulkan
		"-i", job.InputPath, // Input file
		"-vf", fmt.Sprintf("libplacebo=w=iw*2:h=ih*2:custom_shader_path=%s", shaderPath),
		"-c:v", "libx264", // Video codec
		"-crf", "18", // Quality (lower = better)
		"-preset", "medium", // Encoding speed/quality tradeoff
		"-c:a", "copy", // Copy audio without re-encoding
		job.OutputPath,
	}

	return exec.CommandContext(ctx, "ffmpeg", args...)
}

// Run executes an upscaling job with progress reporting.
// It blocks until the job completes or the context is cancelled.
func (e *Engine) Run(ctx context.Context, job *models.UpscaleJob, cb ProgressCallback) error {
	// Get duration for percentage calculation (used in Plan 02 for progress)
	_, err := ProbeDuration(ctx, job.InputPath)
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

	// Drain stderr in background (actual parsing in Plan 02)
	done := make(chan struct{})
	go func() {
		io.Copy(io.Discard, stderr)
		close(done)
	}()

	// Wait for completion
	err = cmd.Wait()
	<-done // Ensure stderr draining completes

	if err != nil {
		// Clean up partial output on error or cancel
		os.Remove(job.OutputPath)

		// Check if cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg: %w", err)
	}

	return nil
}
