<template>
  <el-table :data="list">
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tag :type="row.online ? 'success' : 'info'">{{ row.online ? '在线' : '离线' }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="hostname" label="主机" />
    <el-table-column prop="os" label="系统" width="120" />
    <el-table-column prop="version" label="版本" width="100" />
    <el-table-column prop="client_ip" label="IP" width="140" />
    <el-table-column label="连接时间" width="180">
      <template #default="{ row }">{{ formatTime(row.connected_at) }}</template>
    </el-table-column>
    <el-table-column label="断开时间" width="180">
      <template #default="{ row }">{{ formatTime(row.disconnected_at) }}</template>
    </el-table-column>
    <el-table-column prop="disconnect_reason" label="原因" width="160" />
  </el-table>
  <el-pagination
    class="pager"
    background
    layout="total, prev, pager, next"
    :total="pager.total"
    v-model:current-page="pager.page"
    @current-change="load"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listSessions } from '../../../api/session'
import type { ClientSession } from '../../../api/types'
import { usePagination } from '../../../composables/usePagination'
import { useSSE } from '../../../composables/useSSE'
import { formatTime } from '../../../utils/format'

const props = defineProps<{ tunnelId: number }>()
const emit = defineEmits<{ refresh: [] }>()
const { pager } = usePagination()
const list = ref<ClientSession[]>([])

async function load() {
  const data = await listSessions(props.tunnelId, pager.page, pager.pageSize)
  list.value = data.list || []
  pager.total = data.total
}

useSSE({
  'session.online': () => { void load(); emit('refresh') },
  'session.offline': () => { void load(); emit('refresh') },
})

onMounted(load)
</script>

<style scoped>
.pager { margin-top: 16px; justify-content: flex-end; }
</style>
