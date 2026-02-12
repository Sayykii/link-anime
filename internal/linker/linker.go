package linker

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"link-anime/internal/database"
	"link-anime/internal/models"
	"link-anime/internal/scanner"
	"link-anime/internal/ws"
)

// Link creates hardlinks from source to destination.
func Link(req models.LinkRequest, downloadDir, mediaDir, moviesDir string, hub *ws.Hub) (*models.LinkResult, error) {
	// Resolve source path
	sourcePath, err := resolveSource(req.Source, downloadDir)
	if err != nil {
		return nil, err
	}

	// Determine destination
	var destDir string
	if req.Type == "movie" {
		destDir = filepath.Join(moviesDir, req.Name)
	} else {
		destDir = filepath.Join(mediaDir, req.Name, fmt.Sprintf("Season %d", req.Season))
	}

	// Check for season subdirectories in source (series only)
	if req.Type == "series" {
		info, err := os.Stat(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("stat source: %w", err)
		}
		if info.IsDir() {
			seasonDirs := scanner.FindSeasonDirs(sourcePath)
			if len(seasonDirs) > 0 {
				return linkMultiSeason(sourcePath, mediaDir, req.Name, seasonDirs, req.DryRun, hub)
			}
		}
	}

	// Single source -> single destination
	return linkSingle(sourcePath, destDir, req, hub)
}

func linkSingle(sourcePath, destDir string, req models.LinkRequest, hub *ws.Hub) (*models.LinkResult, error) {
	result := &models.LinkResult{DestDir: destDir}
	var linkedFiles []string

	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("stat source: %w", err)
	}

	if info.IsDir() {
		// Collect video files
		var videoFiles []string
		entries, err := os.ReadDir(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("read source dir: %w", err)
		}
		for _, e := range entries {
			if !e.IsDir() && scanner.IsVideo(e.Name()) {
				videoFiles = append(videoFiles, filepath.Join(sourcePath, e.Name()))
			}
		}
		sort.Strings(videoFiles)

		if !req.DryRun {
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return nil, fmt.Errorf("create dest dir: %w", err)
			}
		}

		total := len(videoFiles)
		for i, srcFile := range videoFiles {
			filename := filepath.Base(srcFile)
			destFile := filepath.Join(destDir, filename)

			status := linkFile(srcFile, destFile, req.DryRun, result)
			if status == "linked" {
				linkedFiles = append(linkedFiles, destFile)
			}

			if hub != nil {
				hub.Broadcast(models.WSMessage{
					Type: "link:progress",
					Data: models.LinkProgress{
						File:    filename,
						Status:  status,
						Current: i + 1,
						Total:   total,
					},
				})
			}
		}
	} else if scanner.IsVideo(info.Name()) {
		// Single file
		if !req.DryRun {
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return nil, fmt.Errorf("create dest dir: %w", err)
			}
		}

		filename := filepath.Base(sourcePath)
		destFile := filepath.Join(destDir, filename)
		status := linkFile(sourcePath, destFile, req.DryRun, result)
		if status == "linked" {
			linkedFiles = append(linkedFiles, destFile)
		}

		if hub != nil {
			hub.Broadcast(models.WSMessage{
				Type: "link:progress",
				Data: models.LinkProgress{
					File:    filename,
					Status:  status,
					Current: 1,
					Total:   1,
				},
			})
		}
	}

	result.Files = linkedFiles

	// Write history
	if !req.DryRun && result.Linked > 0 {
		if err := writeHistory(req, result, sourcePath); err != nil {
			// Non-fatal
			fmt.Fprintf(os.Stderr, "warning: failed to write history: %v\n", err)
		}
	}

	if hub != nil {
		hub.Broadcast(models.WSMessage{
			Type: "link:complete",
			Data: result,
		})
	}

	return result, nil
}

func linkMultiSeason(sourcePath, mediaDir, showName string, seasonDirs map[int]string, dryRun bool, hub *ws.Hub) (*models.LinkResult, error) {
	combined := &models.LinkResult{
		DestDir: filepath.Join(mediaDir, showName),
	}

	// Sort season numbers
	var seasons []int
	for s := range seasonDirs {
		seasons = append(seasons, s)
	}
	sort.Ints(seasons)

	for _, snum := range seasons {
		sdir := seasonDirs[snum]
		destDir := filepath.Join(mediaDir, showName, fmt.Sprintf("Season %d", snum))

		req := models.LinkRequest{
			Source: filepath.Base(sourcePath),
			Type:   "series",
			Name:   showName,
			Season: snum,
			DryRun: dryRun,
		}

		r, err := linkSingle(sdir, destDir, req, hub)
		if err != nil {
			return nil, fmt.Errorf("season %d: %w", snum, err)
		}

		combined.Linked += r.Linked
		combined.Skipped += r.Skipped
		combined.Failed += r.Failed
		combined.Size += r.Size
		combined.Files = append(combined.Files, r.Files...)
	}

	return combined, nil
}

func linkFile(src, dest string, dryRun bool, result *models.LinkResult) string {
	fileInfo, err := os.Stat(src)
	if err != nil {
		result.Failed++
		return "failed"
	}
	fileSize := fileInfo.Size()

	if _, err := os.Stat(dest); err == nil {
		// Dest exists — check if same inode
		srcStat, err1 := os.Stat(src)
		destStat, err2 := os.Stat(dest)
		if err1 == nil && err2 == nil {
			srcSys, ok1 := srcStat.Sys().(*syscall.Stat_t)
			destSys, ok2 := destStat.Sys().(*syscall.Stat_t)
			if ok1 && ok2 && srcSys.Ino == destSys.Ino {
				result.Skipped++
				return "skipped"
			}
		}
		result.Skipped++
		return "skipped"
	}

	if dryRun {
		result.Linked++
		result.Size += fileSize
		return "linked"
	}

	if err := os.Link(src, dest); err != nil {
		result.Failed++
		return "failed"
	}

	result.Linked++
	result.Size += fileSize
	return "linked"
}

func writeHistory(req models.LinkRequest, result *models.LinkResult, sourcePath string) error {
	var season *int
	if req.Type == "series" {
		season = &req.Season
	}

	var seasonVal sql.NullInt64
	if season != nil {
		seasonVal = sql.NullInt64{Int64: int64(*season), Valid: true}
	}

	res, err := database.DB.Exec(
		`INSERT INTO history (media_type, show_name, season, file_count, total_size, dest_path, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.Type, req.Name, seasonVal, result.Linked, result.Size, result.DestDir, filepath.Base(sourcePath),
	)
	if err != nil {
		return err
	}

	historyID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	for _, f := range result.Files {
		// Determine the source path for this linked file
		srcFile := filepath.Join(sourcePath, filepath.Base(f))
		_, err := database.DB.Exec(
			`INSERT INTO linked_files (history_id, file_path, source_path) VALUES (?, ?, ?)`,
			historyID, f, srcFile,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Unlink removes hardlinks from the library side.
func Unlink(targetDir string) (*models.LinkResult, error) {
	result := &models.LinkResult{DestDir: targetDir}

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !scanner.IsVideo(info.Name()) {
			return nil
		}

		if err := os.Remove(path); err != nil {
			result.Failed++
		} else {
			result.Linked++ // reusing Linked as "removed" count
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Clean up empty directories
	cleanEmptyDirs(targetDir)

	return result, nil
}

// Undo reverses the last link operation.
func Undo() (*models.LinkResult, *models.HistoryEntry, error) {
	// Get the last history entry
	var entry models.HistoryEntry
	var seasonVal sql.NullInt64
	err := database.DB.QueryRow(
		`SELECT id, timestamp, media_type, show_name, season, file_count, total_size, dest_path, source
		 FROM history ORDER BY id DESC LIMIT 1`,
	).Scan(&entry.ID, &entry.Timestamp, &entry.MediaType, &entry.ShowName, &seasonVal,
		&entry.FileCount, &entry.TotalSize, &entry.DestPath, &entry.Source)
	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("no history entries to undo")
	}
	if err != nil {
		return nil, nil, fmt.Errorf("query history: %w", err)
	}
	if seasonVal.Valid {
		s := int(seasonVal.Int64)
		entry.Season = &s
	}

	// Get linked files for this entry
	rows, err := database.DB.Query(
		`SELECT file_path FROM linked_files WHERE history_id = ?`, entry.ID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("query linked files: %w", err)
	}
	defer rows.Close()

	result := &models.LinkResult{DestDir: entry.DestPath}
	var dirsToClean []string

	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			continue
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			result.Skipped++
			continue
		}

		if err := os.Remove(filePath); err != nil {
			result.Failed++
		} else {
			result.Linked++ // removed count
			dir := filepath.Dir(filePath)
			dirsToClean = append(dirsToClean, dir)
		}
	}

	// Clean up empty directories
	seen := make(map[string]bool)
	for _, dir := range dirsToClean {
		if !seen[dir] {
			seen[dir] = true
			cleanEmptyDirs(dir)
		}
	}

	// Remove history entry
	database.DB.Exec("DELETE FROM linked_files WHERE history_id = ?", entry.ID)
	database.DB.Exec("DELETE FROM history WHERE id = ?", entry.ID)

	return result, &entry, nil
}

// GetHistory returns recent history entries.
func GetHistory(limit int) ([]models.HistoryEntry, error) {
	rows, err := database.DB.Query(
		`SELECT id, timestamp, media_type, show_name, season, file_count, total_size, dest_path, source
		 FROM history ORDER BY id DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.HistoryEntry
	for rows.Next() {
		var e models.HistoryEntry
		var seasonVal sql.NullInt64
		err := rows.Scan(&e.ID, &e.Timestamp, &e.MediaType, &e.ShowName, &seasonVal,
			&e.FileCount, &e.TotalSize, &e.DestPath, &e.Source)
		if err != nil {
			continue
		}
		if seasonVal.Valid {
			s := int(seasonVal.Int64)
			e.Season = &s
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// --- helpers ---

func resolveSource(source, downloadDir string) (string, error) {
	// Try exact match first
	path := filepath.Join(downloadDir, source)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Try dot-for-space replacement
	dotName := strings.ReplaceAll(source, " ", ".")
	path = filepath.Join(downloadDir, dotName)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Try space-for-dot replacement
	spaceName := strings.ReplaceAll(source, ".", " ")
	path = filepath.Join(downloadDir, spaceName)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("source not found: %s (searched in %s)", source, downloadDir)
}

func cleanEmptyDirs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Remove empty subdirectories first
	for _, e := range entries {
		if e.IsDir() {
			subdir := filepath.Join(dir, e.Name())
			cleanEmptyDirs(subdir)
		}
	}

	// Re-read after cleaning subdirs
	entries, err = os.ReadDir(dir)
	if err != nil {
		return
	}

	if len(entries) == 0 {
		os.Remove(dir)
	}
}
