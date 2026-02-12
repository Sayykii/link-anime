package shoko

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a Shoko Server API client.
type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// New creates a new Shoko client.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// IsConfigured returns true if the Shoko URL is set.
func (c *Client) IsConfigured() bool {
	return c.baseURL != ""
}

// doRequest executes an authenticated HTTP request.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("apikey", c.apiKey)
	}

	return c.client.Do(req)
}

// ScanImportFolder triggers a scan of import folders in Shoko.
// This is the primary integration point — after linking files into the library,
// we tell Shoko to rescan so it picks up the new episodes.
func (c *Client) ScanImportFolder(folderID int) error {
	if !c.IsConfigured() {
		return fmt.Errorf("shoko not configured")
	}

	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v3/ImportFolder/%d/Scan", folderID), nil)
	if err != nil {
		return fmt.Errorf("shoko scan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shoko scan failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// ScanAllImportFolders triggers a scan of all import folders.
func (c *Client) ScanAllImportFolders() error {
	if !c.IsConfigured() {
		return fmt.Errorf("shoko not configured")
	}

	// Get list of import folders
	folders, err := c.GetImportFolders()
	if err != nil {
		return err
	}

	for _, f := range folders {
		if err := c.ScanImportFolder(f.ID); err != nil {
			return fmt.Errorf("scan folder %d (%s): %w", f.ID, f.Name, err)
		}
	}

	return nil
}

// ImportFolder represents a Shoko import folder.
type ImportFolder struct {
	ID   int    `json:"ID"`
	Name string `json:"Name"`
	Path string `json:"Path"`
}

// GetImportFolders returns the list of import folders.
func (c *Client) GetImportFolders() ([]ImportFolder, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("shoko not configured")
	}

	resp, err := c.doRequest("GET", "/api/v3/ImportFolder", nil)
	if err != nil {
		return nil, fmt.Errorf("shoko get folders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shoko get folders failed (status %d): %s", resp.StatusCode, string(body))
	}

	var folders []ImportFolder
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, fmt.Errorf("shoko decode folders: %w", err)
	}

	return folders, nil
}

// TestConnection checks if the Shoko server is reachable and authenticated.
func (c *Client) TestConnection() error {
	if !c.IsConfigured() {
		return fmt.Errorf("shoko not configured")
	}

	resp, err := c.doRequest("GET", "/api/v3/Init/Status", nil)
	if err != nil {
		return fmt.Errorf("shoko connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("shoko auth failed: invalid API key")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("shoko returned status %d", resp.StatusCode)
	}

	return nil
}
