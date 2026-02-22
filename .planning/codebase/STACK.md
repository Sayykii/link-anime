# Technology Stack

**Analysis Date:** 2026-02-22

## Languages

**Primary:**
- Go 1.24 - Backend API server (`cmd/server/`, `internal/`)
- TypeScript 5.9.3 - Frontend application (`frontend/src/`)

**Secondary:**
- Vue 3 Single File Components (.vue) - UI components (`frontend/src/components/`, `frontend/src/views/`)
- GLSL - Anime4K upscale shaders (`shaders/`)

## Runtime

**Environment:**
- Go 1.24 (backend, compiled to static binary)
- Node.js ^20.19.0 || >=22.12.0 (frontend dev/build)
- Bun 1.x (Docker build environment for frontend)
- Alpine Linux 3.21 (container runtime)

**Package Manager:**
- Go modules (`go.mod`) - no go.sum present, dependencies downloaded at build
- Bun (lockfile: `bun.lock*` optional)
- npm-run-all2 for script orchestration

## Frameworks

**Core:**
- Vue 3.5.28 - Frontend SPA framework
- chi v5 - Go HTTP router (`github.com/go-chi/chi/v5`)
- Pinia 3.0.4 - Vue state management

**Testing:**
- Not detected - no test framework configured for frontend
- Go standard testing (`internal/parser/parser_test.go`)

**Build/Dev:**
- Vite 7.3.1 - Frontend build tool and dev server
- vue-tsc 3.2.4 - TypeScript type checking
- Docker multi-stage build (Bun -> Go -> Alpine)

## Key Dependencies

**Backend (Go):**
- `github.com/go-chi/chi/v5` - HTTP routing with middleware
- `github.com/gorilla/websocket` - WebSocket real-time communication
- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO required)
- `golang.org/x/crypto/bcrypt` - Password hashing

**Frontend (npm):**
- `vue-router` 5.0.2 - Client-side routing
- `@vueuse/core` 14.2.1 - Vue composition utilities
- `@tanstack/vue-table` 8.21.3 - Data table components
- `reka-ui` 2.8.0 - Headless UI components
- `lucide-vue-next` 0.563.0 - Icon library
- `vue-sonner` 2.0.9 - Toast notifications
- `tailwindcss` 4.1.18 - Utility-first CSS
- `class-variance-authority` 0.7.1 - Component variants
- `clsx` 2.1.1 + `tailwind-merge` 3.4.0 - Class name utilities

**Infrastructure:**
- `ffmpeg` with `libplacebo` - Video upscaling pipeline
- Vulkan GPU access - Hardware-accelerated upscaling

## Configuration

**Environment:**
- Configuration via environment variables (prefix: `LA_`)
- Settings persist to SQLite database (override env vars)
- `.env` file support via Docker Compose

**Key Environment Variables:**
- `LA_PORT` - HTTP server port (default: 8787)
- `LA_PASSWORD` - Initial admin password
- `LA_DOWNLOAD_DIR` - Source directory for downloads
- `LA_MEDIA_DIR` - Destination for anime series
- `LA_MOVIES_DIR` - Destination for anime movies
- `LA_QBIT_*` - qBittorrent connection settings
- `LA_SHOKO_*` - Shoko Server connection settings
- `LA_NOTIFY_URL` - Notification webhook URL

**Build:**
- `frontend/vite.config.ts` - Vite configuration with Vue and Tailwind plugins
- `frontend/tsconfig.json` - TypeScript project references setup
- `Dockerfile` - Multi-stage build (frontend -> backend -> runtime)
- Path alias: `@/*` maps to `frontend/src/*`

## Database

**Type:** SQLite 3 (via modernc.org/sqlite)

**Location:** `$LA_DATA_DIR/link-anime.db` (default: `/app/data/link-anime.db`)

**Configuration:**
- WAL journal mode enabled
- 5000ms busy timeout
- Max 1 open connection (SQLite limitation)

**Tables:**
- `settings` - Key-value configuration storage
- `history` - Link operation history
- `linked_files` - Individual file tracking for undo
- `rss_rules` - Auto-download rule definitions
- `rss_matches` - Matched torrents from RSS polling

## Platform Requirements

**Development:**
- Go 1.24+ for backend development
- Node.js 20.19+ or 22.12+ for frontend development
- Bun (optional, used in Docker build)

**Production:**
- Docker with Compose support
- Volume mount for persistent data (`/app/data`)
- Volume mount for media storage (`/data`)
- GPU passthrough optional (`/dev/dri`) for Vulkan upscaling
- Network access to host services via `host.docker.internal`

**Container Runtime:**
- Alpine 3.21 base image
- User/group ID mapping (PUID/PGID) for NAS permissions
- Timezone configuration (TZ)

---

*Stack analysis: 2026-02-22*
