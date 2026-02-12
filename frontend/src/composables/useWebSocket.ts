import { ref, onUnmounted } from 'vue'
import type { WSMessage } from '@/lib/types'

export function useWebSocket() {
  const connected = ref(false)
  const lastMessage = ref<WSMessage | null>(null)
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  const listeners = new Map<string, Set<(data: unknown) => void>>()

  function connect() {
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${window.location.host}/api/ws`

    ws = new WebSocket(url)

    ws.onopen = () => {
      connected.value = true
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        lastMessage.value = msg

        // Dispatch to type-specific listeners
        const typeListeners = listeners.get(msg.type)
        if (typeListeners) {
          for (const fn of typeListeners) {
            fn(msg.data)
          }
        }

        // Also dispatch to wildcard listeners
        const allListeners = listeners.get('*')
        if (allListeners) {
          for (const fn of allListeners) {
            fn(msg)
          }
        }
      } catch {
        // Ignore parse errors
      }
    }

    ws.onclose = () => {
      connected.value = false
      // Auto-reconnect after 3 seconds
      reconnectTimer = setTimeout(connect, 3000)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
    connected.value = false
  }

  function on(type: string, callback: (data: unknown) => void) {
    if (!listeners.has(type)) {
      listeners.set(type, new Set())
    }
    listeners.get(type)!.add(callback)

    // Return cleanup function
    return () => {
      listeners.get(type)?.delete(callback)
    }
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    lastMessage,
    connect,
    disconnect,
    on,
  }
}
