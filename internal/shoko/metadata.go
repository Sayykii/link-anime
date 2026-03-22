package shoko

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// GetSeries returns a paginated list of series from Shoko.
func (c *Client) GetSeries(page, pageSize int) ([]ShokoSeries, int, error) {
	if !c.IsConfigured() {
		return nil, 0, fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Series?page=%d&pageSize=%d&includeDataFrom=AniDB", page, pageSize)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("shoko get series: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("shoko get series failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result listResult[ShokoSeries]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("shoko decode series: %w", err)
	}

	return result.List, result.Total, nil
}

// SearchSeries searches for series by name.
func (c *Client) SearchSeries(query string) ([]ShokoSeries, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Series/Search/%s?includeDataFrom=AniDB&limit=20", query)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("shoko search series: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko search failed (status %d): %s", resp.StatusCode, string(body))
	}

	var results []ShokoSeries
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("shoko decode search: %w", err)
	}

	return results, nil
}

// GetSeriesDetails returns detailed info for a single series.
func (c *Client) GetSeriesDetails(id int) (*ShokoSeries, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Series/%d?includeDataFrom=AniDB", id)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("shoko get series detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko series detail failed (status %d): %s", resp.StatusCode, string(body))
	}

	var series ShokoSeries
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, fmt.Errorf("shoko decode series detail: %w", err)
	}

	return &series, nil
}

// GetSeriesEpisodes returns episodes for a series.
func (c *Client) GetSeriesEpisodes(seriesID int, includeMissing bool) ([]ShokoEpisode, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Series/%d/Episode?pageSize=0&includeMissing=%s&includeDataFrom=AniDB&type=Episode",
		seriesID, strconv.FormatBool(includeMissing))
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("shoko get episodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko episodes failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result listResult[ShokoEpisode]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("shoko decode episodes: %w", err)
	}

	return result.List, nil
}

// GetDashboardStats returns collection statistics.
func (c *Client) GetDashboardStats() (*ShokoDashboardStats, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	resp, err := c.doRequest("GET", "/api/v3/Dashboard/Stats", nil)
	if err != nil {
		return nil, fmt.Errorf("shoko dashboard stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko dashboard stats failed (status %d): %s", resp.StatusCode, string(body))
	}

	var stats ShokoDashboardStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("shoko decode stats: %w", err)
	}

	return &stats, nil
}

// GetRecentlyAddedEpisodes returns recently imported episodes.
func (c *Client) GetRecentlyAddedEpisodes(limit int) ([]ShokoDashboardEpisode, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Dashboard/RecentlyAddedEpisodes?pageSize=%d&page=1", limit)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("shoko recently added: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko recently added failed (status %d): %s", resp.StatusCode, string(body))
	}

	var episodes []ShokoDashboardEpisode
	if err := json.NewDecoder(resp.Body).Decode(&episodes); err != nil {
		return nil, fmt.Errorf("shoko decode recently added: %w", err)
	}

	return episodes, nil
}

// GetContinueWatching returns series the user is mid-way through.
func (c *Client) GetContinueWatching(limit int) ([]ShokoDashboardEpisode, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Dashboard/ContinueWatchingEpisodes?pageSize=%d&page=1", limit)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("shoko continue watching: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko continue watching failed (status %d): %s", resp.StatusCode, string(body))
	}

	var episodes []ShokoDashboardEpisode
	if err := json.NewDecoder(resp.Body).Decode(&episodes); err != nil {
		return nil, fmt.Errorf("shoko decode continue watching: %w", err)
	}

	return episodes, nil
}

// GetCalendar returns upcoming airing episodes.
func (c *Client) GetCalendar(days int) ([]ShokoDashboardEpisode, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Dashboard/AniDBCalendar?numberOfDays=%d&showAll=false", days)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("shoko calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko calendar failed (status %d): %s", resp.StatusCode, string(body))
	}

	var episodes []ShokoDashboardEpisode
	if err := json.NewDecoder(resp.Body).Decode(&episodes); err != nil {
		return nil, fmt.Errorf("shoko decode calendar: %w", err)
	}

	return episodes, nil
}

// GetImage proxies an image from Shoko. Returns the response body and content type.
func (c *Client) GetImage(source, imageType, id string) ([]byte, string, error) {
	if !c.IsConfigured() {
		return nil, "", fmt.Errorf("shoko not configured")
	}

	path := fmt.Sprintf("/api/v3/Image/%s/%s/%s", source, imageType, id)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("shoko image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("shoko image failed (status %d)", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("shoko read image: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return data, contentType, nil
}
