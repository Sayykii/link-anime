# Coding Conventions

**Analysis Date:** 2026-02-22

## Language-Specific Conventions

### Go Backend

**Files:**
- Single-purpose files: `client.go`, `parser.go`, `database.go`
- Handler files suffixed with `_handler.go`: `auth_handler.go`, `library_handler.go`
- Test files follow Go convention: `*_test.go`

**Functions:**
- camelCase for unexported: `parseSeasonNum()`, `linkFile()`, `resolveSource()`
- PascalCase for exported: `ParseReleaseName()`, `Link()`, `GetHistory()`
- Handler methods on Server struct: `func (s *Server) handleLogin(...)`

**Variables:**
- Short names in local scope: `r`, `w`, `req`, `err`, `m`
- Descriptive names for broader scope: `sourcePath`, `destDir`, `sessionDuration`
- Package-level vars use camelCase: `sessions`, `mu`, `DB`

**Types:**
- PascalCase structs: `Server`, `Client`, `Result`, `LinkRequest`
- Lowercase for unexported: `session`, `loginRequest`

### TypeScript Frontend

**Files:**
- Components: PascalCase `.vue` files: `LibraryView.vue`, `Button.vue`
- Composables: camelCase with `use` prefix: `useApi.ts`, `useWebSocket.ts`
- Stores: camelCase: `library.ts`, `auth.ts`
- Types/utils: camelCase: `types.ts`, `utils.ts`

**Functions:**
- camelCase: `fetchShows()`, `handleLogin()`, `formatSize()`
- Vue composables: `useApi()`, `useWebSocket()`, `useLibraryStore()`

**Variables:**
- camelCase: `searchQuery`, `unlinkTarget`, `lastMessage`
- Refs: camelCase: `const connected = ref(false)`

**Types:**
- PascalCase interfaces: `Show`, `Movie`, `LinkRequest`, `Settings`
- Type parameters match backend JSON: `showName`, `mediaType` (not snake_case)

## Code Style

**Go Formatting:**
- Standard `gofmt` (no explicit config files)
- Tabs for indentation
- Import groups: stdlib, then external packages

**TypeScript/Vue Formatting:**
- No ESLint/Prettier config files present
- 2-space indentation observed in Vue files
- Single quotes for imports
- Semicolons used

**Vue Component Structure:**
```vue
<script setup lang="ts">
// Imports
// Composables/stores
// Refs/reactive state
// Computed properties
// Functions
// Lifecycle hooks
</script>

<template>
  <!-- Template -->
</template>
```

## Import Organization

### Go

**Order:**
1. Standard library
2. External packages (chi, bcrypt, sqlite)
3. Internal project packages (link-anime/internal/...)

**Example from `cmd/server/main.go`:**
```go
import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"link-anime/internal/api"
	"link-anime/internal/auth"
	"link-anime/internal/config"
)
```

### TypeScript

**Order:**
1. External packages (vue, vue-router, pinia)
2. Internal components (using `@/` alias)
3. Types (last, often inline with component imports)

**Path Aliases:**
- `@/*` maps to `./src/*` (configured in `tsconfig.json`)

**Example from `LibraryView.vue`:**
```typescript
import { onMounted, ref, computed } from 'vue'
import { useLibraryStore } from '@/stores/library'
import { useApi } from '@/composables/useApi'
import type { UnlinkPreview } from '@/lib/types'
import { Card, CardContent } from '@/components/ui/card'
```

## Error Handling

### Go Backend

**Pattern:** Return errors up the call stack, handle at API boundary
```go
// Internal functions return errors
func (c *Client) Login() error {
    resp, err := c.client.PostForm(...)
    if err != nil {
        return fmt.Errorf("qbit login request: %w", err)
    }
    if resp.StatusCode != 200 {
        return fmt.Errorf("qbit login failed: %s (status %d)", string(body), resp.StatusCode)
    }
    return nil
}

// Handlers convert errors to JSON responses
func (s *Server) handleLink(w http.ResponseWriter, r *http.Request) {
    result, err := linker.Link(req, ...)
    if err != nil {
        jsonError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    jsonOK(w, result)
}
```

**Error wrapping:** Use `fmt.Errorf("context: %w", err)` for wrapped errors

**JSON error responses:** Helper function `jsonError(w, msg, code)` writes:
```json
{"error": "message here"}
```

### TypeScript Frontend

**Pattern:** Custom `ApiError` class with status codes
```typescript
class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}
```

**API error handling:**
```typescript
if (!resp.ok) {
  throw new ApiError(data.error || 'Request failed', resp.status)
}
```

**Component error handling:** Try/catch with toast notifications
```typescript
try {
  await api.unlinkPreview(path)
} catch (err: any) {
  toast.error('Failed to check files', { description: err.message })
}
```

**Auth redirects:** 401 responses redirect to `/login` automatically

## Logging

**Go Framework:** Standard `log` package
```go
log.SetFlags(log.LstdFlags | log.Lshortfile)  // main.go
log.Printf("[link] result: linked=%d ...", result.Linked)
log.Printf("[shoko] Triggering scan for: %s", req.Name)
```

**Patterns:**
- Prefixes for subsystems: `[link]`, `[shoko]`, `[qbit]`
- Printf style with format specifiers
- Non-fatal warnings written to stderr: `fmt.Fprintf(os.Stderr, "warning: ...")`

**Frontend:** No explicit logging framework, uses toast notifications for user feedback

## Comments

**When to Comment:**
- Exported functions and types (Go doc comments)
- Complex regex patterns
- Non-obvious logic

**Go Doc Comments:**
```go
// Result holds the parsed release name and optional season.
type Result struct { ... }

// ParseReleaseName extracts a clean show/movie name and optional season
// from a typical anime release folder name.
func ParseReleaseName(input string) Result { ... }

// Link creates hardlinks from source to destination.
func Link(req models.LinkRequest, ...) (*models.LinkResult, error) { ... }
```

**Inline Comments:**
```go
// Strip video file extension FIRST (before dot conversion)
name = reVideoExt.ReplaceAllString(name, "")

// Dest exists - check if same inode
if _, err := os.Stat(dest); err == nil { ... }
```

**TypeScript:** Minimal comments, types are self-documenting
```typescript
// API types matching Go backend models
export interface Show { ... }
```

## Function Design

**Size:** Functions are compact, typically 20-60 lines

**Parameters:**
- Go: Explicit parameters, no option structs (exception: request structs from JSON)
- TypeScript: Optional parameters have defaults: `getHistory(limit = 50)`

**Return Values:**
- Go: `(result, error)` pattern for fallible operations
- TypeScript: Promises for async operations

## Module Design

**Go Exports:**
- Public functions/types are PascalCase
- One package per directory under `internal/`
- Packages are single-responsibility: `parser`, `linker`, `scanner`, `auth`

**TypeScript Exports:**
- Barrel files (`index.ts`) for UI components
- Named exports for composables/stores

**Example barrel file (`components/ui/button/index.ts`):**
```typescript
export { default as Button } from "./Button.vue"
export const buttonVariants = cva(...)
export type ButtonVariants = VariantProps<typeof buttonVariants>
```

## Vue-Specific Conventions

**Composition API:** All components use `<script setup lang="ts">`

**State Management:**
- Pinia stores for shared state
- Local `ref()` for component-scoped state
- `computed()` for derived values

**Props Pattern:** Using `withDefaults(defineProps<Props>(), {...})`

**Event Handlers:** `@click="methodName"` or inline `@click="() => {}"`

**UI Components:** shadcn/ui pattern with Reka UI primitives
- Variants via `class-variance-authority`
- Styling via Tailwind CSS classes
- Component composition pattern

---

*Convention analysis: 2026-02-22*
