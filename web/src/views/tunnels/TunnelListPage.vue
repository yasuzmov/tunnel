<!-- 穿透记录列表：搜索、分页、新建与删除；点击行进入详情。 -->
<template>
  <div class="page-card">
    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索名称或备注"
        :prefix-icon="Search"
        style="width: 260px"
        clearable
        @clear="search"
        @keyup.enter="search"
      />
      <el-button type="primary" :icon="Plus" @click="openCreate">新建记录</el-button>
    </div>
    <el-table :data="list" v-loading="loading" class="is-clickable" @row-click="goDetail">
      <el-table-column label="名称" min-width="200">
        <template #default="{ row }">
          <div class="name-cell">
            <div class="badge">{{ row.name.slice(0, 1) }}</div>
            <div>
              <div class="name">{{ row.name }}</div>
              <div class="remark">{{ row.remark || '无备注' }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="绑定域名" min-width="220">
        <template #default="{ row }">
          <template v-if="row.domains?.length">
            <el-tag v-for="d in row.domains" :key="d.id" size="small" effect="plain" class="domain-tag">{{ d.domain }}</el-tag>
          </template>
          <span v-else class="muted">未绑定</span>
        </template>
      </el-table-column>
      <el-table-column v-if="auth.isSuper" label="创建者" width="130">
        <template #default="{ row }">
          <span v-if="row.created_by">{{ row.creator_name || `#${row.created_by}` }}</span>
          <span v-else class="muted">历史记录</span>
        </template>
      </el-table-column>
      <el-table-column label="转发日志" width="110" align="center">
        <template #default="{ row }">
          <div @click.stop>
            <el-switch
              :model-value="row.log_enabled"
              :loading="loggingId === row.id"
              @change="(v: boolean | string | number) => onLogging(row, Boolean(v))"
            />
          </div>
        </template>
      </el-table-column>
      <el-table-column label="最近在线" width="180">
        <template #default="{ row }">
          <span :class="{ muted: !row.last_online_at }">{{ row.last_online_at ? formatTime(row.last_online_at) : '从未联通' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" align="right">
        <template #default="{ row }">
          <el-button link type="primary" @click.stop="goDetail(row)">进入</el-button>
          <el-button link type="danger" @click.stop="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="还没有穿透记录" :image-size="90">
          <el-button type="primary" @click="openCreate">立即新建</el-button>
        </el-empty>
      </template>
    </el-table>
    <el-pagination
      class="pager"
      background
      layout="total, prev, pager, next"
      :total="pager.total"
      v-model:current-page="pager.page"
      v-model:page-size="pager.pageSize"
      @current-change="load"
    />
    <el-dialog v-model="show" title="新建穿透记录" width="440px">
      <el-form label-position="top" @submit.prevent="onCreate">
        <el-form-item label="名称" required><el-input v-model="form.name" placeholder="例如：办公室 NAS" maxlength="64" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="3" maxlength="255" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="show = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="onCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { createTunnel, deleteTunnel, listTunnels, setTunnelLogging } from '../../api/tunnel'
import type { Tunnel } from '../../api/types'
import { usePagination } from '../../composables/usePagination'
import { useAuthStore } from '../../stores/auth'
import { formatTime } from '../../utils/format'

const { pager } = usePagination()
const list = ref<Tunnel[]>([])
const loading = ref(false)
const creating = ref(false)
const keyword = ref('')
const loggingId = ref(0)
const show = ref(false)
const form = reactive({ name: '', remark: '' })
const router = useRouter()
const auth = useAuthStore()

async function load() {
  loading.value = true
  try {
    const data = await listTunnels(pager.page, pager.pageSize, keyword.value)
    list.value = data.list || []
    pager.total = data.total
  } finally {
    loading.value = false
  }
}

function search() {
  pager.page = 1
  void load()
}

function openCreate() {
  form.name = ''
  form.remark = ''
  show.value = true
}

async function onCreate() {
  if (!form.name.trim()) {
    ElMessage.warning('请填写名称')
    return
  }
  creating.value = true
  try {
    const res = await createTunnel(form.name.trim(), form.remark)
    show.value = false
    ElMessage.success('已创建')
    await router.push(`/tunnels/${res.tunnel.id}`)
  } finally {
    creating.value = false
  }
}

async function onLogging(row: Tunnel, enabled: boolean) {
  loggingId.value = row.id
  try {
    await setTunnelLogging(row.id, enabled)
    row.log_enabled = enabled
    ElMessage.success(enabled ? '已开始记录转发日志' : '已停止记录转发日志')
  } finally {
    loggingId.value = 0
  }
}

function goDetail(row: Tunnel) {
  router.push(`/tunnels/${row.id}`)
}

async function onDelete(row: Tunnel) {
  await ElMessageBox.confirm(`删除记录「${row.name}」？已绑定的域名配置会一并移除。`, '确认删除', { type: 'warning' })
  await deleteTunnel(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<style scoped>
.name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}
.badge {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-weight: 600;
  color: var(--app-primary);
  background: var(--app-primary-soft);
}
.name {
  font-weight: 600;
  color: var(--app-text);
}
.remark {
  font-size: 12px;
  color: var(--app-text-3);
  margin-top: 2px;
}
.domain-tag {
  margin: 2px 6px 2px 0;
}
.muted {
  color: var(--app-text-3);
}
</style>
