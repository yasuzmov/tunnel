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

/**
 * 修改当前管理员口令。原密码错误时服务端返回 400（非 401），不会触发拦截器跳登录。
 * @param oldPassword 原密码
 * @param newPassword 新密码，8~64 位且不同于原密码
 */
export async function changePassword(oldPassword: string, newPassword: string) {
  await http.put('/auth/password', { old_password: oldPassword, new_password: newPassword })
}
