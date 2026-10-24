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
    <el-button type="primary" :disabled="!tunnel.last_online_at" @click="show = true">添加域名</el-button>
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
  <el-dialog v-model="show" title="绑定域名" width="420px">
    <el-form label-width="90px">
      <el-form-item label="域名"><el-input v-model="form.domain" placeholder="app.example.com" /></el-form-item>
      <el-form-item label="启用 HTTPS"><el-switch v-model="form.https" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="show = false">取消</el-button>
      <el-button type="primary" @click="onCreate">确定</el-button>
    </template>
  </el-dialog>
  <el-dialog v-model="previewVisible" title="Nginx 配置预览" width="720px">
    <pre class="mono">{{ preview }}</pre>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createDomain, deleteDomain, listDomains, previewNginx } from '../../../api/domain'
import type { Tunnel, TunnelDomain } from '../../../api/types'
import { useSSE } from '../../../composables/useSSE'

const props = defineProps<{ tunnel: Tunnel }>()
const list = ref<TunnelDomain[]>([])
const show = ref(false)
const previewVisible = ref(false)
const preview = ref('')
const form = reactive({ domain: '', https: false })

async function load() {
  list.value = await listDomains(props.tunnel.id)
}

async function onCreate() {
  await createDomain(props.tunnel.id, form.domain, form.https)
  show.value = false
  form.domain = ''
  form.https = false
  ElMessage.success('已提交，nginx 配置将自动应用')
  await load()
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
.bar { margin-bottom: 12px; }
pre { white-space: pre-wrap; background: #1d1e2c; color: #e5eaf3; padding: 12px; border-radius: 6px; }
</style>
