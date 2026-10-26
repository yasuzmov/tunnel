import { http } from './http'
import type { AvailableDomain, DomainPoolConfig, ManagedDomain, TunnelDomain } from './types'

export async function listDomains(tunnelId: number) {
  const { data } = await http.get<{ data: TunnelDomain[] }>(`/tunnels/${tunnelId}/domains`)
  return data.data
}

/** 当前管理员可绑定且尚未被占用的域名：普通管理员仅限自己创建的，超级管理员可用全部。 */
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

/** 域名管理：普通管理员只能看到自己创建的域名，超级管理员可见全部。 */
export async function listManagedDomains() {
  const { data } = await http.get<{ data: ManagedDomain[] }>('/domains')
  return data.data
}

/** 域名池配置：创建域名时固定的主域名后缀。 */
export async function getDomainPoolConfig() {
  const { data } = await http.get<{ data: DomainPoolConfig }>('/domains/config')
  return data.data
}

/** 创建域名：只提交前缀，后端拼接为「前缀.主域名」。 */
export async function createManagedDomain(name: string, remark: string) {
  const { data } = await http.post<{ data: ManagedDomain }>('/domains', { name, remark })
  return data.data
}

/** 修改域名：name 为空表示前缀不变，只改备注。 */
export async function updateManagedDomain(id: number, name: string, remark: string) {
  await http.put(`/domains/${id}`, { name, remark })
}

export async function deleteManagedDomain(id: number) {
  await http.delete(`/domains/${id}`)
}
