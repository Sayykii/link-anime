<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import CommandPalette from '@/components/CommandPalette.vue'
import { Toaster } from '@/components/ui/sonner'
import { Sheet, SheetContent, SheetTrigger, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Menu } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const commandOpen = ref(false)
const mobileOpen = ref(false)

function handleKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    if (!auth.authenticated) return
    e.preventDefault()
    commandOpen.value = !commandOpen.value
  }
}

onMounted(async () => {
  // Apply saved theme
  const saved = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  document.documentElement.classList.toggle('dark', saved === 'dark' || (!saved && prefersDark))

  document.addEventListener('keydown', handleKeydown)
  await auth.checkAuth()
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

watch(() => auth.authenticated, (isAuth) => {
  if (!isAuth && !auth.checking && route.meta.public !== true) {
    router.push('/login')
  }
})

watch(() => auth.checking, (isChecking) => {
  if (!isChecking && !auth.authenticated && route.meta.public !== true) {
    router.push('/login')
  }
})

// Close mobile drawer and scroll to top on navigation
watch(() => route.path, () => {
  mobileOpen.value = false
  const main = document.querySelector('main')
  if (main) main.scrollTo({ top: 0 })
})
</script>

<template>
  <div v-if="auth.checking" class="flex h-screen items-center justify-center">
    <div class="text-muted-foreground">Loading...</div>
  </div>

  <div v-else-if="!auth.authenticated || route.meta.public" class="min-h-screen">
    <router-view v-slot="{ Component }">
      <Transition name="page" mode="out-in">
        <component :is="Component" :key="route.path" />
      </Transition>
    </router-view>
  </div>

  <div v-else class="flex h-screen overflow-hidden">
    <!-- Desktop sidebar: hidden on mobile -->
    <div class="hidden md:block">
      <AppSidebar @open-command="commandOpen = true" />
    </div>

    <main class="flex-1 overflow-auto">
      <!-- Mobile header bar -->
      <div class="sticky top-0 z-20 flex items-center gap-3 border-b bg-background/80 backdrop-blur-sm p-3 md:hidden">
        <Sheet v-model:open="mobileOpen">
          <SheetTrigger as-child>
            <Button variant="ghost" size="icon" class="h-9 w-9">
              <Menu class="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" class="w-56 p-0">
            <SheetTitle class="sr-only">Navigation</SheetTitle>
            <AppSidebar @open-command="commandOpen = true; mobileOpen = false" />
          </SheetContent>
        </Sheet>
        <span class="font-display text-lg tracking-wider uppercase">link-anime</span>
      </div>

      <div class="p-6">
        <router-view v-slot="{ Component }">
          <Transition name="page" mode="out-in">
            <component :is="Component" :key="route.path" />
          </Transition>
        </router-view>
      </div>
    </main>
  </div>

  <CommandPalette v-model="commandOpen" />
  <Toaster richColors position="top-right" />
</template>
