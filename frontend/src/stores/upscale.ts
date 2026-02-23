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

  // WebSocket listeners (call once after ws.connect())
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
