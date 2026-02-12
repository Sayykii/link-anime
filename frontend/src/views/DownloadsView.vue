<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useApi } from '@/composables/useApi'
import { useRouter } from 'vue-router'
import type { DownloadItem, TorrentStatus, NyaaResult } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
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
  ExternalLink,
} from 'lucide-vue-next'

const api = useApi()
const router = useRouter()
const activeTab = ref('local')

// Local downloads
const downloads = ref<DownloadItem[]>([])
const loadingDownloads = ref(false)

// Torrents
const torrents = ref<TorrentStatus[]>([])
const loadingTorrents = ref(false)

// Nyaa search
const nyaaQuery = ref('')
const nyaaResults = ref<NyaaResult[]>([])
const searchingNyaa = ref(false)

onMounted(() => {
  loadDownloads()
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

    <Tabs v-model="activeTab" @update:model-value="(v) => { if (v === 'torrents') loadTorrents() }">
      <TabsList>
        <TabsTrigger value="local" class="gap-2">
          <HardDrive class="h-4 w-4" />
          Local Files
        </TabsTrigger>
        <TabsTrigger value="torrents" class="gap-2">
          <Download class="h-4 w-4" />
          Torrents
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

      <!-- Torrents -->
      <TabsContent value="torrents">
        <Card>
          <CardHeader class="flex flex-row items-center justify-between">
            <div>
              <CardTitle>Active Torrents</CardTitle>
              <CardDescription>Torrents from qBittorrent</CardDescription>
            </div>
            <Button variant="outline" size="sm" @click="loadTorrents" class="gap-2">
              <RefreshCw class="h-4 w-4" />
              Refresh
            </Button>
          </CardHeader>
          <CardContent class="p-0">
            <div v-if="loadingTorrents" class="flex items-center gap-2 text-muted-foreground py-8 justify-center">
              <Loader2 class="h-4 w-4 animate-spin" />
            </div>
            <Table v-else>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead class="w-24">Status</TableHead>
                  <TableHead class="w-28">Progress</TableHead>
                  <TableHead class="w-24">Speed</TableHead>
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
                      <Progress :model-value="t.progress * 100" class="w-16 h-2" />
                      <span class="text-xs">{{ (t.progress * 100).toFixed(0) }}%</span>
                    </div>
                  </TableCell>
                  <TableCell class="text-xs whitespace-nowrap">
                    <span v-if="t.dlSpeed > 0">{{ formatSpeed(t.dlSpeed) }}</span>
                    <span v-else class="text-muted-foreground">-</span>
                  </TableCell>
                  <TableCell class="text-xs whitespace-nowrap">{{ formatSize(t.size) }}</TableCell>
                </TableRow>
                <TableRow v-if="!torrents.length && !loadingTorrents">
                  <TableCell colspan="5" class="text-center text-muted-foreground py-8">
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
