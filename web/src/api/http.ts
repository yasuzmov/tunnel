import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

export const http = axios.create({
  baseURL: '/api',
  timeout: 20000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (resp) => {
    const body = resp.data as { code: number; message: string; data: unknown }
    if (body.code !== 0) {
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(new Error(body.message))
    }
    return resp
  },
  (err) => {
    const status = err.response?.status
    const message = err.response?.data?.message || err.message || '网络错误'
    if (status === 401) {
      localStorage.removeItem('token')
      if (router.currentRoute.value.path !== '/login') {
        router.push({ path: '/login' })
      }
    }
    ElMessage.error(message)
    return Promise.reject(err)
  },
)

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}
