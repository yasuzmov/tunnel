import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('../views/login/LoginPage.vue') },
    {
      path: '/',
      component: () => import('../layouts/MainLayout.vue'),
      children: [
        { path: '', redirect: '/dashboard' },
        { path: 'dashboard', component: () => import('../views/dashboard/DashboardPage.vue') },
        { path: 'tunnels', component: () => import('../views/tunnels/TunnelListPage.vue') },
        { path: 'tunnels/:id', component: () => import('../views/tunnels/TunnelDetailPage.vue') },
        { path: 'logs', component: () => import('../views/logs/LogListPage.vue') },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.path === '/login') {
    return true
  }
  if (!auth.token) {
    return '/login'
  }
  if (!auth.user) {
    try {
      await auth.loadMe()
    } catch {
      auth.logout()
      return '/login'
    }
  }
  return true
})

export default router
