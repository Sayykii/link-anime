<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import { useWebSocket } from '@/composables/useWebSocket'
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
} from 'lucide-vue-next'

const api = useApi()
const router = useRouter()
const { connected, connect, on } = useWebSocket()
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
  on('download_complete', (data: unknown) => {
    const t = data as TorrentStatus
    // Refresh local files when a download completes
    loadDownloads()
  })
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
    nyaaResults.value = await api.searchNyaa(nyaaQuery.value)
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

function goToLink(name: string) {
  router.push('/link')
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
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-bold">Downloads</h1>
      <p class="text-muted-foreground">Manage downloads, search Nyaa, and monitor torrents</p>
    </div>

    <Tabs v-model="activeTab">
      <TabsList>
        <TabsTrigger value="local" class="gap-2">
          <HardDrive class="h-4 w-4" />
          Local Files
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
          Nyaa Search
        </TabsTrigger>
      </TabsList>

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
            <div v-else class="space-y-2">
              <div
                v-for="item in downloads"
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
          <CardContent class="p-0">
            <div v-if="loadingTorrents && !torrents.length" class="flex items-center gap-2 text-muted-foreground py-8 justify-center">
              <Loader2 class="h-4 w-4 animate-spin" />
            </div>
            <Table v-else>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead class="w-24">Status</TableHead>
                  <TableHead class="w-36">Progress</TableHead>
                  <TableHead class="w-24">Speed</TableHead>
                  <TableHead class="w-16">ETA</TableHead>
                  <TableHead class="w-20">Size</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="t in torrents" :key="t.hash">
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
                </TableRow>
                <TableRow v-if="!torrents.length && !loadingTorrents">
                  <TableCell colspan="6" class="text-center text-muted-foreground py-8">
                    No torrents found. Is qBittorrent configured?
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
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
            <form @submit.prevent="searchNyaa" class="flex gap-2">
              <Input v-model="nyaaQuery" placeholder="Search anime..." class="flex-1" />
              <Button type="submit" :disabled="searchingNyaa || !nyaaQuery" class="gap-2">
                <Loader2 v-if="searchingNyaa" class="h-4 w-4 animate-spin" />
                <Search v-else class="h-4 w-4" />
                Search
              </Button>
            </form>

            <Table v-if="nyaaResults.length">
              <TableHeader>
                <TableRow>
                  <TableHead>Title</TableHead>
                  <TableHead class="w-20">Size</TableHead>
                  <TableHead class="w-16">S/L</TableHead>
                  <TableHead class="w-24"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="(r, i) in nyaaResults" :key="i">
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
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</template>
