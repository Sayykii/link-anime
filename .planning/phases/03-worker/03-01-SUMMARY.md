---
phase: 03-worker
plan: 01
subsystem: upscale
tags: [go, sqlite, worker, queue]

# Dependency graph
requires:
  - phase: 02-engine
    provides: Engine with Run method, ProgressCallback
provides:
  - Worker struct with Start/Stop lifecycle
  - Database functions for job queries (GetNextPendingJob, UpdateJobStatus, ResetRunningJob)
affects: [03-02, api]

# Tech tracking
tech-stack:
  added: []
  patterns: [DownloadMonitor lifecycle pattern reused]

key-files:
  created:
    - internal/database/upscale.go
    - internal/upscale/worker.go
  modified: []

key-decisions:
  - "Follow DownloadMonitor pattern exactly for lifecycle consistency"
  - "3-second poll interval for responsive job pickup"
  - "FIFO job ordering (oldest pending first)"

patterns-established:
  - "Worker lifecycle: stopCh/done channels with run() goroutine"
  - "Centralized DB functions per domain (upscale.go for upscale jobs)"

requirements-completed: [WRK-01, WRK-02]

# Metrics
duration: 1min
completed: 2026-02-22
---

# Phase 3 Plan 1: Worker Lifecycle Summary

**Worker struct with Start/Stop lifecycle following DownloadMonitor pattern, database functions for FIFO job queue polling**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-22T22:30:40Z
- **Completed:** 2026-02-22T22:31:44Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Created centralized database functions for upscale job queries
- Built Worker struct matching DownloadMonitor lifecycle pattern
- Configured 3-second poll interval per WRK-02 requirement

## Task Commits

Each task was committed atomically:

1. **Task 1: Create database functions for job queries** - `71fe204` (feat)
2. **Task 2: Create Worker struct with Start/Stop lifecycle** - `ecb6a8f` (feat)

## Files Created/Modified
- `internal/database/upscale.go` - GetNextPendingJob, UpdateJobStatus, ResetRunningJob functions
- `internal/upscale/worker.go` - Worker struct with Start/Stop/run/processNext/handleShutdown

## Decisions Made
- Follow DownloadMonitor pattern exactly for lifecycle consistency across the codebase
- 3-second poll interval balances responsiveness with resource usage
- FIFO ordering (ORDER BY created_at ASC) ensures fair job scheduling

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Worker lifecycle foundation complete
- Ready for Plan 02: Job processing implementation (actually running FFmpeg jobs)
- processNext() and handleShutdown() are stubs awaiting Plan 02 implementation

---
*Phase: 03-worker*
*Completed: 2026-02-22*

## Self-Check: PASSED
