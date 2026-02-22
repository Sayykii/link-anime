# Architecture

**Analysis Date:** 2026-02-22

## Pattern Overview

**Overall:** Modular Monolith with Go Backend + Vue SPA Frontend

**Key Characteristics:**
- Single Go binary serves API + embedded Vue SPA
- Domain-driven package organization in `internal/`
- WebSocket hub for real-time progress/status updates
- Cookie-based session authentication
- SQLite for persistence, env vars for config (DB overrides env)

## Layers

**Presentation Layer (API Handlers):**
- Purpose: HTTP request handling, validation, JSON serialization
- Location: `internal/api/`
- Contains: Route handlers, middleware integration, WebSocket endpoint
- Depends on: All domain packages, `ws.Hub`
- Used by: Frontend SPA, external clients

**Domain Services:**
- Purpose: Business logic for specific concerns
- Location: `internal/{domain}/` (linker, scanner, parser, rss, monitor)
- Contains: Core algorithms, file operations, polling loops
- Depends on: `models`, `database`, `ws` (for broadcasts)
- Used by: API handlers

**External Integrations:**
- Purpose: Third-party API clients
- Location: `internal/qbit/`, `internal/shoko/`, `internal/nyaa/`, `internal/notify/`
- Contains: HTTP clients with auth, response parsing
- Depends on: `models` (for return types)
- Used by: API handlers, RSS poller, download monitor

**Data Layer:**
- Purpose: Persistence (SQLite), settings storage
- Location: `internal/database/`
- Contains: Global DB connection, migrations, settings CRUD
- Depends on: None (except stdlib sql)
- Used by: All packages needing persistence

**Real-time Layer:**
- Purpose: WebSocket connection management, broadcasting
- Location: `internal/ws/`
- Contains: Hub (connection registry), Client (goroutine pumps)
- Depends on: None
- Used by: Linker (progress), monitor (torrents), rss (matches), handlers

**Configuration:**
- Purpose: Environment variable loading, defaults
- Location: `internal/config/`
- Contains: Config struct, env var parsing
- Depends on: None
- Used by: `cmd/server/main.go`, API handlers (via Server struct)

**Models:**
- Purpose: Shared data structures (DTOs)
- Location: `internal/models/`
- Contains: Go structs with JSON tags
- Depends on: None
- Used by: All packages

**Frontend (Vue 3 SPA):**
- Purpose: User interface
- Location: `frontend/src/`
- Contains: Views, components, stores (Pinia), composables
- Depends on: Backend API via `/api/*`
- Used by: End users via browser

## Data Flow

**Link Operation Flow:**

1. User selects download item in `LinkWizardView.vue`
2. Frontend POSTs to `/api/link` via `useApi().link()`
3. `api/link_handler.go` validates request, calls `linker.Link()`
4. `linker/linker.go` scans source, creates hardlinks, broadcasts progress via `ws.Hub`
5. Frontend receives `link:progress` WS messages, updates progress bar
6. Linker writes history to SQLite, broadcasts `link:complete`
7. Handler triggers Shoko scan (async) and sends notification

**Real-time Download Monitoring:**

1. `monitor.NewDownloadMonitor()` starts polling goroutine (5s interval)
2. Polls qBittorrent via `qbit.Client.ListTorrents()`
3. Compares progress to previous tick, detects completions
4. Broadcasts `torrent_progress` (all) and `download_complete` (newly finished) via WS
5. Sends external notification on completion

**RSS Auto-Download Flow:**

1. `rss.Poller.Start()` runs poll loop (15m interval)
2. For each enabled rule, searches Nyaa RSS feed
3. Filters by min seeders, resolution
4. Deduplicates via hash stored in `rss_matches` table
5. Adds magnet to qBittorrent if configured
6. Broadcasts `rss_match` via WebSocket

**State Management:**
- Backend: Settings in SQLite (DB > env vars priority), session tokens in memory map
- Frontend: Pinia stores for `auth` and `library` state

## Key Abstractions

**Server (api.Server):**
- Purpose: Dependency injection container for handlers
- Examples: `internal/api/router.go` line 20-27
- Pattern: Struct with pointer fields for all integration clients

**Hub (ws.Hub):**
- Purpose: Fan-out WebSocket messages to all connected clients
- Examples: `internal/ws/hub.go`
- Pattern: Gorilla WebSocket with connection map + mutex

**Client Pattern (qbit, shoko, notify):**
- Purpose: Encapsulated HTTP client with auth
- Examples: `internal/qbit/client.go`, `internal/shoko/client.go`
- Pattern: Struct with baseURL, creds, `*http.Client`; `IsConfigured()` check

**Getter Functions:**
- Purpose: Allow background services to pick up reinitClients() changes
- Examples: `func() *qbit.Client { return server.Qbit }` in `main.go`
- Pattern: Closure returning current pointer, not copy

## Entry Points

**Server Main:**
- Location: `cmd/server/main.go`
- Triggers: `go run ./cmd/server`, Docker container start
- Responsibilities: Load config, init DB, create clients, start pollers/monitors, serve HTTP

**API Router:**
- Location: `internal/api/router.go` `NewRouter()`
- Triggers: Every HTTP request
- Responsibilities: Route matching, middleware chain, static file serving

**WebSocket Endpoint:**
- Location: `internal/api/ws_handler.go` (via `/api/ws`)
- Triggers: Frontend opens WS connection
- Responsibilities: Upgrade HTTP, register client with Hub

**Embedded Frontend:**
- Location: `cmd/server/main.go` `//go:embed all:dist`
- Triggers: `GET /*` (non-API routes)
- Responsibilities: Serve built Vue SPA, fallback to index.html for SPA routing

## Error Handling

**Strategy:** Return errors up the call stack, handlers convert to JSON

**Patterns:**
- Domain functions return `(result, error)` tuples
- Handlers call `jsonError(w, msg, status)` for client errors
- Background services log errors, don't crash (poller, monitor)
- Non-fatal warnings logged to stderr (e.g., history write failure)

**Error Response Format:**
```json
{"error": "human readable message"}
```

## Cross-Cutting Concerns

**Logging:**
- `log.Printf()` throughout with `[domain]` prefixes
- `log.SetFlags(log.LstdFlags | log.Lshortfile)` in main

**Validation:**
- Request body validation in handlers (required fields, enum values)
- File existence checks in linker/scanner

**Authentication:**
- Cookie-based sessions (`link-anime-session`)
- `auth.Middleware` wraps protected routes
- Session cleanup goroutine runs hourly

---

*Architecture analysis: 2026-02-22*
