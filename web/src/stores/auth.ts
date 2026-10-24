import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchMe, login as loginApi } from '../api/auth'
import type { AdminUser } from '../api/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<AdminUser | null>(null)

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

  return { token, user, login, loadMe, logout }
})
