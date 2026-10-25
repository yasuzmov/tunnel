/**
 * 左侧菜单与权限键的唯一前端定义。
 * 顺序即侧栏展示顺序；key 必须与后端 service/menu.go 保持一致。
 */
import type { MenuKey } from '../api/types'

export interface MenuDef {
  key: MenuKey
  label: string
  path: string
}

export const MENU_DEFS: MenuDef[] = [
  { key: 'dashboard', label: '仪表盘', path: '/dashboard' },
  { key: 'tunnels', label: '穿透记录', path: '/tunnels' },
  { key: 'domains', label: '域名管理', path: '/domains' },
  { key: 'logs', label: '请求日志', path: '/logs' },
  { key: 'admins', label: '管理员管理', path: '/admins' },
]

/** 普通管理员可被授予的菜单；管理员管理涉及账号安全，仅超级管理员可用。 */
export const GRANTABLE_MENUS = MENU_DEFS.filter((item) => item.key !== 'admins')

/** 未配置菜单权限的普通管理员默认可见菜单。 */
export const DEFAULT_ADMIN_MENUS: MenuKey[] = ['dashboard', 'tunnels', 'logs']

export const ALL_MENU_KEYS: MenuKey[] = MENU_DEFS.map((item) => item.key)

export function menuLabel(key: MenuKey) {
  return MENU_DEFS.find((item) => item.key === key)?.label ?? key
}
