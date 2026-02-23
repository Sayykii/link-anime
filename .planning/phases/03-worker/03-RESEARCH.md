# Phase 3: Worker - Research

**Researched:** 2026-02-23
**Domain:** Go queue worker with graceful lifecycle
**Confidence:** HIGH

## Summary

Phase 3 implements a single-job queue worker that processes upscale jobs sequentially. The worker polls the database for pending jobs, runs them through the existing upscale engine (Phase 2), and broadcasts progress via WebSocket. The project already has two excellent patterns to follow: `DownloadMonitor` for Start/Stop lifecycle and `rss.Poller` for database polling with graceful shutdown.

The implementation is straightforward because:
1. All complex FFmpeg/progress logic already exists in `internal/upscale/` (Phase 2)
2. The `DownloadMonitor` pattern provides exact lifecycle blueprint to copy
3. SQLite with single connection already handles concurrency safety
4. WebSocket hub is already available for progress broadcasts

**Primary recommendation:** Create `internal/upscale/worker.go` following the `DownloadMonitor` pattern exactly, with a 3-second poll interval, oldest-first job selection, and graceful shutdown that resets running jobs to pending.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| WRK-01 | Single-job queue worker with Start/Stop lifecycle | Follow `DownloadMonitor` pattern - uses stopCh/done channels, Start() spawns goroutine, Stop() closes channel and waits |
| WRK-02 | Poll DB every 3s for pending jobs (oldest first) | Use `time.Ticker` like DownloadMonitor, query with `ORDER BY created_at ASC LIMIT 1` |
| WRK-03 | Status transitions (pending->running->completed/failed) | Database functions to update status + timestamps, handle errors from Engine.Run |
| WRK-04 | Graceful shutdown (reset running job to pending) | Track current job ID, on shutdown reset its status before exiting |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `time.Ticker` | stdlib | Polling interval | Already used by DownloadMonitor and rss.Poller |
| `context` | stdlib | Cancellation propagation | Engine.Run already accepts context |
| `sync.Mutex` | stdlib | Protect currentJobID | Follows existing codebase patterns |
| `database/sql` | stdlib | Job queries | Already initialized, single connection |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `link-anime/internal/upscale` | local | Engine for running jobs | Every job execution |
| `link-anime/internal/ws` | local | Progress broadcasts | Real-time UI updates |
| `link-anime/internal/database` | local | Job CRUD | Polling and status updates |
| `link-anime/internal/models` | local | UpscaleJob, UpscaleProgress | Type definitions |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| DB polling | Channel-based queue | More complex, DB polling sufficient for 3s interval with single worker |
| sync.Mutex | atomic.Int64 for jobID | Mutex clearer when protecting multiple fields |

**Installation:** No new dependencies needed - all stdlib and existing internal packages.

## Architecture Patterns

### Recommended Project Structure

```
internal/upscale/
├── engine.go      # Existing - FFmpeg runner
├── progress.go    # Existing - stderr parsing
├── ffprobe.go     # Existing - duration detection  
├── probe.go       # Existing - capability detection
└── worker.go      # NEW - queue worker
```

### Pattern 1: DownloadMonitor Lifecycle (COPY THIS)

**What:** Start/Stop pattern with channel-based signaling
**When to use:** Any background worker that needs graceful shutdown
**Source:** `internal/monitor/monitor.go`

```go
type Worker struct {
    // Dependencies
    engine     *Engine
    hub        *ws.Hub
    shaderDir  string
    
    // Lifecycle channels
    stopCh     chan struct{}
    done       chan struct{}
    
    // Current job tracking (for graceful shutdown)
    currentJob int64
    mu         sync.Mutex
}

func NewWorker(hub *ws.Hub, shaderDir string) *Worker {
    return &Worker{
        engine:    NewEngine(shaderDir),
        hub:       hub,
        shaderDir: shaderDir,
        stopCh:    make(chan struct{}),
        done:      make(chan struct{}),
    }
}

func (w *Worker) Start() {
    go w.run()
}

func (w *Worker) Stop() {
    close(w.stopCh)
    <-w.done
}

func (w *Worker) run() {
    defer close(w.done)
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()
    
    // Initial poll
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
```

### Pattern 2: Context Cancellation for Running Jobs

**What:** Pass cancellable context to Engine.Run for mid-job stop
**When to use:** Graceful shutdown while job is running
**Source:** Existing Engine.Run signature

```go
func (w *Worker) processNext() {
    job := getNextPendingJob() // oldest first
    if job == nil {
        return
    }
    
    // Track for graceful shutdown
    w.mu.Lock()
    w.currentJob = job.ID
    w.mu.Unlock()
    
    // Create cancellable context
    ctx, cancel := context.WithCancel(context.Background())
    w.cancelFunc = cancel  // Store for shutdown
    
    markJobRunning(job.ID)
    
    err := w.engine.Run(ctx, job, func(p models.UpscaleProgress) {
        w.hub.Broadcast(models.WSMessage{
            Type: "upscale_progress",
            Data: p,
        })
    })
    
    w.mu.Lock()
    w.currentJob = 0
    w.mu.Unlock()
    
    if err != nil {
        if ctx.Err() != nil {
            // Cancelled - don't mark failed, shutdown handler resets
            return
        }
        markJobFailed(job.ID, err.Error())
    } else {
        markJobCompleted(job.ID)
    }
}
```

### Pattern 3: Graceful Shutdown Reset

**What:** Reset running job to pending on shutdown for restart pickup
**When to use:** Worker Stop() is called while job is processing
**Source:** WRK-04 requirement

```go
func (w *Worker) handleShutdown() {
    // Cancel any running job
    if w.cancelFunc != nil {
        w.cancelFunc()
    }
    
    // Reset job to pending so it restarts on next run
    w.mu.Lock()
    jobID := w.currentJob
    w.mu.Unlock()
    
    if jobID != 0 {
        log.Printf("[upscale] shutdown: resetting job %d to pending", jobID)
        resetJobToPending(jobID)
    }
}
```

### Anti-Patterns to Avoid

- **Don't use goroutine per job:** Single worker means single goroutine processing loop
- **Don't poll without ticker:** Busy-waiting wastes CPU; ticker provides clean interval
- **Don't ignore context cancellation:** Engine.Run returns ctx.Err() on cancel, handle it
- **Don't mark cancelled jobs as failed:** They should reset to pending for retry

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Job execution | Custom FFmpeg wrapper | Existing `Engine.Run()` | Phase 2 already handles all complexity |
| Progress broadcast | Custom WS code | Existing `hub.Broadcast()` | Already integrated and tested |
| Duration/progress parsing | Parsing FFmpeg output | Existing `progress.go` | Handles \r line endings, throttling |
| Database access | Raw SQL everywhere | Centralized functions | Consistency, easier testing |

**Key insight:** The worker is just glue code connecting existing pieces. Phase 2 did the hard work.

## Common Pitfalls

### Pitfall 1: Not Tracking Current Job ID
**What goes wrong:** Cannot reset job on shutdown, gets stuck in "running" forever
**Why it happens:** Forgot to track which job is being processed
**How to avoid:** Use mutex-protected `currentJob int64` field, set before Engine.Run, clear after
**Warning signs:** Jobs stuck in "running" status after container restart

### Pitfall 2: Blocking on Stop() Forever
**What goes wrong:** Stop() hangs because done channel never closes
**Why it happens:** Forgot `defer close(w.done)` in run(), or deadlock in shutdown
**How to avoid:** Always `defer close(w.done)` at start of run()
**Warning signs:** Container takes >30s to stop

### Pitfall 3: Polling Too Frequently
**What goes wrong:** High CPU/DB load, SQLite lock contention
**Why it happens:** Polling every 100ms instead of 3s
**How to avoid:** Use 3s interval (matches requirement WRK-02), single query per poll
**Warning signs:** High CPU when queue is empty

### Pitfall 4: Race Between Shutdown and Job Completion
**What goes wrong:** Job completes normally but also gets reset to pending (runs twice)
**Why it happens:** Shutdown resets after job already marked completed
**How to avoid:** Only reset if currentJob != 0 AND status is still "running"
**Warning signs:** Same job processed multiple times

### Pitfall 5: Partial Output Files Left on Cancel
**What goes wrong:** Incomplete `*_4k.mkv` files litter disk after cancelled jobs
**Why it happens:** Engine.Run already handles this - removes output on error/cancel
**How to avoid:** Trust existing Engine.Run cleanup (see engine.go:91)
**Warning signs:** N/A - already handled

## Code Examples

### Database Functions Needed

```go
// GetNextPendingJob returns the oldest pending job, or nil if none.
func GetNextPendingJob() (*models.UpscaleJob, error) {
    var job models.UpscaleJob
    var errStr sql.NullString
    var startedAt, completedAt sql.NullTime
    
    err := database.DB.QueryRow(`
        SELECT id, input_path, output_path, preset, status, error, 
               created_at, started_at, completed_at
        FROM upscale_jobs
        WHERE status = 'pending'
        ORDER BY created_at ASC
        LIMIT 1
    `).Scan(&job.ID, &job.InputPath, &job.OutputPath, &job.Preset, 
           &job.Status, &errStr, &job.CreatedAt, &startedAt, &completedAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    if errStr.Valid {
        job.Error = &errStr.String
    }
    if startedAt.Valid {
        job.StartedAt = &startedAt.Time
    }
    if completedAt.Valid {
        job.CompletedAt = &completedAt.Time
    }
    
    return &job, nil
}

// UpdateJobStatus sets job status and timestamps atomically.
func UpdateJobStatus(id int64, status string, errMsg *string) error {
    var err error
    switch status {
    case models.UpscaleStatusRunning:
        _, err = database.DB.Exec(`
            UPDATE upscale_jobs 
            SET status = ?, started_at = CURRENT_TIMESTAMP 
            WHERE id = ?`, status, id)
    case models.UpscaleStatusCompleted:
        _, err = database.DB.Exec(`
            UPDATE upscale_jobs 
            SET status = ?, completed_at = CURRENT_TIMESTAMP 
            WHERE id = ?`, status, id)
    case models.UpscaleStatusFailed:
        _, err = database.DB.Exec(`
            UPDATE upscale_jobs 
            SET status = ?, error = ?, completed_at = CURRENT_TIMESTAMP 
            WHERE id = ?`, status, errMsg, id)
    case models.UpscaleStatusPending:
        // Reset - clear started_at
        _, err = database.DB.Exec(`
            UPDATE upscale_jobs 
            SET status = ?, started_at = NULL 
            WHERE id = ?`, status, id)
    }
    return err
}

// ResetRunningJob resets a specific running job to pending (for shutdown).
func ResetRunningJob(id int64) error {
    result, err := database.DB.Exec(`
        UPDATE upscale_jobs 
        SET status = 'pending', started_at = NULL 
        WHERE id = ? AND status = 'running'`, id)
    if err != nil {
        return err
    }
    affected, _ := result.RowsAffected()
    if affected == 0 {
        // Job already completed/failed - no reset needed
        return nil
    }
    return nil
}
```

### WebSocket Broadcast Types

```go
// Already exists: upscale_progress (from progress callback)
w.hub.Broadcast(models.WSMessage{
    Type: "upscale_progress",
    Data: models.UpscaleProgress{...},
})

// New: upscale_complete (job finished successfully)
w.hub.Broadcast(models.WSMessage{
    Type: "upscale_complete",
    Data: map[string]interface{}{
        "jobId":      job.ID,
        "outputPath": job.OutputPath,
    },
})

// New: upscale_failed (job errored)
w.hub.Broadcast(models.WSMessage{
    Type: "upscale_failed",
    Data: map[string]interface{}{
        "jobId": job.ID,
        "error": errMsg,
    },
})
```

### Integration in main.go

```go
// Create upscale worker (after hub creation)
upscaleWorker := upscale.NewWorker(hub, cfg.ShaderDir)
upscaleWorker.Start()
defer upscaleWorker.Stop()

// Add to Server struct for API access if needed
server.UpscaleWorker = upscaleWorker
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| External queue (Redis) | In-process SQLite polling | Project decision | Simpler deployment, no extra dependency |
| Multiple workers | Single worker | Project constraint | Single GPU = sequential processing |

**Deprecated/outdated:**
- N/A - this is a new implementation following established project patterns

## Open Questions

1. **Should worker be field on api.Server?**
   - What we know: DownloadMonitor and Poller are not on Server, just started in main
   - What's unclear: API handlers might need to check worker status
   - Recommendation: Start simple (not on Server), add if API needs it

2. **Cancel context storage location**
   - What we know: Need to cancel Engine.Run on shutdown
   - What's unclear: Whether to store cancelFunc on Worker or pass channel
   - Recommendation: Store cancelFunc on Worker, simplest approach

## Sources

### Primary (HIGH confidence)
- `internal/monitor/monitor.go` - DownloadMonitor lifecycle pattern
- `internal/rss/poller.go` - Ticker-based polling pattern
- `internal/upscale/engine.go` - Engine.Run signature and behavior
- `internal/models/models.go` - UpscaleJob, UpscaleProgress structs
- `internal/database/database.go` - upscale_jobs schema

### Secondary (MEDIUM confidence)
- `cmd/server/main.go` - Service initialization and shutdown order

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all stdlib + existing internal packages
- Architecture: HIGH - direct copy of existing DownloadMonitor pattern
- Pitfalls: HIGH - common Go concurrency patterns, verified against codebase

**Research date:** 2026-02-23
**Valid until:** 2026-03-23 (stable Go patterns, internal codebase)
