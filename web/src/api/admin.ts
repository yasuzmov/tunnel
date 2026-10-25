import { http } from './http'
import type { AdminAccount, AdminRole } from './types'

export async function listAdmins() {
  const { data } = await http.get<{ data: AdminAccount[] }>('/admins')
  return data.data
}

export async function createAdmin(username: string, password: string, role: AdminRole) {
  const { data } = await http.post<{ data: AdminAccount }>('/admins', { username, password, role })
  return data.data
}

export async function updateAdmin(id: number, username: string, role: AdminRole) {
  const { data } = await http.put<{ data: AdminAccount }>(`/admins/${id}`, { username, role })
  return data.data
}

export async function resetAdminPassword(id: number, password: string) {
  await http.put(`/admins/${id}/password`, { password })
}

export async function setAdminDomains(id: number, domainIds: number[]) {
  await http.put(`/admins/${id}/domains`, { domain_ids: domainIds })
}

export async function deleteAdmin(id: number) {
  await http.delete(`/admins/${id}`)
}
