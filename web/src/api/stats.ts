import { http } from './http'
import type { DashboardStats } from './types'

export async function fetchDashboard() {
  const { data } = await http.get<{ data: DashboardStats }>('/stats/dashboard')
  return data.data
}
