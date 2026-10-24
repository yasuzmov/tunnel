import { http } from './http'
import type { TunnelDomain } from './types'

export async function listDomains(tunnelId: number) {
  const { data } = await http.get<{ data: TunnelDomain[] }>(`/tunnels/${tunnelId}/domains`)
  return data.data
}

export async function createDomain(tunnelId: number, domain: string, httpsEnabled: boolean) {
  const { data } = await http.post<{ data: TunnelDomain }>(`/tunnels/${tunnelId}/domains`, {
    domain,
    https_enabled: httpsEnabled,
  })
  return data.data
}

export async function deleteDomain(tunnelId: number, domainId: number) {
  await http.delete(`/tunnels/${tunnelId}/domains/${domainId}`)
}

export async function previewNginx(tunnelId: number, domainId: number) {
  const { data } = await http.get<{ data: { config: string } }>(
    `/tunnels/${tunnelId}/domains/${domainId}/nginx-preview`,
  )
  return data.data.config
}
