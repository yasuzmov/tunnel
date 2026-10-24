import { onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'

export function useSSE(handlers: Record<string, (payload: unknown) => void>) {
  let source: EventSource | null = null

  function connect() {
    const token = useAuthStore().token
    source = new EventSource(`/api/events?token=${encodeURIComponent(token)}`)
    for (const [type, fn] of Object.entries(handlers)) {
      source.addEventListener(type, (ev) => {
        const msg = ev as MessageEvent<string>
        try {
          fn(JSON.parse(msg.data))
        } catch {
          fn(msg.data)
        }
      })
    }
  }

  onMounted(connect)
  onUnmounted(() => source?.close())
}
