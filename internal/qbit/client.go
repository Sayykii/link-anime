package qbit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"link-anime/internal/models"
)

// Client is a qBittorrent Web API client.
type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
	mu       sync.Mutex
	loggedIn bool
}

// New creates a new qBittorrent client.
func New(baseURL, username, password string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 15 * time.Second,
			Jar:     jar,
		},
	}
}

// Login authenticates with qBittorrent.
func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := url.Values{
		"username": {c.username},
		"password": {c.password},
	}

	resp, err := c.client.PostForm(c.baseURL+"/api/v2/auth/login", data)
	if err != nil {
		return fmt.Errorf("qbit login request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || string(body) != "Ok." {
		return fmt.Errorf("qbit login failed: %s (status %d)", string(body), resp.StatusCode)
	}

	c.loggedIn = true
	return nil
}

// ensureLoggedIn logs in if not already authenticated.
func (c *Client) ensureLoggedIn() error {
	if !c.loggedIn {
		return c.Login()
	}
	return nil
}

// ListTorrents returns torrents, optionally filtered by category.
func (c *Client) ListTorrents(category string) ([]models.TorrentStatus, error) {
	if err := c.ensureLoggedIn(); err != nil {
		return nil, err
	}

	params := url.Values{}
	if category != "" {
		params.Set("category", category)
	}
	params.Set("sort", "added_on")
	params.Set("reverse", "true")

	resp, err := c.client.Get(c.baseURL + "/api/v2/torrents/info?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("qbit list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		// Session expired, re-login
		c.loggedIn = false
		if err := c.Login(); err != nil {
			return nil, err
		}
		return c.ListTorrents(category)
	}

	var raw []struct {
		Name     string  `json:"name"`
		Hash     string  `json:"hash"`
		State    string  `json:"state"`
		Progress float64 `json:"progress"`
		DLSpeed  int64   `json:"dlspeed"`
		ULSpeed  int64   `json:"upspeed"`
		Size     int64   `json:"size"`
		ETA      int     `json:"eta"`
		Ratio    float64 `json:"ratio"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("qbit decode: %w", err)
	}

	torrents := make([]models.TorrentStatus, len(raw))
	for i, t := range raw {
		torrents[i] = models.TorrentStatus{
			Name:     t.Name,
			Hash:     t.Hash,
			State:    t.State,
			Progress: t.Progress,
			DLSpeed:  t.DLSpeed,
			ULSpeed:  t.ULSpeed,
			Size:     t.Size,
			ETA:      t.ETA,
			Ratio:    t.Ratio,
		}
	}

	return torrents, nil
}

// AddMagnet adds a magnet link to qBittorrent.
func (c *Client) AddMagnet(magnet, category, savePath string) error {
	if err := c.ensureLoggedIn(); err != nil {
		return err
	}

	data := url.Values{
		"urls": {magnet},
	}
	if category != "" {
		data.Set("category", category)
	}
	if savePath != "" {
		data.Set("savepath", savePath)
	}

	resp, err := c.client.PostForm(c.baseURL+"/api/v2/torrents/add", data)
	if err != nil {
		return fmt.Errorf("qbit add: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qbit add failed: %s (status %d)", string(body), resp.StatusCode)
	}

	return nil
}

// GetTorrent returns info for a single torrent by hash.
func (c *Client) GetTorrent(hash string) (*models.TorrentStatus, error) {
	if err := c.ensureLoggedIn(); err != nil {
		return nil, err
	}

	params := url.Values{"hashes": {hash}}
	resp, err := c.client.Get(c.baseURL + "/api/v2/torrents/info?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("qbit get torrent: %w", err)
	}
	defer resp.Body.Close()

	var raw []struct {
		Name     string  `json:"name"`
		Hash     string  `json:"hash"`
		State    string  `json:"state"`
		Progress float64 `json:"progress"`
		DLSpeed  int64   `json:"dlspeed"`
		ULSpeed  int64   `json:"upspeed"`
		Size     int64   `json:"size"`
		ETA      int     `json:"eta"`
		Ratio    float64 `json:"ratio"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("qbit decode: %w", err)
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("torrent not found: %s", hash)
	}

	t := raw[0]
	return &models.TorrentStatus{
		Name:     t.Name,
		Hash:     t.Hash,
		State:    t.State,
		Progress: t.Progress,
		DLSpeed:  t.DLSpeed,
		ULSpeed:  t.ULSpeed,
		Size:     t.Size,
		ETA:      t.ETA,
		Ratio:    t.Ratio,
	}, nil
}

// DeleteTorrent deletes a torrent by hash. If deleteFiles is true, also removes downloaded data.
func (c *Client) DeleteTorrent(hash string, deleteFiles bool) error {
	if err := c.ensureLoggedIn(); err != nil {
		return err
	}

	deleteFlag := "false"
	if deleteFiles {
		deleteFlag = "true"
	}

	data := url.Values{
		"hashes":      {hash},
		"deleteFiles": {deleteFlag},
	}

	resp, err := c.client.PostForm(c.baseURL+"/api/v2/torrents/delete", data)
	if err != nil {
		return fmt.Errorf("qbit delete: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// IsConfigured returns true if qBittorrent URL is set.
func (c *Client) IsConfigured() bool {
	return c.baseURL != ""
}
