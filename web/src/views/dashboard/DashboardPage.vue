<!-- 仪表盘：统计卡片 + 接入指引；客户端上下线时自动刷新指标。 -->
<template>
  <div class="page-stack">
    <div class="grid">
      <StatCard v-for="item in cards" :key="item.label" v-bind="item" />
    </div>
    <QuickStartCard />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Connection, DataLine, Monitor, WarningFilled } from '@element-plus/icons-vue'
import { fetchDashboard } from '../../api/stats'
import type { DashboardStats } from '../../api/types'
import { useSSE } from '../../composables/useSSE'
import StatCard from './components/StatCard.vue'
import QuickStartCard from './components/QuickStartCard.vue'

const stats = ref<DashboardStats | null>(null)
const cards = computed(() => {
  const s = stats.value
  return [
    { label: '在线客户端', value: s?.online_sessions ?? 0, icon: Monitor, color: '#16a34a', hint: '实时连接中的隧道' },
    { label: '穿透记录', value: s?.tunnel_count ?? 0, icon: Connection, color: '#3a6ff7', hint: '已创建的记录总数' },
    { label: '今日请求', value: s?.today_requests ?? 0, icon: DataLine, color: '#7c3aed', hint: '经公网入口转发' },
    {
      label: '今日错误率',
      value: `${((s?.error_rate ?? 0) * 100).toFixed(1)}%`,
      icon: WarningFilled,
      color: '#f59e0b',
      hint: `错误 ${s?.today_errors ?? 0} 次（状态码 ≥ 400 或转发失败）`,
    },
  ]
})

async function load() {
  stats.value = await fetchDashboard()
}

useSSE({
  'session.online': () => { void load() },
  'session.offline': () => { void load() },
})

onMounted(load)
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}
</style>
