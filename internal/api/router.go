package api

import (
	"net/http"
	"strings"

	"link-anime/internal/auth"
	"link-anime/internal/config"
	"link-anime/internal/notify"
	"link-anime/internal/qbit"
	"link-anime/internal/rss"
	"link-anime/internal/shoko"
	"link-anime/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server holds all dependencies for the API handlers.
type Server struct {
	Config   *config.Config
	Hub      *ws.Hub
	Qbit     *qbit.Client
	Shoko    *shoko.Client
	Notifier *notify.Notifier
	Poller   *rss.Poller
}

// NewRouter creates the chi router with all routes and middleware.
func NewRouter(s *Server, staticFS http.FileSystem) chi.Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/logout", s.handleLogout)
		r.Get("/auth/check", s.handleAuthCheck)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware)

			// Library
			r.Get("/library/shows", s.handleGetShows)
			r.Get("/library/movies", s.handleGetMovies)
			r.Get("/library/stats", s.handleGetStats)

			// Downloads
			r.Get("/downloads", s.handleGetDownloads)
			r.Get("/downloads/parse", s.handleParseRelease)

			// Link operations
			r.Post("/link", s.handleLink)
			r.Post("/link/preview", s.handleLinkPreview)
			r.Get("/link/unlink/preview", s.handleUnlinkPreview)
			r.Delete("/link/unlink", s.handleUnlink)
			r.Get("/link/undo/preview", s.handleUndoPreview)
			r.Post("/link/undo", s.handleUndo)

			// History
			r.Get("/history", s.handleGetHistory)

			// Settings
			r.Get("/settings", s.handleGetSettings)
			r.Put("/settings", s.handleUpdateSettings)
			r.Post("/settings/password", s.handleChangePassword)

			// qBittorrent
			r.Get("/qbit/torrents", s.handleQbitTorrents)
			r.Post("/qbit/add", s.handleQbitAdd)
			r.Delete("/qbit/delete", s.handleQbitDelete)
			r.Get("/qbit/test", s.handleQbitTest)

			// Nyaa
			r.Get("/nyaa/search", s.handleNyaaSearch)

			// Shoko
			r.Post("/shoko/scan", s.handleShokoScan)
			r.Get("/shoko/test", s.handleShokoTest)

			// RSS Rules
			r.Get("/rss/rules", s.handleListRSSRules)
			r.Get("/rss/rule", s.handleGetRSSRule)
			r.Post("/rss/rules", s.handleCreateRSSRule)
			r.Put("/rss/rules", s.handleUpdateRSSRule)
			r.Delete("/rss/rules", s.handleDeleteRSSRule)
			r.Post("/rss/rules/toggle", s.handleToggleRSSRule)
			r.Get("/rss/matches", s.handleListRSSMatches)
			r.Delete("/rss/matches", s.handleClearRSSMatches)
			r.Post("/rss/poll", s.handleRSSPollNow)

			// WebSocket
			r.Get("/ws", s.handleWS)
		})
	})

	// Serve SPA frontend
	if staticFS != nil {
		r.Get("/*", spaHandler(staticFS))
	}

	return r
}

// spaHandler serves the Vue SPA. It serves static files if they exist,
// otherwise falls back to index.html for client-side routing.
func spaHandler(fs http.FileSystem) http.HandlerFunc {
	fileServer := http.FileServer(fs)

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip API routes
		if strings.HasPrefix(path, "/api") {
			http.NotFound(w, r)
			return
		}

		// Try to open the file
		f, err := fs.Open(path)
		if err != nil {
			// File doesn't exist, serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		// File exists, serve it
		fileServer.ServeHTTP(w, r)
	}
}
