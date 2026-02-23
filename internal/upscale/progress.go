package upscale

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"link-anime/internal/models"
)

// progressRegex matches FFmpeg progress lines: "frame=  123 fps=24.5 ... time=00:05:23.45"
var progressRegex = regexp.MustCompile(`frame=\s*(\d+)\s+fps=\s*([\d.]+).*time=(\d+:\d+:\d+\.\d+)`)

// scanFFmpegLines is a custom bufio.SplitFunc that handles both \r and \n line endings.
// FFmpeg uses \r for in-place progress updates on the same line.
func scanFFmpegLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Look for \r first (FFmpeg uses \r for in-place updates)
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// Fall back to \n
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// At EOF, return remaining data
	if atEOF {
		return len(data), data, nil
	}

	// Request more data
	return 0, nil, nil
}

// parseTimeToSeconds parses FFmpeg time format "HH:MM:SS.ms" to seconds.
// Example: "00:05:23.45" -> 323.45
func parseTimeToSeconds(timeStr string) float64 {
	// Split on : to get hours, minutes, seconds.ms
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}

	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0
	}

	return hours*3600 + minutes*60 + seconds
}

// parseProgress reads FFmpeg stderr output and extracts progress updates.
// It throttles updates to approximately 1 per second and calls the callback
// with each progress update.
func parseProgress(r io.Reader, totalDuration float64, jobID int64, cb ProgressCallback) {
	if cb == nil {
		// No callback, just drain the reader
		io.Copy(io.Discard, r)
		return
	}

	scanner := bufio.NewScanner(r)
	scanner.Split(scanFFmpegLines)

	var lastUpdate time.Time

	for scanner.Scan() {
		line := scanner.Text()

		matches := progressRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		// Throttle updates to ~1 per second
		if time.Since(lastUpdate) < time.Second {
			continue
		}
		lastUpdate = time.Now()

		// Extract values from regex groups
		frame, _ := strconv.Atoi(matches[1])
		fps, _ := strconv.ParseFloat(matches[2], 64)
		timeStr := matches[3]

		// Calculate percentage from time vs total duration
		currentSeconds := parseTimeToSeconds(timeStr)
		var percent float64
		if totalDuration > 0 {
			percent = (currentSeconds / totalDuration) * 100
			if percent > 100 {
				percent = 100
			}
		}

		// Call progress callback
		cb(models.UpscaleProgress{
			JobID:   jobID,
			Frame:   frame,
			FPS:     fps,
			Time:    timeStr,
			Percent: percent,
		})
	}
}

// parseProgressWithCapture is like parseProgress but also captures the last N lines
// of stderr output for error reporting when ffmpeg fails.
func parseProgressWithCapture(r io.Reader, totalDuration float64, jobID int64, cb ProgressCallback) []string {
	const maxLines = 10 // Keep last 10 lines for error context

	scanner := bufio.NewScanner(r)
	scanner.Split(scanFFmpegLines)

	var lastUpdate time.Time
	var lastLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Capture last N lines for error context
		if len(line) > 0 {
			lastLines = append(lastLines, line)
			if len(lastLines) > maxLines {
				lastLines = lastLines[1:]
			}
		}

		if cb == nil {
			continue
		}

		matches := progressRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		// Throttle updates to ~1 per second
		if time.Since(lastUpdate) < time.Second {
			continue
		}
		lastUpdate = time.Now()

		// Extract values from regex groups
		frame, _ := strconv.Atoi(matches[1])
		fps, _ := strconv.ParseFloat(matches[2], 64)
		timeStr := matches[3]

		// Calculate percentage from time vs total duration
		currentSeconds := parseTimeToSeconds(timeStr)
		var percent float64
		if totalDuration > 0 {
			percent = (currentSeconds / totalDuration) * 100
			if percent > 100 {
				percent = 100
			}
		}

		// Call progress callback
		cb(models.UpscaleProgress{
			JobID:   jobID,
			Frame:   frame,
			FPS:     fps,
			Time:    timeStr,
			Percent: percent,
		})
	}

	return lastLines
}
