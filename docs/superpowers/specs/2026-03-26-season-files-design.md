# Season File Visibility in Link Wizard

## Problem

When linking a new season in the Link Wizard, season pills show episode counts but not which files are actually linked. Users can't verify whether specific episodes are present or identify gaps without manually browsing the filesystem.

## Solution

**Add filenames to season data + clickable season pills that expand a file list panel.**

### Backend changes

**Go `Season` struct** (`internal/models/models.go:14-18`) gets a new field:
```go
Files []string `json:"files"`
```

This holds just the filenames (not full paths) of video files in that season directory.

**Scanner** (`internal/scanner/scanner.go`): Replace `countVideos(seasonPath)` with a new `collectVideos(dir string) []string` function that walks the directory and returns sorted video filenames. The `Episodes` count is derived from `len(files)` instead of a separate count. The existing `countVideos` helper remains for use by other callers (movies, downloads).

No new API endpoints needed — the `/shows` endpoint already returns whatever the scanner produces.

### Frontend changes

**TypeScript `Season` type** (`frontend/src/lib/types.ts:10-14`) gets:
```ts
files: string[]
```

**Link Wizard template** (`frontend/src/views/LinkWizardView.vue`):

- Season pills become clickable. Clicking a pill toggles a file list panel below that show's row.
- New ref: `expandedSeason` of type `{ showName: string; seasonNumber: number } | null` — tracks which season's file list is currently visible.
- When a pill is clicked, if it matches `expandedSeason`, collapse (set to null). Otherwise, expand that season.
- The file list renders inside the show button's parent area (after the show row, before the next show), as a compact monospace list with a subtle background.
- `expandedSeason` is cleared when selecting a different show, navigating away, or resetting.

**File list panel contents:**
- Header: "Season N" with file count (e.g. "Season 1 — 24 files")
- Sorted list of filenames in monospace `text-xs` font
- Max height with overflow scroll for seasons with many files
- Subtle border and background to visually separate from the show list

### What doesn't change

- Season pills still show `S1 · 24ep` at a glance (the count is now derived from `len(files)`)
- Auto-fill and duplicate season warning behavior unchanged
- Movies path unchanged
- Library view, Downloads view unchanged
- No new API endpoints

## Files to modify

- `internal/models/models.go:14-18` — add `Files []string` to `Season` struct
- `internal/scanner/scanner.go:70-81` — add `collectVideos()` helper, use it in `ScanLibrary`
- `frontend/src/lib/types.ts:10-14` — add `files: string[]` to `Season` interface
- `frontend/src/views/LinkWizardView.vue` — add `expandedSeason` ref, make pills clickable, render file list panel
