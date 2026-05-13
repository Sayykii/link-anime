<script setup lang="ts">
import { ref, computed } from 'vue'
import type { LinkResult, LinkProgress } from '@/lib/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Check, ArrowRight, ChevronDown, ChevronUp } from 'lucide-vue-next'

const props = defineProps<{
  finalResult: LinkResult | null
  linkProgress: LinkProgress[]
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

const showFileDetails = ref(false)

const skippedFiles = computed(() =>
  props.linkProgress.filter(p => p.status === 'skipped')
)

const failedFiles = computed(() =>
  props.linkProgress.filter(p => p.status === 'failed')
)

const hasIssues = computed(() =>
  skippedFiles.value.length > 0 || failedFiles.value.length > 0
)
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
              <div class="text-2xl font-bold" :class="finalResult.skipped > 0 ? 'text-amber-500' : ''">{{ finalResult.skipped }}</div>
              <div class="text-sm text-muted-foreground">Skipped</div>
            </div>
            <div>
              <div class="text-2xl font-bold" :class="finalResult.failed > 0 ? 'text-destructive' : ''">{{ finalResult.failed }}</div>
              <div class="text-sm text-muted-foreground">Failed</div>
            </div>
          </div>
          <div class="text-sm mt-3">
            <span class="text-muted-foreground">Destination:</span>
            <code class="ml-1 text-xs bg-muted px-1 py-0.5 rounded">{{ finalResult.destDir }}</code>
          </div>
        </div>
      </div>

      <!-- File details (expandable when there are skipped/failed) -->
      <div v-if="hasIssues && linkProgress.length > 0">
        <button
          class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
          @click="showFileDetails = !showFileDetails"
        >
          <ChevronDown v-if="!showFileDetails" class="h-4 w-4" />
          <ChevronUp v-else class="h-4 w-4" />
          View file details
        </button>

        <div v-if="showFileDetails" class="mt-2 space-y-3">
          <div v-if="failedFiles.length > 0">
            <div class="text-xs font-medium text-destructive mb-1">Failed</div>
            <div class="max-h-32 overflow-auto rounded border bg-muted/30 p-2 space-y-0.5">
              <div
                v-for="(f, i) in failedFiles"
                :key="i"
                class="text-xs font-mono text-muted-foreground truncate"
                :title="f.file"
              >
                {{ f.file }}
              </div>
            </div>
          </div>

          <div v-if="skippedFiles.length > 0">
            <div class="text-xs font-medium text-amber-500 mb-1">Skipped (already exist)</div>
            <div class="max-h-32 overflow-auto rounded border bg-muted/30 p-2 space-y-0.5">
              <div
                v-for="(f, i) in skippedFiles"
                :key="i"
                class="text-xs font-mono text-muted-foreground truncate"
                :title="f.file"
              >
                {{ f.file }}
              </div>
            </div>
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
