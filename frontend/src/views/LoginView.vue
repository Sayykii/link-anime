<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Link } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const auth = useAuthStore()
const router = useRouter()
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!password.value) return
  loading.value = true
  try {
    await auth.login(password.value)
    router.push('/')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Login failed'
    toast.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-background">
    <Card class="w-full max-w-sm">
      <CardHeader class="text-center">
        <div class="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-lg bg-primary">
          <Link class="h-6 w-6 text-primary-foreground" />
        </div>
        <CardTitle class="text-2xl">link-anime</CardTitle>
        <CardDescription>Enter your password to continue</CardDescription>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="space-y-2">
            <Label for="password">Password</Label>
            <Input
              id="password"
              type="password"
              v-model="password"
              placeholder="Enter password"
              :disabled="loading"
              autofocus
            />
          </div>
          <Button type="submit" class="w-full" :disabled="loading || !password">
            {{ loading ? 'Signing in...' : 'Sign In' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
