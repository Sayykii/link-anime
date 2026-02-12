<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useApi } from '@/composables/useApi'
import type { HistoryEntry, LinkResult } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { toast } from 'vue-sonner'
import { Undo2, RefreshCw, Clock } from 'lucide-vue-next'

const api = useApi()
const history = ref<HistoryEntry[]>([])
const loading = ref(false)

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

async function handleUndo() {
  try {
    const { result, entry } = await api.undo()
    toast.success(`Undid: ${entry.showName} (${result.linked} files removed)`)
    await loadHistory()
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Undo failed')
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
        <AlertDialog>
          <AlertDialogTrigger as-child>
            <Button variant="outline" size="sm" class="gap-2" :disabled="!history.length">
              <Undo2 class="h-4 w-4" />
              Undo Last
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Undo Last Link?</AlertDialogTitle>
              <AlertDialogDescription v-if="history.length">
                This will remove the hardlinks created for "{{ history[0]?.showName }}"
                ({{ history[0]?.fileCount }} files). The original download files are not affected.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction @click="handleUndo">Undo</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        <Button variant="outline" size="sm" @click="loadHistory" class="gap-2">
          <RefreshCw class="h-4 w-4" />
          Refresh
        </Button>
      </div>
    </div>

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
