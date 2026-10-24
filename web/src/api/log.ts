import { http, type PageData } from './http'
import type { RequestLog } from './types'

export interface LogQuery {
  page: number
  page_size: number
  tunnel_id?: number
  domain?: string
  method?: string
  status?: number
  path?: string
  remote_ip?: string
  start?: string
  end?: string
}

export async function listLogs(params: LogQuery, tunnelId?: number) {
  const url = tunnelId ? `/tunnels/${tunnelId}/logs` : '/logs'
  const { data } = await http.get<{ data: PageData<RequestLog> }>(url, { params })
  return data.data
}

export async function getLog(id: number) {
  const { data } = await http.get<{ data: RequestLog }>(`/logs/${id}`)
  return data.data
}
