<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'
import type { HistoryEntry, UnlinkPreview } from '@/lib/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { toast } from 'vue-sonner'
import { Undo2, RefreshCw, Clock, Loader2, AlertTriangle } from 'lucide-vue-next'

const api = useApi()
const history = ref<HistoryEntry[]>([])
const loading = ref(false)

// Undo safety state
const undoDialogOpen = ref(false)
const undoPreview = ref<UnlinkPreview | null>(null)
const undoEntry = ref<HistoryEntry | null>(null)
const undoLoading = ref(false)
const undoExecuting = ref(false)

onMounted(() => loadHistory())

async function loadHistory() {
  loading.value = true
  try {
    history.value = await api.getHistory(100)
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to load history')
  } finally {
    loading.value = false
  }
}

async function openUndoDialog() {
  undoPreview.value = null
  undoEntry.value = null
  undoLoading.value = true
  undoDialogOpen.value = true

  try {
    const { preview, entry } = await api.undoPreview()
    undoPreview.value = preview
    undoEntry.value = entry
  } catch (err: any) {
    toast.error('Failed to check undo safety', { description: err.message })
    undoDialogOpen.value = false
  } finally {
    undoLoading.value = false
  }
}

const hasUnsafeFiles = computed(() => {
  return undoPreview.value && undoPreview.value.unsafeFiles && undoPreview.value.unsafeFiles.length > 0
})

async function handleUndo(force: boolean) {
  undoExecuting.value = true
  try {
    const { result, entry } = await api.undo(force)
    const removed = result.linked
    const skipped = result.skipped

    if (removed > 0) {
      toast.success(`Undid: ${entry.showName}`, {
        description: `Removed ${removed} file${removed !== 1 ? 's' : ''}${skipped > 0 ? `, skipped ${skipped} unsafe` : ''}`,
      })
    } else if (skipped > 0) {
      toast.warning('No files removed', {
        description: `${skipped} file${skipped !== 1 ? 's' : ''} skipped (only copy, no source)`,
      })
    } else {
      toast.success(`Undid: ${entry.showName}`, { description: 'History entry removed (files already gone)' })
    }

    undoDialogOpen.value = false
    await loadHistory()
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Undo failed')
  } finally {
    undoExecuting.value = false
  }
}

function formatDate(ts: string): string {
  return new Date(ts).toLocaleString()
}

function formatSize(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(2) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  return (bytes / 1024).toFixed(1) + ' KB'
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">History</h1>
        <p class="text-muted-foreground">Recent link operations</p>
      </div>
      <div class="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          class="gap-2"
          :disabled="!history.length"
          @click="openUndoDialog"
        >
          <Undo2 class="h-4 w-4" />
          Undo Last
        </Button>

        <Button variant="outline" size="sm" @click="loadHistory" class="gap-2">
          <RefreshCw class="h-4 w-4" />
          Refresh
        </Button>
      </div>
    </div>

    <!-- Undo confirmation dialog with safety check -->
    <AlertDialog v-model:open="undoDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Undo Last Link?</AlertDialogTitle>
          <AlertDialogDescription v-if="undoLoading" class="flex items-center gap-2">
            <Loader2 class="h-4 w-4 animate-spin" />
            Checking file safety...
          </AlertDialogDescription>
          <AlertDialogDescription v-else-if="undoPreview && undoEntry">
            <div class="space-y-3">
              <p>
                This will undo the link for "<strong>{{ undoEntry.showName }}</strong>"
                ({{ undoPreview.totalFiles }} file{{ undoPreview.totalFiles !== 1 ? 's' : '' }} on disk).
              </p>

              <!-- Safety warning for files that are the only copy -->
              <div
                v-if="hasUnsafeFiles"
                class="rounded-md border border-destructive/50 bg-destructive/10 p-3 space-y-2"
              >
                <div class="flex items-center gap-2 text-destructive font-medium">
                  <AlertTriangle class="h-4 w-4" />
                  Data loss warning
                </div>
                <p class="text-sm">
                  <strong>{{ undoPreview.unsafeFiles!.length }}</strong>
                  file{{ undoPreview.unsafeFiles!.length !== 1 ? 's are' : ' is' }} the
                  <strong>only copy</strong> (source file in downloads no longer exists).
                  Removing {{ undoPreview.unsafeFiles!.length !== 1 ? 'them' : 'it' }} will cause
                  <strong>permanent data loss</strong>.
                </p>
              </div>

              <div v-if="undoPreview.safeFiles && undoPreview.safeFiles.length > 0" class="text-sm text-muted-foreground">
                {{ undoPreview.safeFiles.length }} file{{ undoPreview.safeFiles.length !== 1 ? 's are' : ' is' }}
                safe to remove (hardlinks with source still in downloads).
              </div>

              <div v-if="undoPreview.totalFiles === 0" class="text-sm text-muted-foreground">
                All files are already gone. This will just remove the history entry.
              </div>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter v-if="!undoLoading && undoPreview">
          <AlertDialogCancel :disabled="undoExecuting">Cancel</AlertDialogCancel>
          <!-- If there are unsafe files, show two options -->
          <template v-if="hasUnsafeFiles">
            <Button
              v-if="undoPreview!.safeFiles && undoPreview!.safeFiles.length > 0"
              variant="outline"
              @click="handleUndo(false)"
              :disabled="undoExecuting"
              class="gap-2"
            >
              <Loader2 v-if="undoExecuting" class="h-4 w-4 animate-spin" />
              Remove safe only
            </Button>
            <AlertDialogAction
              @click.prevent="handleUndo(true)"
              :disabled="undoExecuting"
              class="bg-destructive text-destructive-foreground hover:bg-destructive/90 gap-2"
            >
              <Loader2 v-if="undoExecuting" class="h-4 w-4 animate-spin" />
              Remove all (data loss)
            </AlertDialogAction>
          </template>
          <!-- All files are safe or all gone -->
          <AlertDialogAction
            v-else
            @click.prevent="handleUndo(false)"
            :disabled="undoExecuting"
            class="gap-2"
          >
            <Loader2 v-if="undoExecuting" class="h-4 w-4 animate-spin" />
            {{ undoPreview.totalFiles === 0 ? 'Remove entry' : 'Undo' }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Season</TableHead>
              <TableHead>Files</TableHead>
              <TableHead>Size</TableHead>
              <TableHead>Source</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="entry in history" :key="entry.id">
              <TableCell class="whitespace-nowrap text-sm">
                <div class="flex items-center gap-1">
                  <Clock class="h-3 w-3 text-muted-foreground" />
                  {{ formatDate(entry.timestamp) }}
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="outline" class="capitalize">{{ entry.mediaType }}</Badge>
              </TableCell>
              <TableCell class="font-medium">{{ entry.showName }}</TableCell>
              <TableCell>
                <span v-if="entry.season !== undefined && entry.season !== null">S{{ entry.season }}</span>
                <span v-else class="text-muted-foreground">-</span>
              </TableCell>
              <TableCell>{{ entry.fileCount }}</TableCell>
              <TableCell class="whitespace-nowrap">{{ formatSize(entry.totalSize) }}</TableCell>
              <TableCell class="max-w-48 truncate text-sm text-muted-foreground">{{ entry.source }}</TableCell>
            </TableRow>
            <TableRow v-if="!history.length && !loading">
              <TableCell colspan="7" class="text-center text-muted-foreground py-8">
                No history entries yet
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
