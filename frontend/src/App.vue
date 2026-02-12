<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import { Toaster } from '@/components/ui/sonner'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

onMounted(async () => {
  await auth.checkAuth()
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
    <router-view />
  </div>

  <div v-else class="flex h-screen overflow-hidden">
    <AppSidebar />
    <main class="flex-1 overflow-auto p-6">
      <router-view />
    </main>
  </div>

  <Toaster richColors position="top-right" />
</template>
