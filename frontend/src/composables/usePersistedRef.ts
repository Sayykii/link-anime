import { ref, watch, type Ref } from 'vue'

export function usePersistedRef<T>(key: string, defaultValue: T): Ref<T> {
  const storageKey = `la:${key}`
  let initial = defaultValue

  try {
    const stored = localStorage.getItem(storageKey)
    if (stored !== null) {
      initial = JSON.parse(stored)
    }
  } catch {
    // corrupt or missing — use default
  }

  const value = ref(initial) as Ref<T>

  watch(value, (v) => {
    try {
      localStorage.setItem(storageKey, JSON.stringify(v))
    } catch {
      // storage full or unavailable
    }
  }, { deep: true })

  return value
}
