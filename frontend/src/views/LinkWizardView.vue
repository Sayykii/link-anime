<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import { useWebSocket } from '@/composables/useWebSocket'
import { useLibraryStore } from '@/stores/library'
import { useRoute, useRouter } from 'vue-router'
import type { DownloadItem, LinkResult, LinkProgress, Show } from '@/lib/types'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import { ArrowRight } from 'lucide-vue-next'

import SourceStep from '@/components/link-wizard/SourceStep.vue'
import TypeStep from '@/components/link-wizard/TypeStep.vue'
import DetailsStep from '@/components/link-wizard/DetailsStep.vue'
import ConfirmStep from '@/components/link-wizard/ConfirmStep.vue'
import ProgressStep from '@/components/link-wizard/ProgressStep.vue'
import DoneStep from '@/components/link-wizard/DoneStep.vue'

const api = useApi()
const library = useLibraryStore()
const route = useRoute()
const router = useRouter()
const { connect, on, connected } = useWebSocket()

// Wizard state
const step = ref(1)
const loading = ref(false)

// Step 1: Source selection
const downloads = ref<DownloadItem[]>([])
const selectedSource = ref<DownloadItem | null>(null)
const sourceFilter = ref('')
const showLinked = ref(false)

// Bulk link queue
const selectedSources = ref<Set<string>>(new Set())
const bulkQueue = ref<DownloadItem[]>([])
const bulkIndex = ref(0)
const isBulkMode = computed(() => bulkQueue.value.length > 1)
const bulkTotal = computed(() => bulkQueue.value.length)
const bulkCurrent = computed(() => bulkIndex.value + 1)

// Linked file detection
type LinkedSource = { id: number; mediaType: string; showName: string; season?: number }
const linkedSources = ref<Record<string, LinkedSource>>({})

const filteredSources = computed(() => {
  let items = downloads.value
  if (!showLinked.value) {
    items = items.filter(d => !d.linked)
  }
  if (!sourceFilter.value) return items
  const q = sourceFilter.value.toLowerCase().replace(/[.\-_ ]+/g, ' ')
  return items.filter(d =>
    d.name.toLowerCase().replace(/[.\-_ ]+/g, ' ').includes(q)
  )
})

// Step 2: Type
const mediaType = ref<'series' | 'movie'>('series')

// Step 3: Show details
const showName = ref('')
const seasonNumber = ref(1)
const suggestedName = ref('')
const suggestedSeason = ref<number | null>(null)

// Shoko folder mapping for poster matching
const folderMap = ref<Record<string, { shokoId: number; name: string; posterUrl?: string }>>({})
const shokoAvailable = ref(false)
const selectedShow = ref<Show | null>(null)

// Step 4/5: Preview & Progress
const previewResult = ref<LinkResult | null>(null)
const linkProgress = ref<LinkProgress[]>([])
const progressPercent = ref(0)

// Step 6: Final result
const finalResult = ref<LinkResult | null>(null)

const existingMovies = computed(() => library.movies.map(m => m.name))

function findShokoPoster(folderName: string): string | null {
  const entry = folderMap.value[folderName]
  return entry?.posterUrl ?? null
}

const currentPoster = computed(() => findShokoPoster(showName.value))

const nextAvailableSeason = computed(() => {
  if (!selectedShow.value || !selectedShow.value.seasons.length) return null
  return Math.max(...selectedShow.value.seasons.map(s => s.number)) + 1
})

const duplicateSeasonWarning = computed(() => {
  if (!selectedShow.value) return null
  const existing = selectedShow.value.seasons.find(s => s.number === seasonNumber.value)
  if (!existing) return null
  return `Season ${existing.number} already exists (${existing.episodes} episodes)`
})

onMounted(async () => {
  connect()
  await loadDownloads()
  await Promise.all([library.fetchShows(), library.fetchMovies()])

  try {
    folderMap.value = await api.shokoFolderMap()
    shokoAvailable.value = true
  } catch {
    shokoAvailable.value = false
  }

  const sourceParam = route.query.source as string | undefined
  if (sourceParam && downloads.value.length) {
    const match = downloads.value.find(d => d.name === sourceParam)
    if (match) {
      selectSource(match)
      router.replace('/link')
    }
  }
})

on('link:progress', (data) => {
  const p = data as LinkProgress
  linkProgress.value.push(p)
  progressPercent.value = Math.round((p.current / p.total) * 100)
})

on('link:complete', (data) => {
  finalResult.value = data as LinkResult
  step.value = 6
})

async function loadDownloads() {
  loading.value = true
  try {
    const [dl, linked] = await Promise.all([
      api.getDownloads(),
      api.getLinkedSources(),
    ])
    downloads.value = dl
    linkedSources.value = linked
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to load downloads')
  } finally {
    loading.value = false
  }
}

function toggleSource(item: DownloadItem) {
  const set = new Set(selectedSources.value)
  if (set.has(item.path)) {
    set.delete(item.path)
  } else {
    set.add(item.path)
  }
  selectedSources.value = set
}

function startLinking() {
  const queue = downloads.value.filter(d => selectedSources.value.has(d.path))
  if (queue.length === 0) return
  bulkQueue.value = queue
  bulkIndex.value = 0
  beginSource(queue[0])
}

function selectSource(item: DownloadItem) {
  bulkQueue.value = [item]
  bulkIndex.value = 0
  selectedSources.value = new Set([item.path])
  beginSource(item)
}

function beginSource(item: DownloadItem) {
  selectedSource.value = item
  parseName(item.name)
  step.value = 2
}

async function parseName(name: string) {
  try {
    const result = await api.parseRelease(name)
    suggestedName.value = result.name
    suggestedSeason.value = result.season
    showName.value = result.name
    if (result.season !== null) {
      seasonNumber.value = result.season
    }
  } catch {
    showName.value = name
  }
}

function selectType(type: 'series' | 'movie') {
  mediaType.value = type
  step.value = 3
}

function selectExistingShow(show: Show) {
  showName.value = show.name
  selectedShow.value = show
  if (show.seasons.length > 0) {
    const maxSeason = Math.max(...show.seasons.map(s => s.number))
    seasonNumber.value = maxSeason + 1
  }
}

watch(showName, (name) => {
  if (selectedShow.value && name !== selectedShow.value.name) {
    selectedShow.value = null
  }
})

async function goToConfirm() {
  if (!showName.value || !selectedSource.value) return

  loading.value = true
  try {
    previewResult.value = await api.linkPreview({
      source: selectedSource.value.name,
      type: mediaType.value,
      name: showName.value,
      season: mediaType.value === 'series' ? seasonNumber.value : 0,
      dryRun: true,
    })
    step.value = 4
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Preview failed')
  } finally {
    loading.value = false
  }
}

async function executeLink() {
  if (!selectedSource.value) return

  step.value = 5
  linkProgress.value = []
  progressPercent.value = 0
  finalResult.value = null

  try {
    const result = await api.link({
      source: selectedSource.value.name,
      type: mediaType.value,
      name: showName.value,
      season: mediaType.value === 'series' ? seasonNumber.value : 0,
      dryRun: false,
    })

    if (!finalResult.value) {
      finalResult.value = result
      step.value = 6
    }

    toast.success(`Linked ${result.linked} files`)
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Link failed')
    step.value = 4
  }
}

function nextInQueue() {
  const nextIdx = bulkIndex.value + 1
  if (nextIdx < bulkQueue.value.length) {
    bulkIndex.value = nextIdx
    mediaType.value = 'series'
    showName.value = ''
    selectedShow.value = null
    seasonNumber.value = 1
    previewResult.value = null
    linkProgress.value = []
    finalResult.value = null
    progressPercent.value = 0
    beginSource(bulkQueue.value[nextIdx])
  }
}

const hasMoreInQueue = computed(() => bulkIndex.value + 1 < bulkQueue.value.length)

function reset() {
  step.value = 1
  selectedSource.value = null
  mediaType.value = 'series'
  showName.value = ''
  selectedShow.value = null
  seasonNumber.value = 1
  previewResult.value = null
  linkProgress.value = []
  finalResult.value = null
  progressPercent.value = 0
  sourceFilter.value = ''
  selectedSources.value = new Set()
  bulkQueue.value = []
  bulkIndex.value = 0
  loadDownloads()
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="font-display text-3xl tracking-wider uppercase">Link Wizard</h1>
      <p class="text-muted-foreground text-sm mt-1">Hardlink anime from downloads to your media library</p>
    </div>

    <!-- Bulk queue indicator -->
    <div v-if="isBulkMode && step > 1" class="flex items-center gap-2 text-sm">
      <Badge variant="secondary" class="gap-1">
        {{ bulkCurrent }} / {{ bulkTotal }}
      </Badge>
      <span class="text-muted-foreground">{{ selectedSource?.name }}</span>
    </div>

    <!-- Step indicator -->
    <div class="flex items-center gap-1 sm:gap-2 text-sm overflow-x-auto">
      <template
        v-for="(s, i) in [
          { num: 1, label: 'Source' },
          { num: 2, label: 'Type' },
          { num: 3, label: 'Details' },
          { num: 4, label: 'Confirm' },
        ]"
        :key="s.num"
      >
        <button
          :disabled="s.num >= step || step >= 5"
          @click="s.num < step && step < 5 ? step = s.num : null"
          class="shrink-0"
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
        </button>
        <ArrowRight v-if="i < 3" class="h-3 w-3 text-muted-foreground shrink-0" />
      </template>
    </div>

    <!-- Steps -->
    <SourceStep
      v-if="step === 1"
      :downloads="downloads"
      :loading="loading"
      v-model:source-filter="sourceFilter"
      :show-linked="showLinked"
      :filtered-sources="filteredSources"
      :selected-sources="selectedSources"
      :linked-sources="linkedSources"
      @update:show-linked="showLinked = $event"
      @toggle-source="toggleSource"
      @select-source="selectSource"
      @start-linking="startLinking"
    />

    <TypeStep
      v-if="step === 2"
      :source-name="selectedSource?.name ?? ''"
      @select-type="selectType"
      @back="step = 1"
    />

    <DetailsStep
      v-if="step === 3"
      :media-type="mediaType"
      v-model:show-name="showName"
      v-model:season-number="seasonNumber"
      :suggested-name="suggestedName"
      :suggested-season="suggestedSeason"
      :shoko-available="shokoAvailable"
      :current-poster="currentPoster"
      :shows="library.shows"
      :existing-movies="existingMovies"
      :selected-show="selectedShow"
      :next-available-season="nextAvailableSeason"
      :duplicate-season-warning="duplicateSeasonWarning"
      :loading="loading"
      :folder-map="folderMap"
      @select-existing-show="selectExistingShow"
      @confirm="goToConfirm"
      @back="step = 2"
    />

    <ConfirmStep
      v-if="step === 4"
      :selected-source="selectedSource"
      :media-type="mediaType"
      :show-name="showName"
      :season-number="seasonNumber"
      :preview-result="previewResult"
      :shoko-available="shokoAvailable"
      :current-poster="currentPoster"
      @execute="executeLink"
      @back="step = 3"
    />

    <ProgressStep
      v-if="step === 5"
      :progress-percent="progressPercent"
      :link-progress="linkProgress"
    />

    <DoneStep
      v-if="step === 6"
      :final-result="finalResult"
      :show-name="showName"
      :shoko-available="shokoAvailable"
      :current-poster="currentPoster"
      :has-more-in-queue="hasMoreInQueue"
      :bulk-index="bulkIndex"
      :bulk-total="bulkTotal"
      @next-in-queue="nextInQueue"
      @reset="reset"
      @view-library="router.push('/library')"
    />
  </div>
</template>
