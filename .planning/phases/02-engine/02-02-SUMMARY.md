---
phase: 02-engine
plan: 02
subsystem: upscale
tags: [ffmpeg, progress, stderr, regex, go]

# Dependency graph
requires:
  - phase: 02-engine/01
    provides: Engine struct with Run method and ProbeDuration
provides:
  - FFmpeg stderr progress parsing
  - parseProgress function with throttled callback updates
  - scanFFmpegLines for \r and \n handling
  - parseTimeToSeconds for HH:MM:SS.ms conversion
affects: [queue worker progress broadcasts, WebSocket integration]

# Tech tracking
tech-stack:
  added: [bufio.Scanner with custom SplitFunc]
  patterns: [throttled callback pattern, regex extraction from process output]

key-files:
  created:
    - internal/upscale/progress.go
  modified:
    - internal/upscale/engine.go

key-decisions:
  - "Throttle progress updates to ~1/second to avoid flooding WebSocket"
  - "Handle both \r and \n line endings (FFmpeg uses \r for in-place updates)"
  - "parseProgress handles nil callback internally (no-op drain)"

patterns-established:
  - "Progress callback pattern: throttled updates with percentage calculation"
  - "Custom bufio.SplitFunc for non-standard line endings"

requirements-completed: [ENG-02]

# Metrics
duration: 1min
completed: 2026-02-22
---

# Phase 02 Plan 02: Progress Parsing Summary

**FFmpeg stderr progress parsing with frame/fps/time extraction and throttled callback updates for real-time WebSocket broadcasts**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-22T22:10:23Z
- **Completed:** 2026-02-22T22:11:25Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Created progress parser with regex matching FFmpeg output format
- Implemented custom bufio.SplitFunc for \r and \n line endings
- Integrated parseProgress into Engine.Run goroutine
- Percentage calculated from current time vs total duration via ProbeDuration

## Task Commits

Each task was committed atomically:

1. **Task 1: Create progress parser** - `e040c7a` (feat)
2. **Task 2: Integrate progress parsing into Engine.Run** - `04104ae` (feat)

**Plan metadata:** `c711887` (docs: complete plan)

## Files Created/Modified
- `internal/upscale/progress.go` - Progress parser with progressRegex, scanFFmpegLines, parseTimeToSeconds, parseProgress
- `internal/upscale/engine.go` - Updated Run method to call parseProgress with duration

## Decisions Made
- Throttle updates to 1/second to avoid flooding WebSocket clients
- Handle both \r (FFmpeg in-place) and \n line endings in scanner
- parseProgress handles nil callback by draining to io.Discard

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Engine phase complete with all progress parsing in place
- Ready for Phase 3: Queue worker that uses Engine.Run with progress callback
- Progress updates will be forwarded to WebSocket in Phase 5

---
*Phase: 02-engine*
*Completed: 2026-02-22*

## Self-Check: PASSED

- [x] internal/upscale/progress.go exists
- [x] internal/upscale/engine.go exists
- [x] Commit e040c7a (Task 1) found
- [x] Commit 04104ae (Task 2) found
