<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useApi } from '@/composables/useApi'
import type { Settings } from '@/lib/types'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { toast } from 'vue-sonner'
import { Save, TestTube, KeyRound, Loader2, Copy, RefreshCw, Trash2, Puzzle } from 'lucide-vue-next'

const api = useApi()
const settings = ref<Settings>({
  qbitUrl: '',
  qbitUser: '',
  qbitPass: '',
  qbitCategory: '',
  shokoUrl: '',
  shokoApiKey: '',
  notifyUrl: '',
  downloadDir: '',
  mediaDir: '',
  moviesDir: '',
})
const loading = ref(false)
const saving = ref(false)

// Password change
const currentPassword = ref('')
const newPassword = ref('')
const changingPassword = ref(false)
const testingQbit = ref(false)
const testingShoko = ref(false)
const apiKey = ref('')
const generatingKey = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const [s, key] = await Promise.all([api.getSettings(), api.getAPIKey()])
    settings.value = s
    apiKey.value = key.apiKey
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to load settings')
  } finally {
    loading.value = false
  }
})

async function saveSettings() {
  saving.value = true
  try {
    await api.updateSettings(settings.value)
    toast.success('Settings saved')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to save settings')
  } finally {
    saving.value = false
  }
}

async function testQbit() {
  testingQbit.value = true
  try {
    await api.testQbit()
    toast.success('qBittorrent connection successful')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'qBittorrent test failed')
  } finally {
    testingQbit.value = false
  }
}

async function testShoko() {
  testingShoko.value = true
  try {
    await api.testShoko()
    toast.success('Shoko connection successful')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Shoko test failed')
  } finally {
    testingShoko.value = false
  }
}

async function generateAPIKey() {
  generatingKey.value = true
  try {
    const res = await api.generateAPIKey()
    apiKey.value = res.apiKey
    toast.success('API key generated')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to generate API key')
  } finally {
    generatingKey.value = false
  }
}

async function revokeAPIKey() {
  try {
    await api.deleteAPIKey()
    apiKey.value = ''
    toast.success('API key revoked')
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to revoke API key')
  }
}

function copyAPIKey() {
  navigator.clipboard.writeText(apiKey.value)
  toast.success('Copied to clipboard')
}

async function changePassword() {
  if (!currentPassword.value || !newPassword.value) return
  changingPassword.value = true
  try {
    await api.changePassword(currentPassword.value, newPassword.value)
    toast.success('Password changed')
    currentPassword.value = ''
    newPassword.value = ''
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : 'Failed to change password')
  } finally {
    changingPassword.value = false
  }
}
</script>

<template>
  <div class="space-y-6 max-w-2xl">
    <div>
      <h1 class="text-3xl font-bold">Settings</h1>
      <p class="text-muted-foreground">Configure paths, integrations, and notifications</p>
    </div>

    <div v-if="loading" class="flex items-center gap-2 text-muted-foreground py-8 justify-center">
      <Loader2 class="h-4 w-4 animate-spin" />
      Loading settings...
    </div>

    <template v-else>
      <!-- Paths -->
      <Card glass>
        <CardHeader>
          <CardTitle>Paths</CardTitle>
          <CardDescription>Directories for downloads and media library</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>Download Directory</Label>
            <Input v-model="settings.downloadDir" placeholder="/data/downloads/complete/anime" />
          </div>
          <div class="space-y-2">
            <Label>Media Directory (Series)</Label>
            <Input v-model="settings.mediaDir" placeholder="/data/media/anime" />
          </div>
          <div class="space-y-2">
            <Label>Movies Directory</Label>
            <Input v-model="settings.moviesDir" placeholder="/data/media/anime-movies" />
          </div>
        </CardContent>
      </Card>

      <!-- qBittorrent -->
      <Card glass>
        <CardHeader>
          <CardTitle>qBittorrent</CardTitle>
          <CardDescription>Connect to qBittorrent for download management</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>URL</Label>
            <Input v-model="settings.qbitUrl" placeholder="http://qbittorrent:8080" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label>Username</Label>
              <Input v-model="settings.qbitUser" />
            </div>
            <div class="space-y-2">
              <Label>Password</Label>
              <Input v-model="settings.qbitPass" type="password" />
            </div>
          </div>
          <div class="space-y-2">
            <Label>Category</Label>
            <Input v-model="settings.qbitCategory" placeholder="anime" />
          </div>
          <Button variant="outline" size="sm" @click="testQbit" :disabled="testingQbit" class="gap-2">
            <Loader2 v-if="testingQbit" class="h-4 w-4 animate-spin" />
            <TestTube v-else class="h-4 w-4" />
            {{ testingQbit ? 'Testing...' : 'Test Connection' }}
          </Button>
        </CardContent>
      </Card>

      <!-- Shoko -->
      <Card glass>
        <CardHeader>
          <CardTitle>Shoko Server</CardTitle>
          <CardDescription>Trigger import scans after linking</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>URL</Label>
            <Input v-model="settings.shokoUrl" placeholder="http://shoko:8111" />
          </div>
          <div class="space-y-2">
            <Label>API Key</Label>
            <Input v-model="settings.shokoApiKey" type="password" />
          </div>
          <Button variant="outline" size="sm" @click="testShoko" :disabled="testingShoko" class="gap-2">
            <Loader2 v-if="testingShoko" class="h-4 w-4 animate-spin" />
            <TestTube v-else class="h-4 w-4" />
            {{ testingShoko ? 'Testing...' : 'Test Connection' }}
          </Button>
        </CardContent>
      </Card>

      <!-- Notifications -->
      <Card glass>
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>Receive notifications when content is linked (Discord, ntfy, or generic webhook)</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>Webhook URL</Label>
            <Input v-model="settings.notifyUrl" placeholder="https://discord.com/api/webhooks/... or https://ntfy.sh/topic" />
          </div>
        </CardContent>
      </Card>

      <!-- Save -->
      <Button @click="saveSettings" :disabled="saving" class="gap-2">
        <Save class="h-4 w-4" />
        {{ saving ? 'Saving...' : 'Save Settings' }}
      </Button>

      <Separator />

      <!-- API Key -->
      <Card glass>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Puzzle class="h-5 w-5" />
            API Key
          </CardTitle>
          <CardDescription>
            Generate an API key for external tools like the Tampermonkey userscript to add torrents from nyaa.si
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <template v-if="apiKey">
            <div class="flex items-center gap-2">
              <Input :model-value="apiKey" readonly class="font-mono text-xs" />
              <Button variant="outline" size="icon" @click="copyAPIKey" title="Copy">
                <Copy class="h-4 w-4" />
              </Button>
            </div>
            <div class="flex gap-2">
              <Button variant="outline" size="sm" @click="generateAPIKey" :disabled="generatingKey" class="gap-2">
                <RefreshCw class="h-4 w-4" />
                Regenerate
              </Button>
              <Button variant="destructive" size="sm" @click="revokeAPIKey" class="gap-2">
                <Trash2 class="h-4 w-4" />
                Revoke
              </Button>
            </div>
          </template>
          <template v-else>
            <p class="text-sm text-muted-foreground">No API key generated yet.</p>
            <Button variant="outline" size="sm" @click="generateAPIKey" :disabled="generatingKey" class="gap-2">
              <KeyRound class="h-4 w-4" />
              {{ generatingKey ? 'Generating...' : 'Generate API Key' }}
            </Button>
          </template>
        </CardContent>
      </Card>

      <!-- Change Password -->
      <Card glass>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <KeyRound class="h-5 w-5" />
            Change Password
          </CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>Current Password</Label>
            <Input v-model="currentPassword" type="password" />
          </div>
          <div class="space-y-2">
            <Label>New Password</Label>
            <Input v-model="newPassword" type="password" />
          </div>
          <Button
            variant="outline"
            @click="changePassword"
            :disabled="!currentPassword || !newPassword || changingPassword"
          >
            {{ changingPassword ? 'Changing...' : 'Change Password' }}
          </Button>
        </CardContent>
      </Card>
    </template>
  </div>
</template>
