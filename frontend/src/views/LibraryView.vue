<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useLibraryStore } from '@/stores/library'
import { useApi } from '@/composables/useApi'
import type { UnlinkPreview } from '@/lib/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Search, RefreshCw, Tv, Film, Trash2, Loader2, AlertTriangle, X, ArrowUpDown } from 'lucide-vue-next'
import EmptyState from '@/components/EmptyState.vue'
import { toast } from 'vue-sonner'

const api = useApi()
const router = useRouter()
const library = useLibraryStore()
const searchQuery = ref('')
const activeTab = ref('shows')
const sortBy = ref('name-asc')

// Unlink state
const unlinkDialogOpen = ref(false)
const unlinkTarget = ref<{ name: string; path: string; type: 'show' | 'season' | 'movie' }>()
const unlinkPreview = ref<UnlinkPreview | null>(null)
const unlinkLoading = ref(false)
const unlinkExecuting = ref(false)

onMounted(() => {
  library.fetchShows()
  library.fetchMovies()
})

function sortItems<T extends { name: string }>(items: T[], sort: string, getEpisodes?: (i: T) => number, getSeasons?: (i: T) => number): T[] {
  const sorted = [...items]
  switch (sort) {
    case 'name-asc': sorted.sort((a, b) => a.name.localeCompare(b.name)); break
    case 'name-desc': sorted.sort((a, b) => b.name.localeCompare(a.name)); break
    case 'episodes-desc': if (getEpisodes) sorted.sort((a, b) => getEpisodes(b) - getEpisodes(a)); break
    case 'episodes-asc': if (getEpisodes) sorted.sort((a, b) => getEpisodes(a) - getEpisodes(b)); break
    case 'seasons-desc': if (getSeasons) sorted.sort((a, b) => getSeasons(b) - getSeasons(a)); break
    case 'seasons-asc': if (getSeasons) sorted.sort((a, b) => getSeasons(a) - getSeasons(b)); break
  }
  return sorted
}

const filteredShows = computed(() => {
  let items = library.shows
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    items = items.filter(s => s.name.toLowerCase().includes(q))
  }
  return sortItems(items, sortBy.value, s => s.episodes, s => s.seasons.length)
})

const filteredMovies = computed(() => {
  let items = library.movies
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    items = items.filter(m => m.name.toLowerCase().includes(q))
  }
  return sortItems(items, sortBy.value)
})

function refresh() {
  library.fetchShows()
  library.fetchMovies()
}

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

const hasUnsafeFiles = computed(() => {
  return unlinkPreview.value && unlinkPreview.value.unsafeFiles && unlinkPreview.value.unsafeFiles.length > 0
})

async function executeUnlink(force: boolean) {
  if (!unlinkTarget.value) return
  unlinkExecuting.value = true

  try {
    const result = await api.unlink(unlinkTarget.value.path, force)
    const removed = result.linked // reused as removed count
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
        <h1 class="text-3xl font-bold">Library</h1>
        <p class="text-muted-foreground">Browse your anime collection</p>
      </div>
      <Button variant="outline" size="sm" @click="refresh" class="gap-2">
        <RefreshCw class="h-4 w-4" />
        Refresh
      </Button>
    </div>

    <!-- Filter bar -->
    <div class="sticky-filter flex flex-col sm:flex-row sm:items-center gap-3">
      <div class="relative flex-1 max-w-sm">
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
      <Select v-model="sortBy">
        <SelectTrigger class="w-44 h-9">
          <ArrowUpDown class="h-3.5 w-3.5 mr-1.5 text-muted-foreground" />
          <SelectValue placeholder="Sort by..." />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="name-asc">Name A-Z</SelectItem>
          <SelectItem value="name-desc">Name Z-A</SelectItem>
          <SelectItem value="episodes-desc">Episodes (most)</SelectItem>
          <SelectItem value="episodes-asc">Episodes (fewest)</SelectItem>
          <SelectItem value="seasons-desc">Seasons (most)</SelectItem>
          <SelectItem value="seasons-asc">Seasons (fewest)</SelectItem>
        </SelectContent>
      </Select>
    </div>

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
                <TableRow v-if="!filteredShows.length && searchQuery">
                  <TableCell colspan="4">
                    <EmptyState
                      :icon="Search"
                      :heading="`No results for &quot;${searchQuery}&quot;`"
                      action-label="Clear filter"
                      action-variant="outline"
                      @action="searchQuery = ''"
                    />
                  </TableCell>
                </TableRow>
                <TableRow v-if="!filteredShows.length && !searchQuery">
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
                <TableRow v-if="!filteredMovies.length && searchQuery">
                  <TableCell colspan="3">
                    <EmptyState
                      :icon="Search"
                      :heading="`No results for &quot;${searchQuery}&quot;`"
                      action-label="Clear filter"
                      action-variant="outline"
                      @action="searchQuery = ''"
                    />
                  </TableCell>
                </TableRow>
                <TableRow v-if="!filteredMovies.length && !searchQuery">
                  <TableCell colspan="3">
                    <EmptyState
                      :icon="Film"
                      heading="No movies yet"
                      description="Link anime movies from your downloads to build your collection"
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

    <!-- Unlink confirmation dialog -->
    <AlertDialog v-model:open="unlinkDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Unlink {{ unlinkTarget?.name }}?</AlertDialogTitle>
          <AlertDialogDescription v-if="unlinkLoading" class="flex items-center gap-2">
            <Loader2 class="h-4 w-4 animate-spin" />
            Checking file safety...
          </AlertDialogDescription>
          <AlertDialogDescription v-else-if="unlinkPreview">
            <div class="space-y-3">
              <p>
                This will remove <strong>{{ unlinkPreview.totalFiles }}</strong>
                video file{{ unlinkPreview.totalFiles !== 1 ? 's' : '' }} from the library.
              </p>

              <!-- Safety warning for files that are the only copy -->
              <div
                v-if="hasUnsafeFiles"
                class="rounded-md border border-destructive/50 bg-destructive/10 p-3 space-y-2"
              >
                <div class="flex items-center gap-2 text-destructive font-medium">
                  <AlertTriangle class="h-4 w-4" />
                  Data loss warning
                </div>
                <p class="text-sm">
                  <strong>{{ unlinkPreview.unsafeFiles!.length }}</strong>
                  file{{ unlinkPreview.unsafeFiles!.length !== 1 ? 's are' : ' is' }} the
                  <strong>only copy</strong> (source file in downloads no longer exists).
                  Removing {{ unlinkPreview.unsafeFiles!.length !== 1 ? 'them' : 'it' }} will cause
                  <strong>permanent data loss</strong>.
                </p>
              </div>

              <div v-if="unlinkPreview.safeFiles && unlinkPreview.safeFiles.length > 0" class="text-sm text-muted-foreground">
                {{ unlinkPreview.safeFiles.length }} file{{ unlinkPreview.safeFiles.length !== 1 ? 's are' : ' is' }}
                safe to remove (hardlinks with source still in downloads).
              </div>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter v-if="!unlinkLoading && unlinkPreview">
          <AlertDialogCancel :disabled="unlinkExecuting">Cancel</AlertDialogCancel>
          <!-- If there are unsafe files, show two options -->
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
              Remove all (data loss)
            </AlertDialogAction>
          </template>
          <!-- All files are safe -->
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
