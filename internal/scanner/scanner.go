package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"link-anime/internal/models"
)

// videoRe matches video file extensions.
var videoRe *regexp.Regexp

// InitVideoExtensions compiles the video extension regex.
func InitVideoExtensions(exts []string) {
	pattern := `(?i)\.(?:` + strings.Join(exts, "|") + `)$`
	videoRe = regexp.MustCompile(pattern)
}

func init() {
	// Default, overridden by InitVideoExtensions if called.
	InitVideoExtensions([]string{"mkv", "mp4", "avi"})
}

// IsVideo checks if a filename has a video extension.
func IsVideo(name string) bool {
	return videoRe.MatchString(name)
}

// ScanLibrary returns all shows in the media directory.
func ScanLibrary(mediaDir string) ([]models.Show, error) {
	entries, err := os.ReadDir(mediaDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read media dir: %w", err)
	}

	var shows []models.Show
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		showPath := filepath.Join(mediaDir, entry.Name())
		show := models.Show{
			Name: entry.Name(),
			Path: showPath,
		}

		// Scan seasons
		seasonEntries, err := os.ReadDir(showPath)
		if err != nil {
			continue
		}

		for _, se := range seasonEntries {
			if !se.IsDir() {
				continue
			}

			seasonNum := parseSeasonDir(se.Name())
			if seasonNum < 0 {
				// Count loose videos at show root that might be in non-season dirs
				continue
			}

			seasonPath := filepath.Join(showPath, se.Name())
			epCount := countVideos(seasonPath)
			show.Seasons = append(show.Seasons, models.Season{
				Number:   seasonNum,
				Path:     seasonPath,
				Episodes: epCount,
			})
			show.Episodes += epCount
		}

		// Also count loose videos at show root
		show.Episodes += countVideosFlat(showPath)

		sort.Slice(show.Seasons, func(i, j int) bool {
			return show.Seasons[i].Number < show.Seasons[j].Number
		})

		shows = append(shows, show)
	}

	sort.Slice(shows, func(i, j int) bool {
		return strings.ToLower(shows[i].Name) < strings.ToLower(shows[j].Name)
	})

	return shows, nil
}

// ScanMovies returns all movies in the movies directory.
func ScanMovies(moviesDir string) ([]models.Movie, error) {
	entries, err := os.ReadDir(moviesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read movies dir: %w", err)
	}

	var movies []models.Movie
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moviePath := filepath.Join(moviesDir, entry.Name())
		fileCount := countVideos(moviePath)
		movies = append(movies, models.Movie{
			Name:  entry.Name(),
			Path:  moviePath,
			Files: fileCount,
		})
	}

	sort.Slice(movies, func(i, j int) bool {
		return strings.ToLower(movies[i].Name) < strings.ToLower(movies[j].Name)
	})

	return movies, nil
}

// ScanDownloads returns all downloadable items (folders + loose video files).
func ScanDownloads(downloadDir string) ([]models.DownloadItem, error) {
	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read download dir: %w", err)
	}

	var items []models.DownloadItem
	for _, entry := range entries {
		fullPath := filepath.Join(downloadDir, entry.Name())

		if entry.IsDir() {
			vc := countVideos(fullPath)
			size := dirSize(fullPath)
			items = append(items, models.DownloadItem{
				Name:       entry.Name(),
				Path:       fullPath,
				IsDir:      true,
				VideoCount: vc,
				Size:       size,
			})
		} else if IsVideo(entry.Name()) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			items = append(items, models.DownloadItem{
				Name:       entry.Name(),
				Path:       fullPath,
				IsDir:      false,
				VideoCount: 1,
				Size:       info.Size(),
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items, nil
}

// LibrarySize calculates total video file size across media + movies dirs.
func LibrarySize(mediaDir, moviesDir string) int64 {
	var total int64
	for _, dir := range []string{mediaDir, moviesDir} {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && IsVideo(info.Name()) {
				total += info.Size()
			}
			return nil
		})
	}
	return total
}

// FindSeasonDirs detects season subdirectories within a source path.
// Returns a map of season number -> directory path.
func FindSeasonDirs(sourcePath string) map[int]string {
	result := make(map[int]string)
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		snum := parseSeasonDir(entry.Name())
		if snum >= 0 {
			result[snum] = filepath.Join(sourcePath, entry.Name())
		}
	}
	return result
}

// CountVideosIn counts video files in a directory (recursively within that dir only for depth 1).
func CountVideosIn(dir string) int {
	return countVideos(dir)
}

// --- helpers ---

var reSeasonDir = regexp.MustCompile(`(?i)^(?:season\s*0*(\d+)|s0*(\d+))$`)

func parseSeasonDir(name string) int {
	m := reSeasonDir.FindStringSubmatch(name)
	if m == nil {
		return -1
	}
	numStr := m[1]
	if numStr == "" {
		numStr = m[2]
	}
	n, err := strconv.Atoi(numStr)
	if err != nil {
		return -1
	}
	return n
}

func countVideos(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && IsVideo(info.Name()) {
			count++
		}
		return nil
	})
	return count
}

func countVideosFlat(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && IsVideo(e.Name()) {
			count++
		}
	}
	return count
}

func dirSize(dir string) int64 {
	var total int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}
