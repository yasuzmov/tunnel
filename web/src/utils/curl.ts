import type { RequestLog } from '../api/types'

export function toCurl(log: RequestLog) {
  const url = `${log.scheme || 'http'}://${log.host}${log.path}${log.query_string ? `?${log.query_string}` : ''}`
  const parts = [`curl -X ${log.method || 'GET'} '${url}'`]
  let headers: Record<string, string[]> = {}
  try {
    headers = JSON.parse(log.request_headers || '{}')
  } catch {
    headers = {}
  }
  for (const [k, vals] of Object.entries(headers)) {
    if (['host', 'content-length'].includes(k.toLowerCase())) {
      continue
    }
    for (const v of vals) {
      parts.push(`  -H '${k}: ${v.replace(/'/g, `'\\''`)}'`)
    }
  }
  const body = log.body && decodeMaybe(log.body.request_body)
  if (body) {
    parts.push(`  --data-raw '${body.replace(/'/g, `'\\''`)}'`)
  }
  return parts.join(' \\\n')
}

function decodeMaybe(raw?: string | number[]) {
  if (raw == null) {
    return ''
  }
  if (typeof raw === 'string') {
    return raw
  }
  return new TextDecoder().decode(Uint8Array.from(raw))
}
