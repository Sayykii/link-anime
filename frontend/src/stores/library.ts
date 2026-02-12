import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '@/composables/useApi'
import type { Show, Movie, DownloadItem, LibraryStats } from '@/lib/types'

export const useLibraryStore = defineStore('library', () => {
  const api = useApi()

  const shows = ref<Show[]>([])
  const movies = ref<Movie[]>([])
  const downloads = ref<DownloadItem[]>([])
  const stats = ref<LibraryStats | null>(null)
  const loading = ref(false)

  async function fetchShows() {
    loading.value = true
    try {
      shows.value = await api.getShows()
    } finally {
      loading.value = false
    }
  }

  async function fetchMovies() {
    loading.value = true
    try {
      movies.value = await api.getMovies()
    } finally {
      loading.value = false
    }
  }

  async function fetchDownloads() {
    loading.value = true
    try {
      downloads.value = await api.getDownloads()
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      stats.value = await api.getStats()
    } catch {
      // ignore
    }
  }

  async function refreshAll() {
    await Promise.all([fetchShows(), fetchMovies(), fetchStats()])
  }

  return {
    shows,
    movies,
    downloads,
    stats,
    loading,
    fetchShows,
    fetchMovies,
    fetchDownloads,
    fetchStats,
    refreshAll,
  }
})
