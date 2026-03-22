<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useLibraryStore } from '@/stores/library'
import { useApi } from '@/composables/useApi'
import { formatSize, seriesPosterUrl, dashboardPosterUrl } from '@/lib/utils'
import type { ShokoDashboard } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tv,
  Film,
  HardDrive,
  Hash,
  Layers,
  Link,
  Eye,
  Clock,
  AlertCircle,
  Play,
  Calendar,
  Loader2,
} from 'lucide-vue-next'

const api = useApi()
const library = useLibraryStore()
const router = useRouter()

const dashboard = ref<ShokoDashboard | null>(null)
const shokoAvailable = ref(false)
const loading = ref(true)

onMounted(async () => {
  // Fetch basic stats in parallel with Shoko dashboard
  library.fetchStats()

  try {
    const data = await api.shokoDashboard()
    dashboard.value = data
    shokoAvailable.value = true
  } catch {
    // Shoko not configured — fall back to basic stats
    shokoAvailable.value = false
  } finally {
    loading.value = false
  }
})

const stats = computed(() => dashboard.value?.stats)
const recentlyAdded = computed(() => dashboard.value?.recentlyAdded?.slice(0, 8) ?? [])
const continueWatching = computed(() => dashboard.value?.continueWatching?.slice(0, 8) ?? [])
const calendar = computed(() => dashboard.value?.calendar?.slice(0, 10) ?? [])

// Group calendar by day
const calendarByDay = computed(() => {
  const groups: Record<string, typeof calendar.value> = {}
  for (const ep of calendar.value) {
    const day = ep.AirDate ?? 'Unknown'
    if (!groups[day]) groups[day] = []
    groups[day].push(ep)
  }
  return groups
})

function formatCalendarDay(dateStr: string): string {
  if (dateStr === 'Unknown') return dateStr
  const d = new Date(dateStr + 'T00:00:00')
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const diff = Math.round((d.getTime() - today.getTime()) / 86400000)
  if (diff === 0) return 'Today'
  if (diff === 1) return 'Tomorrow'
  return d.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
}

function truncate(text: string, max: number): string {
  return text.length > max ? text.slice(0, max) + '...' : text
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="font-display text-3xl tracking-wider uppercase">Dashboard</h1>
      <p class="text-muted-foreground text-sm mt-1">Overview of your anime library</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-6">
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
      <div class="grid gap-4 md:grid-cols-2">
        <Card glass>
          <CardHeader><Skeleton class="h-5 w-32" /></CardHeader>
          <CardContent class="flex gap-3">
            <Skeleton v-for="i in 4" :key="i" class="h-40 w-28 rounded-md" />
          </CardContent>
        </Card>
        <Card glass>
          <CardHeader><Skeleton class="h-5 w-32" /></CardHeader>
          <CardContent class="flex gap-3">
            <Skeleton v-for="i in 4" :key="i" class="h-40 w-28 rounded-md" />
          </CardContent>
        </Card>
      </div>
    </div>

    <!-- Shoko-powered dashboard -->
    <template v-else-if="shokoAvailable && dashboard">
      <!-- Stats row -->
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card glass>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">Series</CardTitle>
            <div class="gradient-icon"><Tv class="h-4 w-4" /></div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold tabular-nums">{{ stats?.SeriesCount ?? 0 }}</div>
            <p v-if="stats?.FinishedSeries" class="text-xs text-muted-foreground mt-1">{{ stats.FinishedSeries }} finished</p>
          </CardContent>
        </Card>
        <Card glass>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">Episodes</CardTitle>
            <div class="gradient-icon"><Hash class="h-4 w-4" /></div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold tabular-nums">{{ stats?.FileCount ?? 0 }}</div>
            <p class="text-xs text-muted-foreground mt-1">{{ formatSize(stats?.FileSize ?? 0) }}</p>
          </CardContent>
        </Card>
        <Card glass>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">Watched</CardTitle>
            <div class="gradient-icon"><Eye class="h-4 w-4" /></div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold tabular-nums">{{ stats?.WatchedEpisodes ?? 0 }}</div>
            <p v-if="stats?.WatchedHours" class="text-xs text-muted-foreground mt-1">{{ Math.round(stats.WatchedHours) }} hours</p>
          </CardContent>
        </Card>
        <Card glass>
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <CardTitle class="text-sm font-medium">Missing</CardTitle>
            <div class="gradient-icon"><AlertCircle class="h-4 w-4" /></div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold tabular-nums">{{ stats?.MissingEpisodesCollecting ?? 0 }}</div>
            <p class="text-xs text-muted-foreground mt-1">episodes to collect</p>
          </CardContent>
        </Card>
      </div>

      <!-- Continue Watching -->
      <Card v-if="continueWatching.length" glass>
        <CardHeader class="pb-3">
          <CardTitle class="flex items-center gap-2 text-base">
            <Play class="h-4 w-4" />
            Continue Watching
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="flex gap-3 overflow-x-auto pb-2 -mx-1 px-1">
            <button
              v-for="ep in continueWatching"
              :key="ep.IDs.ID"
              class="flex-shrink-0 w-32 group cursor-pointer text-left"
              @click="router.push(`/library/series/${ep.IDs.ShokoSeries}`)"
            >
              <div class="relative aspect-[2/3] rounded-md overflow-hidden bg-muted mb-2">
                <img
                  v-if="dashboardPosterUrl(ep.SeriesPoster)"
                  :src="dashboardPosterUrl(ep.SeriesPoster)!"
                  :alt="ep.SeriesTitle"
                  class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
                  loading="lazy"
                />
                <div v-else class="w-full h-full flex items-center justify-center">
                  <Tv class="h-8 w-8 text-muted-foreground/40" />
                </div>
                <div class="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/80 to-transparent p-2">
                  <Badge variant="secondary" class="text-xs">
                    EP {{ ep.Number }}
                  </Badge>
                </div>
              </div>
              <p class="text-xs font-medium leading-tight line-clamp-2">{{ ep.SeriesTitle }}</p>
            </button>
          </div>
        </CardContent>
      </Card>

      <!-- Recently Added + Calendar side by side -->
      <div class="grid gap-4 lg:grid-cols-2">
        <!-- Recently Added -->
        <Card v-if="recentlyAdded.length" glass>
          <CardHeader class="pb-3">
            <CardTitle class="flex items-center gap-2 text-base">
              <Clock class="h-4 w-4" />
              Recently Added
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-1">
            <button
              v-for="ep in recentlyAdded"
              :key="ep.IDs.ID"
              class="flex items-center gap-3 w-full rounded-md p-2 hover:bg-muted/50 transition-colors cursor-pointer text-left"
              @click="router.push(`/library/series/${ep.IDs.ShokoSeries}`)"
            >
              <div class="flex-shrink-0 w-10 h-14 rounded overflow-hidden bg-muted">
                <img
                  v-if="dashboardPosterUrl(ep.SeriesPoster)"
                  :src="dashboardPosterUrl(ep.SeriesPoster)!"
                  :alt="ep.SeriesTitle"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium truncate">{{ ep.SeriesTitle }}</p>
                <p class="text-xs text-muted-foreground truncate">
                  Episode {{ ep.Number }}<span v-if="ep.Title"> - {{ ep.Title }}</span>
                </p>
              </div>
              <Badge variant="outline" class="text-xs flex-shrink-0">EP {{ ep.Number }}</Badge>
            </button>
          </CardContent>
        </Card>

        <!-- Airing Calendar -->
        <Card v-if="calendar.length" glass>
          <CardHeader class="pb-3">
            <CardTitle class="flex items-center gap-2 text-base">
              <Calendar class="h-4 w-4" />
              Airing This Week
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <div v-for="(eps, day) in calendarByDay" :key="day">
              <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1.5">{{ formatCalendarDay(day) }}</p>
              <div class="space-y-1">
                <div
                  v-for="ep in eps"
                  :key="ep.IDs.ID"
                  class="flex items-center gap-2 text-sm py-1"
                >
                  <div class="flex-shrink-0 w-7 h-10 rounded overflow-hidden bg-muted">
                    <img
                      v-if="dashboardPosterUrl(ep.SeriesPoster)"
                      :src="dashboardPosterUrl(ep.SeriesPoster)!"
                      class="w-full h-full object-cover"
                      loading="lazy"
                    />
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm truncate">{{ ep.SeriesTitle }}</p>
                  </div>
                  <Badge variant="outline" class="text-xs flex-shrink-0">EP {{ ep.Number }}</Badge>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Quick actions -->
      <Card glass>
        <CardContent class="flex gap-3 pt-6">
          <Button @click="router.push('/link')" class="gap-2">
            <Link class="h-4 w-4" />
            Link New Content
          </Button>
          <Button variant="outline" @click="router.push('/library')" class="gap-2">
            <Tv class="h-4 w-4" />
            Browse Library
          </Button>
        </CardContent>
      </Card>
    </template>

    <!-- Fallback: basic dashboard when Shoko is not configured -->
    <template v-else>
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card
          v-for="(stat, index) in [
            { label: 'Shows', value: library.stats?.shows ?? 0, icon: Tv },
            { label: 'Seasons', value: library.stats?.seasons ?? 0, icon: Layers },
            { label: 'Episodes', value: library.stats?.episodes ?? 0, icon: Hash },
            { label: 'Movies', value: library.stats?.movies ?? 0, icon: Film },
          ]"
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

      <Card glass class="stagger-fade-in" style="animation-delay: 320ms">
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Total Library Size</CardTitle>
          <div class="gradient-icon"><HardDrive class="h-4 w-4" /></div>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatSize(library.stats?.size ?? 0) }}</div>
        </CardContent>
      </Card>

      <Card glass>
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
        </CardContent>
      </Card>
    </template>
  </div>
</template>
