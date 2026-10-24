<template>
  <div class="page-card">
    <h3>仪表盘</h3>
    <el-row :gutter="16">
      <el-col :span="6" v-for="item in cards" :key="item.label">
        <el-card>
          <div class="label">{{ item.label }}</div>
          <div class="value">{{ item.value }}</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchDashboard } from '../../api/stats'
import type { DashboardStats } from '../../api/types'

const stats = ref<DashboardStats | null>(null)
const cards = computed(() => [
  { label: '在线客户端', value: stats.value?.online_sessions ?? 0 },
  { label: '穿透记录', value: stats.value?.tunnel_count ?? 0 },
  { label: '今日请求', value: stats.value?.today_requests ?? 0 },
  { label: '今日错误率', value: `${(((stats.value?.error_rate ?? 0) * 100).toFixed(1))}%` },
])

onMounted(async () => {
  stats.value = await fetchDashboard()
})
</script>

<style scoped>
h3 { margin: 0 0 16px; }
.label { color: #909399; }
.value { font-size: 28px; font-weight: 700; margin-top: 8px; }
</style>
