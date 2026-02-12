package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port     int
	DataDir  string
	Password string

	// Paths (inside container these are under /data)
	DownloadDir string
	MediaDir    string
	MoviesDir   string

	// Video extensions to match
	VideoExtensions []string

	// qBittorrent
	QbitURL      string
	QbitUser     string
	QbitPass     string
	QbitCategory string

	// Shoko Server
	ShokoURL    string
	ShokoAPIKey string

	// Notifications
	NotifyURL string
}

func Load() *Config {
	return &Config{
		Port:     envInt("LA_PORT", 8787),
		DataDir:  envStr("LA_DATA_DIR", "/app/data"),
		Password: envStr("LA_PASSWORD", "changeme"),

		DownloadDir: envStr("LA_DOWNLOAD_DIR", "/data/downloads/complete/anime"),
		MediaDir:    envStr("LA_MEDIA_DIR", "/data/media/anime"),
		MoviesDir:   envStr("LA_MOVIES_DIR", "/data/media/anime-movies"),

		VideoExtensions: []string{"mkv", "mp4", "avi"},

		QbitURL:      envStr("LA_QBIT_URL", ""),
		QbitUser:     envStr("LA_QBIT_USER", ""),
		QbitPass:     envStr("LA_QBIT_PASS", ""),
		QbitCategory: envStr("LA_QBIT_CATEGORY", "anime"),

		ShokoURL:    envStr("LA_SHOKO_URL", ""),
		ShokoAPIKey: envStr("LA_SHOKO_APIKEY", ""),

		NotifyURL: envStr("LA_NOTIFY_URL", ""),
	}
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
