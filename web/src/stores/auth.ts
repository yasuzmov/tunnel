import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchMe, login as loginApi } from '../api/auth'
import type { AdminUser, MenuKey } from '../api/types'
import { ALL_MENU_KEYS, DEFAULT_ADMIN_MENUS, MENU_DEFS } from '../utils/menus'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<AdminUser | null>(null)
  const isSuper = computed(() => user.value?.role === 'super_admin')
  /** 当前账号实际可见的左侧菜单；超级管理员始终拥有全部菜单。 */
  const menus = computed<MenuKey[]>(() => {
    const current = user.value
    if (!current) {
      return []
    }
    if (current.role === 'super_admin') {
      return ALL_MENU_KEYS
    }
    return current.menus?.length ? current.menus : DEFAULT_ADMIN_MENUS
  })
  const can = (menu: MenuKey) => menus.value.includes(menu)
  /** 登录后或访问无权限页面时落到第一个可见菜单。 */
  const homePath = computed(() => MENU_DEFS.find((item) => can(item.key))?.path ?? '/login')

  async function login(username: string, password: string) {
    const res = await loginApi(username, password)
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
  }

  async function loadMe() {
    if (!token.value) {
      return
    }
    user.value = await fetchMe()
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  return { token, user, isSuper, menus, can, homePath, login, loadMe, logout }
})
