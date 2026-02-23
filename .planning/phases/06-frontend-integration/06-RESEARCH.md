# Phase 6: Frontend Integration - Research

**Researched:** 2026-02-23
**Domain:** Vue 3 Composables, Pinia State Management, WebSocket Integration
**Confidence:** HIGH

## Summary

Phase 6 integrates the upscale API and WebSocket messages into the existing Vue 3 frontend. The project already has well-established patterns: `useApi.ts` composable for REST calls, `useWebSocket.ts` composable with typed listener registration, and Pinia stores using composition API style. TypeScript interfaces for upscale entities already exist in `types.ts`.

The implementation is straightforward: add 6 methods to `useApi()`, register 3 WebSocket listeners in a new `useUpscaleStore`, and maintain reactive state for job list and progress. No new libraries required—all patterns are established.

**Primary recommendation:** Extend existing `useApi.ts` with upscale methods, create `stores/upscale.ts` Pinia store following `library.ts` pattern, wire WebSocket listeners in store initialization.

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| FE-01 | useApi methods for all upscale endpoints | Add 6 methods to existing `useApi.ts`: list, create, get, delete, cancel, probe |
| FE-02 | WebSocket listeners for upscale_progress/complete/failed | Use existing `useWebSocket.on()` API with typed handlers in store |

</phase_requirements>

## Standard Stack

### Core (Already Installed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vue 3 | ^3.5.28 | Reactive framework | Project foundation |
| Pinia | ^3.0.4 | State management | Project standard for stores |
| TypeScript | ~5.9.3 | Type safety | All frontend code is typed |

### Supporting (Already Available)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @vueuse/core | ^14.2.1 | Vue utilities | Reactive utilities if needed |

### No New Dependencies

This phase requires zero new npm packages. All patterns exist in the codebase.

## Architecture Patterns

### Existing Project Structure (Frontend)

```
frontend/src/
├── composables/
│   ├── useApi.ts          # REST API wrapper (extend this)
│   └── useWebSocket.ts    # WebSocket client (use as-is)
├── stores/
│   ├── auth.ts            # Auth state
│   └── library.ts         # Library state (follow this pattern)
├── lib/
│   └── types.ts           # TypeScript interfaces (already has upscale types)
└── views/                 # Components use stores
```

### Pattern 1: API Method Addition

**What:** Add typed methods to `useApi()` returning object
**When to use:** All REST endpoint integrations
**Existing Example:**
```typescript
// From useApi.ts - follow this exact pattern
export function useApi() {
  return {
    // ... existing methods
    
    // Upscale endpoints to add
    listUpscaleJobs: () => request<UpscaleJob[]>('GET', '/upscale/jobs'),
    createUpscaleJob: (inputPath: string, preset: string) =>
      request<UpscaleJob>('POST', '/upscale/jobs', { inputPath, preset }),
    getUpscaleJob: (id: number) => request<UpscaleJob>('GET', `/upscale/jobs/${id}`),
    deleteUpscaleJob: (id: number) => request<{ deleted: boolean }>('DELETE', `/upscale/jobs/${id}`),
    cancelUpscaleJob: (id: number) => request<{ cancelled: boolean }>('POST', `/upscale/jobs/${id}/cancel`),
    probeUpscale: () => request<ProbeResult>('GET', '/upscale/probe'),
  }
}
```

### Pattern 2: Pinia Store with Composition API

**What:** Store using `defineStore` with setup function returning refs and actions
**When to use:** State that needs to be shared across components
**Existing Example (library.ts):**
```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '@/composables/useApi'
import type { UpscaleJob, UpscaleProgress } from '@/lib/types'

export const useUpscaleStore = defineStore('upscale', () => {
  const api = useApi()

  const jobs = ref<UpscaleJob[]>([])
  const progress = ref<Record<number, UpscaleProgress>>({}) // jobId -> progress
  const loading = ref(false)
  const probeResult = ref<ProbeResult | null>(null)

  async function fetchJobs() {
    loading.value = true
    try {
      jobs.value = await api.listUpscaleJobs()
    } finally {
      loading.value = false
    }
  }

  // ... other actions

  return { jobs, progress, loading, probeResult, fetchJobs, /* ... */ }
})
```

### Pattern 3: WebSocket Listener Registration

**What:** Use `useWebSocket().on(type, callback)` for typed message handling
**When to use:** Real-time updates from backend
**Existing API:**
```typescript
// useWebSocket returns:
// - connected: Ref<boolean>
// - on(type: string, callback: (data: unknown) => void): () => void

// Usage in store or component:
const ws = useWebSocket()
ws.connect()

// Register typed handlers
ws.on('upscale_progress', (data) => {
  const progress = data as UpscaleProgress
  // Update reactive state
})

ws.on('upscale_complete', (data) => {
  const { jobId, outputPath } = data as { jobId: number; outputPath: string }
  // Refresh job list or update specific job
})

ws.on('upscale_failed', (data) => {
  const { jobId, error } = data as { jobId: number; error: string }
  // Update job status, show notification
})
```

### Anti-Patterns to Avoid

- **Direct fetch() calls:** Always use `useApi()` for consistency and error handling
- **Non-reactive state updates:** Always mutate `.value` on refs, not the underlying object
- **Forgetting to type cast WebSocket data:** The `data` parameter is `unknown`, must cast to typed interface
- **Multiple WebSocket connections:** Use singleton pattern (connect once at app mount)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| API requests | Custom fetch wrapper | `useApi.ts` `request()` function | Already handles auth, errors, redirects |
| WebSocket reconnection | Custom reconnect logic | `useWebSocket.ts` | Already implements 3s auto-reconnect |
| State management | Component local state for shared data | Pinia store | Components need shared job state |

**Key insight:** All infrastructure exists. This phase is purely additive—no new patterns or utilities needed.

## Common Pitfalls

### Pitfall 1: WebSocket Connection Timing

**What goes wrong:** WebSocket listeners registered before connection established
**Why it happens:** Store initialized before `ws.connect()` called
**How to avoid:** Call `ws.connect()` in App.vue on mount, or use lazy listener registration
**Warning signs:** Missing updates, console shows `ws.readyState !== OPEN`

### Pitfall 2: Stale Progress State

**What goes wrong:** Progress bar shows old data after job completes
**Why it happens:** Progress record not cleared on complete/failed events
**How to avoid:** Delete progress entry when job finishes:
```typescript
ws.on('upscale_complete', (data) => {
  const { jobId } = data as { jobId: number }
  delete progress.value[jobId]
  // Also refresh job list
})
```
**Warning signs:** Progress at 100% but status shows completed

### Pitfall 3: Race Condition on Job Creation

**What goes wrong:** New job not immediately visible in list
**Why it happens:** `createUpscaleJob` returns job, but `fetchJobs` called in parallel races
**How to avoid:** Either add returned job directly to state, or await creation before refresh:
```typescript
async function createJob(inputPath: string, preset: string) {
  const job = await api.createUpscaleJob(inputPath, preset)
  jobs.value.unshift(job) // Add directly, don't refetch
}
```
**Warning signs:** Job appears after delay or requires manual refresh

### Pitfall 4: TypeScript Type Mismatch

**What goes wrong:** Type errors between Go JSON and TypeScript interfaces
**Why it happens:** Camel case vs snake case, optional field handling
**How to avoid:** TypeScript interfaces already match Go models (verified in types.ts)
**Warning signs:** Runtime property access errors, type narrowing failures

## Code Examples

### Complete Store Example

```typescript
// stores/upscale.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'
import { useWebSocket } from '@/composables/useWebSocket'
import type { UpscaleJob, UpscaleProgress, ProbeResult } from '@/lib/types'

export const useUpscaleStore = defineStore('upscale', () => {
  const api = useApi()
  const ws = useWebSocket()

  // State
  const jobs = ref<UpscaleJob[]>([])
  const progress = ref<Record<number, UpscaleProgress>>({})
  const probeResult = ref<ProbeResult | null>(null)
  const loading = ref(false)

  // Computed
  const runningJob = computed(() => jobs.value.find(j => j.status === 'running'))
  const pendingJobs = computed(() => jobs.value.filter(j => j.status === 'pending'))
  const pipelineAvailable = computed(() =>
    probeResult.value?.FFmpegFound &&
    probeResult.value?.LibplaceboOK &&
    !!probeResult.value?.VulkanDevice
  )

  // Actions
  async function fetchJobs() {
    loading.value = true
    try {
      jobs.value = await api.listUpscaleJobs()
    } finally {
      loading.value = false
    }
  }

  async function createJob(inputPath: string, preset: string) {
    const job = await api.createUpscaleJob(inputPath, preset)
    jobs.value.unshift(job)
    return job
  }

  async function deleteJob(id: number) {
    await api.deleteUpscaleJob(id)
    jobs.value = jobs.value.filter(j => j.id !== id)
  }

  async function cancelJob(id: number) {
    await api.cancelUpscaleJob(id)
    const job = jobs.value.find(j => j.id === id)
    if (job) job.status = 'cancelled'
    delete progress.value[id]
  }

  async function probe() {
    probeResult.value = await api.probeUpscale()
  }

  // WebSocket handlers (call setupListeners once)
  function setupListeners() {
    ws.on('upscale_progress', (data) => {
      const p = data as UpscaleProgress
      progress.value[p.jobId] = p
    })

    ws.on('upscale_complete', (data) => {
      const { jobId, outputPath } = data as { jobId: number; outputPath: string }
      delete progress.value[jobId]
      const job = jobs.value.find(j => j.id === jobId)
      if (job) {
        job.status = 'completed'
        job.outputPath = outputPath
      }
    })

    ws.on('upscale_failed', (data) => {
      const { jobId, error } = data as { jobId: number; error: string }
      delete progress.value[jobId]
      const job = jobs.value.find(j => j.id === jobId)
      if (job) {
        job.status = 'failed'
        job.error = error
      }
    })
  }

  return {
    // State
    jobs,
    progress,
    probeResult,
    loading,
    // Computed
    runningJob,
    pendingJobs,
    pipelineAvailable,
    // Actions
    fetchJobs,
    createJob,
    deleteJob,
    cancelJob,
    probe,
    setupListeners,
  }
})
```

### ProbeResult TypeScript Interface

```typescript
// Add to lib/types.ts if not present
export interface ProbeResult {
  FFmpegFound: boolean
  LibplaceboOK: boolean
  VulkanDevice: string  // empty if unavailable
}
```

### API Backend Contract Reference

| Endpoint | Method | Request Body | Response |
|----------|--------|--------------|----------|
| `/api/upscale/jobs` | GET | - | `UpscaleJob[]` |
| `/api/upscale/jobs` | POST | `{ inputPath, preset }` | `UpscaleJob` |
| `/api/upscale/jobs/{id}` | GET | - | `UpscaleJob` |
| `/api/upscale/jobs/{id}` | DELETE | - | `{ deleted: true }` |
| `/api/upscale/jobs/{id}/cancel` | POST | - | `{ cancelled: true }` |
| `/api/upscale/probe` | GET | - | `ProbeResult` |

### WebSocket Message Contract Reference

| Type | Data Shape |
|------|------------|
| `upscale_progress` | `{ jobId, frame, fps, time, percent }` |
| `upscale_complete` | `{ jobId, outputPath }` |
| `upscale_failed` | `{ jobId, error }` |

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Options API stores | Composition API stores | Pinia 2.0+ | Project uses composition style |
| Separate composable per domain | Pinia stores for shared state | Project convention | Use stores not composables for job state |

**Current Project Conventions:**
- Composables (`useApi`, `useWebSocket`) for utilities
- Pinia stores for shared reactive state
- Types defined in `lib/types.ts`

## Open Questions

1. **WebSocket Listener Lifecycle**
   - What we know: `useWebSocket` has `on()` method, returns cleanup function
   - What's unclear: Best place to call `setupListeners()` (App.vue vs store instantiation)
   - Recommendation: Call `setupListeners()` in App.vue `onMounted`, after `ws.connect()`

2. **Probe Caching Strategy**
   - What we know: Probe checks FFmpeg/Vulkan availability
   - What's unclear: Should probe be cached or called every view mount?
   - Recommendation: Call once on store init, cache result (hardware rarely changes)

## Sources

### Primary (HIGH confidence)
- `frontend/src/composables/useApi.ts` - Verified API pattern
- `frontend/src/composables/useWebSocket.ts` - Verified WebSocket pattern
- `frontend/src/stores/library.ts` - Verified Pinia store pattern
- `frontend/src/lib/types.ts` - Verified TypeScript interfaces
- `internal/api/upscale_handler.go` - Backend API contract
- `internal/upscale/worker.go` - WebSocket message shapes

### Secondary (MEDIUM confidence)
- `internal/upscale/probe.go` - ProbeResult struct definition

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - No new deps, all patterns verified in codebase
- Architecture: HIGH - Direct extension of existing patterns
- Pitfalls: HIGH - Based on code review of existing implementation

**Research date:** 2026-02-23
**Valid until:** 60 days (stable patterns, no external dependencies)
