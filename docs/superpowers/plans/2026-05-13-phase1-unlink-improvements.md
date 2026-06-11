# Phase 1: Unlink Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Shoko rescan after unlink/undo operations and simplify the unlink confirmation dialogs.

**Architecture:** Backend adds async Shoko scan + notifications to existing unlink/undo handlers. Frontend extracts a shared FileSafetyWarning component and rewrites dialog copy in both LibraryView and HistoryView.

**Tech Stack:** Go (backend handlers, Shoko client), Vue 3 + TypeScript (frontend components), shadcn-vue (UI primitives)

---

### Task 1: Add Shoko rescan + notification to handleUnlink

**Files:**
- Modify: `internal/api/link_handler.go:112-134`

- [ ] **Step 1: Add Shoko rescan and notification to handleUnlink**

Replace the current `handleUnlink` function body (lines 112-134) with:

```go
func (s *Server) handleUnlink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path  string `json:"path"`
		Force bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		jsonError(w, "path is required", http.StatusBadRequest)
		return
	}

	result, err := linker.Unlink(req.Path, req.Force)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.Linked > 0 {
		name := filepath.Base(req.Path)

		if s.Notifier != nil {
			s.Notifier.Send("Unlinked: "+name, "Removed from library", []notify.Field{
				{Name: "Files", Value: fmt.Sprintf("%d removed", result.Linked)},
			}, "red")
		}

		if s.Shoko != nil && s.Shoko.IsConfigured() {
			go func() {
				log.Printf("[shoko] Triggering rescan after unlink: %s", name)
				if err := s.Shoko.ScanAllImportFolders(); err != nil {
					log.Printf("[shoko] Rescan after unlink failed: %v", err)
				}
			}()
		}
	}

	jsonOK(w, result)
}
```

- [ ] **Step 2: Add `path/filepath` to imports if not present**

Ensure the import block in `link_handler.go` includes:
```go
"path/filepath"
```

- [ ] **Step 3: Verify build**

Run: `cd /home/desktop/CodingProjects/link-anime && go build ./...`
Expected: no errors

---

### Task 2: Add Shoko rescan + notification to handleUndo

**Files:**
- Modify: `internal/api/link_handler.go:159-176`

- [ ] **Step 1: Add Shoko rescan and notification to handleUndo**

Replace the current `handleUndo` function body (lines 159-176) with:

```go
func (s *Server) handleUndo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Force bool `json:"force"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, entry, err := linker.Undo(req.Force)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if result.Linked > 0 {
		if s.Notifier != nil {
			s.Notifier.Send("Undid: "+entry.ShowName, "Removed from library", []notify.Field{
				{Name: "Files", Value: fmt.Sprintf("%d removed", result.Linked)},
			}, "red")
		}

		if s.Shoko != nil && s.Shoko.IsConfigured() {
			go func() {
				log.Printf("[shoko] Triggering rescan after undo: %s", entry.ShowName)
				if err := s.Shoko.ScanAllImportFolders(); err != nil {
					log.Printf("[shoko] Rescan after undo failed: %v", err)
				}
			}()
		}
	}

	jsonOK(w, map[string]interface{}{
		"result": result,
		"entry":  entry,
	})
}
```

- [ ] **Step 2: Verify build**

Run: `cd /home/desktop/CodingProjects/link-anime && go build ./...`
Expected: no errors

---

### Task 3: Create FileSafetyWarning.vue shared component

**Files:**
- Create: `frontend/src/components/FileSafetyWarning.vue`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/FileSafetyWarning.vue`:

```vue
<script setup lang="ts">
import { computed } from 'vue'
import type { UnlinkPreview } from '@/lib/types'
import { AlertTriangle } from 'lucide-vue-next'

const props = defineProps<{
  preview: UnlinkPreview
  showZeroNote?: boolean
}>()

const hasUnsafe = computed(() =>
  props.preview.unsafeFiles && props.preview.unsafeFiles.length > 0
)

const safeCount = computed(() =>
  props.preview.safeFiles?.length ?? 0
)

const unsafeCount = computed(() =>
  props.preview.unsafeFiles?.length ?? 0
)
</script>

<template>
  <div class="space-y-3">
    <p>
      <strong>{{ preview.totalFiles }}</strong>
      video file{{ preview.totalFiles !== 1 ? 's' : '' }} will be removed from your library.
    </p>

    <div
      v-if="hasUnsafe"
      class="rounded-md border border-destructive/50 bg-destructive/10 p-3 space-y-2"
    >
      <div class="flex items-center gap-2 text-destructive font-medium">
        <AlertTriangle class="h-4 w-4" />
        Data loss warning
      </div>
      <p class="text-sm">
        <strong>{{ unsafeCount }}</strong>
        {{ unsafeCount !== 1 ? 'files are' : 'file is' }} the only copy &mdash;
        the original in your downloads is gone. Removing
        {{ unsafeCount !== 1 ? 'them' : 'it' }} means
        {{ unsafeCount !== 1 ? "they're" : "it's" }} gone for good.
      </p>
    </div>

    <div v-if="safeCount > 0" class="text-sm text-muted-foreground">
      {{ safeCount }} file{{ safeCount !== 1 ? 's' : '' }}
      still {{ safeCount !== 1 ? 'have' : 'has' }} the original in downloads.
    </div>

    <div v-if="showZeroNote && preview.totalFiles === 0" class="text-sm text-muted-foreground">
      All files are already gone. This will just remove the history entry.
    </div>
  </div>
</template>
```

- [ ] **Step 2: Verify frontend build**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npx vue-tsc --noEmit`
Expected: no type errors

---

### Task 4: Replace LibraryView.vue unlink dialog with FileSafetyWarning

**Files:**
- Modify: `frontend/src/views/LibraryView.vue`

- [ ] **Step 1: Add import**

Add to the imports in `LibraryView.vue`:
```typescript
import FileSafetyWarning from '@/components/FileSafetyWarning.vue'
```

Remove `AlertTriangle` from the lucide import (no longer used directly in this file).

- [ ] **Step 2: Replace dialog description content**

Replace the `<AlertDialogDescription v-else-if="unlinkPreview">` block (lines 515-540) with:

```vue
          <AlertDialogDescription v-else-if="unlinkPreview" as="div">
            <FileSafetyWarning :preview="unlinkPreview" />
          </AlertDialogDescription>
```

- [ ] **Step 3: Simplify button labels**

Replace the footer button section. Change `"Remove all (data loss)"` to `"Remove all"`:

In the `<AlertDialogAction>` inside the `v-if="hasUnsafeFiles"` template (line 555-562), change the label text from:
```
Remove all (data loss)
```
to:
```
Remove all
```

- [ ] **Step 4: Remove hasUnsafeFiles computed if only used in template**

Keep `hasUnsafeFiles` computed — it's still used to conditionally render the two-button vs one-button footer.

- [ ] **Step 5: Verify frontend build**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npx vue-tsc --noEmit`
Expected: no type errors

---

### Task 5: Replace HistoryView.vue undo dialog with FileSafetyWarning

**Files:**
- Modify: `frontend/src/views/HistoryView.vue`

- [ ] **Step 1: Add import**

Add to the imports in `HistoryView.vue`:
```typescript
import FileSafetyWarning from '@/components/FileSafetyWarning.vue'
```

Remove `AlertTriangle` from the lucide import (no longer used directly in this file).

- [ ] **Step 2: Replace dialog description content**

Replace the `<AlertDialogDescription v-else-if="undoPreview && undoEntry">` block (lines 171-205) with:

```vue
          <AlertDialogDescription v-else-if="undoPreview && undoEntry" as="div">
            <p class="mb-3">
              Undo link for "<strong>{{ undoEntry.showName }}</strong>":
            </p>
            <FileSafetyWarning :preview="undoPreview" :show-zero-note="true" />
          </AlertDialogDescription>
```

- [ ] **Step 3: Simplify button labels**

Change `"Remove all (data loss)"` to `"Remove all"` in the undo dialog footer (line 221-228).

- [ ] **Step 4: Verify frontend build**

Run: `cd /home/desktop/CodingProjects/link-anime/frontend && npx vue-tsc --noEmit`
Expected: no type errors

---

### Task 6: Manual smoke test

- [ ] **Step 1: Start dev servers**

Run backend and frontend dev servers.

- [ ] **Step 2: Test unlink flow**

1. Go to Library view
2. Click trash icon on a movie or show
3. Verify dialog shows simplified text from FileSafetyWarning
4. Verify button labels are "Remove" (all safe) or "Remove safe only" / "Remove all" (mixed)
5. Execute unlink — check backend logs for `[shoko] Triggering rescan after unlink`

- [ ] **Step 3: Test undo flow**

1. Go to History view
2. Click "Undo Last"
3. Verify dialog shows entry name + FileSafetyWarning with showZeroNote
4. Execute undo — check backend logs for `[shoko] Triggering rescan after undo`
