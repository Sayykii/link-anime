# Phase 7: Frontend UI - Research

**Researched:** 2026-02-23
**Domain:** Vue 3 Components, Dialog/Modal Patterns, Real-time Progress Display
**Confidence:** HIGH

## Summary

Phase 7 builds the user interface for triggering, monitoring, and managing upscale jobs from the Downloads view. The infrastructure is complete: Phase 6 delivered a fully-functional `useUpscaleStore` with all API methods, computed properties (`pipelineAvailable`, `runningJob`, `pendingJobs`), and WebSocket listeners (`setupListeners()`). The types (`UpscaleJob`, `UpscaleProgress`, `ProbeResult`) already exist in `types.ts`.

The implementation follows established patterns from `DownloadsView.vue` — which already uses Tabs, Table components, Progress bars, Dialog/AlertDialog for confirmations, and Badge for status. The view has ~671 lines and demonstrates all required UI patterns: tabbed layout, real-time WebSocket updates, action buttons with loading states, and toast notifications.

**Primary recommendation:** Extend DownloadsView with a fourth "Upscale Queue" tab. Add upscale button to local file items (conditionally on `pipelineAvailable`). Use Dialog for preset picker. Display 4K badge on items with completed upscale by cross-referencing jobs.

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| UI-01 | Upscale button on download items (when probe available) | Conditional render on `pipelineAvailable` computed, follow existing "Link" button pattern |
| UI-02 | Preset picker dialog (Fast/Balanced/Quality) | Use Dialog component with radio/select for 3 presets, submit calls `createJob()` |
| UI-03 | Upscale Queue tab with job list | Add fourth TabsTrigger/TabsContent to existing Tabs structure |
| UI-04 | Progress bar + FPS for running jobs (via WebSocket) | Use Progress component + `progress.value[job.id]` from store |
| UI-05 | Cancel button for running/pending jobs | Button calls `cancelJob(id)`, disabled for completed/failed |
| UI-06 | Delete button for completed/failed jobs | Button calls `deleteJob(id)`, only for terminal statuses |
| UI-07 | 4K badge on items with completed upscale | Cross-reference `jobs.value` by inputPath to show badge |

</phase_requirements>

## Standard Stack

### Core (Already Installed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vue 3 | ^3.5.28 | Reactive framework | Project foundation |
| Pinia | ^3.0.4 | State management | `useUpscaleStore` already created |
| reka-ui | ^2.8.0 | Headless UI components | Dialog, Progress, Tabs primitives |
| lucide-vue-next | ^0.563.0 | Icons | Consistent iconography |
| vue-sonner | ^2.0.9 | Toast notifications | User feedback |

### UI Components (Already Available)

| Component | Location | Use Case |
|-----------|----------|----------|
| Dialog | `@/components/ui/dialog` | Preset picker modal |
| Tabs | `@/components/ui/tabs` | Queue tab addition |
| Progress | `@/components/ui/progress` | Job progress bar |
| Badge | `@/components/ui/badge` | Status badges, 4K badge |
| Button | `@/components/ui/button` | Action buttons |
| Table | `@/components/ui/table` | Job list display |
| AlertDialog | `@/components/ui/alert-dialog` | Cancel/delete confirmations |

### No New Dependencies

All UI components exist. No npm packages needed.

## Architecture Patterns

### Existing Project Structure (Relevant Files)

```
frontend/src/
├── views/
│   └── DownloadsView.vue     # Extend this (add queue tab, upscale button)
├── stores/
│   └── upscale.ts            # Already complete from Phase 6
├── composables/
│   └── useWebSocket.ts       # WebSocket singleton
├── components/ui/
│   ├── dialog/               # Preset picker
│   ├── progress/             # Job progress
│   └── badge/                # Status + 4K badge
└── lib/
    └── types.ts              # UpscaleJob, UpscaleProgress types
```

### Pattern 1: Conditional Action Button (UI-01)

**What:** Button that appears only when feature is available
**Existing Example (from DownloadsView line 382-385):**
```vue
<Button size="sm" variant="outline" @click="goToLink(item.name)" class="gap-1 shrink-0">
  <Link class="h-3 w-3" />
  Link
</Button>
```
**Upscale Implementation:**
```vue
<Button 
  v-if="upscaleStore.pipelineAvailable" 
  size="sm" 
  variant="outline" 
  @click="openUpscaleDialog(item)"
  class="gap-1 shrink-0"
>
  <Sparkles class="h-3 w-3" />
  Upscale
</Button>
```

### Pattern 2: Dialog with Form (UI-02)

**What:** Modal dialog with form inputs and submit action
**Use reka-ui Dialog primitives (already wrapped in project):**
```vue
<Dialog v-model:open="presetDialogOpen">
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Upscale Quality</DialogTitle>
      <DialogDescription>Choose upscaling preset for {{ selectedFile?.name }}</DialogDescription>
    </DialogHeader>
    
    <div class="space-y-3 py-4">
      <label 
        v-for="preset in presets" 
        :key="preset.value"
        class="flex items-start gap-3 rounded-lg border p-3 cursor-pointer"
        :class="{ 'border-primary bg-primary/5': selectedPreset === preset.value }"
      >
        <input 
          type="radio" 
          v-model="selectedPreset" 
          :value="preset.value"
          class="mt-1"
        />
        <div>
          <div class="font-medium">{{ preset.label }}</div>
          <div class="text-sm text-muted-foreground">{{ preset.description }}</div>
        </div>
      </label>
    </div>
    
    <DialogFooter>
      <Button variant="outline" @click="presetDialogOpen = false">Cancel</Button>
      <Button @click="submitUpscale" :disabled="!selectedPreset || submitting">
        <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
        Queue Upscale
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

### Pattern 3: Tab Addition (UI-03)

**What:** Adding a tab to existing Tabs component
**Current Tabs (DownloadsView line 303-324):**
```vue
<Tabs v-model="activeTab">
  <TabsList>
    <TabsTrigger value="local">Local Files</TabsTrigger>
    <TabsTrigger value="torrents">Torrents</TabsTrigger>
    <TabsTrigger value="nyaa">Nyaa</TabsTrigger>
  </TabsList>
  <!-- TabsContent blocks follow -->
</Tabs>
```
**Add fourth tab:**
```vue
<TabsTrigger value="upscale" class="gap-2">
  <Sparkles class="h-4 w-4" />
  Upscale Queue
  <span v-if="upscaleStore.runningJob" class="ml-1 h-2 w-2 rounded-full bg-green-500 animate-pulse"></span>
</TabsTrigger>
```

### Pattern 4: Real-time Progress Display (UI-04)

**What:** Progress bar with live stats from WebSocket
**Use store's reactive progress state:**
```vue
<template v-if="job.status === 'running' && upscaleStore.progress[job.id]">
  <div class="flex items-center gap-2">
    <Progress :model-value="upscaleStore.progress[job.id].percent" class="w-20 h-2" />
    <span class="text-xs tabular-nums">
      {{ upscaleStore.progress[job.id].percent.toFixed(1) }}%
    </span>
    <span class="text-xs text-muted-foreground tabular-nums">
      {{ upscaleStore.progress[job.id].fps.toFixed(1) }} fps
    </span>
  </div>
</template>
```

### Pattern 5: Conditional Action Buttons (UI-05, UI-06)

**What:** Different buttons based on job status
**Implementation:**
```vue
<template v-if="job.status === 'running' || job.status === 'pending'">
  <Button size="sm" variant="ghost" @click="handleCancel(job.id)">
    <X class="h-4 w-4" />
  </Button>
</template>
<template v-else-if="job.status === 'completed' || job.status === 'failed'">
  <Button size="sm" variant="ghost" @click="handleDelete(job.id)">
    <Trash2 class="h-4 w-4" />
  </Button>
</template>
```

### Pattern 6: Cross-Reference Badge (UI-07)

**What:** Badge shown on download items that have completed upscale
**Implementation approach:**
```typescript
// In script setup
const completedUpscalePaths = computed(() => {
  const paths = new Set<string>()
  for (const job of upscaleStore.jobs) {
    if (job.status === 'completed') {
      paths.add(job.inputPath)
    }
  }
  return paths
})

function hasUpscaled(item: DownloadItem): boolean {
  // Check if any file in this download has a completed upscale
  // For directories, would need to check all video files
  return completedUpscalePaths.value.has(item.path)
}
```
**Template:**
```vue
<Badge v-if="hasUpscaled(item)" variant="secondary" class="text-xs">
  4K
</Badge>
```

### Anti-Patterns to Avoid

- **Direct API calls in template:** Always go through store actions
- **Manual WebSocket subscription:** Store's `setupListeners()` already handles it
- **Forgetting to call probe():** Must probe on mount to determine `pipelineAvailable`
- **Polling for progress:** Use WebSocket updates, not API polling

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Modal/dialog | Custom overlay | `Dialog` component | Accessibility, focus trap, ESC handling |
| Progress bar | `<div>` with width | `Progress` component | Proper ARIA, animation |
| Radio group | Checkboxes | Native `<input type="radio">` or reka-ui RadioGroup | Form semantics |
| Toast notifications | Custom alert | `vue-sonner` via `toast()` | Queue, dismissal, styling |

**Key insight:** The UI component library already handles all complex interaction patterns. Focus on composition, not recreation.

## Common Pitfalls

### Pitfall 1: Store Not Initialized

**What goes wrong:** `pipelineAvailable` always false, no jobs shown
**Why it happens:** Forgot to call `upscaleStore.probe()` and `upscaleStore.fetchJobs()` on mount
**How to avoid:**
```typescript
onMounted(async () => {
  // Existing DownloadsView mount logic...
  
  // Add upscale initialization
  await upscaleStore.probe()
  await upscaleStore.fetchJobs()
  upscaleStore.setupListeners()
})
```
**Warning signs:** Upscale button never appears, queue always empty

### Pitfall 2: WebSocket Listeners Not Registered

**What goes wrong:** Progress bar never updates, status stuck
**Why it happens:** `setupListeners()` not called after WebSocket connect
**How to avoid:** Call in onMounted after existing `connect()`:
```typescript
onMounted(() => {
  // ... existing code ...
  connect()
  // ... existing on() calls ...
  
  // Add upscale listeners
  upscaleStore.setupListeners()
})
```
**Warning signs:** Jobs start but progress stays at 0%

### Pitfall 3: Stale Job List After Create

**What goes wrong:** New job doesn't appear in queue immediately
**Why it happens:** Store's `createJob()` already adds to `jobs` array — but you might refetch
**How to avoid:** Don't call `fetchJobs()` after create; trust store's optimistic update
**Warning signs:** Job appears, disappears, reappears

### Pitfall 4: 4K Badge Path Mismatch

**What goes wrong:** Badge doesn't appear even though upscale completed
**Why it happens:** `inputPath` in job doesn't match `item.path` format
**How to avoid:** Normalize paths or check for containment:
```typescript
function hasUpscaled(item: DownloadItem): boolean {
  return upscaleStore.jobs.some(
    j => j.status === 'completed' && j.inputPath.includes(item.path)
  )
}
```
**Warning signs:** Badge never shows, or shows on wrong items

### Pitfall 5: Dialog State Not Reset

**What goes wrong:** Dialog shows stale file/preset from previous open
**Why it happens:** Forgot to reset `selectedPreset` and `selectedFile` on close
**How to avoid:**
```typescript
function openUpscaleDialog(item: DownloadItem) {
  selectedFile.value = item
  selectedPreset.value = 'balanced' // Default
  presetDialogOpen.value = true
}
```
**Warning signs:** Wrong file name shown, or preset from last time selected

## Code Examples

### Preset Configuration

```typescript
// Presets matching backend (internal/upscale/probe.go)
const presets = [
  { 
    value: 'fast', 
    label: 'Fast', 
    description: 'Quick upscale with basic enhancement. Good for batch processing.' 
  },
  { 
    value: 'balanced', 
    label: 'Balanced', 
    description: 'Best quality/speed tradeoff. Recommended for most content.' 
  },
  { 
    value: 'quality', 
    label: 'Quality', 
    description: 'Maximum quality, slower processing. For your favorites.' 
  },
]
```

### Status Badge Styling

```typescript
function statusVariant(status: UpscaleStatus): 'default' | 'secondary' | 'outline' | 'destructive' {
  switch (status) {
    case 'running': return 'default'
    case 'pending': return 'outline'
    case 'completed': return 'secondary'
    case 'failed': return 'destructive'
    case 'cancelled': return 'outline'
    default: return 'outline'
  }
}

function statusLabel(status: UpscaleStatus): string {
  return status.charAt(0).toUpperCase() + status.slice(1)
}
```

### Complete Queue Tab Structure

```vue
<TabsContent value="upscale">
  <Card>
    <CardHeader class="flex flex-row items-center justify-between">
      <div>
        <CardTitle>Upscale Queue</CardTitle>
        <CardDescription>
          {{ upscaleStore.jobs.length }} job{{ upscaleStore.jobs.length !== 1 ? 's' : '' }}
          <template v-if="upscaleStore.runningJob"> — 1 running</template>
        </CardDescription>
      </div>
      <Button variant="outline" size="sm" @click="upscaleStore.fetchJobs()" class="gap-2">
        <RefreshCw class="h-4 w-4" />
        Refresh
      </Button>
    </CardHeader>
    <CardContent>
      <!-- Empty state -->
      <div v-if="!upscaleStore.jobs.length" class="text-center text-muted-foreground py-8">
        No upscale jobs. Select a file and click "Upscale" to get started.
      </div>
      
      <!-- Job table -->
      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead>File</TableHead>
            <TableHead class="w-24">Preset</TableHead>
            <TableHead class="w-24">Status</TableHead>
            <TableHead class="w-40">Progress</TableHead>
            <TableHead class="w-10"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="job in upscaleStore.jobs" :key="job.id">
            <!-- File name -->
            <TableCell class="font-medium max-w-sm truncate">
              {{ extractFileName(job.inputPath) }}
            </TableCell>
            
            <!-- Preset -->
            <TableCell class="capitalize">{{ job.preset }}</TableCell>
            
            <!-- Status -->
            <TableCell>
              <Badge :variant="statusVariant(job.status)">
                {{ statusLabel(job.status) }}
              </Badge>
            </TableCell>
            
            <!-- Progress -->
            <TableCell>
              <template v-if="job.status === 'running' && upscaleStore.progress[job.id]">
                <div class="flex items-center gap-2">
                  <Progress :model-value="upscaleStore.progress[job.id].percent" class="w-20 h-2" />
                  <span class="text-xs tabular-nums">
                    {{ upscaleStore.progress[job.id].percent.toFixed(1) }}%
                  </span>
                  <span class="text-xs text-muted-foreground">
                    {{ upscaleStore.progress[job.id].fps.toFixed(1) }} fps
                  </span>
                </div>
              </template>
              <template v-else-if="job.status === 'failed'">
                <span class="text-xs text-destructive truncate max-w-32" :title="job.error">
                  {{ job.error }}
                </span>
              </template>
              <span v-else class="text-muted-foreground">-</span>
            </TableCell>
            
            <!-- Actions -->
            <TableCell>
              <Button 
                v-if="job.status === 'running' || job.status === 'pending'"
                size="sm" 
                variant="ghost" 
                @click="handleCancel(job.id)"
              >
                <X class="h-4 w-4" />
              </Button>
              <Button 
                v-else
                size="sm" 
                variant="ghost" 
                @click="handleDelete(job.id)"
              >
                <Trash2 class="h-4 w-4" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </CardContent>
  </Card>
</TabsContent>
```

### Helper Functions

```typescript
function extractFileName(path: string): string {
  return path.split('/').pop() || path
}

async function handleCancel(id: number) {
  try {
    await upscaleStore.cancelJob(id)
    toast.success('Job cancelled')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to cancel job')
  }
}

async function handleDelete(id: number) {
  try {
    await upscaleStore.deleteJob(id)
    toast.success('Job removed')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to delete job')
  }
}

async function submitUpscale() {
  if (!selectedFile.value || !selectedPreset.value) return
  submitting.value = true
  try {
    await upscaleStore.createJob(selectedFile.value.path, selectedPreset.value)
    toast.success('Job queued', { description: selectedFile.value.name })
    presetDialogOpen.value = false
    activeTab.value = 'upscale' // Switch to queue tab
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to queue upscale')
  } finally {
    submitting.value = false
  }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Options API components | Composition API with `<script setup>` | Vue 3.2+ | Project standard, follow it |
| Vuex | Pinia | Project uses Pinia 3.x | Store already exists |
| Custom modals | reka-ui Dialog primitives | Project convention | Use existing components |

**Current Project Conventions:**
- All views use `<script setup lang="ts">` with Composition API
- Toast for user feedback (`toast.success()`, `toast.error()`)
- Loading states with `Loader2` spinner icon
- Tables for list data, responsive with mobile card fallback
- AlertDialog for destructive confirmations

## Open Questions

1. **Directory Upscale Handling**
   - What we know: Downloads can be directories containing multiple videos
   - What's unclear: Should upscale button on directory queue all videos?
   - Recommendation: v1 — only show upscale button for single video files, not directories. Add batch support in v2.

2. **4K Badge Granularity**
   - What we know: Jobs track `inputPath` and `outputPath`
   - What's unclear: For a directory download, how to show partial completion?
   - Recommendation: Show "4K" badge only when all videos in download have completed upscale. For v1, skip badge on directories entirely — show only on individual files.

3. **Queue Sort Order**
   - What we know: Backend returns jobs in creation order (oldest first)
   - What's unclear: Should running job always be at top?
   - Recommendation: Display as-returned. Running job will naturally be the oldest pending that got picked up. Could add client-side sort if needed.

## Sources

### Primary (HIGH confidence)
- `frontend/src/views/DownloadsView.vue` — Existing UI patterns (671 lines of reference)
- `frontend/src/stores/upscale.ts` — Store API, all methods available
- `frontend/src/lib/types.ts` — TypeScript interfaces
- `frontend/src/components/ui/dialog/` — Dialog component structure
- `frontend/src/components/ui/progress/` — Progress component
- `internal/upscale/probe.go` — Preset definitions (fast/balanced/quality)
- `internal/api/upscale_handler.go` — API contract validation

### Secondary (MEDIUM confidence)
- reka-ui documentation — Dialog, Progress primitives

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — All components already exist in project
- Architecture: HIGH — Direct extension of existing DownloadsView patterns
- Pitfalls: HIGH — Based on code review of store and existing view

**Research date:** 2026-02-23
**Valid until:** 90 days (stable patterns, no external dependencies)
