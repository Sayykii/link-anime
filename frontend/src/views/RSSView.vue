<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useApi } from '@/composables/useApi'
import type { RSSRule, RSSMatch } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import {
  Rss,
  Plus,
  Pencil,
  Trash2,
  RefreshCw,
  Power,
  PowerOff,
  Loader2,
  Eraser,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const api = useApi()

const rules = ref<RSSRule[]>([])
const matches = ref<RSSMatch[]>([])
const loading = ref(true)
const polling = ref(false)
const activeTab = ref('rules')
const filterRuleId = ref<number | null>(null)

// Dialog state
const showRuleDialog = ref(false)
const editingRule = ref<RSSRule | null>(null)
const savingRule = ref(false)

// Delete confirmation
const showDeleteDialog = ref(false)
const deletingRule = ref<RSSRule | null>(null)

// Clear matches confirmation
const showClearDialog = ref(false)
const clearingRule = ref<RSSRule | null>(null)

// Form fields
const form = ref({
  name: '',
  query: '',
  showName: '',
  season: 1,
  mediaType: 'series' as 'series' | 'movie',
  minSeeders: 1,
  resolution: '',
  enabled: true,
})

const filteredMatches = computed(() => {
  if (!filterRuleId.value) return matches.value
  return matches.value.filter(m => m.ruleId === filterRuleId.value)
})

onMounted(async () => {
  await loadData()
})

async function loadData() {
  loading.value = true
  try {
    const [rulesData, matchesData] = await Promise.all([
      api.listRSSRules(),
      api.listRSSMatches(),
    ])
    rules.value = rulesData
    matches.value = matchesData
  } catch (err: any) {
    toast.error('Failed to load RSS data', { description: err.message })
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingRule.value = null
  form.value = {
    name: '',
    query: '',
    showName: '',
    season: 1,
    mediaType: 'series',
    minSeeders: 1,
    resolution: '',
    enabled: true,
  }
  showRuleDialog.value = true
}

function openEditDialog(rule: RSSRule) {
  editingRule.value = rule
  form.value = {
    name: rule.name,
    query: rule.query,
    showName: rule.showName,
    season: rule.season,
    mediaType: rule.mediaType as 'series' | 'movie',
    minSeeders: rule.minSeeders,
    resolution: rule.resolution,
    enabled: rule.enabled,
  }
  showRuleDialog.value = true
}

async function saveRule() {
  if (!form.value.name || !form.value.query || !form.value.showName) {
    toast.error('Please fill in name, query, and show name')
    return
  }

  savingRule.value = true
  try {
    if (editingRule.value) {
      await api.updateRSSRule({
        ...editingRule.value,
        ...form.value,
      })
      toast.success('Rule updated')
    } else {
      await api.createRSSRule(form.value)
      toast.success('Rule created')
    }
    showRuleDialog.value = false
    await loadData()
  } catch (err: any) {
    toast.error('Failed to save rule', { description: err.message })
  } finally {
    savingRule.value = false
  }
}

async function toggleRule(rule: RSSRule) {
  try {
    await api.toggleRSSRule(rule.id, !rule.enabled)
    rule.enabled = !rule.enabled
    toast.success(rule.enabled ? 'Rule enabled' : 'Rule disabled')
  } catch (err: any) {
    toast.error('Failed to toggle rule', { description: err.message })
  }
}

function confirmDelete(rule: RSSRule) {
  deletingRule.value = rule
  showDeleteDialog.value = true
}

async function deleteRule() {
  if (!deletingRule.value) return
  try {
    await api.deleteRSSRule(deletingRule.value.id)
    toast.success('Rule deleted')
    showDeleteDialog.value = false
    await loadData()
  } catch (err: any) {
    toast.error('Failed to delete rule', { description: err.message })
  }
}

function confirmClear(rule: RSSRule) {
  clearingRule.value = rule
  showClearDialog.value = true
}

async function clearMatches() {
  if (!clearingRule.value) return
  try {
    await api.clearRSSMatches(clearingRule.value.id)
    toast.success('Matches cleared')
    showClearDialog.value = false
    await loadData()
  } catch (err: any) {
    toast.error('Failed to clear matches', { description: err.message })
  }
}

async function pollNow() {
  polling.value = true
  try {
    await api.rssPollNow()
    toast.success('Poll triggered — checking feeds now')
    // Refresh data after a short delay to allow poll to complete
    setTimeout(() => loadData(), 3000)
  } catch (err: any) {
    toast.error('Failed to trigger poll', { description: err.message })
  } finally {
    polling.value = false
  }
}

function formatDate(dateStr: string | undefined) {
  if (!dateStr) return 'Never'
  const d = new Date(dateStr)
  return d.toLocaleString()
}

function statusColor(status: string) {
  switch (status) {
    case 'downloaded': return 'default'
    case 'linked': return 'secondary'
    case 'failed': return 'destructive'
    case 'pending': return 'outline'
    default: return 'default'
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-3xl tracking-wider uppercase">RSS Watch</h1>
        <p class="text-muted-foreground text-sm mt-1">
          Auto-download new episodes from Nyaa RSS feeds
        </p>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" size="sm" :disabled="polling" @click="pollNow">
          <Loader2 v-if="polling" class="mr-2 h-4 w-4 animate-spin" />
          <RefreshCw v-else class="mr-2 h-4 w-4" />
          Poll Now
        </Button>
        <Button size="sm" @click="openCreateDialog">
          <Plus class="mr-2 h-4 w-4" />
          New Rule
        </Button>
      </div>
    </div>

    <Tabs v-model="activeTab">
      <TabsList>
        <TabsTrigger value="rules">
          <Rss class="mr-2 h-4 w-4" />
          Rules ({{ rules.length }})
        </TabsTrigger>
        <TabsTrigger value="matches">
          Matches ({{ matches.length }})
        </TabsTrigger>
      </TabsList>

      <!-- Rules Tab -->
      <TabsContent value="rules" class="space-y-4">
        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        </div>

        <div v-else-if="rules.length === 0" class="text-center py-12 text-muted-foreground">
          <Rss class="h-12 w-12 mx-auto mb-4 opacity-40" />
          <p class="text-lg">No RSS rules yet</p>
          <p class="text-sm mt-1">Create a rule to auto-download new episodes from Nyaa</p>
          <Button class="mt-4" @click="openCreateDialog">
            <Plus class="mr-2 h-4 w-4" />
            Create First Rule
          </Button>
        </div>

        <div v-else class="grid gap-4">
          <Card v-for="rule in rules" :key="rule.id" :class="{ 'opacity-50': !rule.enabled }">
            <CardHeader class="pb-3">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <CardTitle class="text-base">{{ rule.name }}</CardTitle>
                  <Badge v-if="rule.enabled" variant="default" class="text-xs">Active</Badge>
                  <Badge v-else variant="secondary" class="text-xs">Disabled</Badge>
                  <Badge v-if="rule.matchCount > 0" variant="outline" class="text-xs">
                    {{ rule.matchCount }} match{{ rule.matchCount !== 1 ? 'es' : '' }}
                  </Badge>
                </div>
                <div class="flex gap-1">
                  <Button variant="ghost" size="icon" class="h-8 w-8" @click="toggleRule(rule)"
                    :title="rule.enabled ? 'Disable' : 'Enable'">
                    <Power v-if="rule.enabled" class="h-4 w-4" />
                    <PowerOff v-else class="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" class="h-8 w-8" @click="openEditDialog(rule)" title="Edit">
                    <Pencil class="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" class="h-8 w-8" @click="confirmClear(rule)"
                    title="Clear matches" :disabled="rule.matchCount === 0">
                    <Eraser class="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" class="h-8 w-8 text-destructive hover:text-destructive"
                    @click="confirmDelete(rule)" title="Delete">
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <CardDescription class="mt-1">
                Query: <code class="bg-muted px-1 py-0.5 rounded text-xs">{{ rule.query }}</code>
                <span v-if="rule.resolution" class="ml-2">
                  Res: <code class="bg-muted px-1 py-0.5 rounded text-xs">{{ rule.resolution }}</code>
                </span>
                <span class="ml-2">
                  Min seeders: <code class="bg-muted px-1 py-0.5 rounded text-xs">{{ rule.minSeeders }}</code>
                </span>
              </CardDescription>
            </CardHeader>
            <CardContent class="pt-0">
              <div class="flex gap-6 text-sm text-muted-foreground">
                <div>
                  <span class="font-medium text-foreground">Show:</span> {{ rule.showName }}
                </div>
                <div v-if="rule.mediaType === 'series'">
                  <span class="font-medium text-foreground">Season:</span> {{ rule.season }}
                </div>
                <div>
                  <span class="font-medium text-foreground">Type:</span> {{ rule.mediaType }}
                </div>
                <div>
                  <span class="font-medium text-foreground">Last check:</span> {{ formatDate(rule.lastCheck) }}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </TabsContent>

      <!-- Matches Tab -->
      <TabsContent value="matches" class="space-y-4">
        <!-- Filter by rule -->
        <div class="flex items-center gap-4">
          <Label>Filter by rule:</Label>
          <Select v-model="filterRuleId as any">
            <SelectTrigger class="w-64">
              <SelectValue placeholder="All rules" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="null as any">All rules</SelectItem>
              <SelectItem v-for="rule in rules" :key="rule.id" :value="rule.id as any">
                {{ rule.name }}
              </SelectItem>
            </SelectContent>
          </Select>
          <Button v-if="filterRuleId" variant="outline" size="sm" @click="filterRuleId = null">
            Clear filter
          </Button>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        </div>

        <div v-else-if="filteredMatches.length === 0" class="text-center py-12 text-muted-foreground">
          <p class="text-lg">No matches yet</p>
          <p class="text-sm mt-1">Matches appear when RSS polls find new torrents matching your rules</p>
        </div>

        <div v-else class="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead class="w-32">Rule</TableHead>
                <TableHead class="w-28">Status</TableHead>
                <TableHead class="w-40">Matched</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="match in filteredMatches" :key="match.id">
                <TableCell class="max-w-md truncate font-mono text-xs" :title="match.title">
                  {{ match.title }}
                </TableCell>
                <TableCell>
                  <Badge variant="outline" class="text-xs">{{ match.ruleName }}</Badge>
                </TableCell>
                <TableCell>
                  <Badge :variant="statusColor(match.status) as any" class="text-xs capitalize">
                    {{ match.status }}
                  </Badge>
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  {{ formatDate(match.matched) }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </TabsContent>
    </Tabs>

    <!-- Create/Edit Rule Dialog -->
    <Dialog v-model:open="showRuleDialog">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{{ editingRule ? 'Edit Rule' : 'Create Rule' }}</DialogTitle>
          <DialogDescription>
            {{ editingRule ? 'Update this RSS watch rule.' : 'Create a new RSS watch rule to auto-download from Nyaa.' }}
          </DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <Label for="name">Rule Name</Label>
            <Input id="name" v-model="form.name" placeholder="e.g. Dandadan Weekly" />
          </div>

          <div class="grid gap-2">
            <Label for="query">Nyaa Search Query</Label>
            <Input id="query" v-model="form.query" placeholder="e.g. [SubsPlease] Dandadan 1080p" />
            <p class="text-xs text-muted-foreground">Same query you'd type on nyaa.si search</p>
          </div>

          <Separator />

          <div class="grid gap-2">
            <Label for="showName">Show Name (Library folder)</Label>
            <Input id="showName" v-model="form.showName" placeholder="e.g. Dandadan" />
            <p class="text-xs text-muted-foreground">The folder name in your media library</p>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label>Media Type</Label>
              <Select v-model="form.mediaType">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="series">Series</SelectItem>
                  <SelectItem value="movie">Movie</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div v-if="form.mediaType === 'series'" class="grid gap-2">
              <Label for="season">Season Number</Label>
              <Input id="season" type="number" v-model.number="form.season" min="1" />
            </div>
          </div>

          <Separator />

          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label for="minSeeders">Min Seeders</Label>
              <Input id="minSeeders" type="number" v-model.number="form.minSeeders" min="0" />
            </div>
            <div class="grid gap-2">
              <Label for="resolution">Resolution Filter</Label>
              <Select v-model="form.resolution">
                <SelectTrigger>
                  <SelectValue placeholder="Any" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">Any</SelectItem>
                  <SelectItem value="1080p">1080p</SelectItem>
                  <SelectItem value="720p">720p</SelectItem>
                  <SelectItem value="480p">480p</SelectItem>
                  <SelectItem value="2160p">4K (2160p)</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" @click="showRuleDialog = false">Cancel</Button>
          <Button @click="saveRule" :disabled="savingRule">
            <Loader2 v-if="savingRule" class="mr-2 h-4 w-4 animate-spin" />
            {{ editingRule ? 'Update' : 'Create' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirmation -->
    <AlertDialog v-model:open="showDeleteDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Rule</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete "{{ deletingRule?.name }}"?
            This will also delete all {{ deletingRule?.matchCount || 0 }} match records.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction @click="deleteRule" class="bg-destructive text-destructive-foreground hover:bg-destructive/90">
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- Clear Matches Confirmation -->
    <AlertDialog v-model:open="showClearDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Clear Matches</AlertDialogTitle>
          <AlertDialogDescription>
            Clear all match history for "{{ clearingRule?.name }}"?
            The poller will re-detect previously matched torrents on the next poll.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction @click="clearMatches">Clear</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
