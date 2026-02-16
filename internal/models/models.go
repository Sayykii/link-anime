package models

import "time"

// Show represents an anime series in the media library.
type Show struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Seasons  []Season `json:"seasons"`
	Episodes int      `json:"episodes"`
}

// Season represents a season directory within a show.
type Season struct {
	Number   int    `json:"number"`
	Path     string `json:"path"`
	Episodes int    `json:"episodes"`
}

// Movie represents a movie in the movies library.
type Movie struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Files int    `json:"files"`
}

// DownloadItem represents a folder or file in the download directory.
type DownloadItem struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	IsDir      bool   `json:"isDir"`
	VideoCount int    `json:"videoCount"`
	Size       int64  `json:"size"`
}

// LinkRequest is the payload for creating hardlinks.
type LinkRequest struct {
	Source string `json:"source"` // folder/file name in downloads dir
	Type   string `json:"type"`   // "series" or "movie"
	Name   string `json:"name"`   // show name or movie name
	Season int    `json:"season"` // season number (series only)
	DryRun bool   `json:"dryRun"`
}

// LinkResult describes the outcome of a link operation.
type LinkResult struct {
	Linked  int      `json:"linked"`
	Skipped int      `json:"skipped"`
	Failed  int      `json:"failed"`
	Size    int64    `json:"size"`
	DestDir string   `json:"destDir"`
	Files   []string `json:"files"`
}

// HistoryEntry records a past link operation.
type HistoryEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	MediaType string    `json:"mediaType"`
	ShowName  string    `json:"showName"`
	Season    *int      `json:"season,omitempty"`
	FileCount int       `json:"fileCount"`
	TotalSize int64     `json:"totalSize"`
	DestPath  string    `json:"destPath"`
	Source    string    `json:"source"`
}

// LinkedFile records a single hardlinked file for undo.
type LinkedFile struct {
	ID         int64  `json:"id"`
	HistoryID  int64  `json:"historyId"`
	FilePath   string `json:"filePath"`
	SourcePath string `json:"sourcePath"`
}

// ParseResult is the output of release name parsing.
type ParseResult struct {
	Name   string `json:"name"`
	Season *int   `json:"season,omitempty"`
}

// LibraryStats gives an overview of the library.
type LibraryStats struct {
	Shows    int   `json:"shows"`
	Seasons  int   `json:"seasons"`
	Episodes int   `json:"episodes"`
	Movies   int   `json:"movies"`
	Size     int64 `json:"size"`
}

// Settings represents user-configurable settings stored in DB.
type Settings struct {
	QbitURL      string `json:"qbitUrl"`
	QbitUser     string `json:"qbitUser"`
	QbitPass     string `json:"qbitPass"`
	QbitCategory string `json:"qbitCategory"`
	ShokoURL     string `json:"shokoUrl"`
	ShokoAPIKey  string `json:"shokoApiKey"`
	NotifyURL    string `json:"notifyUrl"`
	DownloadDir  string `json:"downloadDir"`
	MediaDir     string `json:"mediaDir"`
	MoviesDir    string `json:"moviesDir"`
}

// WSMessage is a typed WebSocket message.
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// LinkProgress is sent over WebSocket during linking.
type LinkProgress struct {
	File    string `json:"file"`
	Status  string `json:"status"` // "linked", "skipped", "failed"
	Current int    `json:"current"`
	Total   int    `json:"total"`
}

// TorrentStatus represents a torrent in qBittorrent.
type TorrentStatus struct {
	Name     string  `json:"name"`
	Hash     string  `json:"hash"`
	State    string  `json:"state"`
	Progress float64 `json:"progress"`
	DLSpeed  int64   `json:"dlSpeed"`
	ULSpeed  int64   `json:"ulSpeed"`
	Size     int64   `json:"size"`
	ETA      int     `json:"eta"`
	Ratio    float64 `json:"ratio"`
}

// NyaaResult represents a search result from Nyaa.
type NyaaResult struct {
	Title    string `json:"title"`
	Magnet   string `json:"magnet"`
	Size     string `json:"size"`
	Seeders  int    `json:"seeders"`
	Leechers int    `json:"leechers"`
}

// RSSRule defines an auto-download rule.
type RSSRule struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Query      string     `json:"query"`
	ShowName   string     `json:"showName"`
	Season     int        `json:"season"`
	MediaType  string     `json:"mediaType"`
	MinSeeders int        `json:"minSeeders"`
	Resolution string     `json:"resolution,omitempty"`
	Enabled    bool       `json:"enabled"`
	LastCheck  *time.Time `json:"lastCheck,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	MatchCount int        `json:"matchCount"` // populated by queries, not stored
}

// RSSMatch records a torrent matched by an RSS rule.
type RSSMatch struct {
	ID       int64     `json:"id"`
	RuleID   int64     `json:"ruleId"`
	Title    string    `json:"title"`
	Hash     string    `json:"hash"`
	Matched  time.Time `json:"matched"`
	Status   string    `json:"status"`             // "downloaded", "linked", "failed"
	RuleName string    `json:"ruleName,omitempty"` // populated by join queries
}
