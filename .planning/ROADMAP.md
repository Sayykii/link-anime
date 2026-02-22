# Roadmap: Link-Anime Upscaling

## Overview

Add AI-powered video upscaling (1080p→4K) to Link-Anime. The journey builds from database foundation through FFmpeg engine, queue worker, REST API, WebSocket integration, and finally frontend UI — each phase delivering verifiable capability that enables the next.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foundation** - Database schema, Go models, TypeScript interfaces (completed 2026-02-22)
- [x] **Phase 2: Engine** - FFmpeg runner with libplacebo upscaling and progress parsing (completed 2026-02-22)
- [x] **Phase 3: Worker** - Single-job queue worker with Start/Stop lifecycle (completed 2026-02-22)
- [ ] **Phase 4: API** - REST endpoints for job CRUD and pipeline probe
- [ ] **Phase 5: WebSocket** - Real-time progress/status broadcasts
- [ ] **Phase 6: Frontend Integration** - API client methods and WebSocket listeners
- [ ] **Phase 7: Frontend UI** - Downloads view components and Upscale Queue tab

## Phase Details

### Phase 1: Foundation
**Goal**: Database schema and models exist for upscale job persistence
**Depends on**: Nothing (first phase)
**Requirements**: DB-01, DB-02, DB-03, DB-04
**Success Criteria** (what must be TRUE):
  1. `upscale_jobs` table exists in SQLite with status enum (pending/running/completed/failed/cancelled)
  2. UpscaleJob Go model serializes to JSON matching API conventions
  3. UpscaleProgress model captures frame/fps/time/percent fields
  4. TypeScript interfaces exist mirroring Go models for frontend type safety
**Plans**: 2 plans

Plans:
- [ ] 01-01-PLAN.md — Add upscale_jobs table migration
- [ ] 01-02-PLAN.md — Create Go models and TypeScript interfaces

### Phase 2: Engine
**Goal**: FFmpeg can upscale a video file with progress reporting
**Depends on**: Phase 1 (needs models for progress struct)
**Requirements**: ENG-01, ENG-02, ENG-03, ENG-04
**Success Criteria** (what must be TRUE):
  1. FFmpeg command executes with libplacebo + Anime4K shader via Vulkan
  2. Progress callback receives frame/fps/time updates from stderr parsing
  3. ffprobe detects input duration for percentage calculation
  4. Context cancellation kills FFmpeg process and cleans up partial output
**Plans**: 2 plans

Plans:
- [ ] 02-01-PLAN.md — Engine struct with FFmpeg runner and ffprobe duration detection
- [ ] 02-02-PLAN.md — Progress parsing from FFmpeg stderr

### Phase 3: Worker
**Goal**: Queue worker processes jobs sequentially with graceful lifecycle
**Depends on**: Phase 2 (needs engine to run jobs)
**Requirements**: WRK-01, WRK-02, WRK-03, WRK-04
**Success Criteria** (what must be TRUE):
  1. Worker starts/stops following DownloadMonitor pattern
  2. Pending jobs are picked up within 3 seconds (oldest first)
  3. Job status transitions correctly (pending→running→completed/failed)
  4. Graceful shutdown resets running job to pending for restart
**Plans**: 2 plans

Plans:
- [ ] 03-01-PLAN.md — Queue worker with Start/Stop lifecycle and DB polling
- [ ] 03-02-PLAN.md — Job processing with status transitions and main.go integration

### Phase 4: API
**Goal**: REST endpoints expose job management to frontend
**Depends on**: Phase 3 (needs worker for cancel operations)
**Requirements**: API-01, API-02, API-03, API-04, API-05, API-06
**Success Criteria** (what must be TRUE):
  1. GET /api/upscale/jobs returns all jobs (sorted by created_at)
  2. POST /api/upscale/jobs validates input file exists and queues job
  3. DELETE /api/upscale/jobs/{id} removes pending/completed/failed jobs
  4. POST /api/upscale/jobs/{id}/cancel stops running job
  5. GET /api/upscale/probe returns pipeline availability (Vulkan, FFmpeg, shaders)
**Plans**: 2 plans

Plans:
- [ ] 04-01-PLAN.md — Database CRUD functions and job list/create/get/delete handlers
- [ ] 04-02-PLAN.md — Cancel endpoint with Worker.CancelJob and probe endpoint

### Phase 5: WebSocket
**Goal**: Verify and document WebSocket message contracts for frontend
**Depends on**: Phase 3 (worker emits progress events)
**Requirements**: WS-01, WS-02, WS-03
**Success Criteria** (what must be TRUE):
  1. upscale_progress messages broadcast at ~1s interval during encode
  2. upscale_complete message broadcasts on successful completion
  3. upscale_failed message broadcasts on error (with error message)
**Plans**: 1 plan

Plans:
- [ ] 05-01-PLAN.md — Verify WebSocket broadcasts and add TypeScript message contracts

### Phase 6: Frontend Integration
**Goal**: Frontend can communicate with upscale API and receive real-time updates
**Depends on**: Phase 4, Phase 5 (needs API and WebSocket)
**Requirements**: FE-01, FE-02
**Success Criteria** (what must be TRUE):
  1. useApi has methods for all upscale endpoints (list, create, get, delete, cancel, probe)
  2. WebSocket listeners handle upscale_progress/complete/failed messages
  3. Pinia store (or composable) maintains upscale jobs state
**Plans**: TBD

Plans:
- [ ] 06-01: API client methods and WebSocket listeners

### Phase 7: Frontend UI
**Goal**: Users can trigger, monitor, and manage upscale jobs from Downloads view
**Depends on**: Phase 6 (needs API integration)
**Requirements**: UI-01, UI-02, UI-03, UI-04, UI-05, UI-06, UI-07
**Success Criteria** (what must be TRUE):
  1. Upscale button appears on download items when pipeline available
  2. Preset picker dialog offers Fast/Balanced/Quality options
  3. Upscale Queue tab shows all jobs with status
  4. Running jobs display progress bar + FPS (updated via WebSocket)
  5. Cancel button stops running/pending jobs; Delete button removes completed/failed
  6. 4K badge appears on items with completed upscale
**Plans**: TBD

Plans:
- [ ] 07-01: Upscale button and preset picker
- [ ] 07-02: Upscale Queue tab
- [ ] 07-03: Job controls and 4K badge

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 2/2 | Complete   | 2026-02-22 |
| 2. Engine | 2/2 | Complete   | 2026-02-22 |
| 3. Worker | 2/2 | Complete   | 2026-02-22 |
| 4. API | 2/2 | Complete | 2026-02-22 |
| 5. WebSocket | 0/1 | Not started | - |
| 6. Frontend Integration | 0/1 | Not started | - |
| 7. Frontend UI | 0/3 | Not started | - |

---
*Roadmap created: 2026-02-22*
