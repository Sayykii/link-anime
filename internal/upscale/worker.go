package upscale

import (
	"context"
	"log"
	"sync"
	"time"

	"link-anime/internal/database"
	"link-anime/internal/ws"
)

// Worker polls the database for pending upscale jobs and processes them.
// It follows the same lifecycle pattern as DownloadMonitor.
type Worker struct {
	engine *Engine
	hub    *ws.Hub

	// Lifecycle channels (matches DownloadMonitor)
	stopCh chan struct{}
	done   chan struct{}

	// Current job tracking for graceful shutdown
	currentJob int64
	cancelFunc context.CancelFunc
	mu         sync.Mutex
}

// NewWorker creates a new upscale worker.
func NewWorker(hub *ws.Hub, shaderDir string) *Worker {
	return &Worker{
		engine: NewEngine(shaderDir),
		hub:    hub,
		stopCh: make(chan struct{}),
		done:   make(chan struct{}),
	}
}

// Start begins the worker loop in a goroutine.
func (w *Worker) Start() {
	go w.run()
}

// Stop signals the worker to shut down and waits for it.
func (w *Worker) Stop() {
	close(w.stopCh)
	<-w.done
}

func (w *Worker) run() {
	defer close(w.done)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Do an initial poll right away
	w.processNext()

	for {
		select {
		case <-w.stopCh:
			w.handleShutdown()
			return
		case <-ticker.C:
			w.processNext()
		}
	}
}

func (w *Worker) processNext() {
	job, err := database.GetNextPendingJob()
	if err != nil {
		log.Printf("[upscale] failed to get pending job: %v", err)
		return
	}

	if job == nil {
		return // No pending jobs
	}

	log.Printf("[upscale] found pending job %d: %s", job.ID, job.InputPath)
	// Actual processing added in Plan 02
}

func (w *Worker) handleShutdown() {
	log.Printf("[upscale] worker shutting down")
	// Graceful shutdown added in Plan 02
}
