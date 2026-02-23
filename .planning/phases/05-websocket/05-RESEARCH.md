# Phase 5: WebSocket - Research

**Researched:** 2026-02-23
**Domain:** WebSocket real-time messaging (Go/TypeScript)
**Confidence:** HIGH

## Summary

**Key Finding:** Phase 5 requirements are ALREADY IMPLEMENTED in Phase 3 (Worker). The worker.go file already broadcasts all three required WebSocket message types (`upscale_progress`, `upscale_complete`, `upscale_failed`) through the existing Hub infrastructure.

This phase should focus on **verification and documentation** rather than new implementation. The WebSocket broadcasts are functional but have not been formally verified with integration tests or documented with message contracts.

**Primary recommendation:** Create verification tests and document the message contracts for frontend consumption. The implementation is complete; this phase formalizes it.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| WS-01 | upscale_progress broadcasts during encode (~1s interval) | **Already implemented** in worker.go (lines 97-101) using progress.go throttling |
| WS-02 | upscale_complete broadcast on success | **Already implemented** in worker.go (lines 127-132) |
| WS-03 | upscale_failed broadcast on error | **Already implemented** in worker.go (lines 117-124) |

### Evidence of Implementation

```go
// worker.go - Progress broadcast (WS-01)
err = w.engine.Run(ctx, job, func(p models.UpscaleProgress) {
    w.hub.Broadcast(models.WSMessage{
        Type: "upscale_progress",
        Data: p,
    })
})

// worker.go - Complete broadcast (WS-02)
w.hub.Broadcast(models.WSMessage{
    Type: "upscale_complete",
    Data: map[string]interface{}{"jobId": job.ID, "outputPath": job.OutputPath},
})

// worker.go - Failed broadcast (WS-03)
w.hub.Broadcast(models.WSMessage{
    Type: "upscale_failed",
    Data: map[string]interface{}{"jobId": job.ID, "error": errStr},
})
```

The ~1s throttling for WS-01 is implemented in progress.go (lines 94-98):
```go
// Throttle updates to ~1 per second
if time.Since(lastUpdate) < time.Second {
    continue
}
```
</phase_requirements>

## Standard Stack

### Core (Already In Use)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gorilla/websocket | existing | WebSocket connections | Already used by Hub, well-maintained |
| internal/ws.Hub | existing | Broadcast management | Project's existing pattern for all WS messages |

### Supporting (Already In Use)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | Message serialization | All message marshaling |
| sync.RWMutex | stdlib | Thread-safe client map | Hub's concurrent access |

**No new dependencies needed** - Phase 5 uses existing infrastructure.

## Architecture Patterns

### Current Project Structure (WebSocket)
```
internal/
├── ws/
│   └── hub.go           # WebSocket hub with Broadcast()
├── upscale/
│   ├── worker.go        # Emits WS messages during job processing
│   └── progress.go      # Throttles progress updates to ~1/s
├── models/
│   └── models.go        # WSMessage, UpscaleProgress types
└── monitor/
    └── monitor.go       # Reference: torrent_progress broadcast pattern
```

### Pattern: Typed Message Broadcasting (Already Implemented)
**What:** All WebSocket messages use consistent `WSMessage{Type, Data}` envelope
**When to use:** Every broadcast to clients
**Example (existing):**
```go
// Source: internal/ws/hub.go
type WSMessage struct {
    Type string      `json:"type"`
    Data interface{} `json:"data,omitempty"`
}

// Source: internal/upscale/worker.go
w.hub.Broadcast(models.WSMessage{
    Type: "upscale_progress",
    Data: models.UpscaleProgress{...},
})
```

### Pattern: Progress Throttling (Already Implemented)
**What:** Limit high-frequency events to prevent client overload
**When to use:** FFmpeg progress (hundreds of updates/second) → ~1/second
**Example (existing):**
```go
// Source: internal/upscale/progress.go
var lastUpdate time.Time
for scanner.Scan() {
    // Throttle updates to ~1 per second
    if time.Since(lastUpdate) < time.Second {
        continue
    }
    lastUpdate = time.Now()
    cb(models.UpscaleProgress{...})
}
```

### Pattern: Reference Implementation (DownloadMonitor)
**What:** Use existing torrent_progress pattern as template
**Why:** Consistency, proven approach
**Example (existing):**
```go
// Source: internal/monitor/monitor.go
m.hub.Broadcast(models.WSMessage{
    Type: "torrent_progress",
    Data: models.TorrentProgress{
        Torrents:  torrents,
        Completed: completed,
    },
})
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| WebSocket management | Custom WS handling | internal/ws.Hub | Already handles connections, cleanup, broadcasting |
| Message format | Custom envelope | models.WSMessage | Project convention, frontend expects this |
| Progress throttling | Additional debouncing | progress.go existing | Already limits to ~1/s |

**Key insight:** This phase should NOT add new code. The infrastructure exists and is tested implicitly through other features (torrent monitoring). Focus on verification and documentation.

## Common Pitfalls

### Pitfall 1: Duplicate Implementation
**What goes wrong:** Adding new WebSocket code when it already exists
**Why it happens:** Requirements listed separately from implementation phase
**How to avoid:** Verify existing code before writing new code
**Warning signs:** Code that looks similar to existing broadcasts

### Pitfall 2: Inconsistent Message Format
**What goes wrong:** upscale messages use different structure than torrent messages
**Why it happens:** Not following established patterns
**How to avoid:** Match models.WSMessage envelope exactly
**Warning signs:** Frontend needing special handling for upscale vs torrent messages
**Status:** ALREADY CORRECT - same WSMessage envelope used

### Pitfall 3: Missing Error Context
**What goes wrong:** upscale_failed broadcasts don't include job ID
**Why it happens:** Only including error message, not identifier
**How to avoid:** Always include jobId in error payload
**Warning signs:** Frontend can't update correct job on failure
**Status:** ALREADY CORRECT - includes jobId and error in map

### Pitfall 4: No Throttling on Progress
**What goes wrong:** Sending hundreds of updates per second
**Why it happens:** FFmpeg outputs progress very frequently
**How to avoid:** Throttle in progress parsing, not hub
**Warning signs:** Client performance issues, high WS traffic
**Status:** ALREADY CORRECT - progress.go throttles to ~1/s

## Code Examples

### Message Type Contracts (Document for Frontend)

**upscale_progress** (WS-01):
```typescript
// Sent ~1/second during encoding
interface UpscaleProgressMessage {
  type: 'upscale_progress'
  data: {
    jobId: number
    frame: number
    fps: number
    time: string    // e.g., "00:05:23.45"
    percent: number // 0-100
  }
}
```

**upscale_complete** (WS-02):
```typescript
// Sent once when job finishes successfully
interface UpscaleCompleteMessage {
  type: 'upscale_complete'
  data: {
    jobId: number
    outputPath: string
  }
}
```

**upscale_failed** (WS-03):
```typescript
// Sent once when job fails
interface UpscaleFailedMessage {
  type: 'upscale_failed'
  data: {
    jobId: number
    error: string
  }
}
```

### Frontend Listener Pattern (for Phase 6)
```typescript
// Using existing useWebSocket composable
const ws = useWebSocket()
ws.connect()

// Register typed listeners
ws.on('upscale_progress', (data: UpscaleProgress) => {
  // Update job progress in store
  updateJobProgress(data.jobId, data.percent, data.fps)
})

ws.on('upscale_complete', (data: { jobId: number, outputPath: string }) => {
  // Mark job complete, refresh job list
  markJobComplete(data.jobId)
})

ws.on('upscale_failed', (data: { jobId: number, error: string }) => {
  // Mark job failed with error
  markJobFailed(data.jobId, data.error)
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Polling for status | WebSocket push | Already implemented | Real-time updates without polling |
| Unthrottled progress | ~1s throttling | Implemented in Phase 3 | Prevents client overload |

**Already current:** No deprecated patterns in use.

## Open Questions

1. **Integration Testing**
   - What we know: Code exists and compiles
   - What's unclear: No automated tests for WS message flow
   - Recommendation: Phase 5 could add verification tests (optional given time constraints)

2. **Message Contract Documentation**
   - What we know: Messages are broadcast correctly
   - What's unclear: Frontend team needs clear contract documentation
   - Recommendation: Phase 5 plan should include documenting message shapes for Phase 6

3. **Error Handling Edge Cases**
   - What we know: Basic error broadcast works
   - What's unclear: What happens if Hub.Broadcast fails (client disconnected mid-send)?
   - Recommendation: Hub already handles this via channel select/default. No action needed.

## Verification Plan (Recommended Phase 5 Scope)

Since implementation is complete, Phase 5 should verify and document:

### Verification Tasks
1. **Manual Test:** Start server, connect WebSocket, queue upscale job, verify all 3 message types
2. **Code Review:** Confirm message payloads match TypeScript types
3. **Document:** Create message contract documentation for frontend

### Verification Commands
```bash
# Verify broadcasts exist in worker
grep -n "upscale_progress\|upscale_complete\|upscale_failed" internal/upscale/worker.go

# Verify throttling in progress parser
grep -n "time.Since(lastUpdate)" internal/upscale/progress.go

# Verify types exist for frontend
grep -n "UpscaleProgress" frontend/src/lib/types.ts
```

## Sources

### Primary (HIGH confidence)
- `internal/upscale/worker.go` - Direct code inspection shows all broadcasts implemented
- `internal/upscale/progress.go` - Direct code inspection shows ~1s throttling
- `internal/ws/hub.go` - Direct code inspection shows Hub.Broadcast pattern
- `internal/models/models.go` - Direct code inspection shows WSMessage, UpscaleProgress types
- `frontend/src/lib/types.ts` - Direct code inspection shows TypeScript interfaces exist

### Secondary (MEDIUM confidence)
- `internal/monitor/monitor.go` - Reference pattern for torrent_progress broadcasts (proven in production)

### Tertiary (LOW confidence)
- None - All findings based on direct code inspection

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing project infrastructure, no new libraries
- Architecture: HIGH - Following established patterns (DownloadMonitor)
- Implementation: HIGH - Direct code inspection confirms all requirements met
- Pitfalls: HIGH - Based on actual code review, not assumptions

**Research date:** 2026-02-23
**Valid until:** N/A (based on current codebase state, not external libraries)

## Recommendation for Planner

**Phase 5 is essentially complete.** The planner should create a single verification plan that:

1. Confirms all 3 message types broadcast correctly (quick manual test or script)
2. Documents the message contracts in a format useful for Phase 6 frontend work
3. Updates REQUIREMENTS.md to mark WS-01, WS-02, WS-03 as complete

This phase could be as simple as a 15-minute verification task, or could include writing automated tests if desired. The implementation work was done in Phase 3.
