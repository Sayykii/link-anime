# UI/UX Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 11 UI/UX issues across the site: consistency, mobile support, loading states, wizard UX, and general polish.

**Architecture:** All changes are frontend-only (Vue 3 + Tailwind + shadcn-vue). No backend or API changes needed. Each task is an independent UI fix that can be committed separately.

**Tech Stack:** Vue 3, TypeScript, Tailwind CSS v4, shadcn-vue (Reka UI), Lucide icons

---

### Task 1: Fix Downloads heading to use display font

**Files:**
- Modify: `frontend/src/views/DownloadsView.vue:365`

The Downloads page uses `text-3xl font-bold` while every other page uses `font-display text-3xl tracking-wider uppercase`. The CSS base layer already applies `font-display` and `uppercase` to all `h1` elements, but `font-bold` overrides the display font weight. The subtitle also lacks the `text-sm mt-1` styling used elsewhere.

- [ ] **Step 1: Fix heading classes**

Change line 365 from:
```html
<h1 class="text-3xl font-bold">Downloads</h1>
<p class="text-muted-foreground">Manage downloads, search Nyaa, and monitor torrents</p>
```
to:
```html
<h1 class="font-display text-3xl tracking-wider uppercase">Downloads</h1>
<p class="text-muted-foreground text-sm mt-1">Manage downloads, search Nyaa, and monitor torrents</p>
```

- [ ] **Step 2: Verify visually**

Run: `cd frontend && npm run dev`
Check Downloads page heading matches Dashboard/Library/Link Wizard style.

- [ ] **Step 3: Commit**

```
fix: use consistent display font heading on Downloads page
```

---

### Task 2: Add mobile sidebar with hamburger menu + sheet drawer

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/layout/AppSidebar.vue`

Currently the sidebar is always visible at `w-56`. On mobile, it eats ~30% of viewport. Solution: hide sidebar on `md:` breakpoint and show a hamburger button that opens a Sheet (slide-over drawer) containing the same nav.

- [ ] **Step 1: Add mobile hamburger and sheet to App.vue**

In `App.vue`, add imports for `Sheet`, `SheetContent`, `SheetTrigger` from `@/components/ui/sheet`, `Button` from `@/components/ui/button`, and `Menu` from `lucide-vue-next`. Add a `mobileOpen` ref.

Replace the authenticated layout (`<div v-else class="flex h-screen overflow-hidden">`) with:

```html
<div v-else class="flex h-screen overflow-hidden">
  <!-- Desktop sidebar: hidden on mobile -->
  <div class="hidden md:block">
    <AppSidebar @open-command="commandOpen = true" />
  </div>

  <main class="flex-1 overflow-auto">
    <!-- Mobile header bar -->
    <div class="sticky top-0 z-20 flex items-center gap-3 border-b bg-background/80 backdrop-blur-sm p-3 md:hidden">
      <Sheet v-model:open="mobileOpen">
        <SheetTrigger as-child>
          <Button variant="ghost" size="icon" class="h-9 w-9">
            <Menu class="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" class="w-56 p-0">
          <SheetTitle class="sr-only">Navigation</SheetTitle>
          <AppSidebar @open-command="commandOpen = true; mobileOpen = false" @navigate="mobileOpen = false" />
        </SheetContent>
      </Sheet>
      <span class="font-display text-lg tracking-wider uppercase">link-anime</span>
    </div>

    <div class="p-6">
      <router-view v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" :key="route.path" />
        </Transition>
      </router-view>
    </div>
  </main>
</div>
```

- [ ] **Step 2: Emit navigate event from AppSidebar**

In `AppSidebar.vue`, add `'navigate'` to the `defineEmits` array. On each nav button's `@click`, also emit `'navigate'`:

```ts
defineEmits<{
  'open-command': []
  'navigate': []
}>()
```

For each nav button, change `@click="router.push(item.path)"` to:
```html
@click="router.push(item.path); $emit('navigate')"
```

Do the same for the Settings button at the bottom.

- [ ] **Step 3: Close mobile drawer on route change**

In `App.vue`, add a watcher:
```ts
watch(() => route.path, () => { mobileOpen.value = false })
```

- [ ] **Step 4: Verify**

Test on a narrow viewport (< 768px): sidebar should be hidden, hamburger visible. Clicking hamburger opens sheet with full nav. Clicking a nav item closes the sheet and navigates.

On desktop (>= 768px): sidebar visible as before, no hamburger.

- [ ] **Step 5: Commit**

```
feat: add mobile sidebar drawer with hamburger menu
```

---

### Task 3: Make wizard step indicator responsive and clickable

**Files:**
- Modify: `frontend/src/views/LinkWizardView.vue:225-233`

The step indicator uses inline badges+arrows that wrap badly on mobile and aren't clickable.

- [ ] **Step 1: Replace step indicator with responsive clickable version**

Replace the step indicator block (lines 225-233) with:

```html
<!-- Step indicator -->
<div class="flex items-center gap-1 sm:gap-2 text-sm overflow-x-auto">
  <button
    v-for="(s, i) in [
      { num: 1, label: 'Source' },
      { num: 2, label: 'Type' },
      { num: 3, label: 'Details' },
      { num: 4, label: 'Confirm' },
    ]"
    :key="s.num"
    class="flex items-center gap-1 sm:gap-2 shrink-0"
    :disabled="s.num >= step || step >= 5"
    @click="s.num < step && step < 5 ? step = s.num : null"
  >
    <Badge
      :variant="step >= s.num ? 'default' : 'outline'"
      :class="[
        s.num < step && step < 5 ? 'cursor-pointer hover:bg-primary/80' : '',
        'transition-colors'
      ]"
    >
      <span class="hidden sm:inline">{{ s.num }}. {{ s.label }}</span>
      <span class="sm:hidden">{{ s.num }}</span>
    </Badge>
    <ArrowRight v-if="i < 3" class="h-3 w-3 text-muted-foreground shrink-0" />
  </button>
</div>
```

This makes badges:
- Clickable to jump back to completed steps (but not forward, and disabled during progress/done)
- Show only the step number on mobile, full label on sm+
- Horizontally scrollable if needed

- [ ] **Step 2: Verify**

Test wizard flow: complete steps 1-3, click step 1 badge — should jump back. Click step 4 badge while on step 3 — should not jump forward. On mobile, badges show numbers only.

- [ ] **Step 3: Commit**

```
feat: make wizard step indicator clickable and responsive
```

---

### Task 4: Add search/filter to wizard source selection (Step 1)

**Files:**
- Modify: `frontend/src/views/LinkWizardView.vue`

When there are many downloads, Step 1 has no way to filter them.

- [ ] **Step 1: Add source filter state**

Add a ref near the other wizard state (around line 30):
```ts
const sourceFilter = ref('')
```

Add a computed for filtered downloads (after the `downloads` ref):
```ts
const filteredSources = computed(() => {
  if (!sourceFilter.value) return downloads.value
  const q = sourceFilter.value.toLowerCase().replace(/[.\-_ ]+/g, ' ')
  return downloads.value.filter(d =>
    d.name.toLowerCase().replace(/[.\-_ ]+/g, ' ').includes(q)
  )
})
```

Import `Search` and `X` (they are already imported on line 17).

Also add `computed` to the import from vue (already imported on line 2).

- [ ] **Step 2: Add search input to Step 1 card**

In the Step 1 template, after the `<CardDescription>` and before the loading/empty/list content, add a search field inside `<CardContent>` at the top:

```html
<CardContent>
  <!-- Source filter -->
  <div v-if="downloads.length > 5" class="relative mb-3">
    <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
    <Input v-model="sourceFilter" placeholder="Filter downloads..." class="pl-9 h-9" />
    <button
      v-if="sourceFilter"
      class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
      @click="sourceFilter = ''"
    >
      <X class="h-4 w-4" />
    </button>
  </div>

  <!-- ...existing loading/empty/list content, but use filteredSources instead of downloads -->
```

Change the `v-for="item in downloads"` to `v-for="item in filteredSources"`.

Add a no-results empty state after the existing empty state for no downloads:
```html
<div v-else-if="!filteredSources.length" class="text-center text-muted-foreground py-8">
  No downloads matching "{{ sourceFilter }}"
  <div class="mt-2">
    <Button variant="outline" size="sm" @click="sourceFilter = ''">Clear filter</Button>
  </div>
</div>
```

- [ ] **Step 3: Reset filter on wizard reset**

In the `reset()` function, add:
```ts
sourceFilter.value = ''
```

- [ ] **Step 4: Commit**

```
feat: add search filter to link wizard source selection
```

---

### Task 5: Add skeleton loading to library poster grid

**Files:**
- Modify: `frontend/src/views/LibraryView.vue:179-182`

Replace the plain spinner with poster skeleton cards matching the grid layout.

- [ ] **Step 1: Import Skeleton**

Add `Skeleton` to the imports from `@/components/ui/skeleton`:
```ts
import { Skeleton } from '@/components/ui/skeleton'
```

- [ ] **Step 2: Replace spinner with skeleton grid**

Replace lines 179-182:
```html
<!-- Loading -->
<div v-if="loading" class="flex items-center justify-center py-12">
  <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
</div>
```

With:
```html
<!-- Loading skeleton grid -->
<div v-if="loading" class="space-y-4">
  <div class="flex gap-2">
    <Skeleton class="h-9 w-28 rounded-md" />
    <Skeleton class="h-9 w-28 rounded-md" />
  </div>
  <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
    <div v-for="i in 12" :key="i" class="space-y-2">
      <Skeleton class="aspect-[2/3] rounded-lg w-full" />
      <Skeleton class="h-4 w-3/4" />
      <Skeleton class="h-3 w-1/2" />
    </div>
  </div>
</div>
```

This shows fake tab triggers + a 12-card poster skeleton grid matching the real layout.

- [ ] **Step 3: Commit**

```
feat: add skeleton loading state for library poster grid
```

---

### Task 6: Add scroll-to-top on page navigation

**Files:**
- Modify: `frontend/src/router/index.ts`

Currently, navigating between pages preserves scroll position which can be disorienting.

- [ ] **Step 1: Add scrollBehavior to router**

The `scrollBehavior` option on `createRouter` doesn't work here because the scrollable element is `<main>`, not `window`. Instead, handle it in `App.vue`.

In `App.vue`, add a watcher that scrolls the main element to top on route change. Find the `<main>` element and scroll it:

```ts
watch(() => route.path, () => {
  // Scroll main content area to top on navigation
  const main = document.querySelector('main')
  if (main) main.scrollTo({ top: 0 })
})
```

If a `mobileOpen` watcher already exists from Task 2, combine them:
```ts
watch(() => route.path, () => {
  mobileOpen.value = false
  const main = document.querySelector('main')
  if (main) main.scrollTo({ top: 0 })
})
```

- [ ] **Step 2: Commit**

```
fix: scroll to top on page navigation
```

---

### Task 7: Add infinite scroll / "Load More" to library

**Files:**
- Modify: `frontend/src/views/LibraryView.vue`

The library fetches at most 100 series. Add a "Load More" button for larger libraries.

- [ ] **Step 1: Add pagination state**

Add refs for pagination (near the existing state):
```ts
const shokoPage = ref(1)
const shokoPageSize = 50
const hasMoreSeries = computed(() => shokoSeries.value.length < shokoTotal.value)
const loadingMore = ref(false)
```

- [ ] **Step 2: Update initial fetch to use pageSize**

In `onMounted`, change:
```ts
const data = await api.shokoSeries(1, 100)
```
to:
```ts
const data = await api.shokoSeries(1, shokoPageSize)
```

And in `refresh()`, change the same call and reset `shokoPage.value = 1`.

- [ ] **Step 3: Add loadMore function**

```ts
async function loadMore() {
  loadingMore.value = true
  try {
    shokoPage.value++
    const data = await api.shokoSeries(shokoPage.value, shokoPageSize)
    shokoSeries.value.push(...(data.series || []))
    shokoTotal.value = data.total
  } catch (err: any) {
    toast.error('Failed to load more', { description: err.message })
    shokoPage.value--
  } finally {
    loadingMore.value = false
  }
}
```

- [ ] **Step 4: Add Load More button after poster grid**

After the poster grid `</div>` (the `grid grid-cols-2...` div, around line 267), add:

```html
<div v-if="hasMoreSeries && !searchQuery" class="flex justify-center pt-4">
  <Button variant="outline" @click="loadMore" :disabled="loadingMore" class="gap-2">
    <Loader2 v-if="loadingMore" class="h-4 w-4 animate-spin" />
    Load More ({{ shokoSeries.length }}/{{ shokoTotal }})
  </Button>
</div>
```

- [ ] **Step 5: Commit**

```
feat: add Load More pagination to library poster grid
```

---

### Task 8: De-emphasize "Link" button for already-linked downloads

**Files:**
- Modify: `frontend/src/views/DownloadsView.vue:513-516`

Already-linked items show a green "Linked" badge but the action button is the same style.

- [ ] **Step 1: Change button variant for linked items**

Replace the Link button (around line 513):
```html
<Button size="sm" variant="outline" @click="goToLink(item.name)" class="gap-1 shrink-0">
  <Link class="h-3 w-3" />
  {{ getLinkedEntry(item.name) ? 'Re-link' : 'Link' }}
</Button>
```

With:
```html
<Button
  size="sm"
  :variant="getLinkedEntry(item.name) ? 'ghost' : 'outline'"
  @click="goToLink(item.name)"
  class="gap-1 shrink-0"
>
  <Link class="h-3 w-3" />
  {{ getLinkedEntry(item.name) ? 'Re-link' : 'Link' }}
</Button>
```

This uses `ghost` for already-linked (de-emphasized) and `outline` for unlinked (the primary action).

- [ ] **Step 2: Commit**

```
fix: de-emphasize re-link button for already-linked downloads
```

---

### Task 9: Fix wizard step indicator overflow on mobile

This is already handled in Task 3 (the responsive step indicator). No separate task needed — the `overflow-x-auto` and mobile-only number labels solve this.

---

### Task 10: Fix movies tab to use poster cards when Shoko is available

**Files:**
- Modify: `frontend/src/views/LibraryView.vue`
- Modify: `frontend/src/composables/useApi.ts` (only if a movie poster API exists)

The movies tab always uses a plain table. However, the current API (`library.movies`) only returns filesystem data (name, path, files count) with no poster info. Shoko's folder map may contain movie entries.

- [ ] **Step 1: Check if folder map has movie posters**

In `LibraryView.vue`, import `seriesPosterUrl` (already imported). The `folderMap` from the link wizard maps folder names to Shoko data including poster URLs. However, LibraryView doesn't currently load this map.

Add to imports:
```ts
import { seriesPosterUrl } from '@/lib/utils'
```
(Already imported.)

Add state and fetch the folder map:
```ts
const folderMap = ref<Record<string, { shokoId: number; name: string; posterUrl?: string }>>({})

// In onMounted, after the Shoko series fetch:
try {
  folderMap.value = await api.shokoFolderMap()
} catch { /* ignore */ }
```

Add a helper:
```ts
function moviePosterUrl(movieName: string): string | null {
  const entry = folderMap.value[movieName]
  return entry?.posterUrl ?? null
}
```

- [ ] **Step 2: Replace movies table with poster grid when Shoko is available**

In the Shoko-powered TabsContent for movies (around line 271), replace the plain table with a poster grid + table fallback approach:

```html
<TabsContent value="movies">
  <div v-if="filteredMovies.length === 0" class="py-12">
    <EmptyState
      v-if="searchQuery"
      :icon="Search"
      :heading="`No results for &quot;${searchQuery}&quot;`"
      action-label="Clear filter"
      action-variant="outline"
      @action="searchQuery = ''"
    />
    <EmptyState
      v-else
      :icon="Film"
      heading="No movies yet"
      description="Link anime movies from your downloads"
      action-label="Link New Content"
      @action="router.push('/link')"
    />
  </div>

  <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
    <div
      v-for="movie in filteredMovies"
      :key="movie.path"
      class="group text-left"
    >
      <div class="relative aspect-[2/3] rounded-lg overflow-hidden bg-muted mb-2 ring-1 ring-border/50">
        <img
          v-if="moviePosterUrl(movie.name)"
          :src="moviePosterUrl(movie.name)!"
          :alt="movie.name"
          class="w-full h-full object-cover"
          loading="lazy"
        />
        <div v-else class="w-full h-full flex items-center justify-center">
          <Film class="h-10 w-10 text-muted-foreground/30" />
        </div>
        <div class="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/80 via-black/40 to-transparent p-2 pt-6">
          <Badge variant="secondary" class="text-xs">
            {{ movie.files }} file{{ movie.files !== 1 ? 's' : '' }}
          </Badge>
        </div>
      </div>
      <div class="flex items-start justify-between gap-1">
        <p class="text-sm font-medium leading-tight line-clamp-2">{{ movie.name }}</p>
        <Button
          variant="ghost"
          size="icon"
          class="h-6 w-6 shrink-0 text-muted-foreground hover:text-destructive"
          @click="openUnlinkDialog(movie.name, movie.path, 'movie')"
          title="Unlink movie"
        >
          <Trash2 class="h-3 w-3" />
        </Button>
      </div>
    </div>
  </div>
</TabsContent>
```

- [ ] **Step 3: Commit**

```
feat: show movie posters in library grid instead of plain table
```

---

### Task 11: Fix series detail heading to use display font

**Files:**
- Modify: `frontend/src/views/SeriesDetailView.vue:123`

The heading uses `text-2xl font-bold` which fights the base `h1` display font styles.

- [ ] **Step 1: Fix heading**

Change line 123 from:
```html
<h1 class="text-2xl font-bold">{{ series.Name }}</h1>
```
to:
```html
<h1 class="font-display text-2xl tracking-wider uppercase">{{ series.Name }}</h1>
```

- [ ] **Step 2: Commit**

```
fix: use display font for series detail heading
```

---

### Summary of all changes

| # | Type | What | Files |
|---|------|------|-------|
| 1 | Fix | Downloads heading consistency | `DownloadsView.vue` |
| 2 | Feat | Mobile sidebar drawer | `App.vue`, `AppSidebar.vue` |
| 3 | Feat | Clickable/responsive wizard steps | `LinkWizardView.vue` |
| 4 | Feat | Wizard source filter | `LinkWizardView.vue` |
| 5 | Feat | Library skeleton loading | `LibraryView.vue` |
| 6 | Fix | Scroll to top on navigation | `App.vue` |
| 7 | Feat | Library Load More pagination | `LibraryView.vue` |
| 8 | Fix | De-emphasize re-link button | `DownloadsView.vue` |
| 9 | — | (Covered by Task 3) | — |
| 10 | Feat | Movie poster grid | `LibraryView.vue` |
| 11 | Fix | Series detail heading | `SeriesDetailView.vue` |
