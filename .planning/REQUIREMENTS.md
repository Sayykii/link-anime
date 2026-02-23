# Requirements: Link-Anime Upscaling

**Defined:** 2026-02-22
**Core Value:** Users can upscale their anime downloads to 4K quality through a simple queue-based interface, with real-time progress feedback.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Database & Models

- [x] **DB-01**: Upscale jobs table with status tracking (pending/running/completed/failed/cancelled)
- [x] **DB-02**: UpscaleJob Go model with JSON tags matching existing conventions
- [x] **DB-03**: UpscaleProgress WebSocket payload model
- [x] **DB-04**: TypeScript interfaces mirroring Go models

### Upscale Engine

- [x] **ENG-01**: FFmpeg runner with libplacebo 2x upscaling via Vulkan
- [x] **ENG-02**: Progress parsing from FFmpeg stderr (frame, fps, time)
- [x] **ENG-03**: Duration detection via ffprobe for percentage calculation
- [x] **ENG-04**: Context cancellation support to kill FFmpeg process

### Queue Worker

- [x] **WRK-01**: Single-job queue worker with Start/Stop lifecycle
- [x] **WRK-02**: Poll DB every 3s for pending jobs (oldest first)
- [x] **WRK-03**: Status transitions (pending→running→completed/failed)
- [x] **WRK-04**: Graceful shutdown (reset running job to pending)

### API Endpoints

- [x] **API-01**: GET /api/upscale/jobs — list all jobs
- [x] **API-02**: POST /api/upscale/jobs — queue new job with validation
- [x] **API-03**: GET /api/upscale/jobs/{id} — get single job
- [x] **API-04**: DELETE /api/upscale/jobs/{id} — cancel/delete job
- [x] **API-05**: POST /api/upscale/jobs/{id}/cancel — cancel running job
- [x] **API-06**: GET /api/upscale/probe — check pipeline availability

### WebSocket Messages

- [x] **WS-01**: upscale_progress broadcasts during encode (~1s interval)
- [x] **WS-02**: upscale_complete broadcast on success
- [x] **WS-03**: upscale_failed broadcast on error

### Frontend - Downloads View

- [ ] **UI-01**: Upscale button on download items (when probe available)
- [ ] **UI-02**: Preset picker dialog (Fast/Balanced/Quality)
- [ ] **UI-03**: Upscale Queue tab with job list
- [ ] **UI-04**: Progress bar + FPS for running jobs (via WebSocket)
- [ ] **UI-05**: Cancel button for running/pending jobs
- [ ] **UI-06**: Delete button for completed/failed jobs
- [ ] **UI-07**: 4K badge on items with completed upscale

### Frontend - API Integration

- [x] **FE-01**: useApi methods for all upscale endpoints
- [x] **FE-02**: WebSocket listeners for upscale_progress/complete/failed

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Library Integration

- **LIB-01**: 4K badges on episodes/files with upscaled versions
- **LIB-02**: Revert controls to delete upscaled version

### Advanced Settings

- **ADV-01**: Configurable encoder settings (codec, CRF, preset)
- **ADV-02**: Custom output resolution options
- **ADV-03**: Concurrent job processing limit

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Delete/replace originals | Must preserve for torrent seeding |
| Real-time preview | High complexity, not core value |
| Multiple GPU support | Single GPU sufficient for home use |
| Cloud/remote upscaling | Local processing only |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| DB-01 | Phase 1 | Complete |
| DB-02 | Phase 1 | Complete |
| DB-03 | Phase 1 | Complete |
| DB-04 | Phase 1 | Complete |
| ENG-01 | Phase 2 | Complete |
| ENG-02 | Phase 2 | Complete |
| ENG-03 | Phase 2 | Complete |
| ENG-04 | Phase 2 | Complete |
| WRK-01 | Phase 3 | Complete |
| WRK-02 | Phase 3 | Complete |
| WRK-03 | Phase 3 | Complete |
| WRK-04 | Phase 3 | Complete |
| API-01 | Phase 4 | Complete |
| API-02 | Phase 4 | Complete |
| API-03 | Phase 4 | Complete |
| API-04 | Phase 4 | Complete |
| API-05 | Phase 4 | Complete |
| API-06 | Phase 4 | Complete |
| WS-01 | Phase 5 | Complete |
| WS-02 | Phase 5 | Complete |
| WS-03 | Phase 5 | Complete |
| UI-01 | Phase 7 | Pending |
| UI-02 | Phase 7 | Pending |
| UI-03 | Phase 7 | Pending |
| UI-04 | Phase 7 | Pending |
| UI-05 | Phase 7 | Pending |
| UI-06 | Phase 7 | Pending |
| UI-07 | Phase 7 | Pending |
| FE-01 | Phase 6 | Complete |
| FE-02 | Phase 6 | Complete |

**Coverage:**
- v1 requirements: 30 total
- Mapped to phases: 30 ✓
- Unmapped: 0

---
*Requirements defined: 2026-02-22*
*Last updated: 2026-02-23 after 06-01-PLAN.md completion*
