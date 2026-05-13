<script setup lang="ts">
import type { LinkResult } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Check, ArrowRight } from 'lucide-vue-next'

defineProps<{
  finalResult: LinkResult | null
  showName: string
  shokoAvailable: boolean
  currentPoster: string | null
  hasMoreInQueue: boolean
  bulkIndex: number
  bulkTotal: number
}>()

defineEmits<{
  'next-in-queue': []
  'reset': []
  'view-library': []
}>()
</script>

<template>
  <Card glass>
    <CardHeader>
      <CardTitle class="flex items-center gap-2 text-green-600">
        <Check class="h-5 w-5" />
        Link Complete
      </CardTitle>
    </CardHeader>
    <CardContent class="space-y-4" v-if="finalResult">
      <div class="flex gap-4">
        <div v-if="shokoAvailable && currentPoster" class="flex-shrink-0 w-20">
          <div class="aspect-[2/3] rounded-md overflow-hidden bg-muted">
            <img :src="currentPoster" :alt="showName" class="w-full h-full object-cover" />
          </div>
        </div>
        <div class="flex-1">
          <div class="grid grid-cols-3 gap-4 text-center">
            <div>
              <div class="text-2xl font-bold">{{ finalResult.linked }}</div>
              <div class="text-sm text-muted-foreground">Linked</div>
            </div>
            <div>
              <div class="text-2xl font-bold">{{ finalResult.skipped }}</div>
              <div class="text-sm text-muted-foreground">Skipped</div>
            </div>
            <div>
              <div class="text-2xl font-bold">{{ finalResult.failed }}</div>
              <div class="text-sm text-muted-foreground">Failed</div>
            </div>
          </div>
          <div class="text-sm mt-3">
            <span class="text-muted-foreground">Destination:</span>
            <code class="ml-1 text-xs bg-muted px-1 py-0.5 rounded">{{ finalResult.destDir }}</code>
          </div>
        </div>
      </div>
      <Separator />
      <div class="flex gap-2">
        <Button v-if="hasMoreInQueue" @click="$emit('next-in-queue')" class="gap-2">
          <ArrowRight class="h-4 w-4" />
          Link Next ({{ bulkIndex + 2 }}/{{ bulkTotal }})
        </Button>
        <Button :variant="hasMoreInQueue ? 'outline' : 'default'" @click="$emit('reset')">Link Another</Button>
        <Button variant="outline" @click="$emit('view-library')">View Library</Button>
      </div>
    </CardContent>
  </Card>
</template>
