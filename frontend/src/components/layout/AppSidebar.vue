<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  LayoutDashboard,
  Library,
  Link,
  Download,
  Rss,
  History,
  Settings,
  LogOut,
  Sun,
  Moon,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isDark = ref(false)

const navItems = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  { name: 'Library', path: '/library', icon: Library },
  { name: 'Link', path: '/link', icon: Link },
  { name: 'Downloads', path: '/downloads', icon: Download },
  { name: 'RSS Watch', path: '/rss', icon: Rss },
  { name: 'History', path: '/history', icon: History },
]

onMounted(() => {
  isDark.value = localStorage.getItem('theme') === 'dark' ||
    (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches)
  applyTheme()
})

function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  applyTheme()
}

function applyTheme() {
  document.documentElement.classList.toggle('dark', isDark.value)
}

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
  <aside class="flex h-full w-56 flex-col bg-sidebar text-sidebar-foreground border-r border-sidebar-border">
    <!-- Logo -->
    <div class="flex items-center gap-2 p-4">
      <Link class="h-6 w-6 text-sidebar-primary" />
      <span class="text-xl font-display tracking-wider uppercase">link-anime</span>
    </div>

    <Separator class="bg-sidebar-border" />

    <!-- Navigation -->
    <nav class="flex-1 space-y-1 p-2">
      <Button
        v-for="item in navItems"
        :key="item.path"
        :variant="isActive(item.path) ? 'secondary' : 'ghost'"
        :class="[
          'w-full justify-start gap-2',
          isActive(item.path)
            ? 'bg-sidebar-accent text-sidebar-accent-foreground'
            : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
        ]"
        @click="router.push(item.path)"
      >
        <component :is="item.icon" class="h-4 w-4" />
        {{ item.name }}
      </Button>
    </nav>

    <Separator class="bg-sidebar-border" />

    <!-- Bottom actions -->
    <div class="space-y-1 p-2">
      <Button
        variant="ghost"
        :class="[
          'w-full justify-start gap-2',
          isActive('/settings')
            ? 'bg-sidebar-accent text-sidebar-accent-foreground'
            : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
        ]"
        @click="router.push('/settings')"
      >
        <Settings class="h-4 w-4" />
        Settings
      </Button>
      <Button
        variant="ghost"
        class="w-full justify-start gap-2 text-sidebar-foreground/60 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
        @click="toggleTheme"
      >
        <Sun v-if="isDark" class="h-4 w-4" />
        <Moon v-else class="h-4 w-4" />
        {{ isDark ? 'Light Mode' : 'Dark Mode' }}
      </Button>
      <Button
        variant="ghost"
        class="w-full justify-start gap-2 text-sidebar-foreground/60 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
        @click="handleLogout"
      >
        <LogOut class="h-4 w-4" />
        Logout
      </Button>
    </div>
  </aside>
</template>
