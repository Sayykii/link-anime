<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { seriesPosterUrl } from '@/lib/utils'
import type { ShokoSeries } from '@/lib/types'
import { useLibraryStore } from '@/stores/library'
import type { UnlinkPreview } from '@/lib/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Search, RefreshCw, Tv, Film, Loader2, X, Trash2, MoreVertical } from 'lucide-vue-next'
import EmptyState from '@/components/EmptyState.vue'
import FileSafetyWarning from '@/components/FileSafetyWarning.vue'
import { toast } from 'vue-sonner'

const api = useApi()
const router = useRouter()
const library = useLibraryStore()

const searchQuery = ref('')
const activeTab = ref('shows')
const loading = ref(true)
const shokoAvailable = ref(false)
const shokoSeries = ref<ShokoSeries[]>([])
const shokoTotal = ref(0)

// Pagination
const shokoPage = ref(1)
const shokoPageSize = 50
const hasMoreSeries = computed(() => shokoSeries.value.length < shokoTotal.value)
const loadingMore = ref(false)

// Folder map for movie posters
const folderMap = ref<Record<string, { shokoId: number; name: string; posterUrl?: string }>>({})

// Unlink state (for filesystem fallback)
const unlinkDialogOpen = ref(false)
const unlinkTarget = ref<{ name: string; path: string; type: 'show' | 'season' | 'movie' }>()
const unlinkPreview = ref<UnlinkPreview | null>(null)
const unlinkLoading = ref(false)
const unlinkExecuting = ref(false)

onMounted(async () => {
  // Try Shoko first
  try {
    const data = await api.shokoSeries(1, shokoPageSize)
    shokoSeries.value = data.series || []
    shokoTotal.value = data.total
    shokoAvailable.value = true

    // Load folder map for movie posters
    try {
      folderMap.value = await api.shokoFolderMap()
    } catch { /* ignore */ }
  } catch {
    shokoAvailable.value = false
  }

  // Always load filesystem data as fallback for movies and unlink
  library.fetchShows()
  library.fetchMovies()
  loading.value = false
})

const showPathMap = computed(() => {
  const map: Record<string, string> = {}
  for (const show of library.shows) {
    map[show.name.toLowerCase()] = show.path
  }
  return map
})

function getShowPath(seriesName: string): string | null {
  return showPathMap.value[seriesName.toLowerCase()] ?? null
}

const filteredShokoSeries = computed(() => {
  if (!searchQuery.value) return shokoSeries.value
  const q = searchQuery.value.toLowerCase()
  return shokoSeries.value.filter(s =>
    s.Name.toLowerCase().includes(q) ||
    s.AniDB?.Title?.toLowerCase().includes(q)
  )
})

const filteredShows = computed(() => {
  if (!searchQuery.value) return library.shows
  const q = searchQuery.value.toLowerCase()
  return library.shows.filter(s => s.name.toLowerCase().includes(q))
})

const filteredMovies = computed(() => {
  if (!searchQuery.value) return library.movies
  const q = searchQuery.value.toLowerCase()
  return library.movies.filter(m => m.name.toLowerCase().includes(q))
})

function refresh() {
  loading.value = true
  shokoPage.value = 1
  const promises: Promise<void>[] = []
  if (shokoAvailable.value) {
    promises.push(
      api.shokoSeries(1, shokoPageSize).then(data => {
        shokoSeries.value = data.series || []
        shokoTotal.value = data.total
      })
    )
  }
  promises.push(library.fetchShows() as any, library.fetchMovies() as any)
  Promise.all(promises).finally(() => { loading.value = false })
}

async function loadMore() {
  loadingMore.value = true
  try {
    shokoPage.value++
    const data = await api.shokoSeries(shokoPage.value, shokoPageSize)
    shokoSeries.value.push(...(data.series || []))
    shokoTotal.value = data.total
  } catch (err: any) {
    toast.error('Failed to load more', { description: err.message })
    shokoPage.value--
  } finally {
    loadingMore.value = false
  }
}

function moviePosterUrl(movieName: string): string | null {
  const entry = folderMap.value[movieName]
  return entry?.posterUrl ?? null
}

// Unlink helpers (same as before for filesystem fallback)
const hasUnsafeFiles = computed(() => {
  return unlinkPreview.value && unlinkPreview.value.unsafeFiles && unlinkPreview.value.unsafeFiles.length > 0
})

async function openUnlinkDialog(name: string, path: string, type: 'show' | 'season' | 'movie') {
  unlinkTarget.value = { name, path, type }
  unlinkPreview.value = null
  unlinkLoading.value = true
  unlinkDialogOpen.value = true
  try {
    unlinkPreview.value = await api.unlinkPreview(path)
  } catch (err: any) {
    toast.error('Failed to check files', { description: err.message })
    unlinkDialogOpen.value = false
  } finally {
    unlinkLoading.value = false
  }
}

async function executeUnlink(force: boolean) {
  if (!unlinkTarget.value) return
  unlinkExecuting.value = true
  try {
    const result = await api.unlink(unlinkTarget.value.path, force)
    const removed = result.linked
    const skipped = result.skipped
    if (removed > 0) {
      toast.success(`Unlinked: ${unlinkTarget.value.name}`, {
        description: `Removed ${removed} file${removed !== 1 ? 's' : ''}${skipped > 0 ? `, skipped ${skipped} unsafe` : ''}`,
      })
    } else if (skipped > 0) {
      toast.warning('No files removed', {
        description: `${skipped} file${skipped !== 1 ? 's' : ''} skipped (only copy, no source)`,
      })
    }
    unlinkDialogOpen.value = false
    refresh()
  } catch (err: any) {
    toast.error('Unlink failed', { description: err.message })
  } finally {
    unlinkExecuting.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-3xl tracking-wider uppercase">Library</h1>
        <p class="text-muted-foreground text-sm mt-1">Browse your anime collection</p>
      </div>
      <Button variant="outline" size="sm" @click="refresh" class="gap-2">
        <RefreshCw class="h-4 w-4" />
        Refresh
      </Button>
    </div>

    <!-- Search -->
    <div class="relative max-w-sm">
      <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input v-model="searchQuery" placeholder="Search library..." class="pl-9 h-9" />
      <button
        v-if="searchQuery"
        class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
        @click="searchQuery = ''"
      >
        <X class="h-4 w-4" />
      </button>
    </div>

    <!-- Loading skeleton grid -->
    <div v-if="loading" class="space-y-4">
      <div class="flex gap-2">
        <Skeleton class="h-9 w-28 rounded-md" />
        <Skeleton class="h-9 w-28 rounded-md" />
      </div>
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
        <div v-for="i in 12" :key="i" class="space-y-2">
          <Skeleton class="aspect-[2/3] rounded-lg w-full" />
          <Skeleton class="h-4 w-3/4" />
          <Skeleton class="h-3 w-1/2" />
        </div>
      </div>
    </div>

    <!-- Shoko-powered card grid -->
    <template v-else-if="shokoAvailable">
      <Tabs v-model="activeTab">
        <TabsList>
          <TabsTrigger value="shows" class="gap-2">
            <Tv class="h-4 w-4" />
            Series ({{ shokoTotal }})
          </TabsTrigger>
          <TabsTrigger value="movies" class="gap-2">
            <Film class="h-4 w-4" />
            Movies ({{ library.movies.length }})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="shows">
          <div v-if="filteredShokoSeries.length === 0" class="py-12">
            <EmptyState
              v-if="searchQuery"
              :icon="Search"
              :heading="`No results for &quot;${searchQuery}&quot;`"
              action-label="Clear filter"
              action-variant="outline"
              @action="searchQuery = ''"
            />
            <EmptyState
              v-else
              :icon="Tv"
              heading="No shows yet"
              description="Link anime from your downloads to start building your library"
              action-label="Link New Content"
              @action="router.push('/link')"
            />
          </div>

          <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            <div
              v-for="series in filteredShokoSeries"
              :key="series.IDs.ID"
              class="group text-left cursor-pointer"
              @click="router.push(`/library/series/${series.IDs.ID}`)"
            >
              <!-- Poster -->
              <div class="relative aspect-[2/3] rounded-lg overflow-hidden bg-muted mb-2 ring-1 ring-border/50 group-hover:ring-primary/50 transition-all">
                <img
                  v-if="seriesPosterUrl(series.Images)"
                  :src="seriesPosterUrl(series.Images)!"
                  :alt="series.Name"
                  class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                  loading="lazy"
                />
                <div v-else class="w-full h-full flex items-center justify-center">
                  <Tv class="h-10 w-10 text-muted-foreground/30" />
                </div>

                <!-- Episode count overlay -->
                <div class="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/80 via-black/40 to-transparent p-2 pt-6">
                  <div class="flex items-center gap-1.5">
                    <Badge
                      :variant="(series.Sizes?.Missing?.Episodes ?? 0) > 0 ? 'destructive' : 'secondary'"
                      class="text-xs"
                    >
                      {{ series.Sizes?.Local?.Episodes ?? series.Size }}/{{ series.Sizes?.Total?.Episodes || '?' }}
                    </Badge>
                    <Badge v-if="(series.Sizes?.Watched?.Episodes ?? 0) > 0" variant="outline" class="text-xs text-white border-white/30">
                      {{ series.Sizes.Watched.Episodes }} seen
                    </Badge>
                  </div>
                </div>

                <!-- Rating badge -->
                <div v-if="series.AniDB?.Rating" class="absolute top-2 right-2">
                  <Badge variant="secondary" class="text-xs font-mono">
                    {{ (series.AniDB.Rating.Value / (series.AniDB.Rating.MaxValue / 10)).toFixed(1) }}
                  </Badge>
                </div>

                <!-- Context menu -->
                <div v-if="getShowPath(series.Name)" class="absolute top-2 left-2 opacity-0 group-hover:opacity-100 transition-opacity" @click.stop>
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="secondary" size="icon" class="h-7 w-7 rounded-full shadow-md">
                        <MoreVertical class="h-3.5 w-3.5" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="start">
                      <DropdownMenuItem
                        class="text-destructive"
                        @click="openUnlinkDialog(series.Name, getShowPath(series.Name)!, 'show')"
                      >
                        <Trash2 class="h-4 w-4 mr-2" />
                        Unlink
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>

              <!-- Title -->
              <p class="text-sm font-medium leading-tight line-clamp-2 min-h-[2.5rem] group-hover:text-primary transition-colors">
                {{ series.Name }}
              </p>
              <p v-if="series.AniDB?.Type" class="text-xs text-muted-foreground mt-0.5">{{ series.AniDB.Type }}</p>
            </div>
          </div>

          <!-- Load More -->
          <div v-if="hasMoreSeries && !searchQuery" class="flex justify-center pt-4">
            <Button variant="outline" @click="loadMore" :disabled="loadingMore" class="gap-2">
              <Loader2 v-if="loadingMore" class="h-4 w-4 animate-spin" />
              Load More ({{ shokoSeries.length }}/{{ shokoTotal }})
            </Button>
          </div>
        </TabsContent>

        <!-- Movies tab with poster grid -->
        <TabsContent value="movies">
          <div v-if="filteredMovies.length === 0" class="py-12">
            <EmptyState
              v-if="searchQuery"
              :icon="Search"
              :heading="`No results for &quot;${searchQuery}&quot;`"
              action-label="Clear filter"
              action-variant="outline"
              @action="searchQuery = ''"
            />
            <EmptyState
              v-else
              :icon="Film"
              heading="No movies yet"
              description="Link anime movies from your downloads"
              action-label="Link New Content"
              @action="router.push('/link')"
            />
          </div>

          <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            <div
              v-for="movie in filteredMovies"
              :key="movie.path"
              class="group text-left"
            >
              <div class="relative aspect-[2/3] rounded-lg overflow-hidden bg-muted mb-2 ring-1 ring-border/50">
                <img
                  v-if="moviePosterUrl(movie.name)"
                  :src="moviePosterUrl(movie.name)!"
                  :alt="movie.name"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
                <div v-else class="w-full h-full flex items-center justify-center">
                  <Film class="h-10 w-10 text-muted-foreground/30" />
                </div>
                <div class="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/80 via-black/40 to-transparent p-2 pt-6">
                  <Badge variant="secondary" class="text-xs">
                    {{ movie.files }} file{{ movie.files !== 1 ? 's' : '' }}
                  </Badge>
                </div>
              </div>
              <div class="flex items-start justify-between gap-1">
                <p class="text-sm font-medium leading-tight line-clamp-2 min-h-[2.5rem]">{{ movie.name }}</p>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-6 w-6 shrink-0 text-muted-foreground hover:text-destructive"
                  @click="openUnlinkDialog(movie.name, movie.path, 'movie')"
                  title="Unlink movie"
                >
                  <Trash2 class="h-3 w-3" />
                </Button>
              </div>
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </template>

    <!-- Filesystem fallback (when Shoko is not available) -->
    <template v-else>
      <Tabs v-model="activeTab">
        <TabsList>
          <TabsTrigger value="shows" class="gap-2">
            <Tv class="h-4 w-4" />
            Shows ({{ library.shows.length }})
          </TabsTrigger>
          <TabsTrigger value="movies" class="gap-2">
            <Film class="h-4 w-4" />
            Movies ({{ library.movies.length }})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="shows">
          <Card glass>
            <CardContent class="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Show</TableHead>
                    <TableHead class="w-32">Seasons</TableHead>
                    <TableHead class="w-32">Episodes</TableHead>
                    <TableHead class="w-24 text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="show in filteredShows" :key="show.path">
                    <TableCell class="font-medium">
                      {{ show.name }}
                      <div v-if="show.seasons.length" class="mt-1 flex flex-wrap gap-1">
                        <Badge
                          v-for="season in show.seasons"
                          :key="season.number"
                          variant="secondary"
                          class="text-xs cursor-pointer hover:bg-destructive/20 transition-colors"
                          @click="openUnlinkDialog(`${show.name} - Season ${season.number}`, season.path, 'season')"
                          :title="`Click to unlink Season ${season.number}`"
                        >
                          S{{ season.number }} ({{ season.episodes }})
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>{{ show.seasons.length }}</TableCell>
                    <TableCell>{{ show.episodes }}</TableCell>
                    <TableCell class="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        class="h-8 w-8 text-muted-foreground hover:text-destructive"
                        @click="openUnlinkDialog(show.name, show.path, 'show')"
                        title="Unlink entire show"
                      >
                        <Trash2 class="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                  <TableRow v-if="!filteredShows.length">
                    <TableCell colspan="4">
                      <EmptyState
                        :icon="Tv"
                        heading="No shows yet"
                        description="Link anime from your downloads to start building your library"
                        action-label="Link New Content"
                        @action="router.push('/link')"
                      />
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="movies">
          <Card glass>
            <CardContent class="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Movie</TableHead>
                    <TableHead class="w-32">Files</TableHead>
                    <TableHead class="w-24 text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="movie in filteredMovies" :key="movie.path">
                    <TableCell class="font-medium">{{ movie.name }}</TableCell>
                    <TableCell>{{ movie.files }}</TableCell>
                    <TableCell class="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        class="h-8 w-8 text-muted-foreground hover:text-destructive"
                        @click="openUnlinkDialog(movie.name, movie.path, 'movie')"
                        title="Unlink movie"
                      >
                        <Trash2 class="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                  <TableRow v-if="!filteredMovies.length">
                    <TableCell colspan="3">
                      <EmptyState
                        :icon="Film"
                        heading="No movies yet"
                        description="Link anime movies from your downloads"
                        action-label="Link New Content"
                        @action="router.push('/link')"
                      />
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </template>

    <!-- Unlink confirmation dialog -->
    <AlertDialog v-model:open="unlinkDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Unlink {{ unlinkTarget?.name }}?</AlertDialogTitle>
          <AlertDialogDescription v-if="unlinkLoading" class="flex items-center gap-2">
            <Loader2 class="h-4 w-4 animate-spin" />
            Checking file safety...
          </AlertDialogDescription>
          <AlertDialogDescription v-else-if="unlinkPreview" as="div">
            <FileSafetyWarning :preview="unlinkPreview" />
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter v-if="!unlinkLoading && unlinkPreview">
          <AlertDialogCancel :disabled="unlinkExecuting">Cancel</AlertDialogCancel>
          <template v-if="hasUnsafeFiles">
            <Button
              v-if="unlinkPreview!.safeFiles && unlinkPreview!.safeFiles.length > 0"
              variant="outline"
              @click="executeUnlink(false)"
              :disabled="unlinkExecuting"
              class="gap-2"
            >
              <Loader2 v-if="unlinkExecuting" class="h-4 w-4 animate-spin" />
              Remove safe only
            </Button>
            <AlertDialogAction
              @click.prevent="executeUnlink(true)"
              :disabled="unlinkExecuting"
              class="bg-destructive text-destructive-foreground hover:bg-destructive/90 gap-2"
            >
              <Loader2 v-if="unlinkExecuting" class="h-4 w-4 animate-spin" />
              Remove all
            </AlertDialogAction>
          </template>
          <AlertDialogAction
            v-else
            @click.prevent="executeUnlink(false)"
            :disabled="unlinkExecuting"
            class="bg-destructive text-destructive-foreground hover:bg-destructive/90 gap-2"
          >
            <Loader2 v-if="unlinkExecuting" class="h-4 w-4 animate-spin" />
            Remove
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
