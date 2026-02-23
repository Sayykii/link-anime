---
phase: 06-frontend-integration
plan: 01
subsystem: frontend
tags: [vue, pinia, typescript, websocket, api]

requires:
  - phase: 05-websocket
    provides: WebSocket broadcast types for upscale events
  - phase: 04-api
    provides: Upscale REST API endpoints
provides:
  - useApi upscale methods (6 endpoints)
  - useUpscaleStore Pinia store with reactive state
  - WebSocket listener setup for real-time updates
affects: [07-ui-components, upscale-view]

tech-stack:
  added: []
  patterns: [Pinia composition API store, WebSocket event listeners]

key-files:
  created:
    - frontend/src/stores/upscale.ts
  modified:
    - frontend/src/composables/useApi.ts
    - frontend/src/lib/types.ts

key-decisions:
  - "Progress stored as Record<number, UpscaleProgress> for O(1) lookup by jobId"
  - "setupListeners() called once after ws.connect() - not auto-invoked"

patterns-established:
  - "WebSocket listener setup via store method: store.setupListeners() after connect"

requirements-completed: [FE-01, FE-02]

duration: 1min
completed: 2026-02-23
---

# Phase 06 Plan 01: Upscale API & Store Summary

**Extended useApi with 6 upscale methods and created useUpscaleStore Pinia store with WebSocket event listeners for real-time progress updates**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-22T23:59:50Z
- **Completed:** 2026-02-23T00:00:50Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Added ProbeResult interface and 6 upscale API methods to useApi composable
- Created useUpscaleStore with jobs, progress, probeResult state and computed properties
- Implemented WebSocket listeners for upscale_progress, upscale_complete, upscale_failed events

## Task Commits

Each task was committed atomically:

1. **Task 1: Add upscale API methods and ProbeResult type** - `49fafb1` (feat)
2. **Task 2: Create upscale Pinia store with WebSocket listeners** - `71e9e78` (feat)

## Files Created/Modified
- `frontend/src/lib/types.ts` - Added ProbeResult interface
- `frontend/src/composables/useApi.ts` - Added 6 upscale API methods
- `frontend/src/stores/upscale.ts` - New Pinia store with reactive state and WebSocket handlers

## Decisions Made
- Used Record<number, UpscaleProgress> for progress map for O(1) job lookup
- setupListeners() is a manual call pattern (consistent with App.vue lifecycle)
- Unshift on createJob to show newest jobs first

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

Pre-existing TypeScript error in LinkWizardView.vue (unrelated to upscale changes) - out of scope per deviation rules.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Upscale API methods ready for UI components
- Store ready for consumption in UpscaleView (Phase 7)
- WebSocket listeners ready to be activated via setupListeners()

## Self-Check: PASSED

- [x] frontend/src/stores/upscale.ts exists
- [x] Commit 49fafb1 exists
- [x] Commit 71e9e78 exists

---
*Phase: 06-frontend-integration*
*Completed: 2026-02-23*
