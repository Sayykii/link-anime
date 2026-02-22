---
phase: 02-engine
plan: 01
subsystem: upscale
tags: [ffmpeg, libplacebo, vulkan, anime4k, go]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: UpscaleJob and UpscaleProgress models
provides:
  - Engine struct with Run method for FFmpeg execution
  - ProbeDuration for video duration detection
  - GenerateOutputPath for _4k.mkv naming
affects: [02-02 progress parsing, queue processor]

# Tech tracking
tech-stack:
  added: [os/exec with CommandContext]
  patterns: [subprocess with context cancellation, progress callback pattern]

key-files:
  created:
    - internal/upscale/engine.go
    - internal/upscale/ffprobe.go
  modified: []

key-decisions:
  - "Reuse Presets map from probe.go for shader path mapping"
  - "Always output MKV format regardless of input format"
  - "Drain stderr in goroutine (parsing deferred to Plan 02)"

patterns-established:
  - "Engine pattern: struct with Run(ctx, job, callback) for long-running tasks"
  - "exec.CommandContext for all subprocess execution (enables cancellation)"

requirements-completed: [ENG-01, ENG-03, ENG-04]

# Metrics
duration: 1min
completed: 2026-02-22
---

# Phase 02 Plan 01: FFmpeg Upscaling Engine Summary

**FFmpeg upscaling engine with libplacebo filter, Vulkan GPU acceleration, and Anime4K shader support**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-22T22:06:51Z
- **Completed:** 2026-02-22T22:08:06Z
- **Tasks:** 3
- **Files modified:** 2

## Accomplishments
- Engine struct wrapping FFmpeg with libplacebo 2x upscaling via Vulkan
- ProbeDuration function for video duration detection via ffprobe
- GenerateOutputPath helper creating `_4k.mkv` output paths
- Context cancellation support with partial file cleanup

## Task Commits

Each task was committed atomically:

1. **Task 1: Create ffprobe duration detection** - `343a8fd` (feat)
2. **Task 2: Create Engine struct with FFmpeg runner** - `dca0aa1` (feat)
3. **Task 3: Add GenerateOutputPath helper** - `8d0b228` (feat)

**Plan metadata:** `8ff0a9f` (docs: complete plan)

## Files Created/Modified
- `internal/upscale/ffprobe.go` - ProbeDuration function using ffprobe for video duration
- `internal/upscale/engine.go` - Engine struct with Run method, buildCommand, getShaderPath, GenerateOutputPath

## Decisions Made
- Reuse existing `Presets` map from probe.go rather than duplicating shader mappings
- Always output MKV format regardless of input format (per project decision)
- Defer stderr parsing to Plan 02 - drain stderr in goroutine for now

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Engine foundation complete, ready for Plan 02 (progress parsing)
- Run method signature ready for progress callback integration
- Duration detection in place for percentage calculation

---
*Phase: 02-engine*
*Completed: 2026-02-22*

## Self-Check: PASSED

- [x] internal/upscale/engine.go exists
- [x] internal/upscale/ffprobe.go exists
- [x] Commit 343a8fd (Task 1) found
- [x] Commit dca0aa1 (Task 2) found
- [x] Commit 8d0b228 (Task 3) found
