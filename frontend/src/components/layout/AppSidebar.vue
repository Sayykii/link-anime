<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  LayoutDashboard,
  Library,
  Link,
  Download,
  History,
  Settings,
  LogOut,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const navItems = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  { name: 'Library', path: '/library', icon: Library },
  { name: 'Link', path: '/link', icon: Link },
  { name: 'Downloads', path: '/downloads', icon: Download },
  { name: 'History', path: '/history', icon: History },
]

async function handleLogout() {
  await auth.logout()
  router.push('/login')
}

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <aside class="flex h-full w-56 flex-col border-r bg-card">
    <!-- Logo -->
    <div class="flex items-center gap-2 p-4">
      <Link class="h-6 w-6 text-primary" />
      <span class="text-lg font-semibold">link-anime</span>
    </div>

    <Separator />

    <!-- Navigation -->
    <nav class="flex-1 space-y-1 p-2">
      <Button
        v-for="item in navItems"
        :key="item.path"
        :variant="isActive(item.path) ? 'secondary' : 'ghost'"
        class="w-full justify-start gap-2"
        @click="router.push(item.path)"
      >
        <component :is="item.icon" class="h-4 w-4" />
        {{ item.name }}
      </Button>
    </nav>

    <Separator />

    <!-- Bottom actions -->
    <div class="space-y-1 p-2">
      <Button
        :variant="isActive('/settings') ? 'secondary' : 'ghost'"
        class="w-full justify-start gap-2"
        @click="router.push('/settings')"
      >
        <Settings class="h-4 w-4" />
        Settings
      </Button>
      <Button
        variant="ghost"
        class="w-full justify-start gap-2 text-muted-foreground"
        @click="handleLogout"
      >
        <LogOut class="h-4 w-4" />
        Logout
      </Button>
    </div>
  </aside>
</template>
