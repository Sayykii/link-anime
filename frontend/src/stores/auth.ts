import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '@/composables/useApi'

export const useAuthStore = defineStore('auth', () => {
  const api = useApi()
  const authenticated = ref(false)
  const checking = ref(true)

  async function checkAuth() {
    checking.value = true
    try {
      const result = await api.checkAuth()
      authenticated.value = result.authenticated
    } catch {
      authenticated.value = false
    } finally {
      checking.value = false
    }
  }

  async function login(password: string) {
    await api.login(password)
    authenticated.value = true
  }

  async function logout() {
    try {
      await api.logout()
    } finally {
      authenticated.value = false
    }
  }

  return {
    authenticated,
    checking,
    checkAuth,
    login,
    logout,
  }
})
