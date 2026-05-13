<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Show } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Tv, Search, X, Check } from 'lucide-vue-next'

const props = defineProps<{
  mediaType: 'series' | 'movie'
  showName: string
  seasonNumber: number
  suggestedName: string
  suggestedSeason: number | null
  shokoAvailable: boolean
  currentPoster: string | null
  shows: Show[]
  existingMovies: string[]
  selectedShow: Show | null
  nextAvailableSeason: number | null
  duplicateSeasonWarning: string | null
  loading: boolean
  folderMap: Record<string, { shokoId: number; name: string; posterUrl?: string }>
}>()

const emit = defineEmits<{
  'update:showName': [value: string]
  'update:seasonNumber': [value: number]
  'select-existing-show': [show: Show]
  confirm: []
  back: []
}>()

// Internal state
const showSearch = ref('')
const expandedSeason = ref<{ showName: string; seasonNumber: number } | null>(null)

// Internal computeds
const filteredExistingShows = computed(() => {
  if (!showSearch.value) return props.shows
  const q = showSearch.value.toLowerCase()
  return props.shows.filter(s => s.name.toLowerCase().includes(q))
})

const filteredExistingItems = computed(() => {
  if (!showSearch.value) return props.existingMovies
  const q = showSearch.value.toLowerCase()
  return props.existingMovies.filter(name => name.toLowerCase().includes(q))
})

// Internal functions
function findShokoPoster(name: string): string | null {
  return props.folderMap[name]?.posterUrl ?? null
}

function toggleSeasonFiles(showName: string, seasonNumber: number, event: Event) {
  event.stopPropagation()
  if (expandedSeason.value?.showName === showName && expandedSeason.value?.seasonNumber === seasonNumber) {
    expandedSeason.value = null
  } else {
    expandedSeason.value = { showName, seasonNumber }
  }
}

function selectExisting(show: Show) {
  emit('select-existing-show', show)
  showSearch.value = ''
  expandedSeason.value = null
}
</script>

<template>
  <Card glass>
    <CardHeader>
      <CardTitle>{{ mediaType === 'series' ? 'Series' : 'Movie' }} Details</CardTitle>
      <CardDescription>
        Confirm the name{{ mediaType === 'series' ? ' and season' : '' }}
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <!-- Name input with poster preview -->
      <div class="flex gap-4">
        <!-- Poster preview -->
        <div v-if="shokoAvailable && currentPoster" class="flex-shrink-0 w-20">
          <div class="aspect-[2/3] rounded-md overflow-hidden bg-muted ring-1 ring-border/50">
            <img :src="currentPoster" :alt="showName" class="w-full h-full object-cover" loading="lazy" />
          </div>
        </div>

        <div class="flex-1 space-y-2">
          <Label>Name</Label>
          <Input
            :model-value="showName"
            @update:model-value="$emit('update:showName', $event)"
            placeholder="Show or movie name"
          />
          <p v-if="suggestedName && suggestedName !== showName" class="text-sm text-muted-foreground">
            Suggested: {{ suggestedName }}
            <Button variant="link" size="sm" class="h-auto p-0 ml-1" @click="$emit('update:showName', suggestedName)">
              Use this
            </Button>
          </p>
        </div>
      </div>

      <!-- Existing shows with season pills (series only) -->
      <div v-if="mediaType === 'series' && shows.length" class="space-y-2">
        <Label class="text-xs text-muted-foreground">Or select existing show:</Label>

        <!-- Search filter -->
        <div class="relative">
          <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input v-model="showSearch" placeholder="Filter shows..." class="pl-9 h-9" />
          <button
            v-if="showSearch"
            class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
            @click="showSearch = ''"
          >
            <X class="h-4 w-4" />
          </button>
        </div>

        <!-- Show list with posters and season pills -->
        <div class="max-h-80 overflow-y-auto rounded-md border">
          <div
            v-for="show in filteredExistingShows"
            :key="show.name"
            class="border-b last:border-b-0"
          >
            <button
              class="flex items-center gap-3 w-full p-2 text-left hover:bg-accent transition-colors"
              :class="{ 'bg-accent': showName === show.name }"
              @click="selectExisting(show)"
            >
              <!-- Mini poster -->
              <div class="flex-shrink-0 w-8 h-12 rounded overflow-hidden bg-muted">
                <img
                  v-if="findShokoPoster(show.name)"
                  :src="findShokoPoster(show.name)!"
                  :alt="show.name"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
                <div v-else class="w-full h-full flex items-center justify-center">
                  <Tv class="h-3 w-3 text-muted-foreground/40" />
                </div>
              </div>
              <div class="min-w-0 flex-1">
                <span class="text-sm font-medium truncate block">{{ show.name }}</span>
                <div v-if="show.seasons.length" class="flex gap-1 flex-wrap mt-1">
                  <Badge
                    v-for="season in show.seasons"
                    :key="season.number"
                    variant="outline"
                    class="text-[10px] px-1.5 py-0 cursor-pointer hover:bg-accent"
                    :class="{ 'bg-primary/10 border-primary/40': expandedSeason?.showName === show.name && expandedSeason?.seasonNumber === season.number }"
                    @click="toggleSeasonFiles(show.name, season.number, $event)"
                  >
                    S{{ season.number }} · {{ season.episodes }}ep
                  </Badge>
                </div>
              </div>
              <Check v-if="showName === show.name" class="h-4 w-4 text-primary ml-auto flex-shrink-0" />
            </button>
            <!-- Expanded file list panel -->
            <div
              v-if="show.seasons.some(s => expandedSeason?.showName === show.name && expandedSeason?.seasonNumber === s.number)"
              class="px-3 pb-3 pt-1 bg-muted/30 border-t border-border/50"
            >
              <template v-for="season in show.seasons" :key="season.number">
                <div v-if="expandedSeason?.showName === show.name && expandedSeason?.seasonNumber === season.number">
                  <div class="flex items-center justify-between mb-1.5">
                    <span class="text-xs font-medium text-muted-foreground">Season {{ season.number }} — {{ season.files.length }} file{{ season.files.length !== 1 ? 's' : '' }}</span>
                  </div>
                  <div v-if="season.files.length" class="max-h-40 overflow-y-auto rounded border bg-background/50 p-2">
                    <div
                      v-for="file in season.files"
                      :key="file"
                      class="text-[11px] font-mono text-muted-foreground py-0.5 truncate"
                      :title="file"
                    >
                      {{ file }}
                    </div>
                  </div>
                  <div v-else class="text-xs text-muted-foreground italic">No files</div>
                </div>
              </template>
            </div>
          </div>
          <div v-if="filteredExistingShows.length === 0" class="p-3 text-center text-sm text-muted-foreground">
            No matches
          </div>
        </div>
      </div>

      <!-- Existing movies (movies only) -->
      <div v-if="mediaType === 'movie' && existingMovies.length" class="space-y-2">
        <Label class="text-xs text-muted-foreground">Or select existing movie:</Label>

        <div class="relative">
          <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input v-model="showSearch" placeholder="Filter movies..." class="pl-9 h-9" />
          <button
            v-if="showSearch"
            class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
            @click="showSearch = ''"
          >
            <X class="h-4 w-4" />
          </button>
        </div>

        <div class="max-h-64 overflow-y-auto rounded-md border">
          <button
            v-for="name in filteredExistingItems"
            :key="name"
            class="flex items-center gap-3 w-full p-2 text-left hover:bg-accent transition-colors border-b last:border-b-0"
            :class="{ 'bg-accent': showName === name }"
            @click="$emit('update:showName', name); showSearch = ''"
          >
            <div class="flex-shrink-0 w-8 h-12 rounded overflow-hidden bg-muted">
              <img
                v-if="findShokoPoster(name)"
                :src="findShokoPoster(name)!"
                :alt="name"
                class="w-full h-full object-cover"
                loading="lazy"
              />
              <div v-else class="w-full h-full flex items-center justify-center">
                <Tv class="h-3 w-3 text-muted-foreground/40" />
              </div>
            </div>
            <span class="text-sm font-medium truncate">{{ name }}</span>
            <Check v-if="showName === name" class="h-4 w-4 text-primary ml-auto flex-shrink-0" />
          </button>
          <div v-if="filteredExistingItems.length === 0" class="p-3 text-center text-sm text-muted-foreground">
            No matches
          </div>
        </div>
      </div>

      <div v-if="mediaType === 'series'" class="space-y-2">
        <Label>Season Number</Label>
        <div class="flex items-center gap-2">
          <Input
            :model-value="seasonNumber"
            @update:model-value="$emit('update:seasonNumber', Number($event))"
            type="number"
            min="0"
            max="99"
            class="w-24"
          />
          <span v-if="nextAvailableSeason !== null && seasonNumber === nextAvailableSeason" class="text-xs text-muted-foreground">
            Next available
          </span>
        </div>
        <p v-if="duplicateSeasonWarning" class="text-sm text-amber-500">
          {{ duplicateSeasonWarning }}
        </p>
        <p v-else-if="suggestedSeason !== null && suggestedSeason !== seasonNumber" class="text-sm text-muted-foreground">
          Detected season: {{ suggestedSeason }}
          <Button variant="link" size="sm" class="h-auto p-0 ml-1" @click="$emit('update:seasonNumber', suggestedSeason!)">
            Use this
          </Button>
        </p>
      </div>

      <Separator />

      <div class="flex gap-2">
        <Button variant="ghost" @click="$emit('back')">Back</Button>
        <Button @click="$emit('confirm')" :disabled="!showName || loading">
          {{ loading ? 'Checking...' : 'Preview' }}
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
