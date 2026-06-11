# Phase 1: Unlink Improvements — Shoko Rescan + Dialog Simplification

## A. Shoko Rescan After Unlink/Undo

### Problem
Unlinking removes files from the library directory but never notifies Shoko. Shoko's database retains stale references to files that no longer exist until its next periodic scan.

### Solution
Trigger `ScanAllImportFolders()` in an async goroutine after successful unlink or undo, same pattern as `handleLink` uses for `ImportNewFiles()`.

### Changes

**`internal/api/link_handler.go`**

In `handleUnlink` — after `linker.Unlink()` succeeds and `result.Linked > 0`:
```go
if s.Shoko != nil && s.Shoko.IsConfigured() && result.Linked > 0 {
    go func() {
        log.Printf("[shoko] Triggering rescan after unlink: %s", req.Path)
        if err := s.Shoko.ScanAllImportFolders(); err != nil {
            log.Printf("[shoko] Rescan after unlink failed: %v", err)
        }
    }()
}
```

In `handleUndo` — same pattern after `linker.Undo()` succeeds and `result.Linked > 0`.

Add notification via `s.Notifier` for unlink/undo (parallel to link notifications):
- Title: "Unlinked: {name}"
- Color: "red"
- Fields: Files removed count, skipped count

`handleUnlink` needs access to the target name for notifications. The current request body only has `path` and `force`. We derive the name from the path (last directory component) to avoid changing the API contract.

## B. Unlink Dialog — Simplified Text

### Problem
Current dialog text uses technical language ("only copy", "source file in downloads no longer exists") and redundant labeling ("Remove all (data loss)" when a warning box already explains the risk).

### Solution
Rewrite dialog copy to plain English. Extract shared `FileSafetyWarning.vue` component.

### New Component: `frontend/src/components/FileSafetyWarning.vue`

Props:
- `preview: UnlinkPreview` — the safety check result
- `targetName: string` — what's being unlinked (for the header text)

Encapsulates:
- Total file count message
- Unsafe file warning box (conditional)
- Safe file note (conditional)
- Zero-files-remaining note (conditional, undo only via prop)

### Rewritten Copy

**Main message:**
> "**X** video files will be removed from your library."

**Unsafe warning box** (when unsafe files exist):
> "**Y** of these are the only copy — the original in your downloads is gone. Removing them means they're gone for good."

**Safe note** (when safe files exist):
> "**Z** files still have their original in downloads."

**Zero files note** (undo only, when all files already gone):
> "All files are already gone. This will just remove the history entry."

### Button Labels

When mixed safe/unsafe:
- "Remove safe only" (outline variant, unchanged)
- "Remove all" (destructive variant, no parenthetical)

When all safe:
- "Remove" (destructive variant)

When all gone (undo):
- "Remove entry"

### Files Modified

- `frontend/src/components/FileSafetyWarning.vue` — new shared component
- `frontend/src/views/LibraryView.vue` — replace inline dialog content with FileSafetyWarning
- `frontend/src/views/HistoryView.vue` — replace inline dialog content with FileSafetyWarning
- `internal/api/link_handler.go` — add Shoko rescan + notifications to handleUnlink and handleUndo
