<script setup lang="ts">
import { computed } from 'vue'
import type { UnlinkPreview } from '@/lib/types'
import { AlertTriangle } from 'lucide-vue-next'

const props = defineProps<{
  preview: UnlinkPreview
  showZeroNote?: boolean
}>()

const unsafeCount = computed(() => props.preview.unsafeFiles?.length ?? 0)
const safeCount = computed(() => props.preview.safeFiles?.length ?? 0)
const hasUnsafe = computed(() => unsafeCount.value > 0)
</script>

<template>
  <div class="space-y-3">
    <p>
      <strong>{{ preview.totalFiles }}</strong>
      video file{{ preview.totalFiles !== 1 ? 's' : '' }} will be removed from your library.
    </p>

    <!-- Unsafe files warning -->
    <div
      v-if="hasUnsafe"
      class="rounded-md border border-destructive/50 bg-destructive/10 p-3 space-y-2"
    >
      <div class="flex items-center gap-2 text-destructive font-medium">
        <AlertTriangle class="h-4 w-4" />
        Data loss warning
      </div>
      <p class="text-sm">
        <strong>{{ unsafeCount }}</strong>
        {{ unsafeCount !== 1 ? 'files are' : 'file is' }} the only copy &mdash;
        the original in your downloads is gone.
        Removing {{ unsafeCount !== 1 ? 'them' : 'it' }} means
        {{ unsafeCount !== 1 ? "they're" : "it's" }} gone for good.
      </p>
    </div>

    <!-- Safe files note -->
    <div v-if="safeCount > 0" class="text-sm text-muted-foreground">
      {{ safeCount }} {{ safeCount !== 1 ? 'files' : 'file' }} still
      {{ safeCount !== 1 ? 'have' : 'has' }} the original in downloads.
    </div>

    <!-- Zero files note (undo dialog) -->
    <div v-if="showZeroNote && preview.totalFiles === 0" class="text-sm text-muted-foreground">
      All files are already gone. This will just remove the history entry.
    </div>
  </div>
</template>
