# Phase 1: Foundation - Research

**Researched:** 2026-02-22
**Domain:** SQLite database schema, Go models, TypeScript interfaces
**Confidence:** HIGH

## Summary

Phase 1 establishes the data foundation for upscale job persistence. The existing codebase provides clear patterns to follow: raw `database/sql` with `modernc.org/sqlite` for database access, simple struct-based models with JSON tags in `internal/models/models.go`, and mirrored TypeScript interfaces in `frontend/src/lib/types.ts`.

The project uses a migration-based schema approach with inline SQL strings executed sequentially in `database.go`. No ORM is used - all queries are raw SQL with the `?` placeholder syntax. The WebSocket message pattern (`WSMessage` with `type` and `data` fields) is already established and will inform progress payload design.

**Primary recommendation:** Follow existing patterns exactly - add migration to `migrate()` slice, add models to `models.go`, add interfaces to `types.ts`. No new dependencies required.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| database/sql | stdlib | Database interface | Go standard, already in use |
| modernc.org/sqlite | (indirect) | Pure-Go SQLite driver | Already in go.mod, CGO-free |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | JSON serialization | Go struct ↔ JSON |
| time | stdlib | Timestamps | Go time.Time for dates |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Raw SQL | sqlc, sqlx, GORM | Project already uses raw SQL; consistency > convenience |
| Inline migrations | golang-migrate | Overkill for small schema; project uses inline approach |

**Installation:**
```bash
# No new dependencies needed - all already in go.mod
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── database/
│   └── database.go      # Add migration to migrate() slice
├── models/
│   └── models.go        # Add UpscaleJob, UpscaleProgress structs
frontend/src/lib/
└── types.ts             # Add TypeScript interfaces
```

### Pattern 1: Migration as SQL String
**What:** Migrations are SQL strings in a slice, executed with `CREATE TABLE IF NOT EXISTS`
**When to use:** All schema changes
**Example:**
```go
// Source: internal/database/database.go:44-86
migrations := []string{
    `CREATE TABLE IF NOT EXISTS upscale_jobs (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        input_path TEXT NOT NULL,
        output_path TEXT NOT NULL,
        preset     TEXT NOT NULL DEFAULT 'balanced',
        status     TEXT NOT NULL DEFAULT 'pending',
        error      TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        started_at DATETIME,
        completed_at DATETIME
    )`,
}
```

### Pattern 2: Model with JSON Tags
**What:** Go struct with `json:"camelCase"` tags matching API conventions
**When to use:** All models returned via API
**Example:**
```go
// Source: internal/models/models.go:55-66
type HistoryEntry struct {
    ID        int64     `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    MediaType string    `json:"mediaType"`
    ShowName  string    `json:"showName"`
    Season    *int      `json:"season,omitempty"`
    FileCount int       `json:"fileCount"`
    TotalSize int64     `json:"totalSize"`
    DestPath  string    `json:"destPath"`
    Source    string    `json:"source"`
}
```

### Pattern 3: TypeScript Interface Mirroring
**What:** TypeScript interface with identical field names (camelCase)
**When to use:** All types used by frontend
**Example:**
```typescript
// Source: frontend/src/lib/types.ts:47-57
export interface HistoryEntry {
  id: number
  timestamp: string
  mediaType: string
  showName: string
  season?: number
  fileCount: number
  totalSize: number
  destPath: string
  source: string
}
```

### Pattern 4: WebSocket Message Envelope
**What:** All WS messages use `WSMessage` wrapper with `type` string and `data` payload
**When to use:** All real-time broadcasts
**Example:**
```go
// Source: internal/models/models.go:105-109
type WSMessage struct {
    Type string      `json:"type"`
    Data interface{} `json:"data,omitempty"`
}
```

### Anti-Patterns to Avoid
- **ORM abstractions:** Project uses raw SQL - don't introduce GORM/sqlx patterns
- **Separate migration files:** Inline migrations in migrate() are the convention
- **PascalCase JSON:** All JSON uses camelCase, never PascalCase
- **Optional fields without omitempty:** Use `*Type` with `omitempty` for nullable fields

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Status enum validation | Custom validation | TEXT column + app logic | SQLite lacks enums; check in Go |
| UUID generation | uuid package | INTEGER AUTOINCREMENT | Existing pattern, simpler |
| Time formatting | Custom format | time.Time → automatic RFC3339 | encoding/json handles it |
| Nullable fields | sql.NullString | *string with json omitempty | Cleaner API response |

**Key insight:** The project values simplicity - use SQLite defaults and Go stdlib rather than adding dependencies.

## Common Pitfalls

### Pitfall 1: Status as Go const vs DB string
**What goes wrong:** Defining Go const for statuses that don't match DB string values
**Why it happens:** Tempting to use iota or typed constants
**How to avoid:** Use string constants that match exactly what's stored in DB
**Warning signs:** Type conversion needed when reading from DB
```go
// GOOD
const (
    StatusPending   = "pending"
    StatusRunning   = "running"
    StatusCompleted = "completed"
    StatusFailed    = "failed"
    StatusCancelled = "cancelled"
)

// BAD - don't use iota
type Status int
const (
    StatusPending Status = iota // This requires conversion
)
```

### Pitfall 2: Missing omitempty on optional fields
**What goes wrong:** Null fields serialize as `null` instead of being omitted
**Why it happens:** Forgetting that Go zero values serialize to JSON
**How to avoid:** Use pointer types with `omitempty` for optional fields
**Warning signs:** API responses contain `"field": null` for absent data

### Pitfall 3: TypeScript/Go type mismatch
**What goes wrong:** Go time.Time vs TypeScript string; Go int64 vs TypeScript number
**Why it happens:** Different type systems, manual sync
**How to avoid:** Document Go→TS mapping; time.Time → string, int64 → number (safe up to 2^53)
**Warning signs:** Frontend runtime errors on date parsing

### Pitfall 4: Forgetting IF NOT EXISTS
**What goes wrong:** Migration fails on re-run when table exists
**Why it happens:** Writing CREATE TABLE without IF NOT EXISTS
**How to avoid:** Always use IF NOT EXISTS - migrations run on every startup
**Warning signs:** App crashes on restart with "table already exists"

## Code Examples

Verified patterns from existing codebase:

### Create Table Migration
```go
// Source: internal/database/database.go:44-58
`CREATE TABLE IF NOT EXISTS upscale_jobs (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    input_path   TEXT NOT NULL,
    output_path  TEXT NOT NULL,
    preset       TEXT NOT NULL DEFAULT 'balanced',
    status       TEXT NOT NULL DEFAULT 'pending',
    error        TEXT,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at   DATETIME,
    completed_at DATETIME
)`
```

### Go Model
```go
// Pattern from: internal/models/models.go
type UpscaleJob struct {
    ID          int64      `json:"id"`
    InputPath   string     `json:"inputPath"`
    OutputPath  string     `json:"outputPath"`
    Preset      string     `json:"preset"`
    Status      string     `json:"status"`
    Error       *string    `json:"error,omitempty"`
    CreatedAt   time.Time  `json:"createdAt"`
    StartedAt   *time.Time `json:"startedAt,omitempty"`
    CompletedAt *time.Time `json:"completedAt,omitempty"`
}
```

### WebSocket Progress Model
```go
// Pattern from: internal/models/models.go:111-117
type UpscaleProgress struct {
    JobID   int64   `json:"jobId"`
    Frame   int     `json:"frame"`
    FPS     float64 `json:"fps"`
    Time    string  `json:"time"`    // e.g., "00:05:23"
    Percent float64 `json:"percent"` // 0-100
}
```

### TypeScript Interface
```typescript
// Pattern from: frontend/src/lib/types.ts
export type UpscaleStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

export interface UpscaleJob {
  id: number
  inputPath: string
  outputPath: string
  preset: string
  status: UpscaleStatus
  error?: string
  createdAt: string
  startedAt?: string
  completedAt?: string
}

export interface UpscaleProgress {
  jobId: number
  frame: number
  fps: number
  time: string
  percent: number
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| CGO sqlite3 | modernc.org/sqlite | 2020+ | Pure Go, no CGO, simpler builds |
| Manual JSON tags | Standard convention | Always | camelCase in JSON is de facto |

**Deprecated/outdated:**
- sql.NullString: Use *string with omitempty instead (cleaner JSON)
- Custom time formats: Let encoding/json handle RFC3339

## Open Questions

1. **Progress throttling**
   - What we know: FFmpeg outputs progress rapidly (potentially 10+ times/sec)
   - What's unclear: Should throttling be in Go model or broadcasting layer?
   - Recommendation: This is Phase 5 concern (WebSocket), not Phase 1. Model just defines fields.

2. **Error detail level**
   - What we know: `error` field stores failure reason
   - What's unclear: Full FFmpeg stderr or summarized message?
   - Recommendation: Store summarized error; full stderr can go to logs

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DB-01 | Upscale jobs table with status tracking (pending/running/completed/failed/cancelled) | Migration pattern established; use TEXT for status with 5 allowed values |
| DB-02 | UpscaleJob Go model with JSON tags matching existing conventions | Follow models.go pattern with camelCase JSON tags |
| DB-03 | UpscaleProgress WebSocket payload model | Follow WSMessage/LinkProgress pattern for frame/fps/time/percent |
| DB-04 | TypeScript interfaces mirroring Go models | Follow types.ts pattern with union type for status enum |
</phase_requirements>

## Sources

### Primary (HIGH confidence)
- `internal/database/database.go` - Existing migration and query patterns
- `internal/models/models.go` - Existing model struct patterns
- `frontend/src/lib/types.ts` - Existing TypeScript interface patterns
- `internal/ws/hub.go` - WebSocket broadcast patterns

### Secondary (MEDIUM confidence)
- `go.mod` - Confirms modernc.org/sqlite driver usage
- `frontend/package.json` - Confirms Vue 3 + TypeScript stack

### Tertiary (LOW confidence)
- None - all findings based on codebase inspection

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - directly observed in codebase
- Architecture: HIGH - following exact existing patterns
- Pitfalls: HIGH - based on Go/SQLite common issues and codebase conventions

**Research date:** 2026-02-22
**Valid until:** 2026-04-22 (60 days - stable patterns, no external deps)
