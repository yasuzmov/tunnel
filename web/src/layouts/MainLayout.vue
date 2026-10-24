<!-- 后台主框架：侧栏导航与顶栏退出。 -->
<template>
  <el-container class="wrap">
    <el-aside width="220px" class="aside">
      <div class="brand">内网穿透</div>
      <el-menu :router="true" :default-active="route.path" background-color="#1d1e2c" text-color="#cfd3dc" active-text-color="#409eff">
        <el-menu-item index="/dashboard">仪表盘</el-menu-item>
        <el-menu-item index="/tunnels">穿透记录</el-menu-item>
        <el-menu-item index="/logs">请求日志</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span>{{ auth.user?.username }}</span>
        <el-button link type="primary" @click="onLogout">退出</el-button>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

function onLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.wrap { height: 100%; }
.aside { background: #1d1e2c; color: #fff; }
.brand { height: 56px; display: flex; align-items: center; justify-content: center; font-weight: 700; }
.header { display: flex; align-items: center; justify-content: flex-end; gap: 12px; background: #fff; border-bottom: 1px solid #ebeef5; }
.el-menu { border-right: none; }
</style>
