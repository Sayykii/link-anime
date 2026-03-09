<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { formatSize } from '@/lib/utils'
import type { HistoryEntry, UnlinkPreview } from '@/lib/types'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
import { Undo2, RefreshCw, Clock, Loader2, AlertTriangle, Search, X, History as HistoryIcon } from 'lucide-vue-next'
import EmptyState from '@/components/EmptyState.vue'

const api = useApi()
const router = useRouter()
const history = ref<HistoryEntry[]>([])
const loading = ref(false)
const searchQuery = ref('')
const typeFilter = ref('all') // 'all' | 'series' | 'movie'
const sortOrder = ref('newest') // 'newest' | 'oldest'

const filteredHistory = computed(() => {
  let items = history.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    items = items.filter(e => e.showName.toLowerCase().includes(q) || e.source.toLowerCase().includes(q))
  }
  if (typeFilter.value !== 'all') {
    items = items.filter(e => e.mediaType === typeFilter.value)
  }
  const sorted = [...items]
  if (sortOrder.value === 'oldest') {
    sorted.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
  } else {
    sorted.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
  }
  return sorted
})

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

    <!-- Filter bar -->
    <div class="sticky-filter flex flex-col sm:flex-row sm:items-center gap-3">
      <div class="relative flex-1 max-w-sm">
        <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input v-model="searchQuery" placeholder="Search history..." class="pl-9 h-9" />
        <button
          v-if="searchQuery"
          class="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
          @click="searchQuery = ''"
        >
          <X class="h-4 w-4" />
        </button>
      </div>
      <div class="flex rounded-md border">
        <Button
          v-for="opt in [{ value: 'all', label: 'All' }, { value: 'series', label: 'Shows' }, { value: 'movie', label: 'Movies' }]"
          :key="opt.value"
          :variant="typeFilter === opt.value ? 'default' : 'ghost'"
          size="sm"
          class="rounded-none first:rounded-l-md last:rounded-r-md h-8 px-3 text-xs"
          @click="typeFilter = opt.value"
        >
          {{ opt.label }}
        </Button>
      </div>
      <div class="flex rounded-md border">
        <Button
          v-for="opt in [{ value: 'newest', label: 'Newest' }, { value: 'oldest', label: 'Oldest' }]"
          :key="opt.value"
          :variant="sortOrder === opt.value ? 'default' : 'ghost'"
          size="sm"
          class="rounded-none first:rounded-l-md last:rounded-r-md h-8 px-3 text-xs"
          @click="sortOrder = opt.value"
        >
          {{ opt.label }}
        </Button>
      </div>
    </div>

    <Card glass>
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
            <TableRow v-for="entry in filteredHistory" :key="entry.id">
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
            <TableRow v-if="!filteredHistory.length && !loading && !searchQuery && typeFilter === 'all'">
              <TableCell colspan="7">
                <EmptyState
                  :icon="HistoryIcon"
                  heading="No link history yet"
                  description="Your link operations will appear here"
                  action-label="Link New Content"
                  @action="router.push('/link')"
                />
              </TableCell>
            </TableRow>
            <TableRow v-if="!filteredHistory.length && (searchQuery || typeFilter !== 'all')">
              <TableCell colspan="7">
                <EmptyState
                  :icon="Search"
                  heading="No matching entries"
                  action-label="Clear filters"
                  action-variant="outline"
                  @action="searchQuery = ''; typeFilter = 'all'"
                />
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
