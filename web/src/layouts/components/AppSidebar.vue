<!--
  后台侧边导航。
  Props:
    collapsed (boolean, 默认 false) 是否折叠为仅图标模式。
  详情页（/tunnels/:id）高亮其父级菜单「穿透记录」。
-->
<template>
  <aside class="sidebar" :class="{ collapsed }">
    <div class="brand">
      <div class="logo">
        <el-icon :size="18"><Connection /></el-icon>
      </div>
      <transition name="fade-text">
        <span v-if="!collapsed" class="brand-text">内网穿透</span>
      </transition>
    </div>
    <nav class="nav">
      <router-link
        v-for="item in items"
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
        :title="collapsed ? item.label : undefined"
      >
        <el-icon :size="18"><component :is="item.icon" /></el-icon>
        <span v-if="!collapsed" class="nav-label">{{ item.label }}</span>
      </router-link>
    </nav>
    <div class="foot" v-if="!collapsed">Tunnel Admin</div>
  </aside>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import { useRoute } from 'vue-router'
import { Connection, Document, Link, Odometer, Share, UserFilled } from '@element-plus/icons-vue'
import type { MenuKey } from '../../api/types'
import { useAuthStore } from '../../stores/auth'
import { MENU_DEFS } from '../../utils/menus'

withDefaults(defineProps<{ collapsed?: boolean }>(), { collapsed: false })

const icons: Record<MenuKey, Component> = {
  dashboard: Odometer,
  tunnels: Share,
  domains: Link,
  logs: Document,
  admins: UserFilled,
}
const route = useRoute()
const auth = useAuthStore()
const items = computed(() =>
  MENU_DEFS.filter((item) => auth.can(item.key)).map((item) => ({ ...item, icon: icons[item.key] })),
)

/** 前缀匹配，使子路由也能点亮父菜单。 */
function isActive(path: string) {
  return route.path === path || route.path.startsWith(`${path}/`)
}
</script>

<style scoped>
.sidebar {
  width: 224px;
  flex-shrink: 0;
  background: var(--app-sidebar-bg);
  display: flex;
  flex-direction: column;
  transition: width 0.35s var(--app-ease);
  overflow: hidden;
}
.sidebar.collapsed {
  width: 68px;
}
.brand {
  height: 64px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  color: #fff;
  white-space: nowrap;
}
.logo {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: var(--app-primary);
}
.brand-text {
  font-weight: 700;
  font-size: 16px;
  letter-spacing: 0.5px;
}
.nav {
  flex: 1;
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  height: 42px;
  padding: 0 14px;
  border-radius: 10px;
  color: #a5adbd;
  text-decoration: none;
  white-space: nowrap;
  transition: background 0.2s ease, color 0.2s ease, transform 0.2s var(--app-ease);
}
.nav-item:hover {
  background: var(--app-sidebar-hover);
  color: #fff;
}
.nav-item:active {
  transform: scale(0.98);
}
.nav-item.active {
  color: #fff;
  background: var(--app-primary);
}
.nav-label {
  font-size: 14px;
}
.foot {
  padding: 16px 20px;
  font-size: 12px;
  color: #4b5568;
  white-space: nowrap;
}
.fade-text-enter-active {
  transition: opacity 0.3s ease 0.1s;
}
.fade-text-enter-from {
  opacity: 0;
}
</style>
