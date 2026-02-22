# External Integrations

**Analysis Date:** 2026-02-22

## APIs & External Services

**qBittorrent Web API:**
- Purpose: Torrent download management
- Client: `internal/qbit/client.go`
- Auth: Cookie-based session (`/api/v2/auth/login`)
- Env vars: `LA_QBIT_URL`, `LA_QBIT_USER`, `LA_QBIT_PASS`, `LA_QBIT_CATEGORY`
- Features used:
  - List torrents by category
  - Add magnet links
  - Delete torrents (with optional file removal)
  - Session management with auto-reauth on 403

**Shoko Server API:**
- Purpose: Anime metadata management and library scanning
- Client: `internal/shoko/client.go`
- Auth: API key header (`apikey`)
- Env vars: `LA_SHOKO_URL`, `LA_SHOKO_APIKEY`
- Features used:
  - Get import folder list
  - Trigger folder scans
  - Connection testing

**Nyaa.si RSS:**
- Purpose: Anime torrent search and RSS auto-download
- Client: `internal/nyaa/search.go`
- Auth: None (public API)
- Base URL: `https://nyaa.si`
- Features used:
  - RSS feed search (`?page=rss&q=...&c=1_2`)
  - Anime English-translated category (`c=1_2`)
  - Filter modes: trusted, noremakes
  - Magnet link extraction from result pages

## Notification Services

**Discord Webhooks:**
- Purpose: Rich embed notifications
- Detection: URL contains "discord"
- Format: JSON embed with title, description, fields, color, timestamp
- Implementation: `internal/notify/notify.go`

**ntfy:**
- Purpose: Push notifications
- Detection: URL contains "ntfy"
- Format: POST body with title header
- Implementation: `internal/notify/notify.go`

**Generic Webhook:**
- Purpose: Fallback notification endpoint
- Format: JSON `{ title, message }`
- Implementation: `internal/notify/notify.go`

**Configuration:**
- Env var: `LA_NOTIFY_URL`
- Auto-detection of service type from URL

## Data Storage

**Databases:**
- SQLite 3 (embedded)
- Connection: `$LA_DATA_DIR/link-anime.db`
- Driver: `modernc.org/sqlite` (pure Go, no CGO)
- ORM: None (raw SQL with database/sql)

**File Storage:**
- Local filesystem (hardlink-based)
- Download source: `$LA_DOWNLOAD_DIR`
- Series destination: `$LA_MEDIA_DIR`
- Movies destination: `$LA_MOVIES_DIR`
- Data persistence: `/app/data` volume

**Caching:**
- In-memory session store (Go map with mutex)
- Sessions expire after 24 hours
- Cleanup runs every hour

## Authentication & Identity

**Auth Provider:** Custom (single-user password auth)

**Implementation:** `internal/auth/auth.go`

**Approach:**
- Password hashed with bcrypt (default cost)
- Hash stored in SQLite `settings` table
- Session tokens: 32 random bytes (hex encoded)
- Cookie-based sessions (`link-anime-session`)
- HttpOnly, SameSite=Lax, 24-hour expiry

**API Auth:**
- Middleware: `auth.Middleware` on protected routes
- Public routes: `/api/auth/login`, `/api/auth/logout`, `/api/auth/check`

## Real-time Communication

**WebSocket Hub:**
- Implementation: `internal/ws/hub.go`
- Library: `github.com/gorilla/websocket`
- Endpoint: `/api/ws` (protected)

**Message Types Broadcast:**
- `link:progress` - File linking progress
- `link:complete` - Link operation completed
- `rss_match` - New RSS rule match found
- `downloads:progress` - Torrent download status updates

## Background Services

**RSS Poller:**
- Implementation: `internal/rss/poller.go`
- Interval: 15 minutes (configurable, minimum 5 minutes)
- Runs immediately on start, then on ticker
- Broadcasts matches via WebSocket

**Download Monitor:**
- Implementation: `internal/monitor/monitor.go`
- Interval: 5 seconds
- Polls qBittorrent for torrent status
- Broadcasts progress via WebSocket
- Sends notifications on download completion

## Video Processing

**Upscale Pipeline:**
- Implementation: `internal/upscale/probe.go`
- Tools: FFmpeg with libplacebo filter
- GPU: Vulkan (optional, for hardware acceleration)
- Presets: `shaders/mode-a-{fast,balanced,quality}.glsl`
- Capability probing at startup

## CI/CD & Deployment

**Hosting:**
- Self-hosted via Docker Compose
- Single-container deployment

**CI Pipeline:**
- Not detected (no `.github/workflows/`, no CI config)

**Deployment:**
- `compose.yaml` - Docker Compose service definition
- `deploy.sh` - Deployment script
- `Dockerfile` - Multi-stage build
- `entrypoint.sh` - Container initialization

## Environment Configuration

**Required env vars:**
- `LA_PASSWORD` - Admin password (required for initial setup)

**Optional env vars:**
- `LA_PORT` - Server port (default: 8787)
- `LA_DATA_DIR` - Persistent data directory
- `LA_DOWNLOAD_DIR`, `LA_MEDIA_DIR`, `LA_MOVIES_DIR` - Path configuration
- `LA_QBIT_URL`, `LA_QBIT_USER`, `LA_QBIT_PASS`, `LA_QBIT_CATEGORY` - qBittorrent
- `LA_SHOKO_URL`, `LA_SHOKO_APIKEY` - Shoko Server
- `LA_NOTIFY_URL` - Notification webhook
- `PUID`, `PGID` - Container user/group IDs
- `TZ` - Timezone

**Secrets location:**
- `.env` file (gitignored)
- `.env.example` - Template with placeholder values
- Settings UI persists to SQLite database (survives restarts)

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- Notification webhooks (Discord, ntfy, generic)
- Triggered on: download completion, RSS matches

## Network Architecture

**Internal:**
- HTTP server on configurable port (default: 8787)
- WebSocket on same port (`/api/ws`)
- Frontend SPA embedded in binary via `embed.FS`

**External:**
- qBittorrent: HTTP API (typically `host.docker.internal:8080`)
- Shoko Server: HTTP API (typically `host.docker.internal:8111`)
- Nyaa.si: HTTPS RSS/HTML

**Docker Networking:**
- `host.docker.internal:host-gateway` for accessing host services
- Port 8787 exposed

---

*Integration audit: 2026-02-22*
