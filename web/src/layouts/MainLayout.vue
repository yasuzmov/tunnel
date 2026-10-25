<!--
  后台主框架：侧栏 + 顶栏 + 内容区。
  内容区按路由层级/菜单顺序左右滑动切换；折叠状态持久化到 localStorage。
-->
<template>
  <div class="shell">
    <AppSidebar :collapsed="collapsed" />
    <div class="body">
      <AppHeader
        :collapsed="collapsed"
        :title="route.meta.title"
        :username="auth.user?.username"
        @toggle="toggle"
        @change-password="pwdVisible = true"
        @logout="onLogout"
      />
      <main class="content">
        <router-view v-slot="{ Component, route: r }">
          <transition :name="slideName">
            <div :key="r.fullPath" class="page">
              <component :is="Component" />
            </div>
          </transition>
        </router-view>
      </main>
    </div>
    <ChangePasswordDialog v-model="pwdVisible" @success="onPasswordChanged" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { slideName } from '../router/transition'
import AppSidebar from './components/AppSidebar.vue'
import AppHeader from './components/AppHeader.vue'
import ChangePasswordDialog from './components/ChangePasswordDialog.vue'

const COLLAPSE_KEY = 'sidebar-collapsed'
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const collapsed = ref(localStorage.getItem(COLLAPSE_KEY) === '1')
const pwdVisible = ref(false)

function toggle() {
  collapsed.value = !collapsed.value
  localStorage.setItem(COLLAPSE_KEY, collapsed.value ? '1' : '0')
}

async function onLogout() {
  await ElMessageBox.confirm('确定退出登录？', '提示', { type: 'warning' })
  auth.logout()
  await router.push('/login')
}

/** 改密成功后强制重新登录，让新口令立即生效于当前浏览器。 */
async function onPasswordChanged() {
  ElMessage.success('密码已修改，请使用新密码重新登录')
  auth.logout()
  await router.push('/login')
}
</script>

<style scoped>
.shell {
  display: flex;
  height: 100%;
  background: var(--app-bg);
}
.body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.content {
  position: relative;
  flex: 1;
  overflow: hidden;
}
.page {
  height: 100%;
  overflow-y: auto;
  padding: 20px 24px 32px;
}
</style>
