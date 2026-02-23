<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import { useWebSocket } from '@/composables/useWebSocket'
import { useUpscaleStore } from '@/stores/upscale'
import { useRouter } from 'vue-router'
import type { DownloadItem, TorrentStatus, NyaaResult, TorrentProgress } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import {
  FolderOpen,
  FileVideo,
  Search,
  Download,
  RefreshCw,
  Link,
  Loader2,
  HardDrive,
  Wifi,
  WifiOff,
  Plus,
  MoreVertical,
  Trash2,
  X,
  Sparkles,
} from 'lucide-vue-next'

const api = useApi()
const router = useRouter()
const { connected, connect, on } = useWebSocket()
const upscaleStore = useUpscaleStore()
const activeTab = ref('local')

// Local downloads
const downloads = ref<DownloadItem[]>([])
const loadingDownloads = ref(false)

// Torrents (live via WebSocket)
const torrents = ref<TorrentStatus[]>([])
const loadingTorrents = ref(false)
const wsActive = ref(false)

// Nyaa search
const nyaaQuery = ref('')
const nyaaResults = ref<NyaaResult[]>([])
const searchingNyaa = ref(false)
const nyaaFilter = ref('')

// Magnet link input
const magnetLink = ref('')
const addingMagnet = ref(false)

// Search/filter across all tabs
const searchQuery = ref('')

// Torrent delete state
const deleteDialogOpen = ref(false)
const deleteTarget = ref<TorrentStatus | null>(null)
const deleteWithFiles = ref(false)
const deleting = ref(false)

// Upscale dialog state
const presetDialogOpen = ref(false)
const selectedFile = ref<DownloadItem | null>(null)
const selectedPreset = ref('balanced')
const submitting = ref(false)

// Preset options
const presets = [
  { value: 'fast', label: 'Fast', description: 'Quick upscale with basic enhancement. Good for batch processing.' },
  { value: 'balanced', label: 'Balanced', description: 'Best quality/speed tradeoff. Recommended for most content.' },
  { value: 'quality', label: 'Quality', description: 'Maximum quality, slower processing. For your favorites.' },
]

// Computed for 4K badge - tracks completed upscale paths
const completedUpscalePaths = computed(() => {
  const paths = new Set<string>()
  for (const job of upscaleStore.jobs) {
    if (job.status === 'completed') {
      paths.add(job.inputPath)
    }
  }
  return paths
})

// === Filtering logic ===
function normalize(s: string): string {
  return s.toLowerCase().replace(/[._\-]/g, ' ').replace(/\s+/g, ' ').trim()
}

function matchesFilter(candidate: string, query: string): boolean {
  if (!query) return true
  const normCandidate = normalize(candidate)
  const tokens = normalize(query).split(' ').filter(Boolean)
  return tokens.every(token => normCandidate.includes(token))
}

const filteredDownloads = computed(() => {
  if (!searchQuery.value) return downloads.value
  return downloads.value.filter(d => matchesFilter(d.name, searchQuery.value))
})

const filteredTorrents = computed(() => {
  if (!searchQuery.value) return torrents.value
  return torrents.value.filter(t => matchesFilter(t.name, searchQuery.value))
})

const filteredNyaaResults = computed(() => {
  if (!searchQuery.value) return nyaaResults.value
  return nyaaResults.value.filter(r => matchesFilter(r.title, searchQuery.value))
})

onMounted(() => {
  loadDownloads()

  // Connect WebSocket for live torrent updates
  connect()

  // Listen for torrent progress broadcasts from the download monitor
  on('torrent_progress', (data: unknown) => {
    const progress = data as TorrentProgress
    if (progress.torrents) {
      torrents.value = progress.torrents
      wsActive.value = true
    }

    // Show toast for newly completed downloads
    if (progress.completed && progress.completed.length > 0) {
      for (const t of progress.completed) {
        toast.success('Download Complete', {
          description: `${t.name} (${formatSize(t.size)})`,
        })
      }
    }
  })

  // Also listen for individual download_complete events
  on('download_complete', (_data: unknown) => {
    // Refresh local files when a download completes
    loadDownloads()
  })

  // Initialize upscale store
  upscaleStore.probe()
  upscaleStore.fetchJobs()
  upscaleStore.setupListeners()
})

// When switching to torrents tab, do an initial API fetch if no WS data yet
watch(activeTab, (val) => {
  if (val === 'torrents' && torrents.value.length === 0 && !wsActive.value) {
    loadTorrents()
  }
})

async function loadDownloads() {
  loadingDownloads.value = true
  try {
    downloads.value = await api.getDownloads()
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to load downloads')
  } finally {
    loadingDownloads.value = false
  }
}

async function loadTorrents() {
  loadingTorrents.value = true
  try {
    torrents.value = await api.getQbitTorrents()
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to load torrents')
  } finally {
    loadingTorrents.value = false
  }
}

async function searchNyaa() {
  if (!nyaaQuery.value) return
  searchingNyaa.value = true
  try {
    nyaaResults.value = await api.searchNyaa(nyaaQuery.value, nyaaFilter.value || undefined)
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Nyaa search failed')
  } finally {
    searchingNyaa.value = false
  }
}

async function addTorrent(magnet: string) {
  try {
    await api.addQbitTorrent(magnet)
    toast.success('Torrent added to qBittorrent')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to add torrent')
  }
}

async function addMagnet() {
  const magnet = magnetLink.value.trim()
  if (!magnet) return
  addingMagnet.value = true
  try {
    await api.addQbitTorrent(magnet)
    toast.success('Torrent added to qBittorrent')
    magnetLink.value = ''
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to add torrent')
  } finally {
    addingMagnet.value = false
  }
}

// === Torrent delete ===
function promptDelete(torrent: TorrentStatus, withFiles: boolean) {
  deleteTarget.value = torrent
  deleteWithFiles.value = withFiles
  deleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await api.deleteQbitTorrent(deleteTarget.value.hash, deleteWithFiles.value)
    toast.success(deleteWithFiles.value ? 'Torrent and files removed' : 'Torrent removed')
    // Remove from local list immediately
    torrents.value = torrents.value.filter(t => t.hash !== deleteTarget.value!.hash)
    // Refresh via API if WS not active
    if (!wsActive.value) {
      loadTorrents()
    }
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to delete torrent')
  } finally {
    deleting.value = false
    deleteDialogOpen.value = false
    deleteTarget.value = null
  }
}

function goToLink(name: string) {
  router.push({ path: '/link', query: { source: name } })
}

function formatSize(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(2) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  return (bytes / 1024).toFixed(1) + ' KB'
}

function formatSpeed(bytes: number): string {
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB/s'
  if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB/s'
  return bytes + ' B/s'
}

function formatEta(seconds: number): string {
  if (seconds <= 0 || seconds >= 8640000) return ''
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

function torrentStateLabel(state: string): string {
  const map: Record<string, string> = {
    downloading: 'Downloading',
    stalledDL: 'Stalled',
    uploading: 'Seeding',
    stalledUP: 'Seeding',
    pausedDL: 'Paused',
    pausedUP: 'Completed',
    queuedDL: 'Queued',
    checkingDL: 'Checking',
    error: 'Error',
  }
  return map[state] || state
}

function torrentStateVariant(state: string): 'default' | 'secondary' | 'outline' | 'destructive' {
  if (state === 'downloading') return 'default'
  if (state === 'error') return 'destructive'
  if (state.includes('UP') || state === 'uploading') return 'secondary'
  return 'outline'
}

// === Upscale functions ===
function openUpscaleDialog(item: DownloadItem) {
  selectedFile.value = item
  selectedPreset.value = 'balanced'
  presetDialogOpen.value = true
}

async function submitUpscale() {
  if (!selectedFile.value || !selectedPreset.value) return
  submitting.value = true
  try {
    await upscaleStore.createJob(selectedFile.value.path, selectedPreset.value)
    toast.success('Job queued', { description: selectedFile.value.name })
    presetDialogOpen.value = false
    activeTab.value = 'upscale'
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to queue upscale')
  } finally {
    submitting.value = false
  }
}

function hasUpscaled(item: DownloadItem): boolean {
  return completedUpscalePaths.value.has(item.path)
}

// === Upscale queue helper functions ===
function extractFileName(path: string): string {
  return path.split('/').pop() || path
}

function statusVariant(status: string): 'default' | 'secondary' | 'outline' | 'destructive' {
  switch (status) {
    case 'running': return 'default'
    case 'pending': return 'outline'
    case 'completed': return 'secondary'
    case 'failed': return 'destructive'
    case 'cancelled': return 'outline'
    default: return 'outline'
  }
}

function statusLabel(status: string): string {
  return status.charAt(0).toUpperCase() + status.slice(1)
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
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-bold">Downloads</h1>
      <p class="text-muted-foreground">Manage downloads, search Nyaa, and monitor torrents</p>
    </div>

    <Tabs v-model="activeTab">
      <div class="flex flex-col sm:flex-row sm:items-center gap-3">
        <TabsList>
          <TabsTrigger value="local" class="gap-2">
            <HardDrive class="h-4 w-4" />
            <span class="hidden sm:inline">Local Files</span>
            <span class="sm:hidden">Local</span>
          </TabsTrigger>
          <TabsTrigger value="torrents" class="gap-2">
            <Download class="h-4 w-4" />
            Torrents
            <span
              v-if="wsActive"
              class="ml-1 h-2 w-2 rounded-full bg-green-500 animate-pulse"
              title="Live updates active"
            ></span>
          </TabsTrigger>
          <TabsTrigger value="nyaa" class="gap-2">
            <Search class="h-4 w-4" />
            Nyaa
          </TabsTrigger>
          <TabsTrigger value="upscale" class="gap-2">
            <Sparkles class="h-4 w-4" />
            <span class="hidden sm:inline">Upscale Queue</span>
            <span class="sm:hidden">Upscale</span>
            <span
              v-if="upscaleStore.runningJob"
              class="ml-1 h-2 w-2 rounded-full bg-green-500 animate-pulse"
              title="Upscale in progress"
            ></span>
          </TabsTrigger>
        </TabsList>

        <!-- Global filter input -->
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="Filter results..."
            class="pl-9 h-9"
          />
          <button
            v-if="searchQuery"
            class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
            @click="searchQuery = ''"
          >
            <X class="h-4 w-4" />
          </button>
        </div>
      </div>

      <!-- Local downloads -->
      <TabsContent value="local">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between">
            <div>
              <CardTitle>Downloaded Files</CardTitle>
              <CardDescription>Files in the download directory ready to link</CardDescription>
            </div>
            <Button variant="outline" size="sm" @click="loadDownloads" class="gap-2">
              <RefreshCw class="h-4 w-4" />
              Refresh
            </Button>
          </CardHeader>
          <CardContent>
            <div v-if="loadingDownloads" class="flex items-center gap-2 text-muted-foreground py-8 justify-center">
              <Loader2 class="h-4 w-4 animate-spin" />
            </div>
            <div v-else-if="!downloads.length" class="text-center text-muted-foreground py-8">
              No downloads found
            </div>
            <div v-else-if="!filteredDownloads.length" class="text-center text-muted-foreground py-8">
              No matches for "{{ searchQuery }}"
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="item in filteredDownloads"
                :key="item.path"
                class="flex items-center gap-3 rounded-lg border p-3"
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
                <Badge v-if="hasUpscaled(item)" variant="secondary" class="text-xs shrink-0">
                  4K
                </Badge>
                <Button
                  v-if="upscaleStore.pipelineAvailable && !item.isDir"
                  size="sm"
                  variant="outline"
                  @click="openUpscaleDialog(item)"
                  class="gap-1 shrink-0"
                >
                  <Sparkles class="h-3 w-3" />
                  Upscale
                </Button>
                <Button size="sm" variant="outline" @click="goToLink(item.name)" class="gap-1 shrink-0">
                  <Link class="h-3 w-3" />
                  Link
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <!-- Torrents (live updates via WebSocket) -->
      <TabsContent value="torrents">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between">
            <div>
              <CardTitle class="flex items-center gap-2">
                Active Torrents
                <Badge v-if="wsActive && connected" variant="outline" class="gap-1 text-xs font-normal">
                  <Wifi class="h-3 w-3 text-green-500" />
                  Live
                </Badge>
                <Badge v-else variant="outline" class="gap-1 text-xs font-normal text-muted-foreground">
                  <WifiOff class="h-3 w-3" />
                  Offline
                </Badge>
              </CardTitle>
              <CardDescription>
                {{ wsActive ? 'Auto-updating every 5 seconds' : 'Torrents from qBittorrent' }}
              </CardDescription>
            </div>
            <Button variant="outline" size="sm" @click="loadTorrents" class="gap-2">
              <RefreshCw class="h-4 w-4" />
              Refresh
            </Button>
          </CardHeader>
          <CardContent class="space-y-4">
            <form @submit.prevent="addMagnet" class="flex gap-2">
              <Input v-model="magnetLink" placeholder="Paste magnet link..." class="flex-1 font-mono text-xs" />
              <Button type="submit" :disabled="addingMagnet || !magnetLink.trim()" variant="outline" class="gap-2 shrink-0">
                <Loader2 v-if="addingMagnet" class="h-4 w-4 animate-spin" />
                <Plus v-else class="h-4 w-4" />
                Add
              </Button>
            </form>
            <div v-if="loadingTorrents && !torrents.length" class="flex items-center gap-2 text-muted-foreground py-8 justify-center">
              <Loader2 class="h-4 w-4 animate-spin" />
            </div>

            <!-- Desktop: Table layout -->
            <div v-else class="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead class="w-24">Status</TableHead>
                    <TableHead class="w-36">Progress</TableHead>
                    <TableHead class="w-24">Speed</TableHead>
                    <TableHead class="w-16">ETA</TableHead>
                    <TableHead class="w-20">Size</TableHead>
                    <TableHead class="w-10"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="t in filteredTorrents" :key="t.hash">
                    <TableCell class="font-medium max-w-sm truncate">{{ t.name }}</TableCell>
                    <TableCell>
                      <Badge :variant="torrentStateVariant(t.state)">{{ torrentStateLabel(t.state) }}</Badge>
                    </TableCell>
                    <TableCell>
                      <div class="flex items-center gap-2">
                        <Progress :model-value="t.progress * 100" class="w-20 h-2" />
                        <span class="text-xs tabular-nums">{{ (t.progress * 100).toFixed(1) }}%</span>
                      </div>
                    </TableCell>
                    <TableCell class="text-xs whitespace-nowrap tabular-nums">
                      <span v-if="t.dlSpeed > 0" class="text-green-600 dark:text-green-400">{{ formatSpeed(t.dlSpeed) }}</span>
                      <span v-else class="text-muted-foreground">-</span>
                    </TableCell>
                    <TableCell class="text-xs whitespace-nowrap tabular-nums">
                      <span v-if="t.eta > 0 && t.eta < 8640000">{{ formatEta(t.eta) }}</span>
                      <span v-else class="text-muted-foreground">-</span>
                    </TableCell>
                    <TableCell class="text-xs whitespace-nowrap">{{ formatSize(t.size) }}</TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger as-child>
                          <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                            <MoreVertical class="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem @click="promptDelete(t, false)">
                            <Trash2 class="mr-2 h-4 w-4" />
                            Remove torrent
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem class="text-destructive" @click="promptDelete(t, true)">
                            <Trash2 class="mr-2 h-4 w-4" />
                            Remove + delete files
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                  <TableRow v-if="!filteredTorrents.length && !loadingTorrents && !searchQuery">
                    <TableCell colspan="7" class="text-center text-muted-foreground py-8">
                      No torrents found. Is qBittorrent configured?
                    </TableCell>
                  </TableRow>
                  <TableRow v-if="!filteredTorrents.length && searchQuery && torrents.length">
                    <TableCell colspan="7" class="text-center text-muted-foreground py-8">
                      No matches for "{{ searchQuery }}"
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>

            <!-- Mobile: Stacked card layout -->
            <div class="md:hidden space-y-3">
              <div
                v-for="t in filteredTorrents"
                :key="'m-' + t.hash"
                class="rounded-lg border p-3 space-y-2"
              >
                <div class="flex items-start justify-between gap-2">
                  <div class="font-medium text-sm leading-tight break-all min-w-0 flex-1">{{ t.name }}</div>
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="sm" class="h-7 w-7 p-0 shrink-0">
                        <MoreVertical class="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem @click="promptDelete(t, false)">
                        <Trash2 class="mr-2 h-4 w-4" />
                        Remove torrent
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem class="text-destructive" @click="promptDelete(t, true)">
                        <Trash2 class="mr-2 h-4 w-4" />
                        Remove + delete files
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
                <div class="flex items-center gap-2">
                  <Badge :variant="torrentStateVariant(t.state)" class="text-xs">{{ torrentStateLabel(t.state) }}</Badge>
                  <span class="text-xs text-muted-foreground">{{ formatSize(t.size) }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <Progress :model-value="t.progress * 100" class="flex-1 h-2" />
                  <span class="text-xs tabular-nums shrink-0">{{ (t.progress * 100).toFixed(1) }}%</span>
                </div>
                <div class="flex items-center gap-3 text-xs text-muted-foreground">
                  <span v-if="t.dlSpeed > 0" class="text-green-600 dark:text-green-400">{{ formatSpeed(t.dlSpeed) }}</span>
                  <span v-if="t.eta > 0 && t.eta < 8640000">ETA {{ formatEta(t.eta) }}</span>
                </div>
              </div>
              <div v-if="!filteredTorrents.length && !loadingTorrents && !searchQuery" class="text-center text-muted-foreground py-8">
                No torrents found. Is qBittorrent configured?
              </div>
              <div v-if="!filteredTorrents.length && searchQuery && torrents.length" class="text-center text-muted-foreground py-8">
                No matches for "{{ searchQuery }}"
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <!-- Nyaa search -->
      <TabsContent value="nyaa">
        <Card>
          <CardHeader>
            <CardTitle>Search Nyaa</CardTitle>
            <CardDescription>Find anime torrents on Nyaa.si</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <form @submit.prevent="searchNyaa" class="flex flex-col sm:flex-row gap-2">
              <Input v-model="nyaaQuery" placeholder="Search anime..." class="flex-1" />
              <Select v-model="nyaaFilter">
                <SelectTrigger class="w-full sm:w-40">
                  <SelectValue placeholder="All" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All</SelectItem>
                  <SelectItem value="trusted">Trusted</SelectItem>
                  <SelectItem value="noremakes">No Remakes</SelectItem>
                </SelectContent>
              </Select>
              <Button type="submit" :disabled="searchingNyaa || !nyaaQuery" class="gap-2">
                <Loader2 v-if="searchingNyaa" class="h-4 w-4 animate-spin" />
                <Search v-else class="h-4 w-4" />
                Search
              </Button>
            </form>

            <!-- Desktop table -->
            <div v-if="filteredNyaaResults.length" class="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead class="w-20">Size</TableHead>
                    <TableHead class="w-16">S/L</TableHead>
                    <TableHead class="w-24"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="(r, i) in filteredNyaaResults" :key="i">
                    <TableCell class="font-medium max-w-md truncate">{{ r.title }}</TableCell>
                    <TableCell class="text-xs whitespace-nowrap">{{ r.size }}</TableCell>
                    <TableCell class="text-xs">
                      <span class="text-green-600">{{ r.seeders }}</span>/<span class="text-red-500">{{ r.leechers }}</span>
                    </TableCell>
                    <TableCell>
                      <Button size="sm" variant="outline" @click="addTorrent(r.magnet)" class="gap-1">
                        <Download class="h-3 w-3" />
                        Add
                      </Button>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>

            <!-- Mobile stacked layout -->
            <div v-if="filteredNyaaResults.length" class="md:hidden space-y-3">
              <div
                v-for="(r, i) in filteredNyaaResults"
                :key="'m-' + i"
                class="rounded-lg border p-3 space-y-2"
              >
                <div class="font-medium text-sm leading-tight break-all">{{ r.title }}</div>
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-3 text-xs text-muted-foreground">
                    <span>{{ r.size }}</span>
                    <span>
                      <span class="text-green-600">{{ r.seeders }}</span>/<span class="text-red-500">{{ r.leechers }}</span>
                    </span>
                  </div>
                  <Button size="sm" variant="outline" @click="addTorrent(r.magnet)" class="gap-1">
                    <Download class="h-3 w-3" />
                    Add
                  </Button>
                </div>
              </div>
            </div>

            <div v-if="nyaaResults.length && !filteredNyaaResults.length && searchQuery" class="text-center text-muted-foreground py-8">
              No matches for "{{ searchQuery }}"
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <!-- Upscale Queue -->
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

            <template v-else>
              <!-- Desktop: Table layout -->
              <div class="hidden md:block">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>File</TableHead>
                      <TableHead class="w-24">Preset</TableHead>
                      <TableHead class="w-24">Status</TableHead>
                      <TableHead class="w-48">Progress</TableHead>
                      <TableHead class="w-10"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <TableRow v-for="job in upscaleStore.jobs" :key="job.id">
                      <TableCell class="font-medium max-w-sm truncate">
                        {{ extractFileName(job.inputPath) }}
                      </TableCell>
                      <TableCell class="capitalize">{{ job.preset }}</TableCell>
                      <TableCell>
                        <Badge :variant="statusVariant(job.status)">
                          {{ statusLabel(job.status) }}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <template v-if="job.status === 'running' && upscaleStore.progress[job.id]">
                          <div class="flex items-center gap-2">
                            <Progress :model-value="upscaleStore.progress[job.id]?.percent ?? 0" class="w-20 h-2" />
                            <span class="text-xs tabular-nums">
                              {{ upscaleStore.progress[job.id]?.percent?.toFixed(1) ?? '0.0' }}%
                            </span>
                            <span class="text-xs text-muted-foreground tabular-nums">
                              {{ upscaleStore.progress[job.id]?.fps?.toFixed(1) ?? '0.0' }} fps
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
                      <TableCell>
                        <Button
                          v-if="job.status === 'running' || job.status === 'pending'"
                          size="sm"
                          variant="ghost"
                          @click="handleCancel(job.id)"
                          title="Cancel"
                        >
                          <X class="h-4 w-4" />
                        </Button>
                        <Button
                          v-else
                          size="sm"
                          variant="ghost"
                          @click="handleDelete(job.id)"
                          title="Remove"
                        >
                          <Trash2 class="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </div>

              <!-- Mobile: Stacked card layout -->
              <div class="md:hidden space-y-3">
                <div
                  v-for="job in upscaleStore.jobs"
                  :key="'m-' + job.id"
                  class="rounded-lg border p-3 space-y-2"
                >
                  <div class="flex items-start justify-between gap-2">
                    <div class="font-medium text-sm leading-tight break-all min-w-0 flex-1">
                      {{ extractFileName(job.inputPath) }}
                    </div>
                    <Button
                      v-if="job.status === 'running' || job.status === 'pending'"
                      size="sm"
                      variant="ghost"
                      class="h-7 w-7 p-0 shrink-0"
                      @click="handleCancel(job.id)"
                    >
                      <X class="h-4 w-4" />
                    </Button>
                    <Button
                      v-else
                      size="sm"
                      variant="ghost"
                      class="h-7 w-7 p-0 shrink-0"
                      @click="handleDelete(job.id)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </Button>
                  </div>
                  <div class="flex items-center gap-2">
                    <Badge :variant="statusVariant(job.status)" class="text-xs">
                      {{ statusLabel(job.status) }}
                    </Badge>
                    <span class="text-xs text-muted-foreground capitalize">{{ job.preset }}</span>
                  </div>
                  <template v-if="job.status === 'running' && upscaleStore.progress[job.id]">
                    <div class="flex items-center gap-2">
                      <Progress :model-value="upscaleStore.progress[job.id]?.percent ?? 0" class="flex-1 h-2" />
                      <span class="text-xs tabular-nums shrink-0">
                        {{ upscaleStore.progress[job.id]?.percent?.toFixed(1) ?? '0.0' }}%
                      </span>
                    </div>
                    <div class="text-xs text-muted-foreground">
                      {{ upscaleStore.progress[job.id]?.fps?.toFixed(1) ?? '0.0' }} fps
                    </div>
                  </template>
                  <div v-else-if="job.status === 'failed'" class="text-xs text-destructive truncate">
                    {{ job.error }}
                  </div>
                </div>
              </div>
            </template>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <!-- Delete confirmation dialog -->
    <AlertDialog v-model:open="deleteDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ deleteWithFiles ? 'Remove torrent and delete files?' : 'Remove torrent?' }}</AlertDialogTitle>
          <AlertDialogDescription>
            <template v-if="deleteWithFiles">
              This will remove the torrent from qBittorrent <strong>and permanently delete all downloaded files</strong>.
              This action cannot be undone.
            </template>
            <template v-else>
              This will remove the torrent from qBittorrent. Downloaded files will be kept on disk.
            </template>
            <div v-if="deleteTarget" class="mt-2 p-2 rounded bg-muted text-sm font-mono truncate">
              {{ deleteTarget.name }}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="deleting">Cancel</AlertDialogCancel>
          <AlertDialogAction
            :disabled="deleting"
            :class="deleteWithFiles ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90' : ''"
            @click.prevent="confirmDelete"
          >
            <Loader2 v-if="deleting" class="mr-2 h-4 w-4 animate-spin" />
            {{ deleteWithFiles ? 'Delete' : 'Remove' }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- Upscale preset picker dialog -->
    <Dialog v-model:open="presetDialogOpen">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Upscale Quality</DialogTitle>
          <DialogDescription v-if="selectedFile">
            Choose upscaling preset for {{ selectedFile.name }}
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-3 py-4">
          <label
            v-for="preset in presets"
            :key="preset.value"
            class="flex items-start gap-3 rounded-lg border p-3 cursor-pointer transition-colors"
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
  </div>
</template>
