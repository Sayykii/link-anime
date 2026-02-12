package api

import (
	"link-anime/internal/database"
	"link-anime/internal/notify"
	"link-anime/internal/qbit"
	"link-anime/internal/shoko"
)

// getDownloadDir returns the download directory, preferring DB setting over config.
func (s *Server) getDownloadDir() string {
	if v, err := database.GetSetting("download_dir"); err == nil && v != "" {
		return v
	}
	return s.Config.DownloadDir
}

// getMediaDir returns the media directory, preferring DB setting over config.
func (s *Server) getMediaDir() string {
	if v, err := database.GetSetting("media_dir"); err == nil && v != "" {
		return v
	}
	return s.Config.MediaDir
}

// getMoviesDir returns the movies directory, preferring DB setting over config.
func (s *Server) getMoviesDir() string {
	if v, err := database.GetSetting("movies_dir"); err == nil && v != "" {
		return v
	}
	return s.Config.MoviesDir
}

// newQbitClient creates a new qBittorrent client.
func newQbitClient(url, user, pass string) *qbit.Client {
	return qbit.New(url, user, pass)
}

// newShokoClient creates a new Shoko Server client.
func newShokoClient(url, apiKey string) *shoko.Client {
	return shoko.New(url, apiKey)
}

// newNotifier creates a new notification sender.
func newNotifier(url string) *notify.Notifier {
	return notify.New(url)
}
