package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"link-anime/internal/api"
	"link-anime/internal/auth"
	"link-anime/internal/config"
	"link-anime/internal/database"
	"link-anime/internal/notify"
	"link-anime/internal/qbit"
	"link-anime/internal/rss"
	"link-anime/internal/scanner"
	"link-anime/internal/shoko"
	"link-anime/internal/ws"
)

//go:embed all:dist
var frontendFS embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load config from env vars
	cfg := config.Load()

	// Initialize database
	if err := database.Init(cfg.DataDir); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer database.Close()

	// Initialize password
	if err := auth.InitPassword(cfg.Password); err != nil {
		log.Fatalf("Password init failed: %v", err)
	}

	// Start session cleanup
	auth.CleanupExpiredSessions()

	// Initialize video extension matcher
	scanner.InitVideoExtensions(cfg.VideoExtensions)

	// Create WebSocket hub
	hub := ws.NewHub()

	// Create integration clients
	var qbitClient *qbit.Client
	if cfg.QbitURL != "" {
		qbitClient = qbit.New(cfg.QbitURL, cfg.QbitUser, cfg.QbitPass)
	}

	var shokoClient *shoko.Client
	if cfg.ShokoURL != "" {
		shokoClient = shoko.New(cfg.ShokoURL, cfg.ShokoAPIKey)
	}

	notifier := notify.New(cfg.NotifyURL)

	// Create API server
	server := &api.Server{
		Config:   cfg,
		Hub:      hub,
		Qbit:     qbitClient,
		Shoko:    shokoClient,
		Notifier: notifier,
	}

	// Create RSS poller (getter func reads server.Qbit so reinitClients updates are reflected)
	poller := rss.NewPoller(hub, func() *qbit.Client { return server.Qbit }, 15*time.Minute)
	poller.Start()
	defer poller.Stop()
	server.Poller = poller

	// Embed frontend static files
	var staticFS http.FileSystem
	dist, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		log.Printf("Warning: frontend not embedded, running API-only: %v", err)
	} else {
		staticFS = http.FS(dist)
	}

	// Create router
	router := api.NewRouter(server, staticFS)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("link-anime starting on %s", addr)
	log.Printf("  Download dir: %s", cfg.DownloadDir)
	log.Printf("  Media dir:    %s", cfg.MediaDir)
	log.Printf("  Movies dir:   %s", cfg.MoviesDir)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		httpServer.Close()
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
