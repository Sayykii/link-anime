<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import { useWebSocket } from '@/composables/useWebSocket'
import { useLibraryStore } from '@/stores/library'
import { useRoute, useRouter } from 'vue-router'
import type { DownloadItem, LinkResult, LinkProgress, Show } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Progress } from '@/components/ui/progress'
import { toast } from 'vue-sonner'
import { FolderOpen, FileVideo, Link, ArrowRight, Check, Loader2 } from 'lucide-vue-next'

const api = useApi()
const library = useLibraryStore()
const route = useRoute()
const router = useRouter()
const { connect, on, connected } = useWebSocket()

// Wizard state
const step = ref(1) // 1=source, 2=type, 3=details, 4=confirm, 5=progress, 6=done
const loading = ref(false)

// Step 1: Source selection
const downloads = ref<DownloadItem[]>([])
const selectedSource = ref<DownloadItem | null>(null)

// Step 2: Type
const mediaType = ref<'series' | 'movie'>('series')

// Step 3: Show details
const showName = ref('')
const seasonNumber = ref(1)
const suggestedName = ref('')
const suggestedSeason = ref<number | null>(null)

// Step 4/5: Preview & Progress
const previewResult = ref<LinkResult | null>(null)
const linkProgress = ref<LinkProgress[]>([])
const progressPercent = ref(0)

// Step 6: Final result
const finalResult = ref<LinkResult | null>(null)

// Existing shows for autocomplete
const existingShows = computed(() => library.shows.map(s => s.name))

onMounted(async () => {
  connect()
  await loadDownloads()
  await library.fetchShows()

  // Auto-select source from query param (from Downloads page "Link" button)
  const sourceParam = route.query.source as string | undefined
  if (sourceParam && downloads.value.length) {
    const match = downloads.value.find(d => d.name === sourceParam)
    if (match) {
      selectSource(match)
      // Clean the URL to prevent re-triggering on refresh
      router.replace('/link')
    }
  }
})

// Listen for WebSocket progress
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
    downloads.value = await api.getDownloads()
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to load downloads')
  } finally {
    loading.value = false
  }
}

function selectSource(item: DownloadItem) {
  selectedSource.value = item
  // Auto-parse the release name
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

    // If WS didn't trigger completion, set it manually
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

function reset() {
  step.value = 1
  selectedSource.value = null
  mediaType.value = 'series'
  showName.value = ''
  seasonNumber.value = 1
  previewResult.value = null
  linkProgress.value = []
  finalResult.value = null
  progressPercent.value = 0
  loadDownloads()
}

function formatSize(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(2) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  return (bytes / 1024).toFixed(1) + ' KB'
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-bold">Link Wizard</h1>
      <p class="text-muted-foreground">Hardlink anime from downloads to your media library</p>
    </div>

    <!-- Step indicator -->
    <div class="flex items-center gap-2 text-sm">
      <Badge :variant="step >= 1 ? 'default' : 'outline'">1. Source</Badge>
      <ArrowRight class="h-3 w-3 text-muted-foreground" />
      <Badge :variant="step >= 2 ? 'default' : 'outline'">2. Type</Badge>
      <ArrowRight class="h-3 w-3 text-muted-foreground" />
      <Badge :variant="step >= 3 ? 'default' : 'outline'">3. Details</Badge>
      <ArrowRight class="h-3 w-3 text-muted-foreground" />
      <Badge :variant="step >= 4 ? 'default' : 'outline'">4. Confirm</Badge>
    </div>

    <!-- Step 1: Select source -->
    <Card v-if="step === 1">
      <CardHeader>
        <CardTitle>Select Source</CardTitle>
        <CardDescription>Choose a download to link into your library</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="flex items-center gap-2 text-muted-foreground py-8 justify-center">
          <Loader2 class="h-4 w-4 animate-spin" />
          Loading downloads...
        </div>
        <div v-else-if="!downloads.length" class="text-center text-muted-foreground py-8">
          No downloads found in the download directory
        </div>
        <div v-else class="space-y-2">
          <button
            v-for="item in downloads"
            :key="item.path"
            class="flex w-full items-center gap-3 rounded-lg border p-3 text-left hover:bg-accent transition-colors"
            @click="selectSource(item)"
          >
            <FolderOpen v-if="item.isDir" class="h-5 w-5 text-muted-foreground shrink-0" />
            <FileVideo v-else class="h-5 w-5 text-muted-foreground shrink-0" />
            <div class="min-w-0 flex-1">
              <div class="truncate font-medium">{{ item.name }}</div>
              <div class="text-sm text-muted-foreground">
                {{ item.videoCount }} video{{ item.videoCount !== 1 ? 's' : '' }}
                &middot; {{ formatSize(item.size) }}
              </div>
            </div>
          </button>
        </div>
      </CardContent>
    </Card>

    <!-- Step 2: Select type -->
    <Card v-if="step === 2">
      <CardHeader>
        <CardTitle>Select Type</CardTitle>
        <CardDescription>
          Is "{{ selectedSource?.name }}" a series or a movie?
        </CardDescription>
      </CardHeader>
      <CardContent class="flex gap-4">
        <Button
          size="lg"
          :variant="mediaType === 'series' ? 'default' : 'outline'"
          class="flex-1 gap-2"
          @click="selectType('series')"
        >
          Series
        </Button>
        <Button
          size="lg"
          :variant="mediaType === 'movie' ? 'default' : 'outline'"
          class="flex-1 gap-2"
          @click="selectType('movie')"
        >
          Movie
        </Button>
      </CardContent>
      <CardContent class="pt-0">
        <Button variant="ghost" size="sm" @click="step = 1">Back</Button>
      </CardContent>
    </Card>

    <!-- Step 3: Details -->
    <Card v-if="step === 3">
      <CardHeader>
        <CardTitle>{{ mediaType === 'series' ? 'Series' : 'Movie' }} Details</CardTitle>
        <CardDescription>
          Confirm the name{{ mediaType === 'series' ? ' and season' : '' }}
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label>Name</Label>
          <Input v-model="showName" placeholder="Show or movie name" />
          <p v-if="suggestedName && suggestedName !== showName" class="text-sm text-muted-foreground">
            Suggested: {{ suggestedName }}
            <Button variant="link" size="sm" class="h-auto p-0 ml-1" @click="showName = suggestedName">
              Use this
            </Button>
          </p>
          <!-- Existing shows dropdown -->
          <div v-if="mediaType === 'series' && existingShows.length" class="space-y-1">
            <Label class="text-xs text-muted-foreground">Or select existing show:</Label>
            <Select @update:model-value="(v: string) => showName = v">
              <SelectTrigger>
                <SelectValue placeholder="Select existing show..." />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="name in existingShows" :key="name" :value="name">
                  {{ name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div v-if="mediaType === 'series'" class="space-y-2">
          <Label>Season Number</Label>
          <Input v-model.number="seasonNumber" type="number" min="0" max="99" />
          <p v-if="suggestedSeason !== null && suggestedSeason !== seasonNumber" class="text-sm text-muted-foreground">
            Detected season: {{ suggestedSeason }}
            <Button variant="link" size="sm" class="h-auto p-0 ml-1" @click="seasonNumber = suggestedSeason!">
              Use this
            </Button>
          </p>
        </div>

        <Separator />

        <div class="flex gap-2">
          <Button variant="ghost" @click="step = 2">Back</Button>
          <Button @click="goToConfirm" :disabled="!showName || loading">
            {{ loading ? 'Checking...' : 'Preview' }}
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Step 4: Confirm -->
    <Card v-if="step === 4">
      <CardHeader>
        <CardTitle>Confirm Link</CardTitle>
        <CardDescription>Review before linking</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="text-muted-foreground">Source:</span>
            <div class="font-medium">{{ selectedSource?.name }}</div>
          </div>
          <div>
            <span class="text-muted-foreground">Type:</span>
            <div class="font-medium capitalize">{{ mediaType }}</div>
          </div>
          <div>
            <span class="text-muted-foreground">Name:</span>
            <div class="font-medium">{{ showName }}</div>
          </div>
          <div v-if="mediaType === 'series'">
            <span class="text-muted-foreground">Season:</span>
            <div class="font-medium">{{ seasonNumber }}</div>
          </div>
        </div>

        <Separator />

        <div v-if="previewResult" class="space-y-2">
          <h4 class="font-medium">Preview:</h4>
          <div class="text-sm space-y-1">
            <div>Destination: <code class="text-xs bg-muted px-1 py-0.5 rounded">{{ previewResult.destDir }}</code></div>
            <div>Files to link: <strong>{{ previewResult.linked }}</strong></div>
            <div v-if="previewResult.skipped">Already exists: {{ previewResult.skipped }}</div>
            <div>Total size: {{ formatSize(previewResult.size) }}</div>
          </div>
        </div>

        <Separator />

        <div class="flex gap-2">
          <Button variant="ghost" @click="step = 3">Back</Button>
          <Button @click="executeLink" class="gap-2">
            <Link class="h-4 w-4" />
            Link Now
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Step 5: Progress -->
    <Card v-if="step === 5">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Loader2 class="h-5 w-5 animate-spin" />
          Linking...
        </CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <Progress :model-value="progressPercent" />
        <div class="max-h-48 overflow-auto space-y-1 text-sm font-mono">
          <div v-for="(p, i) in linkProgress" :key="i" class="flex items-center gap-2">
            <Badge
              :variant="p.status === 'linked' ? 'default' : p.status === 'skipped' ? 'secondary' : 'destructive'"
              class="text-xs w-16 justify-center"
            >
              {{ p.status }}
            </Badge>
            <span class="truncate">{{ p.file }}</span>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Step 6: Done -->
    <Card v-if="step === 6">
      <CardHeader>
        <CardTitle class="flex items-center gap-2 text-green-600">
          <Check class="h-5 w-5" />
          Link Complete
        </CardTitle>
      </CardHeader>
      <CardContent class="space-y-4" v-if="finalResult">
        <div class="grid grid-cols-3 gap-4 text-center">
          <div>
            <div class="text-2xl font-bold">{{ finalResult.linked }}</div>
            <div class="text-sm text-muted-foreground">Linked</div>
          </div>
          <div>
            <div class="text-2xl font-bold">{{ finalResult.skipped }}</div>
            <div class="text-sm text-muted-foreground">Skipped</div>
          </div>
          <div>
            <div class="text-2xl font-bold">{{ finalResult.failed }}</div>
            <div class="text-sm text-muted-foreground">Failed</div>
          </div>
        </div>

        <div class="text-sm">
          <span class="text-muted-foreground">Destination:</span>
          <code class="ml-1 text-xs bg-muted px-1 py-0.5 rounded">{{ finalResult.destDir }}</code>
        </div>

        <Separator />

        <Button @click="reset">Link Another</Button>
      </CardContent>
    </Card>
  </div>
</template>
