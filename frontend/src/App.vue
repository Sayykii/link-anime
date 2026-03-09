<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import CommandPalette from '@/components/CommandPalette.vue'
import { Toaster } from '@/components/ui/sonner'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const commandOpen = ref(false)

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
    <AppSidebar @open-command="commandOpen = true" />
    <main class="flex-1 overflow-auto p-6">
      <router-view v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" :key="route.path" />
        </Transition>
      </router-view>
    </main>
  </div>

  <CommandPalette v-model="commandOpen" />
  <Toaster richColors position="top-right" />
</template>
