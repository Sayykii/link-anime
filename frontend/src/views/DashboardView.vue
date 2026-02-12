<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useLibraryStore } from '@/stores/library'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Tv,
  Film,
  HardDrive,
  Hash,
  Layers,
  Link,
} from 'lucide-vue-next'

const library = useLibraryStore()
const router = useRouter()

onMounted(() => {
  library.fetchStats()
})

function formatSize(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(2) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return bytes + ' B'
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-bold">Dashboard</h1>
      <p class="text-muted-foreground">Overview of your anime library</p>
    </div>

    <!-- Stats grid -->
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4" v-if="library.stats">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Shows</CardTitle>
          <Tv class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ library.stats.shows }}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Seasons</CardTitle>
          <Layers class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ library.stats.seasons }}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Episodes</CardTitle>
          <Hash class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ library.stats.episodes }}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Movies</CardTitle>
          <Film class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ library.stats.movies }}</div>
        </CardContent>
      </Card>
    </div>

    <!-- Library size -->
    <Card v-if="library.stats">
      <CardHeader class="flex flex-row items-center justify-between pb-2">
        <CardTitle class="text-sm font-medium">Total Library Size</CardTitle>
        <HardDrive class="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div class="text-2xl font-bold">{{ formatSize(library.stats.size) }}</div>
      </CardContent>
    </Card>

    <!-- Quick actions -->
    <Card>
      <CardHeader>
        <CardTitle>Quick Actions</CardTitle>
      </CardHeader>
      <CardContent class="flex gap-3">
        <Button @click="router.push('/link')" class="gap-2">
          <Link class="h-4 w-4" />
          Link New Content
        </Button>
        <Button variant="outline" @click="router.push('/library')" class="gap-2">
          <Tv class="h-4 w-4" />
          Browse Library
        </Button>
        <Button variant="outline" @click="library.fetchStats()">
          Refresh Stats
        </Button>
      </CardContent>
    </Card>
  </div>
</template>
