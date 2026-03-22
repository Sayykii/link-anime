package shoko

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// SeriesFolderMap maps library folder names to Shoko series IDs.
// It uses Shoko's file locations to determine which series lives in which folder.
type SeriesFolderMap map[string]int // folderName → shokoSeriesID

// shokoFile represents a file in the Shoko API.
type shokoFile struct {
	ID        int             `json:"ID"`
	Locations []shokoLocation `json:"Locations"`
	SeriesIDs []shokoSeriesXRef `json:"SeriesIDs"`
}

type shokoLocation struct {
	ImportFolderID int    `json:"ImportFolderID"`
	RelativePath   string `json:"RelativePath"`
}

type shokoSeriesXRef struct {
	SeriesID struct {
		ID int `json:"ID"`
	} `json:"SeriesID"`
}

// BuildFolderMap builds a mapping from library folder names to Shoko series IDs.
// It fetches files from Shoko and extracts the top-level folder from each file's path.
func (c *Client) BuildFolderMap() (SeriesFolderMap, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	result := make(SeriesFolderMap)

	// Fetch files page by page
	page := 1
	pageSize := 200
	for {
		path := fmt.Sprintf("/api/v3/File?pageSize=%d&page=%d&include=XRefBySeriesID", pageSize, page)
		resp, err := c.doRequest("GET", path, nil)
		if err != nil {
			return nil, fmt.Errorf("shoko get files: %w", err)
		}

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("shoko get files failed (status %d): %s", resp.StatusCode, string(body))
		}

		var fileResult listResult[shokoFile]
		err = json.NewDecoder(resp.Body).Decode(&fileResult)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("shoko decode files: %w", err)
		}

		for _, f := range fileResult.List {
			if len(f.Locations) == 0 || len(f.SeriesIDs) == 0 {
				continue
			}

			// Extract top-level folder from relative path
			relPath := f.Locations[0].RelativePath
			// Normalize separators
			relPath = strings.ReplaceAll(relPath, "\\", "/")
			parts := strings.SplitN(relPath, "/", 2)
			if len(parts) < 1 || parts[0] == "" {
				continue
			}

			folderName := parts[0]
			seriesID := f.SeriesIDs[0].SeriesID.ID

			// Only set if not already mapped (first file wins)
			if _, exists := result[folderName]; !exists {
				result[folderName] = seriesID
			}
		}

		// Check if we've fetched all pages
		if len(fileResult.List) < pageSize {
			break
		}
		page++
	}

	return result, nil
}

// SearchSeriesForFolder tries to find a Shoko series matching a folder name via search.
// Falls back to fuzzy search when file-based matching isn't available.
func (c *Client) SearchSeriesForFolder(folderName string) (*ShokoSeries, error) {
	// Clean up the folder name for search (remove years, brackets, etc.)
	query := folderName
	query = strings.TrimSpace(query)

	results, err := c.SearchSeries(query)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	// Check for exact name match first
	lower := strings.ToLower(folderName)
	for i, s := range results {
		if strings.ToLower(s.Name) == lower {
			return &results[i], nil
		}
		if s.AniDB != nil && strings.ToLower(s.AniDB.Title) == lower {
			return &results[i], nil
		}
	}

	// Return the top search result
	return &results[0], nil
}

// GetFolderSeriesMap builds a complete mapping using file paths first,
// then fills gaps with search-based matching.
func (c *Client) GetFolderSeriesMap(folderNames []string) (map[string]*ShokoSeries, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	// Step 1: Build file-path-based mapping
	folderToID, _ := c.BuildFolderMap() // don't fail if this errors
	if folderToID == nil {
		folderToID = make(SeriesFolderMap)
	}

	// Step 2: Fetch all series we need by ID
	seriesCache := make(map[int]*ShokoSeries)
	result := make(map[string]*ShokoSeries)

	// First pass: resolve folders that have a file-path match
	neededIDs := make(map[int]bool)
	for _, name := range folderNames {
		if id, ok := folderToID[name]; ok {
			neededIDs[id] = true
		}
	}

	// Batch fetch all needed series
	if len(neededIDs) > 0 {
		// Load all series (we likely already have most of them)
		allSeries, _, err := c.GetSeries(1, 100)
		if err == nil {
			for i := range allSeries {
				seriesCache[allSeries[i].IDs.ID] = &allSeries[i]
			}
		}
	}

	// Map folders to series
	for _, name := range folderNames {
		if id, ok := folderToID[name]; ok {
			if s, ok := seriesCache[id]; ok {
				result[name] = s
				continue
			}
			// If not in cache, fetch individually
			s, err := c.GetSeriesDetails(id)
			if err == nil && s != nil {
				result[name] = s
			}
		}
	}

	// Step 3: For unmatched folders, try search
	for _, name := range folderNames {
		if _, ok := result[name]; ok {
			continue // already matched
		}
		s, err := c.SearchSeriesForFolder(name)
		if err == nil && s != nil {
			result[name] = s
		}
	}

	// Store the preferred poster URL in the folder name to series path
	return result, nil
}

// GetPosterForFolder returns the poster URL path components for a folder name.
// Returns source, type, id or empty strings if no poster found.
func GetPreferredPoster(s *ShokoSeries) (source string, imageType string, id int) {
	if s == nil || len(s.Images.Posters) == 0 {
		return "", "", 0
	}
	for _, p := range s.Images.Posters {
		if p.Preferred {
			return p.Source, p.Type, p.ID
		}
	}
	return s.Images.Posters[0].Source, s.Images.Posters[0].Type, s.Images.Posters[0].ID
}

// FolderMapping is the JSON response for a single folder's match.
type FolderMapping struct {
	ShokoID   int    `json:"shokoId"`
	Name      string `json:"name"`             // Shoko's name for the series
	PosterURL string `json:"posterUrl,omitempty"` // proxied poster URL
}

// BuildFolderMappingResponse creates the API response from a folder→series map.
func BuildFolderMappingResponse(m map[string]*ShokoSeries) map[string]*FolderMapping {
	result := make(map[string]*FolderMapping)
	for folder, series := range m {
		if series == nil {
			continue
		}

		fm := &FolderMapping{
			ShokoID: series.IDs.ID,
			Name:    series.Name,
		}

		source, imgType, id := GetPreferredPoster(series)
		if source != "" {
			fm.PosterURL = fmt.Sprintf("/api/shoko/image/%s/%s/%d", source, imgType, id)
		}

		result[folder] = fm
	}
	return result
}

// matchFolder tries to find a folder name in the relative path.
func extractTopFolder(relativePath string) string {
	relativePath = filepath.ToSlash(relativePath)
	parts := strings.SplitN(relativePath, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
