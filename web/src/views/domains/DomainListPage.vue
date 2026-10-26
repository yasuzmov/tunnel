<!-- 域名管理：管理员在固定主域名下创建自己的域名；普通管理员只看到自己创建的，超级管理员可见全部。 -->
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
        :title="tipText"
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
        <el-table-column v-if="auth.isSuper" label="创建者" width="130">
          <template #default="{ row }">
            <span v-if="row.created_by">{{ row.creator_name || `#${row.created_by}` }}</span>
            <span v-else class="muted">历史数据</span>
          </template>
        </el-table-column>
        <el-table-column label="绑定记录" min-width="200">
          <template #default="{ row }">
            <router-link
              v-if="row.bound_tunnel_id && auth.can('tunnels')"
              :to="`/tunnels/${row.bound_tunnel_id}`"
              class="tunnel-link"
            >
              {{ row.bound_tunnel_name || `#${row.bound_tunnel_id}` }}
            </router-link>
            <span v-else-if="row.bound_tunnel_id">{{ row.bound_tunnel_name || `#${row.bound_tunnel_id}` }}</span>
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
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
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
          <el-empty description="还没有创建域名" :image-size="90">
            <el-button type="primary" @click="openCreate">添加第一个域名</el-button>
          </el-empty>
        </template>
      </el-table>
    </div>

    <el-dialog v-model="createVisible" title="添加域名" width="440px">
      <el-form label-position="top" @submit.prevent="onCreate">
        <el-form-item label="域名前缀" required>
          <el-input v-model="createForm.name" placeholder="例如 myapp" maxlength="63">
            <template #append>.{{ baseDomain || '未配置主域名' }}</template>
          </el-input>
          <div class="hint">只能填写前缀：小写字母、数字或连字符，1-63 位，不能以连字符开头或结尾。</div>
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

    <el-dialog v-model="editVisible" title="编辑域名" width="440px">
      <el-form label-position="top" @submit.prevent="onEdit">
        <el-form-item v-if="editForm.prefixEditable || editForm.bound" label="域名前缀">
          <el-input v-model="editForm.name" maxlength="63" :disabled="editForm.bound">
            <template #append>.{{ baseDomain }}</template>
          </el-input>
          <div v-if="editForm.bound" class="hint">该域名已绑定到穿透记录，请先解绑再修改前缀。</div>
        </el-form-item>
        <el-form-item v-else label="域名">
          <el-input :model-value="editForm.domain" disabled />
          <div class="hint">历史域名不在当前主域名下，只能修改备注。</div>
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
  getDomainPoolConfig,
  listManagedDomains,
  updateManagedDomain,
} from '../../api/domain'
import type { ManagedDomain } from '../../api/types'
import { useAuthStore } from '../../stores/auth'
import { formatTime } from '../../utils/format'

const auth = useAuthStore()
const list = ref<ManagedDomain[]>([])
const loading = ref(false)
const saving = ref(false)
const keyword = ref('')
const createVisible = ref(false)
const editVisible = ref(false)
const baseDomain = ref('')
const createForm = reactive({ name: '', remark: '' })
const editForm = reactive({
  id: 0,
  domain: '',
  name: '',
  originalName: '',
  remark: '',
  prefixEditable: false,
  bound: false,
})
const labelRe = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/

const tipText = computed(() => {
  const suffix = baseDomain.value ? `*.${baseDomain.value}` : '主域名'
  return auth.isSuper
    ? `域名统一在 ${suffix} 下创建，只需填写前缀。超级管理员可查看全部域名并绑定任意未使用的域名。`
    : `域名统一在 ${suffix} 下创建，只需填写前缀。这里只显示你自己创建的域名，也只有它们能绑定到你的穿透记录。`
})

/** 校验前缀，返回错误提示；合法时返回空字符串。 */
function checkLabel(raw: string) {
  const name = raw.trim().toLowerCase()
  if (!name) return '请填写域名前缀'
  if (name.includes('.')) return '只能填写前缀，不能包含点号'
  if (!labelRe.test(name)) return '前缀只能包含小写字母、数字或连字符，1-63 位，且不能以连字符开头或结尾'
  return ''
}

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

async function loadConfig() {
  try {
    baseDomain.value = (await getDomainPoolConfig()).base_domain
  } catch {
    baseDomain.value = ''
  }
}

function openCreate() {
  createForm.name = ''
  createForm.remark = ''
  createVisible.value = true
}

async function onCreate() {
  if (!baseDomain.value) {
    ElMessage.warning('服务端未配置主域名，暂时无法创建域名')
    return
  }
  const err = checkLabel(createForm.name)
  if (err) {
    ElMessage.warning(err)
    return
  }
  saving.value = true
  try {
    await createManagedDomain(createForm.name.trim().toLowerCase(), createForm.remark.trim())
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
  editForm.name = row.name
  editForm.originalName = row.name
  editForm.remark = row.remark
  editForm.prefixEditable = !!row.name && !!baseDomain.value
  editForm.bound = !!row.bound_tunnel_id && !!row.name
  editVisible.value = true
}

async function onEdit() {
  let name = ''
  if (editForm.prefixEditable && !editForm.bound) {
    const err = checkLabel(editForm.name)
    if (err) {
      ElMessage.warning(err)
      return
    }
    const next = editForm.name.trim().toLowerCase()
    if (next !== editForm.originalName) name = next
  }
  saving.value = true
  try {
    await updateManagedDomain(editForm.id, name, editForm.remark.trim())
    editVisible.value = false
    ElMessage.success('已保存')
    await load()
  } finally {
    saving.value = false
  }
}

async function onDelete(row: ManagedDomain) {
  await ElMessageBox.confirm(`删除域名 ${row.domain}？删除后该域名可被重新创建。`, '确认删除', {
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

onMounted(() => {
  loadConfig()
  load()
})
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
.hint {
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--app-text-3);
}
</style>
