import { http, type PageData } from './http'
import type { Tunnel, TunnelDetail } from './types'

export async function listTunnels(page: number, pageSize: number, keyword = '') {
  const { data } = await http.get<{ data: PageData<Tunnel> }>('/tunnels', {
    params: { page, page_size: pageSize, keyword },
  })
  return data.data
}

export async function getTunnel(id: number) {
  const { data } = await http.get<{ data: TunnelDetail }>(`/tunnels/${id}`)
  return data.data
}

export async function createTunnel(name: string, remark: string) {
  const { data } = await http.post<{ data: TunnelDetail }>('/tunnels', { name, remark })
  return data.data
}

export async function updateTunnel(id: number, name: string, remark: string) {
  const { data } = await http.put<{ data: TunnelDetail }>(`/tunnels/${id}`, { name, remark })
  return data.data
}

export async function deleteTunnel(id: number) {
  await http.delete(`/tunnels/${id}`)
}

export async function resetTunnelKey(id: number) {
  const { data } = await http.post<{ data: TunnelDetail }>(`/tunnels/${id}/reset-key`)
  return data.data
}
