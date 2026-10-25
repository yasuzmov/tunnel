<!-- 穿透记录详情：顶部概要 + 四个标签页，各标签页展示由子组件承担。 -->
<template>
  <div class="page-stack" v-loading="!detail">
    <div class="page-card hero">
      <button class="back" type="button" @click="router.push('/tunnels')" aria-label="返回列表">
        <el-icon :size="18"><ArrowLeft /></el-icon>
      </button>
      <div class="info">
        <h2>{{ detail?.tunnel.name || '穿透详情' }}</h2>
        <p>{{ detail?.tunnel.remark || '无备注' }}</p>
      </div>
      <el-tag v-if="detail" :type="detail.tunnel.last_online_at ? 'success' : 'info'" effect="light" round>
        {{ detail.tunnel.last_online_at ? `最近在线 ${formatTime(detail.tunnel.last_online_at)}` : '从未联通' }}
      </el-tag>
    </div>
    <div class="page-card">
      <el-tabs v-model="tab" class="tabs">
        <el-tab-pane label="基本信息" name="basic">
          <BasicInfoTab v-if="detail" :detail="detail" @refresh="load" />
        </el-tab-pane>
        <el-tab-pane label="客户端联通记录" name="sessions">
          <SessionsTab v-if="detail" :tunnel-id="detail.tunnel.id" @refresh="load" />
        </el-tab-pane>
        <el-tab-pane label="域名配置" name="domains">
          <DomainsTab v-if="detail" :tunnel="detail.tunnel" @refresh="load" />
        </el-tab-pane>
        <el-tab-pane label="请求日志" name="logs" lazy>
          <LogsTab v-if="detail" :tunnel-id="detail.tunnel.id" :enabled="detail.tunnel.log_enabled" @refresh="load" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getTunnel } from '../../api/tunnel'
import type { TunnelDetail } from '../../api/types'
import { formatTime } from '../../utils/format'
import BasicInfoTab from './components/BasicInfoTab.vue'
import SessionsTab from './components/SessionsTab.vue'
import DomainsTab from './components/DomainsTab.vue'
import LogsTab from './components/LogsTab.vue'

const route = useRoute()
const router = useRouter()
const detail = ref<TunnelDetail | null>(null)
const tab = ref('basic')

async function load() {
  detail.value = await getTunnel(Number(route.params.id))
}

onMounted(load)
</script>

<style scoped>
.hero {
  display: flex;
  align-items: center;
  gap: 16px;
}
.back {
  width: 38px;
  height: 38px;
  border: 1px solid var(--app-border);
  border-radius: 10px;
  background: #fff;
  color: var(--app-text-2);
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: all 0.2s var(--app-ease);
}
.back:hover {
  color: var(--app-primary);
  border-color: var(--app-primary);
  transform: translateX(-2px);
}
.info {
  flex: 1;
  min-width: 0;
}
h2 {
  margin: 0;
  font-size: 20px;
}
p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--app-text-3);
}
.tabs :deep(.el-tab-pane) {
  animation: tab-in 0.35s var(--app-ease);
}
@keyframes tab-in {
  from { transform: translate3d(16px, 0, 0); opacity: 0; }
  to { transform: none; opacity: 1; }
}
</style>
