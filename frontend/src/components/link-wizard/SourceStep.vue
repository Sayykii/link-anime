<script setup lang="ts">
import type { DownloadItem } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { FolderOpen, FileVideo, Link, Loader2, Search, X, Eye, EyeOff, Check, CheckCircle } from 'lucide-vue-next'
import { formatSize } from '@/lib/utils'

type LinkedSource = { id: number; mediaType: string; showName: string; season?: number }

defineProps<{
  downloads: DownloadItem[]
  loading: boolean
  sourceFilter: string
  showLinked: boolean
  filteredSources: DownloadItem[]
  selectedSources: Set<string>
  linkedSources: Record<string, LinkedSource>
}>()

defineEmits<{
  'update:sourceFilter': [value: string]
  'update:showLinked': [value: boolean]
  'toggle-source': [item: DownloadItem]
  'select-source': [item: DownloadItem]
  'start-linking': []
}>()
</script>

<template>
  <Card glass>
    <CardHeader class="flex flex-row items-center justify-between">
      <div>
        <CardTitle>Select Source</CardTitle>
        <CardDescription>Choose downloads to link into your library</CardDescription>
      </div>
      <Button
        v-if="selectedSources.size > 0"
        @click="$emit('start-linking')"
        class="gap-2 shrink-0"
      >
        <Link class="h-4 w-4" />
        Link {{ selectedSources.size }} selected
      </Button>
    </CardHeader>
    <CardContent>
      <div v-if="loading" class="flex items-center gap-2 text-muted-foreground py-8 justify-center">
        <Loader2 class="h-4 w-4 animate-spin" />
        Loading downloads...
      </div>
      <div v-else-if="!downloads.length" class="text-center text-muted-foreground py-8">
        No downloads found in the download directory
      </div>
      <template v-else>
        <div class="flex items-center gap-2 mb-3">
          <div v-if="downloads.length > 5" class="relative flex-1">
            <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input :model-value="sourceFilter" @update:model-value="$emit('update:sourceFilter', $event)" placeholder="Filter downloads..." class="pl-9 h-9" />
            <button
              v-if="sourceFilter"
              class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
              @click="$emit('update:sourceFilter', '')"
            >
              <X class="h-4 w-4" />
            </button>
          </div>
          <Button
            variant="outline"
            size="sm"
            class="gap-1.5 shrink-0 h-9"
            @click="$emit('update:showLinked', !showLinked)"
          >
            <Eye v-if="showLinked" class="h-3.5 w-3.5" />
            <EyeOff v-else class="h-3.5 w-3.5" />
            <span class="hidden sm:inline">{{ showLinked ? 'Showing linked' : 'Hiding linked' }}</span>
          </Button>
        </div>
        <div v-if="sourceFilter && !filteredSources.length" class="text-center text-muted-foreground py-8">
          No downloads matching "{{ sourceFilter }}"
          <div class="mt-2">
            <Button variant="outline" size="sm" @click="$emit('update:sourceFilter', '')">Clear filter</Button>
          </div>
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="item in filteredSources"
            :key="item.path"
            class="flex w-full items-center gap-3 rounded-lg border p-3 transition-colors"
            :class="[
              item.linked ? 'border-green-500/30 bg-green-500/5' : '',
              selectedSources.has(item.path) ? 'ring-2 ring-primary border-primary' : '',
            ]"
          >
            <button
              class="shrink-0 h-5 w-5 rounded border-2 flex items-center justify-center transition-colors"
              :class="selectedSources.has(item.path) ? 'bg-primary border-primary text-primary-foreground' : 'border-muted-foreground/40 hover:border-primary'"
              @click.stop="$emit('toggle-source', item)"
            >
              <Check v-if="selectedSources.has(item.path)" class="h-3 w-3" />
            </button>
            <button
              class="flex items-center gap-3 min-w-0 flex-1 text-left hover:opacity-80"
              @click="$emit('select-source', item)"
            >
              <FolderOpen v-if="item.isDir" class="h-5 w-5 shrink-0" :class="item.linked ? 'text-green-500' : 'text-muted-foreground'" />
              <FileVideo v-else class="h-5 w-5 shrink-0" :class="item.linked ? 'text-green-500' : 'text-muted-foreground'" />
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span class="truncate font-medium">{{ item.name }}</span>
                  <Badge v-if="item.linked" variant="outline" class="shrink-0 gap-1 text-green-600 border-green-500/30 text-xs">
                    <CheckCircle class="h-3 w-3" />
                    Linked
                  </Badge>
                </div>
                <div class="text-sm text-muted-foreground">
                  {{ item.videoCount }} video{{ item.videoCount !== 1 ? 's' : '' }}
                  &middot; {{ formatSize(item.size) }}
                </div>
              </div>
            </button>
          </div>
        </div>
      </template>
    </CardContent>
  </Card>
</template>
