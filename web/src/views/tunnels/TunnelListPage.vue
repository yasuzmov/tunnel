<template>
  <div class="page-card">
    <div class="bar">
      <el-input v-model="keyword" placeholder="搜索名称或备注" style="width: 240px" clearable @clear="load" @keyup.enter="load" />
      <el-button type="primary" @click="openCreate">新建记录</el-button>
    </div>
    <el-table :data="list" v-loading="loading" @row-click="goDetail">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="最近在线" width="180">
        <template #default="{ row }">{{ formatTime(row.last_online_at) }}</template>
      </el-table-column>
      <el-table-column label="域名数" width="90">
        <template #default="{ row }">{{ row.domains?.length || 0 }}</template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click.stop="goDetail(row)">进入</el-button>
          <el-button link type="danger" @click.stop="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
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
    <el-dialog v-model="show" title="新建穿透记录" width="420px">
      <el-form label-width="70px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="show = false">取消</el-button>
        <el-button type="primary" @click="onCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createTunnel, deleteTunnel, listTunnels } from '../../api/tunnel'
import type { Tunnel } from '../../api/types'
import { usePagination } from '../../composables/usePagination'
import { formatTime } from '../../utils/format'

const { pager } = usePagination()
const list = ref<Tunnel[]>([])
const loading = ref(false)
const keyword = ref('')
const show = ref(false)
const form = reactive({ name: '', remark: '' })
const router = useRouter()

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

function openCreate() {
  form.name = ''
  form.remark = ''
  show.value = true
}

async function onCreate() {
  const res = await createTunnel(form.name, form.remark)
  show.value = false
  ElMessage.success('已创建')
  await router.push(`/tunnels/${res.tunnel.id}`)
}

function goDetail(row: Tunnel) {
  router.push(`/tunnels/${row.id}`)
}

async function onDelete(row: Tunnel) {
  await ElMessageBox.confirm(`删除记录「${row.name}」？`, '确认')
  await deleteTunnel(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<style scoped>
.bar { display: flex; justify-content: space-between; margin-bottom: 16px; }
.pager { margin-top: 16px; justify-content: flex-end; }
</style>
