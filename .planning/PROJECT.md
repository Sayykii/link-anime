# Link-Anime Upscaling Feature

## What This Is

Video upscaling feature for an existing anime media management application. Adds AI-powered 2x upscaling using FFmpeg + libplacebo with Anime4K shaders, enabling 1080p→4K conversion for locally downloaded anime files while preserving originals for torrent seeding.

## Core Value

Users can upscale their anime downloads to 4K quality through a simple queue-based interface, with real-time progress feedback.

## Requirements

### Validated

<!-- Existing capabilities from codebase -->

- ✓ Go backend with chi router, SQLite persistence — existing
- ✓ Vue 3 frontend with Pinia, WebSocket integration — existing
- ✓ Download monitoring with real-time progress via WebSocket — existing
- ✓ qBittorrent integration for torrent management — existing
- ✓ FFmpeg available in container with libplacebo + Vulkan — existing

### Active

<!-- Current scope for upscaling feature -->

- [ ] Database schema for upscale job queue
- [ ] Upscale job models (Go + TypeScript)
- [ ] FFmpeg runner with progress parsing
- [ ] Single-job queue worker with graceful shutdown
- [ ] REST API for job CRUD + cancel
- [ ] WebSocket messages for real-time progress
- [ ] Upscale button in Downloads view with preset picker
- [ ] Upscale queue tab showing all jobs
- [ ] 4K badge on items with completed upscale

### Out of Scope

- Library view integration (4K badges, revert controls) — deferred to future phase
- Configurable encoder settings — hardcoded defaults sufficient for v1
- Concurrent job processing — single GPU, sequential queue
- Custom output resolution — 2x upscale only
- Deleting/replacing originals — must preserve for seeding

## Context

**Existing patterns to follow:**
- Download monitor lifecycle (Start/Stop pattern)
- WebSocket hub broadcasting for real-time updates
- `jsonOK(w, data)` / `jsonError(w, msg, code)` for API responses
- Cookie-based session auth via `auth.Middleware`
- Pinia stores + composables for frontend state

**Technical environment:**
- FFmpeg command: `ffmpeg -init_hw_device vulkan -i <input> -vf "libplacebo=w=iw*2:h=ih*2:custom_shader_path=<shader>" -c:v libx265 -crf 16 -preset slow -c:a copy -c:s copy <output>`
- Output naming: `filename_4k.mkv` alongside original
- Progress parsing from FFmpeg stderr (frame=, fps=, time=)
- Duration via ffprobe for percentage calculation

## Constraints

- **Seeding**: Cannot modify/delete original files — torrents must remain seedable
- **Single GPU**: One upscale job at a time (sequential queue)
- **Container**: Must work within existing Docker setup with Vulkan passthrough

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Write alongside (`_4k.mkv`) not replace | Preserves originals for seeding | — Pending |
| libx265 -crf 16 -preset slow | Good quality/size balance for anime | — Pending |
| Sequential queue (1 job) | Single GPU, simpler implementation | — Pending |
| 2x resolution hardcoded | Covers 1080p→4K use case, simpler UI | — Pending |
| Phase 3 focused on Downloads only | Ship faster, defer Library integration | — Pending |

---
*Last updated: 2026-02-22 after initialization*
