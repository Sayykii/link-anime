<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLibraryStore } from '@/stores/library'
import { formatSize } from '@/lib/utils'
import { useCountUp } from '@/composables/useCountUp'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tv,
  Film,
  HardDrive,
  Hash,
  Layers,
  Link,
  Settings,
} from 'lucide-vue-next'

const library = useLibraryStore()
const router = useRouter()

onMounted(() => {
  library.fetchStats()
})

// Reactive source values
const showsCount = computed(() => library.stats?.shows ?? 0)
const seasonsCount = computed(() => library.stats?.seasons ?? 0)
const episodesCount = computed(() => library.stats?.episodes ?? 0)
const moviesCount = computed(() => library.stats?.movies ?? 0)
const sizeBytes = computed(() => library.stats?.size ?? 0)

// Animated counters with stagger
const animatedShows = useCountUp(showsCount, { delay: 0 })
const animatedSeasons = useCountUp(seasonsCount, { delay: 100 })
const animatedEpisodes = useCountUp(episodesCount, { delay: 200 })
const animatedMovies = useCountUp(moviesCount, { delay: 300 })
const animatedSize = useCountUp(sizeBytes, { delay: 400 })

const statCards = computed(() => [
  { label: 'Shows', value: Math.round(animatedShows.value), icon: Tv, delay: 0 },
  { label: 'Seasons', value: Math.round(animatedSeasons.value), icon: Layers, delay: 1 },
  { label: 'Episodes', value: Math.round(animatedEpisodes.value), icon: Hash, delay: 2 },
  { label: 'Movies', value: Math.round(animatedMovies.value), icon: Film, delay: 3 },
])

const hasStats = computed(() => library.stats !== null)
</script>

<template>
  <div class="space-y-6">
    <div class="menacing">
      <h1 class="text-3xl font-bold">Dashboard</h1>
      <p class="text-muted-foreground">Overview of your anime library</p>
    </div>

    <!-- Skeleton loading state -->
    <template v-if="!hasStats">
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card v-for="i in 4" :key="i" glass>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <Skeleton class="h-4 w-16" />
            <Skeleton class="h-4 w-4 rounded" />
          </CardHeader>
          <CardContent>
            <Skeleton class="h-8 w-20" />
          </CardContent>
        </Card>
      </div>
      <Card glass>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <Skeleton class="h-4 w-32" />
          <Skeleton class="h-4 w-4 rounded" />
        </CardHeader>
        <CardContent>
          <Skeleton class="h-8 w-24" />
        </CardContent>
      </Card>
    </template>

    <!-- Stats grid -->
    <template v-else>
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card
          v-for="(stat, index) in statCards"
          :key="stat.label"
          glass
          class="stagger-fade-in"
          :style="{ animationDelay: `${index * 80}ms` }"
        >
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">{{ stat.label }}</CardTitle>
            <div class="gradient-icon">
              <component :is="stat.icon" class="h-4 w-4" />
            </div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold tabular-nums">{{ stat.value }}</div>
          </CardContent>
        </Card>
      </div>

      <!-- Library size -->
      <Card
        glass
        class="stagger-fade-in"
        :style="{ animationDelay: '320ms' }"
      >
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Total Library Size</CardTitle>
          <div class="gradient-icon">
            <HardDrive class="h-4 w-4" />
          </div>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatSize(Math.round(animatedSize)) }}</div>
        </CardContent>
      </Card>
    </template>

    <!-- Quick actions -->
    <Card
      glass
      class="stagger-fade-in"
      :style="{ animationDelay: hasStats ? '400ms' : '0ms' }"
    >
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
