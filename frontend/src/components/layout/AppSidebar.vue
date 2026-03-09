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
  Search,
} from 'lucide-vue-next'

defineEmits<{
  'open-command': []
}>()

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isDark = ref(false)
const logoClickCount = ref(0)
const showMenacing = ref(false)
let clickTimer: ReturnType<typeof setTimeout> | null = null

function handleLogoClick() {
  logoClickCount.value++
  if (clickTimer) clearTimeout(clickTimer)
  clickTimer = setTimeout(() => { logoClickCount.value = 0 }, 500)

  if (logoClickCount.value >= 3) {
    logoClickCount.value = 0
    showMenacing.value = true
    setTimeout(() => { showMenacing.value = false }, 2500)
  }
}

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

const isMac = typeof navigator !== 'undefined' && navigator.userAgent.includes('Mac')

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <aside class="flex h-full w-56 flex-col bg-sidebar text-sidebar-foreground border-r border-sidebar-border">
    <!-- Logo -->
    <div
      class="flex items-center gap-2 p-4 cursor-pointer group transition-all duration-300"
      @click="handleLogoClick"
    >
      <Link class="h-6 w-6 text-sidebar-primary transition-all duration-300 group-hover:drop-shadow-[0_0_8px_var(--sidebar-primary)]" />
      <span class="text-xl font-display tracking-wider uppercase transition-all duration-300 group-hover:text-sidebar-primary group-hover:drop-shadow-[0_0_6px_var(--sidebar-primary)]">link-anime</span>
      <span class="text-xs text-sidebar-primary/0 group-hover:text-sidebar-primary/30 transition-all duration-500 font-display select-none">&#x30B4;</span>
    </div>

    <!-- Full-screen menacing easter egg overlay -->
    <Teleport to="body">
      <Transition name="page">
        <div
          v-if="showMenacing"
          class="fixed inset-0 z-[99998] pointer-events-none flex items-center justify-center"
        >
          <div class="menacing-overlay text-[12rem] font-display text-primary/10 select-none leading-none tracking-widest">
            &#x30B4;&#x30B4;&#x30B4;
          </div>
        </div>
      </Transition>
    </Teleport>

    <Separator class="bg-sidebar-border" />

    <!-- Search hint -->
    <button
      class="mx-2 mt-2 flex items-center gap-2 rounded-md border border-sidebar-border bg-sidebar-accent/50 px-3 py-1.5 text-xs text-sidebar-foreground/50 transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
      @click="$emit('open-command')"
    >
      <Search class="h-3 w-3" />
      <span class="flex-1 text-left">Search...</span>
      <kbd class="pointer-events-none rounded border border-sidebar-border bg-sidebar px-1 py-0.5 font-mono text-[10px] leading-none text-sidebar-foreground/40">
        {{ isMac ? '\u2318' : 'Ctrl+' }}K
      </kbd>
    </button>

    <!-- Navigation -->
    <nav class="flex-1 space-y-1 p-2">
      <Button
        v-for="item in navItems"
        :key="item.path"
        :variant="isActive(item.path) ? 'secondary' : 'ghost'"
        :class="[
          'w-full justify-start gap-2 transition-all duration-200',
          isActive(item.path)
            ? 'bg-sidebar-accent text-sidebar-accent-foreground border-l-2 border-sidebar-primary'
            : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground border-l-2 border-transparent'
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
