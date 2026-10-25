<!--
  后台顶栏：折叠按钮、当前页标题、用户菜单。
  Props:
    collapsed (boolean) 侧栏当前是否折叠，用于切换图标。
    title (string, 默认 '') 当前页标题。
    username (string, 默认 '') 当前登录用户名。
  Emits:
    toggle            点击折叠按钮。
    change-password   选择「修改密码」。
    logout            选择「退出登录」。
-->
<template>
  <header class="header">
    <div class="left">
      <button class="icon-btn" type="button" @click="emit('toggle')" aria-label="折叠侧栏">
        <el-icon :size="18"><Expand v-if="collapsed" /><Fold v-else /></el-icon>
      </button>
      <transition name="title-swap" mode="out-in">
        <h1 :key="title" class="title">{{ title }}</h1>
      </transition>
    </div>
    <el-dropdown trigger="click" @command="onCommand">
      <div class="user">
        <div class="avatar">{{ initial }}</div>
        <span class="name">{{ username }}</span>
        <el-icon :size="12"><ArrowDown /></el-icon>
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="password" :icon="Lock">修改密码</el-dropdown-item>
          <el-dropdown-item command="logout" :icon="SwitchButton" divided>退出登录</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowDown, Expand, Fold, Lock, SwitchButton } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{ collapsed: boolean; title?: string; username?: string }>(), {
  title: '',
  username: '',
})
const emit = defineEmits<{ toggle: []; 'change-password': []; logout: [] }>()

const initial = computed(() => (props.username || '?').slice(0, 1).toUpperCase())

function onCommand(cmd: string) {
  if (cmd === 'password') emit('change-password')
  else if (cmd === 'logout') emit('logout')
}
</script>

<style scoped>
.header {
  height: 64px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px 0 16px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: saturate(180%) blur(12px);
  border-bottom: 1px solid var(--app-border);
}
.left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.icon-btn {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--app-text-2);
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: background 0.2s ease;
}
.icon-btn:hover {
  background: var(--app-primary-soft);
  color: var(--app-primary);
}
.title {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
}
.user {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px 4px 4px;
  border-radius: 999px;
  cursor: pointer;
  color: var(--app-text);
  transition: background 0.2s ease;
}
.user:hover {
  background: #f1f3f9;
}
.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-weight: 600;
  color: #fff;
  background: var(--app-primary);
}
.name {
  font-size: 14px;
}
.title-swap-enter-active,
.title-swap-leave-active {
  transition: transform 0.25s var(--app-ease), opacity 0.2s ease;
}
.title-swap-enter-from {
  transform: translateY(8px);
  opacity: 0;
}
.title-swap-leave-to {
  transform: translateY(-8px);
  opacity: 0;
}
</style>
