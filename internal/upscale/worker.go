package upscale

import (
	"context"
	"log"
	"sync"
	"time"

	"link-anime/internal/database"
	"link-anime/internal/models"
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
		log.Printf("[upscale] error getting pending job: %v", err)
		return
	}
	if job == nil {
		return // No pending jobs
	}

	log.Printf("[upscale] processing job %d: %s", job.ID, job.InputPath)

	// Track job for graceful shutdown
	w.mu.Lock()
	w.currentJob = job.ID
	ctx, cancel := context.WithCancel(context.Background())
	w.cancelFunc = cancel
	w.mu.Unlock()

	// Mark job as running
	if err := database.UpdateJobStatus(job.ID, models.UpscaleStatusRunning, nil); err != nil {
		log.Printf("[upscale] failed to update job %d to running: %v", job.ID, err)
		return
	}

	// Run upscaling with progress callback
	err = w.engine.Run(ctx, job, func(p models.UpscaleProgress) {
		w.hub.Broadcast(models.WSMessage{
			Type: "upscale_progress",
			Data: p,
		})
	})

	// Clear current job
	w.mu.Lock()
	w.currentJob = 0
	w.cancelFunc = nil
	w.mu.Unlock()

	// Handle result
	if err != nil {
		if ctx.Err() != nil {
			// Cancelled by shutdown - don't mark failed, handleShutdown resets
			log.Printf("[upscale] job %d cancelled", job.ID)
			return
		}
		errStr := err.Error()
		database.UpdateJobStatus(job.ID, models.UpscaleStatusFailed, &errStr)
		w.hub.Broadcast(models.WSMessage{
			Type: "upscale_failed",
			Data: map[string]interface{}{"jobId": job.ID, "error": errStr},
		})
		log.Printf("[upscale] job %d failed: %v", job.ID, err)
		return
	}

	database.UpdateJobStatus(job.ID, models.UpscaleStatusCompleted, nil)
	w.hub.Broadcast(models.WSMessage{
		Type: "upscale_complete",
		Data: map[string]interface{}{"jobId": job.ID, "outputPath": job.OutputPath},
	})
	log.Printf("[upscale] job %d completed: %s", job.ID, job.OutputPath)
}

// CancelJob cancels a job by ID if it's currently running.
// Returns true if the job was cancelled, false if not running or different job.
func (w *Worker) CancelJob(id int64) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.currentJob != id || w.cancelFunc == nil {
		return false
	}

	w.cancelFunc()
	return true
}

func (w *Worker) handleShutdown() {
	w.mu.Lock()
	jobID := w.currentJob
	cancel := w.cancelFunc
	w.mu.Unlock()

	if cancel != nil {
		log.Printf("[upscale] cancelling running job %d", jobID)
		cancel()
	}

	if jobID != 0 {
		// Reset job to pending so it restarts on next run
		if err := database.ResetRunningJob(jobID); err != nil {
			log.Printf("[upscale] failed to reset job %d: %v", jobID, err)
		} else {
			log.Printf("[upscale] reset job %d to pending for restart", jobID)
		}
	}

	log.Printf("[upscale] worker shutdown complete")
}
