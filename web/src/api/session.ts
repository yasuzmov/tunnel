import { http, type PageData } from './http'
import type { ClientSession } from './types'

export async function listSessions(tunnelId: number, page: number, pageSize: number) {
  const { data } = await http.get<{ data: PageData<ClientSession> }>(`/tunnels/${tunnelId}/sessions`, {
    params: { page, page_size: pageSize },
  })
  return data.data
}
