/**
 * 路由表与登录守卫。
 * meta.title 用于顶栏标题与浏览器标签；meta.depth / meta.order 决定页面滑动方向。
 */
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { updateSlideDirection } from './transition'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    depth?: number
    order?: number
    /** 仅超级管理员可见/可访问。 */
    superOnly?: boolean
  }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('../views/login/LoginPage.vue'), meta: { title: '登录' } },
    {
      path: '/',
      component: () => import('../layouts/MainLayout.vue'),
      children: [
        { path: '', redirect: '/dashboard' },
        {
          path: 'dashboard',
          component: () => import('../views/dashboard/DashboardPage.vue'),
          meta: { title: '仪表盘', depth: 1, order: 0 },
        },
        {
          path: 'tunnels',
          component: () => import('../views/tunnels/TunnelListPage.vue'),
          meta: { title: '穿透记录', depth: 1, order: 1 },
        },
        {
          path: 'tunnels/:id',
          component: () => import('../views/tunnels/TunnelDetailPage.vue'),
          meta: { title: '穿透详情', depth: 2, order: 1 },
        },
        {
          path: 'domains',
          component: () => import('../views/domains/DomainListPage.vue'),
          meta: { title: '域名管理', depth: 1, order: 2, superOnly: true },
        },
        {
          path: 'logs',
          component: () => import('../views/logs/LogListPage.vue'),
          meta: { title: '请求日志', depth: 1, order: 3 },
        },
        {
          path: 'admins',
          component: () => import('../views/admins/AdminListPage.vue'),
          meta: { title: '管理员管理', depth: 1, order: 4, superOnly: true },
        },
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
  if (to.meta.superOnly && !auth.isSuper) {
    return '/dashboard'
  }
  return true
})

router.afterEach((to, from) => {
  updateSlideDirection(to, from)
  document.title = to.meta.title ? `${to.meta.title} · 内网穿透` : '内网穿透'
})

export default router
