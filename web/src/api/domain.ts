import { http } from './http'
import type { AvailableDomain, ManagedDomain, TunnelDomain } from './types'

export async function listDomains(tunnelId: number) {
  const { data } = await http.get<{ data: TunnelDomain[] }>(`/tunnels/${tunnelId}/domains`)
  return data.data
}

/** 当前管理员可绑定且尚未被占用的域名（受域名池与授权限制）。 */
export async function listAvailableDomains() {
  const { data } = await http.get<{ data: AvailableDomain[] }>('/domains/available')
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

/** 统一域名池（仅超级管理员）。 */
export async function listManagedDomains() {
  const { data } = await http.get<{ data: ManagedDomain[] }>('/domains')
  return data.data
}

export async function createManagedDomain(domain: string, remark: string) {
  const { data } = await http.post<{ data: ManagedDomain }>('/domains', { domain, remark })
  return data.data
}

export async function updateManagedDomain(id: number, remark: string) {
  await http.put(`/domains/${id}`, { remark })
}

export async function deleteManagedDomain(id: number) {
  await http.delete(`/domains/${id}`)
}
