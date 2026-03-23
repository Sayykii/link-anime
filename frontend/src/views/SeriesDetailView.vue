<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { seriesPosterUrl } from '@/lib/utils'
import type { ShokoSeries, ShokoEpisode } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  ArrowLeft,
  Tv,
  Eye,
  EyeOff,
  AlertCircle,
  Star,
  Calendar,
  Hash,
  Loader2,
} from 'lucide-vue-next'

const api = useApi()
const route = useRoute()
const router = useRouter()

const series = ref<ShokoSeries | null>(null)
const episodes = ref<ShokoEpisode[]>([])
const loading = ref(true)
const episodesLoading = ref(true)

const seriesId = computed(() => Number(route.params.id))

onMounted(async () => {
  try {
    const [s, eps] = await Promise.all([
      api.shokoSeriesDetail(seriesId.value),
      api.shokoSeriesEpisodes(seriesId.value, true),
    ])
    series.value = s
    episodes.value = eps || []
  } catch (err: any) {
    // Shoko might not be available
  } finally {
    loading.value = false
    episodesLoading.value = false
  }
})

const posterUrl = computed(() => {
  if (!series.value) return null
  return seriesPosterUrl(series.value.Images)
})

const rating = computed(() => {
  if (!series.value?.AniDB?.Rating) return null
  const r = series.value.AniDB.Rating
  return (r.Value / (r.MaxValue / 10)).toFixed(1)
})

const episodeStats = computed(() => {
  const local = episodes.value.filter(e => e.Size > 0).length
  const missing = episodes.value.filter(e => e.Size === 0).length
  const watched = episodes.value.filter(e => e.Watched !== null).length
  return { local, missing, watched, total: episodes.value.length }
})

function formatAirDate(dateStr?: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="space-y-6">
    <!-- Back button -->
    <Button variant="ghost" size="sm" @click="router.push('/library')" class="gap-2 -ml-2">
      <ArrowLeft class="h-4 w-4" />
      Back to Library
    </Button>

    <!-- Loading -->
    <div v-if="loading" class="space-y-6">
      <div class="flex gap-6">
        <Skeleton class="w-48 h-72 rounded-lg flex-shrink-0" />
        <div class="flex-1 space-y-3">
          <Skeleton class="h-8 w-64" />
          <Skeleton class="h-4 w-32" />
          <Skeleton class="h-20 w-full" />
        </div>
      </div>
    </div>

    <template v-else-if="series">
      <!-- Hero section -->
      <div class="flex flex-col sm:flex-row gap-6">
        <!-- Poster -->
        <div class="flex-shrink-0 w-40 sm:w-48">
          <div class="aspect-[2/3] rounded-lg overflow-hidden bg-muted ring-1 ring-border/50">
            <img
              v-if="posterUrl"
              :src="posterUrl"
              :alt="series.Name"
              class="w-full h-full object-cover"
            />
            <div v-else class="w-full h-full flex items-center justify-center">
              <Tv class="h-12 w-12 text-muted-foreground/30" />
            </div>
          </div>
        </div>

        <!-- Info -->
        <div class="flex-1 min-w-0">
          <h1 class="font-display text-2xl tracking-wider uppercase">{{ series.Name }}</h1>

          <div class="flex flex-wrap items-center gap-2 mt-2">
            <Badge v-if="series.AniDB?.Type" variant="outline">{{ series.AniDB.Type }}</Badge>
            <Badge v-if="rating" variant="secondary" class="gap-1">
              <Star class="h-3 w-3" />
              {{ rating }}
            </Badge>
            <Badge v-if="series.AniDB?.AirDate" variant="outline" class="gap-1">
              <Calendar class="h-3 w-3" />
              {{ formatAirDate(series.AniDB.AirDate) }}
            </Badge>
          </div>

          <!-- Episode stats -->
          <div class="flex gap-4 mt-4 text-sm">
            <div class="flex items-center gap-1.5">
              <Hash class="h-4 w-4 text-muted-foreground" />
              <span><strong>{{ episodeStats.local }}</strong>/{{ episodeStats.total }} episodes</span>
            </div>
            <div v-if="episodeStats.watched" class="flex items-center gap-1.5">
              <Eye class="h-4 w-4 text-muted-foreground" />
              <span><strong>{{ episodeStats.watched }}</strong> watched</span>
            </div>
            <div v-if="episodeStats.missing" class="flex items-center gap-1.5 text-destructive">
              <AlertCircle class="h-4 w-4" />
              <span><strong>{{ episodeStats.missing }}</strong> missing</span>
            </div>
          </div>

          <!-- Description -->
          <p v-if="series.AniDB?.Description" class="text-sm text-muted-foreground mt-4 line-clamp-4">
            {{ series.AniDB.Description.replace(/<[^>]*>/g, '').replace(/`/g, "'") }}
          </p>
        </div>
      </div>

      <!-- Episodes -->
      <Card glass>
        <CardHeader class="pb-3">
          <CardTitle class="text-base">Episodes</CardTitle>
        </CardHeader>
        <CardContent class="p-0">
          <div v-if="episodesLoading" class="flex items-center justify-center py-8">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
          <Table v-else>
            <TableHeader>
              <TableRow>
                <TableHead class="w-16">#</TableHead>
                <TableHead>Title</TableHead>
                <TableHead class="w-28">Air Date</TableHead>
                <TableHead class="w-20 text-center">Status</TableHead>
                <TableHead class="w-20 text-center">Watched</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="ep in episodes"
                :key="ep.IDs.ID"
                :class="{ 'opacity-50': ep.Size === 0 }"
              >
                <TableCell class="font-mono text-sm tabular-nums">
                  {{ ep.AniDB?.EpisodeNumber ?? ep.IDs.ID }}
                </TableCell>
                <TableCell>
                  <div>
                    <span class="text-sm font-medium">{{ ep.Name || ep.AniDB?.Title || 'Unknown' }}</span>
                  </div>
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  {{ formatAirDate(ep.AniDB?.AirDate) }}
                </TableCell>
                <TableCell class="text-center">
                  <Badge v-if="ep.Size > 0" variant="secondary" class="text-xs">Local</Badge>
                  <Badge v-else variant="destructive" class="text-xs">Missing</Badge>
                </TableCell>
                <TableCell class="text-center">
                  <Eye v-if="ep.Watched" class="h-4 w-4 text-primary mx-auto" />
                  <EyeOff v-else class="h-4 w-4 text-muted-foreground/30 mx-auto" />
                </TableCell>
              </TableRow>
              <TableRow v-if="episodes.length === 0">
                <TableCell colspan="5" class="text-center py-8 text-muted-foreground">
                  No episodes found
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </template>

    <!-- Error state -->
    <div v-else class="text-center py-12 text-muted-foreground">
      <Tv class="h-12 w-12 mx-auto mb-4 opacity-40" />
      <p class="font-medium text-foreground">Series not found</p>
      <p class="text-sm mt-1">Shoko may not be configured or the series ID is invalid</p>
      <Button variant="outline" class="mt-4" @click="router.push('/library')">
        Back to Library
      </Button>
    </div>
  </div>
</template>
