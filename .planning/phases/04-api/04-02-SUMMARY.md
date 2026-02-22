---
phase: 04-api
plan: 02
subsystem: api
tags: [rest, cancel, probe, go, chi, upscale]

# Dependency graph
requires:
  - phase: 04-api/01
    provides: Upscale handlers and Server struct
provides:
  - Cancel endpoint for running upscale jobs
  - Probe endpoint for pipeline availability check
  - Worker.CancelJob method for external job cancellation
affects: [05-frontend, 06-integration]

# Tech tracking
tech-stack:
  added: []
  patterns: [cancel via context, probe pattern]

key-files:
  created: []
  modified:
    - internal/upscale/worker.go
    - internal/api/router.go
    - internal/api/upscale_handler.go

key-decisions:
  - "Cancel returns 409 Conflict if job no longer running (race condition handling)"
  - "Probe returns ProbeResult struct with FFmpegFound, LibplaceboOK, VulkanDevice"

patterns-established:
  - "Cancel uses Worker.CancelJob for goroutine-safe cancellation"

requirements-completed: [API-05, API-06]

# Metrics
duration: 1min
completed: 2026-02-22
---

# Phase 04 Plan 02: Upscale Cancel and Probe Endpoints Summary

**Cancel endpoint for running jobs with context cancellation and probe endpoint returning FFmpeg/libplacebo/Vulkan availability**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-22T22:55:42Z
- **Completed:** 2026-02-22T22:56:48Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Worker.CancelJob(id) method for external job cancellation
- POST /api/upscale/jobs/{id}/cancel endpoint with proper error handling
- GET /api/upscale/probe endpoint returning pipeline status
- Server.Worker field for handler access to upscale worker

## Task Commits

Each task was committed atomically:

1. **Task 1: Add CancelJob method to Worker** - `46b7c94` (feat)
2. **Task 2: Add Server.Worker field and cancel/probe handlers** - `9d8af9f` (feat)

## Files Created/Modified
- `internal/upscale/worker.go` - Added CancelJob method for external cancellation
- `internal/api/router.go` - Added Worker field and cancel/probe routes
- `internal/api/upscale_handler.go` - Added handleCancelUpscaleJob and handleUpscaleProbe handlers

## Decisions Made
- Cancel returns 409 Conflict if job finished before cancellation (race condition)
- Probe endpoint reuses existing upscale.Probe() function

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All API endpoints complete for upscale feature
- Phase 04 complete, ready for 05-frontend

## Self-Check: PASSED

---
*Phase: 04-api*
*Completed: 2026-02-22*
