# Codebase Structure

**Analysis Date:** 2026-02-22

## Directory Layout

```
link-anime/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/                 # Go packages (not importable externally)
│   ├── api/                  # HTTP handlers and router
│   ├── auth/                 # Password hashing, sessions
│   ├── config/               # Environment variable loading
│   ├── database/             # SQLite connection and migrations
│   ├── linker/               # Hardlink creation logic
│   ├── models/               # Shared data structures
│   ├── monitor/              # Download progress polling
│   ├── notify/               # External notification sender
│   ├── nyaa/                 # Nyaa RSS search client
│   ├── parser/               # Release name parsing
│   ├── qbit/                 # qBittorrent API client
│   ├── rss/                  # RSS rule polling
│   ├── scanner/              # Library/download directory scanning
│   ├── shoko/                # Shoko Server API client
│   ├── upscale/              # FFmpeg/libplacebo probe (future)
│   └── ws/                   # WebSocket hub and clients
├── frontend/                 # Vue 3 SPA
│   ├── src/
│   │   ├── assets/           # CSS (Tailwind entry)
│   │   ├── components/       # Vue components
│   │   │   ├── layout/       # App layout (sidebar)
│   │   │   └── ui/           # shadcn/ui components
│   │   ├── composables/      # Vue composition API hooks
│   │   ├── lib/              # Utilities and types
│   │   ├── router/           # Vue Router config
│   │   ├── stores/           # Pinia state stores
│   │   ├── views/            # Page-level components
│   │   ├── App.vue           # Root component
│   │   └── main.ts           # Vue app bootstrap
│   ├── public/               # Static assets
│   ├── package.json          # npm dependencies
│   └── vite.config.ts        # Vite build config
├── shaders/                  # GLSL shaders (upscale feature)
│   └── src/                  # Shader source files
├── .planning/                # Project planning docs
│   └── codebase/             # Architecture/convention docs
├── compose.yaml              # Docker Compose config
├── Dockerfile                # Multi-stage build
├── deploy.sh                 # Deployment script
├── entrypoint.sh             # Container entrypoint
├── go.mod                    # Go module definition
└── README.md                 # Project documentation
```

## Directory Purposes

**`cmd/server/`:**
- Purpose: Application entry point
- Contains: Single `main.go` with initialization sequence
- Key files: `main.go` (136 lines)

**`internal/api/`:**
- Purpose: HTTP layer - all API endpoints
- Contains: Router setup, individual handler files per domain
- Key files:
  - `router.go` - Chi router with route definitions
  - `*_handler.go` - Domain-specific handlers (auth, library, link, downloads, qbit, nyaa, rss, settings, history, ws)
  - `helpers.go` - Directory getters, client constructors

**`internal/models/`:**
- Purpose: Shared data structures between packages
- Contains: DTOs with JSON tags for API serialization
- Key files: `models.go` (single file with all structs)

**`internal/linker/`:**
- Purpose: Core hardlinking business logic
- Contains: Link, unlink, undo operations with history tracking
- Key files: `linker.go` (560 lines - largest domain file)

**`internal/scanner/`:**
- Purpose: Filesystem scanning for library/downloads
- Contains: Video file detection, season directory parsing
- Key files: `scanner.go`

**`internal/parser/`:**
- Purpose: Anime release name parsing
- Contains: Regex-based extraction of show name, season, year
- Key files: `parser.go`, `parser_test.go`

**`internal/rss/`:**
- Purpose: RSS auto-download rules and polling
- Contains: Poller, rule CRUD, match tracking
- Key files: `poller.go`

**`internal/monitor/`:**
- Purpose: qBittorrent download progress monitoring
- Contains: Polling loop, completion detection
- Key files: `monitor.go`

**`internal/qbit/`:**
- Purpose: qBittorrent Web API client
- Contains: Login, torrent listing, add/delete operations
- Key files: `client.go`

**`internal/shoko/`:**
- Purpose: Shoko Server API client
- Contains: Import folder scanning trigger
- Key files: `client.go`

**`internal/nyaa/`:**
- Purpose: Nyaa RSS feed search
- Contains: RSS parsing, magnet extraction
- Key files: `search.go`

**`internal/notify/`:**
- Purpose: External notifications (Discord, ntfy)
- Contains: Auto-detect notification service, send messages
- Key files: `notify.go`

**`internal/database/`:**
- Purpose: SQLite persistence
- Contains: Connection management, migrations, settings CRUD
- Key files: `database.go`

**`internal/auth/`:**
- Purpose: Authentication and session management
- Contains: Password hashing (bcrypt), session tokens, middleware
- Key files: `auth.go`

**`internal/config/`:**
- Purpose: Configuration loading
- Contains: Env var parsing with defaults
- Key files: `config.go`

**`internal/ws/`:**
- Purpose: WebSocket infrastructure
- Contains: Hub for broadcasting, Client for connection management
- Key files: `hub.go`

**`frontend/src/views/`:**
- Purpose: Page-level Vue components
- Contains: One `.vue` file per route
- Key files:
  - `DashboardView.vue` - Stats overview
  - `LibraryView.vue` - Shows/movies browser
  - `LinkWizardView.vue` - Main linking UI
  - `DownloadsView.vue` - qBittorrent torrents
  - `RSSView.vue` - Auto-download rules
  - `HistoryView.vue` - Past link operations
  - `SettingsView.vue` - Configuration
  - `LoginView.vue` - Authentication

**`frontend/src/stores/`:**
- Purpose: Pinia state management
- Contains: Global state for auth and library data
- Key files: `auth.ts`, `library.ts`

**`frontend/src/composables/`:**
- Purpose: Reusable Vue composition functions
- Contains: API client, WebSocket hook
- Key files: `useApi.ts`, `useWebSocket.ts`

**`frontend/src/components/ui/`:**
- Purpose: UI component library (shadcn/ui + reka-ui)
- Contains: Barrel-exported component sets
- Key files: `button/`, `card/`, `dialog/`, `table/`, etc.

**`frontend/src/lib/`:**
- Purpose: Utilities and type definitions
- Contains: TypeScript types mirroring Go models, helper functions
- Key files: `types.ts`, `utils.ts`

## Key File Locations

**Entry Points:**
- `cmd/server/main.go`: Go server bootstrap
- `frontend/src/main.ts`: Vue app bootstrap
- `frontend/src/App.vue`: Root Vue component

**Configuration:**
- `internal/config/config.go`: Go config struct and loading
- `frontend/vite.config.ts`: Vite build configuration
- `compose.yaml`: Docker Compose services
- `.env.example`: Environment variable template

**Core Logic:**
- `internal/linker/linker.go`: Hardlink creation/undo
- `internal/parser/parser.go`: Release name parsing
- `internal/scanner/scanner.go`: Filesystem scanning
- `internal/rss/poller.go`: RSS rule processing

**API:**
- `internal/api/router.go`: Route definitions
- `internal/api/link_handler.go`: Link operations
- `frontend/src/composables/useApi.ts`: Frontend API client

**Data:**
- `internal/models/models.go`: All Go structs
- `frontend/src/lib/types.ts`: TypeScript interfaces
- `internal/database/database.go`: Schema migrations

## Naming Conventions

**Files:**
- Go: `snake_case.go` (e.g., `link_handler.go`, `parser_test.go`)
- Vue: `PascalCase.vue` (e.g., `DashboardView.vue`, `AppSidebar.vue`)
- TypeScript: `camelCase.ts` (e.g., `useApi.ts`, `types.ts`)

**Directories:**
- Go packages: `lowercase` single word (e.g., `linker`, `qbit`, `ws`)
- Vue: `kebab-case` for UI components (e.g., `alert-dialog/`)
- Views: `PascalCaseView.vue` suffix pattern

**Functions/Methods:**
- Go: `PascalCase` for exported, `camelCase` for unexported
- TypeScript: `camelCase` for all functions

**Types:**
- Go: `PascalCase` structs (e.g., `LinkRequest`, `TorrentStatus`)
- TypeScript: `PascalCase` interfaces matching Go (e.g., `LinkRequest`)

## Where to Add New Code

**New API Endpoint:**
1. Handler: `internal/api/{domain}_handler.go` (new or existing)
2. Route: Add to `internal/api/router.go` under appropriate group
3. Model: Add structs to `internal/models/models.go`
4. Frontend type: Add to `frontend/src/lib/types.ts`
5. API method: Add to `frontend/src/composables/useApi.ts`

**New Page/View:**
1. View component: `frontend/src/views/{Name}View.vue`
2. Route: Add to `frontend/src/router/index.ts`
3. Navigation: Update `frontend/src/components/layout/AppSidebar.vue`

**New Domain Service:**
1. Package: `internal/{domain}/` with main file
2. Wire up in: `cmd/server/main.go`
3. Expose via: New handler in `internal/api/`

**New Integration Client:**
1. Client: `internal/{service}/client.go`
2. Models: Add types to `internal/models/models.go`
3. Server field: Add to `api.Server` struct
4. Init: Add to `main.go` initialization

**New UI Component:**
1. Component: `frontend/src/components/ui/{name}/` directory
2. Export: Add barrel `index.ts` in component directory

**New Composable:**
1. File: `frontend/src/composables/use{Name}.ts`
2. Export function following `use*` naming pattern

## Special Directories

**`cmd/server/dist/` (at build time):**
- Purpose: Embedded frontend build output
- Generated: Yes (by `vite build`)
- Committed: No (populated during Docker build)

**`.planning/codebase/`:**
- Purpose: Architecture and convention documentation
- Generated: Yes (by GSD mapping)
- Committed: Yes

**`frontend/node_modules/`:**
- Purpose: npm dependencies
- Generated: Yes
- Committed: No

**`shaders/`:**
- Purpose: GLSL shaders for video upscaling (future feature)
- Generated: No
- Committed: Yes

---

*Structure analysis: 2026-02-22*
