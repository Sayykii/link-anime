# Season File Visibility — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show filenames per season in the Link Wizard via clickable season pills that expand a file list panel.

**Architecture:** Backend adds a `Files []string` field to the `Season` model and a `collectVideos()` scanner helper to populate it. Frontend adds the `files` field to the TypeScript type and makes season pills clickable with an expandable file list panel in the existing shows list.

**Tech Stack:** Go (backend models + scanner), Vue 3 / TypeScript / Tailwind CSS (frontend)

**Spec:** `docs/superpowers/specs/2026-03-26-season-files-design.md`

---

## File Map

- **Modify:** `internal/models/models.go:14-18` — add `Files` field to `Season` struct
- **Modify:** `internal/scanner/scanner.go:70-81` — add `collectVideos()`, use in `ScanLibrary`
- **Modify:** `frontend/src/lib/types.ts:10-14` — add `files` field to `Season` interface
- **Modify:** `frontend/src/views/LinkWizardView.vue` — expandable file list panel

No new files. No new API endpoints.

---

### Task 1: Add `Files` field to Go `Season` model

**Files:**
- Modify: `internal/models/models.go:14-18`

- [ ] **Step 1: Add `Files` field to `Season` struct**

In `internal/models/models.go`, the current `Season` struct (lines 14-18):
```go
type Season struct {
	Number   int    `json:"number"`
	Path     string `json:"path"`
	Episodes int    `json:"episodes"`
}
```

Change to:
```go
type Season struct {
	Number   int      `json:"number"`
	Path     string   `json:"path"`
	Episodes int      `json:"episodes"`
	Files    []string `json:"files"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/desktop/CodingProjects/link-anime && go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```
feat: add Files field to Season model
```

---

### Task 2: Add `collectVideos()` scanner helper and use it in `ScanLibrary`

**Files:**
- Modify: `internal/scanner/scanner.go:70-81` (ScanLibrary season loop)
- Modify: `internal/scanner/scanner.go` (add helper after existing helpers)

- [ ] **Step 1: Add `collectVideos()` helper function**

Add this function after the existing `countVideosFlat` function (after line 296 in `internal/scanner/scanner.go`):

```go
// collectVideos walks a directory and returns sorted video filenames (names only, not paths).
func collectVideos(dir string) []string {
	var files []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && IsVideo(info.Name()) {
			files = append(files, info.Name())
		}
		return nil
	})
	sort.Strings(files)
	return files
}
```

- [ ] **Step 2: Update `ScanLibrary` to use `collectVideos`**

In `internal/scanner/scanner.go`, replace the season scanning block (lines 74-81):
```go
			seasonPath := filepath.Join(showPath, se.Name())
			epCount := countVideos(seasonPath)
			show.Seasons = append(show.Seasons, models.Season{
				Number:   seasonNum,
				Path:     seasonPath,
				Episodes: epCount,
			})
			show.Episodes += epCount
```

With:
```go
			seasonPath := filepath.Join(showPath, se.Name())
			files := collectVideos(seasonPath)
			show.Seasons = append(show.Seasons, models.Season{
				Number:   seasonNum,
				Path:     seasonPath,
				Episodes: len(files),
				Files:    files,
			})
			show.Episodes += len(files)
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /home/desktop/CodingProjects/link-anime && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```
feat: collect video filenames per season in scanner
```

---

### Task 3: Add `files` field to TypeScript `Season` type

**Files:**
- Modify: `frontend/src/lib/types.ts:10-14`

- [ ] **Step 1: Add `files` field to `Season` interface**

In `frontend/src/lib/types.ts`, the current `Season` interface (lines 10-14):
```ts
export interface Season {
  number: number
  path: string
  episodes: number
}
```

Change to:
```ts
export interface Season {
  number: number
  path: string
  episodes: number
  files: string[]
}
```

- [ ] **Step 2: Verify types compile**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npx vue-tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```
feat: add files field to Season TypeScript type
```

---

### Task 4: Add expandable file list panel to Link Wizard

**Files:**
- Modify: `frontend/src/views/LinkWizardView.vue:76` (state)
- Modify: `frontend/src/views/LinkWizardView.vue:229-238` (selectExistingShow + reset/nextInQueue)
- Modify: `frontend/src/views/LinkWizardView.vue:558-598` (template — season pills + file panel)

- [ ] **Step 1: Add `expandedSeason` ref**

In `frontend/src/views/LinkWizardView.vue`, after `selectedShow` (line 76), add:
```ts
const expandedSeason = ref<{ showName: string; seasonNumber: number } | null>(null)
```

- [ ] **Step 2: Add `toggleSeasonFiles` function**

After the `watch(showName, ...)` block (around line 243), add:
```ts
function toggleSeasonFiles(showName: string, seasonNumber: number, event: Event) {
  event.stopPropagation() // Don't trigger the parent show button click
  if (expandedSeason.value?.showName === showName && expandedSeason.value?.seasonNumber === seasonNumber) {
    expandedSeason.value = null
  } else {
    expandedSeason.value = { showName, seasonNumber }
  }
}
```

- [ ] **Step 3: Clear `expandedSeason` in `selectExistingShow`**

In the `selectExistingShow` function (line 229), add `expandedSeason.value = null` after `selectedShow.value = show`:
```ts
function selectExistingShow(show: Show) {
  showName.value = show.name
  selectedShow.value = show
  expandedSeason.value = null
  showSearch.value = ''
  // Auto-fill season to next available
  if (show.seasons.length > 0) {
    const maxSeason = Math.max(...show.seasons.map(s => s.number))
    seasonNumber.value = maxSeason + 1
  }
}
```

- [ ] **Step 4: Clear `expandedSeason` in `reset()` and `nextInQueue()`**

In `reset()` (around line 320), add after `selectedShow.value = null`:
```ts
selectedShow.value = null
expandedSeason.value = null
```

In `nextInQueue()` (around line 302), add after `selectedShow.value = null`:
```ts
selectedShow.value = null
expandedSeason.value = null
```

- [ ] **Step 5: Replace the season pills template to make them clickable**

In the template, find the show list inside the series block (around lines 558-598). Replace the entire `<div class="max-h-64 overflow-y-auto rounded-md border">` block with this version that wraps each show in a `<div>` (instead of a `<button>`) so the file panel can render between show rows, and makes pills clickable:

Replace (lines 559-598):
```html
          <div class="max-h-64 overflow-y-auto rounded-md border">
            <button
              v-for="show in filteredExistingShows"
              :key="show.name"
              class="flex items-center gap-3 w-full p-2 text-left hover:bg-accent transition-colors border-b last:border-b-0"
              :class="{ 'bg-accent': showName === show.name }"
              @click="selectExistingShow(show)"
            >
              <!-- Mini poster -->
              <div class="flex-shrink-0 w-8 h-12 rounded overflow-hidden bg-muted">
                <img
                  v-if="findShokoPoster(show.name)"
                  :src="findShokoPoster(show.name)!"
                  :alt="show.name"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
                <div v-else class="w-full h-full flex items-center justify-center">
                  <Tv class="h-3 w-3 text-muted-foreground/40" />
                </div>
              </div>
              <div class="min-w-0 flex-1">
                <span class="text-sm font-medium truncate block">{{ show.name }}</span>
                <div v-if="show.seasons.length" class="flex gap-1 flex-wrap mt-1">
                  <Badge
                    v-for="season in show.seasons"
                    :key="season.number"
                    variant="outline"
                    class="text-[10px] px-1.5 py-0"
                  >
                    S{{ season.number }} · {{ season.episodes }}ep
                  </Badge>
                </div>
              </div>
              <Check v-if="showName === show.name" class="h-4 w-4 text-primary ml-auto flex-shrink-0" />
            </button>
            <div v-if="filteredExistingShows.length === 0" class="p-3 text-center text-sm text-muted-foreground">
              No matches
            </div>
          </div>
```

With:
```html
          <div class="max-h-80 overflow-y-auto rounded-md border">
            <div
              v-for="show in filteredExistingShows"
              :key="show.name"
              class="border-b last:border-b-0"
            >
              <button
                class="flex items-center gap-3 w-full p-2 text-left hover:bg-accent transition-colors"
                :class="{ 'bg-accent': showName === show.name }"
                @click="selectExistingShow(show)"
              >
                <!-- Mini poster -->
                <div class="flex-shrink-0 w-8 h-12 rounded overflow-hidden bg-muted">
                  <img
                    v-if="findShokoPoster(show.name)"
                    :src="findShokoPoster(show.name)!"
                    :alt="show.name"
                    class="w-full h-full object-cover"
                    loading="lazy"
                  />
                  <div v-else class="w-full h-full flex items-center justify-center">
                    <Tv class="h-3 w-3 text-muted-foreground/40" />
                  </div>
                </div>
                <div class="min-w-0 flex-1">
                  <span class="text-sm font-medium truncate block">{{ show.name }}</span>
                  <div v-if="show.seasons.length" class="flex gap-1 flex-wrap mt-1">
                    <Badge
                      v-for="season in show.seasons"
                      :key="season.number"
                      variant="outline"
                      class="text-[10px] px-1.5 py-0 cursor-pointer hover:bg-accent"
                      :class="{ 'bg-primary/10 border-primary/40': expandedSeason?.showName === show.name && expandedSeason?.seasonNumber === season.number }"
                      @click="toggleSeasonFiles(show.name, season.number, $event)"
                    >
                      S{{ season.number }} · {{ season.episodes }}ep
                    </Badge>
                  </div>
                </div>
                <Check v-if="showName === show.name" class="h-4 w-4 text-primary ml-auto flex-shrink-0" />
              </button>
              <!-- Expanded file list panel -->
              <div
                v-if="show.seasons.some(s => expandedSeason?.showName === show.name && expandedSeason?.seasonNumber === s.number)"
                class="px-3 pb-3 pt-1 bg-muted/30 border-t border-border/50"
              >
                <template v-for="season in show.seasons" :key="season.number">
                  <div v-if="expandedSeason?.showName === show.name && expandedSeason?.seasonNumber === season.number">
                    <div class="flex items-center justify-between mb-1.5">
                      <span class="text-xs font-medium text-muted-foreground">Season {{ season.number }} — {{ season.files.length }} file{{ season.files.length !== 1 ? 's' : '' }}</span>
                    </div>
                    <div v-if="season.files.length" class="max-h-40 overflow-y-auto rounded border bg-background/50 p-2">
                      <div
                        v-for="file in season.files"
                        :key="file"
                        class="text-[11px] font-mono text-muted-foreground py-0.5 truncate"
                        :title="file"
                      >
                        {{ file }}
                      </div>
                    </div>
                    <div v-else class="text-xs text-muted-foreground italic">No files</div>
                  </div>
                </template>
              </div>
            </div>
            <div v-if="filteredExistingShows.length === 0" class="p-3 text-center text-sm text-muted-foreground">
              No matches
            </div>
          </div>
```

- [ ] **Step 6: Verify types compile**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npx vue-tsc --noEmit`
Expected: No errors

- [ ] **Step 7: Manual smoke test**

Run both backend and frontend:
- `cd /home/desktop/CodingProjects/link-anime && go run ./cmd/server`
- `cd /home/desktop/CodingProjects/link-anime/frontend && npm run dev`

Verify in browser:
1. Go to Link Wizard, select a source, pick "Series"
2. Existing shows list still shows season pills with episode counts
3. Clicking a season pill expands a file list below that show row
4. File list shows sorted filenames in monospace font
5. Clicking the same pill again collapses it
6. Clicking a different pill collapses the previous and expands the new one
7. Selecting a show (clicking the row, not a pill) clears any expanded panel
8. Movies path still works normally

- [ ] **Step 8: Commit**

```
feat: add expandable file list panel to season pills in link wizard
```
