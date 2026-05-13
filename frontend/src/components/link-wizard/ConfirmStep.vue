<script setup lang="ts">
import type { LinkResult } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Link } from 'lucide-vue-next'
import { formatSize } from '@/lib/utils'

defineProps<{
  selectedSource: { name: string } | null
  mediaType: 'series' | 'movie'
  showName: string
  seasonNumber: number
  previewResult: LinkResult | null
  shokoAvailable: boolean
  currentPoster: string | null
}>()

defineEmits<{
  execute: []
  back: []
}>()
</script>

<template>
  <Card glass>
    <CardHeader>
      <CardTitle>Confirm Link</CardTitle>
      <CardDescription>Review before linking</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <div class="flex gap-4">
        <div v-if="shokoAvailable && currentPoster" class="flex-shrink-0 w-24">
          <div class="aspect-[2/3] rounded-md overflow-hidden bg-muted ring-1 ring-border/50">
            <img :src="currentPoster" :alt="showName" class="w-full h-full object-cover" />
          </div>
        </div>
        <div class="flex-1 grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="text-muted-foreground">Source:</span>
            <div class="font-medium">{{ selectedSource?.name }}</div>
          </div>
          <div>
            <span class="text-muted-foreground">Type:</span>
            <div class="font-medium capitalize">{{ mediaType }}</div>
          </div>
          <div>
            <span class="text-muted-foreground">Name:</span>
            <div class="font-medium">{{ showName }}</div>
          </div>
          <div v-if="mediaType === 'series'">
            <span class="text-muted-foreground">Season:</span>
            <div class="font-medium">{{ seasonNumber }}</div>
          </div>
        </div>
      </div>
      <Separator />
      <div v-if="previewResult" class="space-y-2">
        <h4 class="font-medium">Preview:</h4>
        <div class="text-sm space-y-1">
          <div>Destination: <code class="text-xs bg-muted px-1 py-0.5 rounded">{{ previewResult.destDir }}</code></div>
          <div>Files to link: <strong>{{ previewResult.linked }}</strong></div>
          <div v-if="previewResult.skipped">Already exists: {{ previewResult.skipped }}</div>
          <div>Total size: {{ formatSize(previewResult.size) }}</div>
        </div>
      </div>
      <Separator />
      <div class="flex gap-2">
        <Button variant="ghost" @click="$emit('back')">Back</Button>
        <Button @click="$emit('execute')" class="gap-2">
          <Link class="h-4 w-4" />
          Link Now
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
