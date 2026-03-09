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
	"link-anime/internal/monitor"
	"link-anime/internal/notify"
	"link-anime/internal/qbit"
	"link-anime/internal/rss"
	"link-anime/internal/scanner"
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

	// Create API server (clients initialized as nil, ReinitClients populates them)
	server := &api.Server{
		Config: cfg,
		Hub:    hub,
	}

	// Initialize integration clients from DB settings (first) or env vars (fallback).
	// This ensures credentials saved via the Settings UI survive container restarts.
	server.ReinitClients()

	// Create RSS poller (getter func reads server.Qbit so reinitClients updates are reflected)
	poller := rss.NewPoller(hub, func() *qbit.Client { return server.Qbit }, 15*time.Minute)
	poller.Start()
	defer poller.Stop()
	server.Poller = poller

	// Create download monitor (polls qBit every 5s, broadcasts progress via WS)
	dlMonitor := monitor.NewDownloadMonitor(
		hub,
		func() *qbit.Client { return server.Qbit },
		func() *notify.Notifier { return server.Notifier },
		func() string {
			// Prefer DB setting over config
			if v, err := database.GetSetting("qbit_category"); err == nil && v != "" {
				return v
			}
			return cfg.QbitCategory
		},
		5*time.Second,
	)
	dlMonitor.Start()
	defer dlMonitor.Stop()

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
