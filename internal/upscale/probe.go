package upscale

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// ShaderDir is the path where Anime4K pipeline shaders are stored.
const ShaderDir = "/app/shaders"

// Presets maps preset names to their shader file paths.
var Presets = map[string]string{
	"fast":     ShaderDir + "/mode-a-fast.glsl",
	"balanced": ShaderDir + "/mode-a-balanced.glsl",
	"quality":  ShaderDir + "/mode-a-quality.glsl",
}

// ProbeResult holds the results of the upscale capability check.
type ProbeResult struct {
	FFmpegFound  bool
	LibplaceboOK bool
	VulkanDevice string // empty if no Vulkan device detected
}

// Probe checks whether the upscale pipeline (ffmpeg + libplacebo) is available.
// It returns a ProbeResult and a nil error if ffmpeg is found (even if libplacebo
// or Vulkan are not available). It returns a non-nil error only if ffmpeg cannot
// be executed at all.
func Probe() (*ProbeResult, error) {
	result := &ProbeResult{}

	// 1. Check ffmpeg exists and can list filters
	out, err := exec.Command("ffmpeg", "-filters").CombinedOutput()
	if err != nil {
		return result, fmt.Errorf("ffmpeg not found or failed: %w", err)
	}
	result.FFmpegFound = true

	// 2. Check libplacebo filter is available
	if strings.Contains(string(out), "libplacebo") {
		result.LibplaceboOK = true
	}

	// 3. Try to detect Vulkan device (best-effort, non-fatal)
	vulkanOut, err := exec.Command("ffmpeg", "-init_hw_device", "vulkan", "-f", "lavfi", "-i",
		"nullsrc=s=64x64:d=1", "-frames:v", "1", "-f", "null", "-").CombinedOutput()
	output := string(vulkanOut)
	log.Printf("[probe] Vulkan test: err=%v outputLen=%d", err, len(output))
	if len(output) > 0 {
		// Print last 500 chars of output for debugging
		start := 0
		if len(output) > 500 {
			start = len(output) - 500
		}
		log.Printf("[probe] Vulkan output (last 500 chars): %s", output[start:])
	}
	if err == nil {
		result.VulkanDevice = "available"
	} else {
		// Check if output contains success indicators despite non-zero exit
		if strings.Contains(output, "frame=") && strings.Contains(output, "Lsize=") {
			// FFmpeg processed frames successfully - Vulkan is working
			result.VulkanDevice = "available"
		} else if strings.Contains(output, "vulkan") {
			result.VulkanDevice = "detected (init may have failed)"
		}
	}

	return result, nil
}
