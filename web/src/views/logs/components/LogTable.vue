<!--
  请求日志表格：筛选 + 分页 + 详情抽屉；收到 log.created 事件实时刷新。
  Props:
    tunnelId (number, 可选) 传入时只看该穿透记录的日志，并隐藏「记录」维度。
-->
<template>
  <el-form :inline="true" class="filters" @submit.prevent="search">
    <el-form-item>
      <el-select v-model="q.method" placeholder="方法" clearable style="width: 110px">
        <el-option v-for="m in methods" :key="m" :label="m" :value="m" />
      </el-select>
    </el-form-item>
    <el-form-item><el-input v-model="q.status" placeholder="状态码" clearable style="width: 100px" /></el-form-item>
    <el-form-item><el-input v-model="q.path" placeholder="路径" clearable style="width: 180px" /></el-form-item>
    <el-form-item><el-input v-model="q.remote_ip" placeholder="来源 IP" clearable style="width: 140px" /></el-form-item>
    <el-form-item><el-input v-model="q.domain" placeholder="域名" clearable style="width: 180px" /></el-form-item>
    <el-form-item>
      <el-button type="primary" :icon="Search" native-type="submit">查询</el-button>
      <el-button :icon="RefreshRight" @click="resetFilters">重置</el-button>
    </el-form-item>
  </el-form>
  <el-table :data="list" class="is-clickable" @row-click="open">
    <el-table-column label="方法" width="90">
      <template #default="{ row }">
        <el-tag :type="methodTagType(row.method)" effect="light" size="small" disable-transitions>
          {{ row.is_websocket ? 'WS' : row.method }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="80">
      <template #default="{ row }">
        <el-tag :type="statusTagType(row.status)" effect="plain" size="small" round disable-transitions>
          {{ row.status || 'ERR' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="host" label="Host" width="180" show-overflow-tooltip />
    <el-table-column label="路径" min-width="220" show-overflow-tooltip>
      <template #default="{ row }"><span class="mono">{{ row.path }}{{ row.query_string ? `?${row.query_string}` : '' }}</span></template>
    </el-table-column>
    <el-table-column prop="remote_ip" label="来源 IP" width="130" />
    <el-table-column label="耗时" width="90" align="right">
      <template #default="{ row }"><span :class="['dur', durationClass(row.duration_ms)]">{{ row.duration_ms }} ms</span></template>
    </el-table-column>
    <el-table-column label="时间" width="170">
      <template #default="{ row }">{{ formatTime(row.received_at) }}</template>
    </el-table-column>
  </el-table>
  <el-pagination
    class="pager"
    background
    layout="total, prev, pager, next"
    :total="pager.total"
    v-model:current-page="pager.page"
    @current-change="load"
  />
  <LogDetailDrawer v-model="drawer" :log-id="currentId" />
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RefreshRight, Search } from '@element-plus/icons-vue'
import { listLogs } from '../../../api/log'
import type { RequestLog } from '../../../api/types'
import { usePagination } from '../../../composables/usePagination'
import { useSSE } from '../../../composables/useSSE'
import { formatTime } from '../../../utils/format'
import { durationClass, methodTagType, statusTagType } from '../../../utils/httpTag'
import LogDetailDrawer from './LogDetailDrawer.vue'

const props = defineProps<{ tunnelId?: number }>()
const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS', 'HEAD']
const { pager } = usePagination()
const list = ref<RequestLog[]>([])
const drawer = ref(false)
const currentId = ref(0)
const q = reactive({ method: '', status: '', path: '', remote_ip: '', domain: '' })

async function load() {
  const data = await listLogs({
    page: pager.page,
    page_size: pager.pageSize,
    method: q.method || undefined,
    status: q.status ? Number(q.status) : undefined,
    path: q.path || undefined,
    remote_ip: q.remote_ip || undefined,
    domain: q.domain || undefined,
    tunnel_id: props.tunnelId,
  }, props.tunnelId)
  list.value = data.list || []
  pager.total = data.total
}

function search() {
  pager.page = 1
  void load()
}

function resetFilters() {
  Object.assign(q, { method: '', status: '', path: '', remote_ip: '', domain: '' })
  search()
}

function open(row: RequestLog) {
  currentId.value = row.id
  drawer.value = true
}

/** 仅在第一页时实时刷新，避免用户翻页查看历史时列表被顶走。 */
useSSE({
  'log.created': () => { if (pager.page === 1) void load() },
})

onMounted(load)
</script>

<style scoped>
.filters {
  margin-bottom: 4px;
}
.filters :deep(.el-form-item) {
  margin-right: 10px;
}
.dur {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
}
.dur.fast { color: #16a34a; }
.dur.slow { color: #d97706; }
.dur.very-slow { color: #dc2626; }
</style>
