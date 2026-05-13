<script setup lang="ts">
import type { LinkProgress } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Loader2 } from 'lucide-vue-next'

defineProps<{
  progressPercent: number
  linkProgress: LinkProgress[]
}>()
</script>

<template>
  <Card glass>
    <CardHeader>
      <CardTitle class="flex items-center gap-2">
        <Loader2 class="h-5 w-5 animate-spin" />
        Linking...
      </CardTitle>
    </CardHeader>
    <CardContent class="space-y-4">
      <Progress :model-value="progressPercent" />
      <div class="max-h-48 overflow-auto space-y-1 text-sm font-mono">
        <div v-for="(p, i) in linkProgress" :key="i" class="flex items-center gap-2">
          <Badge
            :variant="p.status === 'linked' ? 'default' : p.status === 'skipped' ? 'secondary' : 'destructive'"
            class="text-xs w-16 justify-center"
          >
            {{ p.status }}
          </Badge>
          <span class="truncate">{{ p.file }}</span>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
