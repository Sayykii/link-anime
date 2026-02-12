<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useLibraryStore } from '@/stores/library'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Search, RefreshCw, Tv, Film } from 'lucide-vue-next'

const library = useLibraryStore()
const searchQuery = ref('')
const activeTab = ref('shows')

onMounted(() => {
  library.fetchShows()
  library.fetchMovies()
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
  library.fetchShows()
  library.fetchMovies()
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

    <!-- Search -->
    <div class="relative max-w-sm">
      <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input v-model="searchQuery" placeholder="Search library..." class="pl-9" />
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
        <Card>
          <CardContent class="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Show</TableHead>
                  <TableHead class="w-32">Seasons</TableHead>
                  <TableHead class="w-32">Episodes</TableHead>
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
                        class="text-xs"
                      >
                        S{{ season.number }} ({{ season.episodes }})
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell>{{ show.seasons.length }}</TableCell>
                  <TableCell>{{ show.episodes }}</TableCell>
                </TableRow>
                <TableRow v-if="!filteredShows.length">
                  <TableCell colspan="3" class="text-center text-muted-foreground py-8">
                    {{ searchQuery ? 'No shows match your search' : 'No shows in library' }}
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="movies">
        <Card>
          <CardContent class="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Movie</TableHead>
                  <TableHead class="w-32">Files</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="movie in filteredMovies" :key="movie.path">
                  <TableCell class="font-medium">{{ movie.name }}</TableCell>
                  <TableCell>{{ movie.files }}</TableCell>
                </TableRow>
                <TableRow v-if="!filteredMovies.length">
                  <TableCell colspan="2" class="text-center text-muted-foreground py-8">
                    {{ searchQuery ? 'No movies match your search' : 'No movies in library' }}
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
