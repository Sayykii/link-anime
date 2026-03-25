# Season Visibility in Link Wizard — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show existing seasons inline when selecting a show in the Link Wizard, and auto-fill the season number to the next available.

**Architecture:** All changes are in `LinkWizardView.vue`. The series template gets its own computed (`filteredExistingShows`) returning full `Show` objects so season data is accessible. The movies template continues using the existing `filteredExistingItems` (strings). A `selectedShow` ref tracks which show is selected for the duplicate season warning.

**Tech Stack:** Vue 3 (Composition API), TypeScript, shadcn-vue Badge component, Tailwind CSS

**Spec:** `docs/superpowers/specs/2026-03-25-season-visibility-design.md`

---

## File Map

- **Modify:** `frontend/src/views/LinkWizardView.vue` — all changes live here
  - Script: new computed, updated function, new ref
  - Template: split existing shows list for series vs movies, add season pills, add hints/warnings

No new files. No backend changes. No other frontend files.

---

### Task 1: Add series-specific computed and update selectExistingShow

**Files:**
- Modify: `frontend/src/views/LinkWizardView.vue:1-8` (imports)
- Modify: `frontend/src/views/LinkWizardView.vue:66-68` (state)
- Modify: `frontend/src/views/LinkWizardView.vue:85-104` (computeds)
- Modify: `frontend/src/views/LinkWizardView.vue:208-211` (selectExistingShow)

- [ ] **Step 1: Add `Show` to the import from types**

At line 8, the existing import is:
```ts
import type { DownloadItem, LinkResult, LinkProgress } from '@/lib/types'
```
Change to:
```ts
import type { DownloadItem, LinkResult, LinkProgress, Show } from '@/lib/types'
```

- [ ] **Step 2: Add `selectedShow` ref**

After `showSearch` (line 75), add:
```ts
const selectedShow = ref<Show | null>(null)
```

- [ ] **Step 3: Add `filteredExistingShows` computed**

After `filteredExistingItems` (line 104), add:
```ts
// Series-specific: returns full Show objects so template can access seasons
const filteredExistingShows = computed(() => {
  const shows = library.shows
  if (!showSearch.value) return shows
  const q = showSearch.value.toLowerCase()
  return shows.filter(s => s.name.toLowerCase().includes(q))
})
```

- [ ] **Step 4: Add `duplicateSeasonWarning` computed**

After `filteredExistingShows`, add:
```ts
// Warning when user enters a season number that already exists
const duplicateSeasonWarning = computed(() => {
  if (!selectedShow.value) return null
  const existing = selectedShow.value.seasons.find(s => s.number === seasonNumber.value)
  if (!existing) return null
  return `Season ${existing.number} already exists (${existing.episodes} episodes)`
})
```

- [ ] **Step 5: Update `selectExistingShow` to accept a `Show` and auto-fill season**

Replace the existing function (lines 208-211):
```ts
function selectExistingShow(name: string) {
  showName.value = name
  showSearch.value = ''
}
```
With:
```ts
function selectExistingShow(show: Show) {
  showName.value = show.name
  selectedShow.value = show
  showSearch.value = ''
  // Auto-fill season to next available
  if (show.seasons.length > 0) {
    const maxSeason = Math.max(...show.seasons.map(s => s.number))
    seasonNumber.value = maxSeason + 1
  }
}
```

- [ ] **Step 6: Add watcher to clear `selectedShow` when name is manually edited**

After the `selectExistingShow` function, add:
```ts
// Clear selectedShow when user manually changes the name (so duplicate warning doesn't show stale data)
watch(showName, (name) => {
  if (selectedShow.value && name !== selectedShow.value.name) {
    selectedShow.value = null
  }
})
```

- [ ] **Step 7: Clear `selectedShow` in `reset()` and `nextInQueue()`**

In the `reset()` function (around line 280), add `selectedShow.value = null` after `showName.value = ''`:
```ts
showName.value = ''
selectedShow.value = null
```

In `nextInQueue()` (around line 267), add `selectedShow.value = null` after `showName.value = ''`:
```ts
showName.value = ''
selectedShow.value = null
```

- [ ] **Step 8: Commit (no type check yet — template still references old signature, fixed in Task 2)**

```
feat: add season-aware computed and selectExistingShow
```

---

### Task 2: Split template — series shows list with season pills

**Files:**
- Modify: `frontend/src/views/LinkWizardView.vue:505-551` (existing shows template)

- [ ] **Step 1: Replace the existing shows/movies template block**

Replace the entire block from line 505 (`<!-- Existing shows/movies with posters -->`) through line 551 (closing `</div>` of the show list section) with two conditional blocks — one for series, one for movies.

**Series block** (when `mediaType === 'series'` and `library.shows.length > 0`):

```html
        <!-- Existing shows with season pills (series only) -->
        <div v-if="mediaType === 'series' && library.shows.length" class="space-y-2">
          <Label class="text-xs text-muted-foreground">Or select existing show:</Label>

          <!-- Search filter -->
          <div class="relative">
            <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input v-model="showSearch" placeholder="Filter shows..." class="pl-9 h-9" />
            <button
              v-if="showSearch"
              class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
              @click="showSearch = ''"
            >
              <X class="h-4 w-4" />
            </button>
          </div>

          <!-- Show list with posters and season pills -->
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
        </div>

        <!-- Existing movies (movies only, unchanged behavior) -->
        <div v-if="mediaType === 'movie' && existingMovies.length" class="space-y-2">
          <Label class="text-xs text-muted-foreground">Or select existing movie:</Label>

          <div class="relative">
            <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input v-model="showSearch" placeholder="Filter movies..." class="pl-9 h-9" />
            <button
              v-if="showSearch"
              class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
              @click="showSearch = ''"
            >
              <X class="h-4 w-4" />
            </button>
          </div>

          <div class="max-h-64 overflow-y-auto rounded-md border">
            <button
              v-for="name in filteredExistingItems"
              :key="name"
              class="flex items-center gap-3 w-full p-2 text-left hover:bg-accent transition-colors border-b last:border-b-0"
              :class="{ 'bg-accent': showName === name }"
              @click="showName = name; showSearch = ''"
            >
              <div class="flex-shrink-0 w-8 h-12 rounded overflow-hidden bg-muted">
                <img
                  v-if="findShokoPoster(name)"
                  :src="findShokoPoster(name)!"
                  :alt="name"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
                <div v-else class="w-full h-full flex items-center justify-center">
                  <Tv class="h-3 w-3 text-muted-foreground/40" />
                </div>
              </div>
              <span class="text-sm font-medium truncate">{{ name }}</span>
              <Check v-if="showName === name" class="h-4 w-4 text-primary ml-auto flex-shrink-0" />
            </button>
            <div v-if="filteredExistingItems.length === 0" class="p-3 text-center text-sm text-muted-foreground">
              No matches
            </div>
          </div>
        </div>
```

- [ ] **Step 2: Verify the app compiles**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npx vue-tsc --noEmit`
Expected: No type errors

- [ ] **Step 3: Commit**

```
feat: split shows/movies template, add season pills to show list
```

---

### Task 3: Add season auto-fill hint and duplicate warning

**Files:**
- Modify: `frontend/src/views/LinkWizardView.vue:553-562` (season number input section)

- [ ] **Step 1: Update the season number input section**

Replace the existing season input block (the `<div v-if="mediaType === 'series'"` block around lines 553-562):

```html
        <div v-if="mediaType === 'series'" class="space-y-2">
          <Label>Season Number</Label>
          <div class="flex items-center gap-2">
            <Input v-model.number="seasonNumber" type="number" min="0" max="99" class="w-24" />
            <span v-if="selectedShow && selectedShow.seasons.length" class="text-xs text-muted-foreground">
              Next available
            </span>
          </div>
          <p v-if="duplicateSeasonWarning" class="text-sm text-amber-500">
            {{ duplicateSeasonWarning }}
          </p>
          <p v-else-if="suggestedSeason !== null && suggestedSeason !== seasonNumber" class="text-sm text-muted-foreground">
            Detected season: {{ suggestedSeason }}
            <Button variant="link" size="sm" class="h-auto p-0 ml-1" @click="seasonNumber = suggestedSeason!">
              Use this
            </Button>
          </p>
        </div>
```

- [ ] **Step 2: Verify the app compiles**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npx vue-tsc --noEmit`
Expected: No type errors

- [ ] **Step 3: Manual smoke test**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npm run dev`

Verify in browser:
1. Go to Link Wizard, select a source, pick "Series"
2. Existing shows list shows season pills under each show name
3. Clicking an existing show auto-fills the season number to next available
4. "Next available" hint appears beside the season input
5. Manually entering an existing season number shows the amber warning
6. Switching to "Movie" type still shows movies list without season pills
7. New shows (typed into the name field) still work normally with no warning

- [ ] **Step 4: Commit**

```
feat: add season auto-fill hint and duplicate season warning
```
