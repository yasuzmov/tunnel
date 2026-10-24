import { http } from './http'
import type { AdminUser } from './types'

export async function login(username: string, password: string) {
  const { data } = await http.post<{ data: { token: string; user: AdminUser } }>('/auth/login', {
    username,
    password,
  })
  return data.data
}

export async function fetchMe() {
  const { data } = await http.get<{ data: AdminUser }>('/auth/me')
  return data.data
}
