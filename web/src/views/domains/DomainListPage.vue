<!-- 统一域名池：超级管理员在这里登记可穿透域名，并查看绑定情况。 -->
<template>
  <div class="page-stack">
    <div class="page-card">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索域名或备注"
          :prefix-icon="Search"
          style="width: 260px"
          clearable
        />
        <el-button type="primary" :icon="Plus" @click="openCreate">添加域名</el-button>
      </div>
      <el-alert
        title="只有这里登记的域名才能绑定到穿透记录；普通管理员还需要在「管理员管理」中被授权。"
        type="info"
        show-icon
        :closable="false"
        class="tip"
      />
      <el-table :data="filtered" v-loading="loading">
        <el-table-column label="域名" min-width="240">
          <template #default="{ row }">
            <div class="domain">{{ row.domain }}</div>
            <div class="remark">{{ row.remark || '无备注' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="绑定记录" min-width="200">
          <template #default="{ row }">
            <router-link v-if="row.bound_tunnel_id" :to="`/tunnels/${row.bound_tunnel_id}`" class="tunnel-link">
              {{ row.bound_tunnel_name || `#${row.bound_tunnel_id}` }}
            </router-link>
            <span v-else class="muted">未绑定</span>
          </template>
        </el-table-column>
        <el-table-column label="HTTPS / 证书" width="150">
          <template #default="{ row }">
            <template v-if="row.bound_tunnel_id">
              <el-tag :type="row.https_enabled ? 'success' : 'info'" effect="plain" size="small">
                {{ row.https_enabled ? '已启用' : '未启用' }}
              </el-tag>
              <div class="cert">{{ certLabel(row.cert_status) }}</div>
            </template>
            <span v-else class="muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="Nginx" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.bound_tunnel_id" :type="nginxType(row.nginx_status)" effect="light" size="small">
              {{ row.nginx_status }}
            </el-tag>
            <span v-else class="muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" align="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑备注</el-button>
            <el-button
              link
              type="danger"
              :disabled="!!row.bound_tunnel_id"
              :title="row.bound_tunnel_id ? '已绑定到穿透记录，请先在记录详情中解绑' : ''"
              @click="onDelete(row)"
            >删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="还没有可穿透域名" :image-size="90">
            <el-button type="primary" @click="openCreate">添加第一个域名</el-button>
          </el-empty>
        </template>
      </el-table>
    </div>

    <el-dialog v-model="createVisible" title="添加域名" width="440px">
      <el-form label-position="top" @submit.prevent="onCreate">
        <el-form-item label="域名" required>
          <el-input v-model="createForm.domain" placeholder="app.example.com" maxlength="253" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="createForm.remark" placeholder="例如：给运营团队使用" maxlength="255" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onCreate">添加</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editVisible" title="编辑备注" width="440px">
      <el-form label-position="top" @submit.prevent="onEdit">
        <el-form-item label="域名">
          <el-input :model-value="editForm.domain" disabled />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" maxlength="255" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import {
  createManagedDomain,
  deleteManagedDomain,
  listManagedDomains,
  updateManagedDomain,
} from '../../api/domain'
import type { ManagedDomain } from '../../api/types'
import { formatTime } from '../../utils/format'

const list = ref<ManagedDomain[]>([])
const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const createVisible = ref(false)
const editVisible = ref(false)
const createForm = reactive({ domain: '', remark: '' })
const editForm = reactive({ id: 0, domain: '', remark: '' })

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return list.value
  return list.value.filter(
    (d) => d.domain.toLowerCase().includes(kw) || d.remark.toLowerCase().includes(kw),
  )
})

async function load() {
  loading.value = true
  try {
    list.value = await listManagedDomains()
  } finally {
    loading.value = false
  }
}

function openCreate() {
  createForm.domain = ''
  createForm.remark = ''
  createVisible.value = true
}

async function onCreate() {
  if (!createForm.domain.trim()) {
    ElMessage.warning('请填写域名')
    return
  }
  saving.value = true
  try {
    await createManagedDomain(createForm.domain.trim(), createForm.remark.trim())
    createVisible.value = false
    ElMessage.success('已添加')
    await load()
  } finally {
    saving.value = false
  }
}

function openEdit(row: ManagedDomain) {
  editForm.id = row.id
  editForm.domain = row.domain
  editForm.remark = row.remark
  editVisible.value = true
}

async function onEdit() {
  saving.value = true
  try {
    await updateManagedDomain(editForm.id, editForm.remark.trim())
    editVisible.value = false
    ElMessage.success('已保存')
    await load()
  } finally {
    saving.value = false
  }
}

async function onDelete(row: ManagedDomain) {
  await ElMessageBox.confirm(`删除域名 ${row.domain}？删除后普通管理员将无法再使用它。`, '确认删除', {
    type: 'warning',
  })
  await deleteManagedDomain(row.id)
  ElMessage.success('已删除')
  await load()
}

function certLabel(status: string) {
  const map: Record<string, string> = {
    none: '未申请',
    pending: '申请中',
    issued: '已签发',
    failed: '申请失败',
  }
  return map[status] || status || '-'
}

function nginxType(status: string) {
  if (status === 'applied') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'pending') return 'warning'
  return 'info'
}

onMounted(load)
</script>

<style scoped>
.tip {
  margin-bottom: 16px;
}
.domain {
  font-weight: 600;
  color: var(--app-text);
}
.remark {
  margin-top: 2px;
  font-size: 12px;
  color: var(--app-text-3);
}
.tunnel-link {
  color: var(--app-primary);
  text-decoration: none;
}
.tunnel-link:hover {
  text-decoration: underline;
}
.cert {
  margin-top: 4px;
  font-size: 12px;
  color: var(--app-text-3);
}
.muted {
  color: var(--app-text-3);
}
</style>
