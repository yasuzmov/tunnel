<template>
  <el-alert
    v-if="!tunnel.last_online_at"
    title="请先用客户端配置地址和 Key 并测试连通性。联通成功后即可绑定域名。"
    type="warning"
    show-icon
    :closable="false"
    style="margin-bottom: 16px"
  />
  <div class="bar">
    <el-button type="primary" :disabled="!tunnel.last_online_at" @click="openCreate">添加域名</el-button>
    <router-link v-if="auth.isSuper" to="/domains" class="manage-link">管理域名</router-link>
    <span class="bar-tip">只能绑定域名列表中已登记的域名</span>
  </div>
  <el-table :data="list">
    <el-table-column prop="domain" label="域名" />
    <el-table-column label="HTTPS" width="90">
      <template #default="{ row }">{{ row.https_enabled ? '是' : '否' }}</template>
    </el-table-column>
    <el-table-column prop="cert_status" label="证书" width="110" />
    <el-table-column prop="nginx_status" label="Nginx" width="110" />
    <el-table-column prop="last_error" label="错误" />
    <el-table-column label="操作" width="200">
      <template #default="{ row }">
        <el-button link type="primary" @click="onPreview(row)">查看配置</el-button>
        <el-button link type="danger" @click="onDelete(row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
  <el-dialog v-model="show" title="绑定域名" width="460px">
    <el-form label-width="90px">
      <el-form-item label="域名">
        <el-select
          v-model="form.domain"
          filterable
          placeholder="选择已授权的域名"
          style="width: 100%"
          :loading="loadingAvailable"
        >
          <el-option
            v-for="d in available"
            :key="d.id"
            :label="d.remark ? `${d.domain}（${d.remark}）` : d.domain"
            :value="d.domain"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="启用 HTTPS"><el-switch v-model="form.https" /></el-form-item>
    </el-form>
    <el-alert
      v-if="!loadingAvailable && !available.length"
      title="暂无可绑定的域名，请联系超级管理员在「域名管理」中添加并授权。"
      type="info"
      show-icon
      :closable="false"
    />
    <template #footer>
      <el-button @click="show = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="onCreate">确定</el-button>
    </template>
  </el-dialog>
  <el-dialog v-model="previewVisible" title="Nginx 配置预览" width="720px">
    <pre class="mono">{{ preview }}</pre>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createDomain,
  deleteDomain,
  listAvailableDomains,
  listDomains,
  previewNginx,
} from '../../../api/domain'
import type { AvailableDomain, Tunnel, TunnelDomain } from '../../../api/types'
import { useSSE } from '../../../composables/useSSE'
import { useAuthStore } from '../../../stores/auth'

const props = defineProps<{ tunnel: Tunnel }>()
const auth = useAuthStore()
const list = ref<TunnelDomain[]>([])
const available = ref<AvailableDomain[]>([])
const show = ref(false)
const saving = ref(false)
const loadingAvailable = ref(false)
const previewVisible = ref(false)
const preview = ref('')
const form = reactive({ domain: '', https: false })

async function load() {
  list.value = await listDomains(props.tunnel.id)
}

async function openCreate() {
  form.domain = ''
  form.https = false
  show.value = true
  loadingAvailable.value = true
  try {
    available.value = await listAvailableDomains()
  } finally {
    loadingAvailable.value = false
  }
}

async function onCreate() {
  if (!form.domain) {
    ElMessage.warning('请选择域名')
    return
  }
  saving.value = true
  try {
    await createDomain(props.tunnel.id, form.domain, form.https)
    show.value = false
    ElMessage.success('已提交，nginx 配置将自动应用')
    await load()
  } finally {
    saving.value = false
  }
}

async function onDelete(row: TunnelDomain) {
  await ElMessageBox.confirm(`删除域名 ${row.domain}？`, '确认')
  await deleteDomain(props.tunnel.id, row.id)
  await load()
}

async function onPreview(row: TunnelDomain) {
  preview.value = await previewNginx(props.tunnel.id, row.id)
  previewVisible.value = true
}

useSSE({
  'domain.updated': () => { void load() },
})

onMounted(load)
</script>

<style scoped>
.bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.manage-link {
  font-size: 13px;
  color: var(--app-primary);
  text-decoration: none;
}
.manage-link:hover {
  text-decoration: underline;
}
.bar-tip {
  font-size: 12px;
  color: var(--app-text-3);
}
pre {
  white-space: pre-wrap;
  background: #1d1e2c;
  color: #e5eaf3;
  padding: 12px;
  border-radius: 6px;
}
</style>
