export function formatTime(v?: string) {
  if (!v) {
    return '-'
  }
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) {
    return v
  }
  return d.toLocaleString()
}

export function formatBytes(n: number) {
  if (n < 1024) {
    return `${n} B`
  }
  if (n < 1024 * 1024) {
    return `${(n / 1024).toFixed(1)} KB`
  }
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}

export function prettyJSON(raw: string) {
  if (!raw) {
    return ''
  }
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

export function decodeBody(raw?: string | number[]) {
  if (raw == null) {
    return ''
  }
  if (typeof raw === 'string') {
    try {
      const bin = atob(raw)
      const bytes = Uint8Array.from(bin, (c) => c.charCodeAt(0))
      return new TextDecoder().decode(bytes)
    } catch {
      return raw
    }
  }
  try {
    return new TextDecoder().decode(Uint8Array.from(raw))
  } catch {
    return ''
  }
}
