<!-- 管理员管理：账号、角色与可用域名授权，仅超级管理员可访问。 -->
<template>
  <div class="page-stack">
    <div class="page-card">
      <div class="toolbar">
        <div class="hint">超级管理员拥有全部菜单与域名；普通管理员只能访问被授权的菜单，并绑定被授权的域名。</div>
        <el-button type="primary" :icon="Plus" @click="openCreate">添加管理员</el-button>
      </div>
      <el-table :data="list" v-loading="loading">
        <el-table-column label="管理员" min-width="230">
          <template #default="{ row }">
            <div class="user-cell">
              <div class="avatar" :class="{ super: row.role === 'super_admin' }">{{ row.username.slice(0, 1).toUpperCase() }}</div>
              <div>
                <div class="name">
                  {{ row.username }}
                  <span v-if="row.id === auth.user?.id" class="self">当前登录</span>
                </div>
                <div class="meta">创建于 {{ formatTime(row.created_at) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="130">
          <template #default="{ row }">
            <el-tag :type="row.role === 'super_admin' ? 'warning' : 'info'" effect="light" size="small">
              {{ roleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="可见菜单" min-width="240">
          <template #default="{ row }">
            <el-tag v-if="row.role === 'super_admin'" type="success" effect="plain" size="small">全部菜单</el-tag>
            <template v-else-if="row.menus?.length">
              <el-tag v-for="m in row.menus" :key="m" size="small" effect="plain" class="menu-tag">{{ menuLabel(m) }}</el-tag>
            </template>
            <span v-else class="muted">默认菜单</span>
          </template>
        </el-table-column>
        <el-table-column label="可用域名" min-width="280">
          <template #default="{ row }">
            <el-tag v-if="row.role === 'super_admin'" type="success" effect="plain" size="small">全部域名</el-tag>
            <template v-else-if="row.domain_names.length">
              <el-tag v-for="d in row.domain_names" :key="d" size="small" effect="plain" class="domain-tag">{{ d }}</el-tag>
            </template>
            <span v-else class="muted">未授权任何域名</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" align="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="openPermissions(row)">域名权限</el-button>
            <el-button link type="primary" @click="openReset(row)">重置密码</el-button>
            <el-button
              link
              type="danger"
              :disabled="row.id === auth.user?.id"
              :title="row.id === auth.user?.id ? '不能删除当前登录账号' : ''"
              @click="onDelete(row)"
            >删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="还没有管理员账号" :image-size="90">
            <el-button type="primary" @click="openCreate">添加管理员</el-button>
          </el-empty>
        </template>
      </el-table>
    </div>

    <el-dialog v-model="formVisible" :title="formForm.id ? '编辑管理员' : '添加管理员'" width="440px">
      <el-form label-position="top" @submit.prevent="onSubmit">
        <el-form-item label="用户名" required>
          <el-input v-model="formForm.username" maxlength="64" placeholder="例如：ops" />
        </el-form-item>
        <el-form-item label="角色" required>
          <el-select v-model="formForm.role" :disabled="formForm.id > 0 && formForm.id === auth.user?.id" style="width: 100%">
            <el-option label="普通管理员（按菜单与域名授权）" value="admin" />
            <el-option label="超级管理员（全部权限）" value="super_admin" />
          </el-select>
          <div v-if="formForm.id === auth.user?.id" class="field-tip">不能修改自己的角色</div>
        </el-form-item>
        <el-form-item v-if="formForm.role !== 'super_admin'" label="可见菜单" required>
          <el-checkbox-group v-model="formForm.menus" class="menu-options">
            <el-checkbox v-for="m in GRANTABLE_MENUS" :key="m.key" :value="m.key">{{ m.label }}</el-checkbox>
          </el-checkbox-group>
          <div class="field-tip">至少选择一个菜单；「管理员管理」仅超级管理员可用。</div>
        </el-form-item>
        <el-form-item v-else label="可见菜单">
          <span class="muted">超级管理员默认拥有全部菜单</span>
        </el-form-item>
        <el-form-item v-if="!formForm.id" label="初始密码" required>
          <el-input v-model="formForm.password" type="password" show-password maxlength="64" placeholder="8~64 位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSubmit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permVisible" title="域名权限" width="520px">
      <div class="perm-head">
        <div>
          <div class="perm-user">{{ permForm.username }}</div>
          <div class="perm-sub">勾选该管理员可以绑定到穿透记录的域名</div>
        </div>
        <div v-if="permForm.role !== 'super_admin'" class="perm-actions">
          <el-button link type="primary" @click="selectAll">全选</el-button>
          <el-button link @click="clearAll">清空</el-button>
        </div>
      </div>
      <el-alert
        v-if="permForm.role === 'super_admin'"
        title="超级管理员默认拥有全部域名，无需单独授权。"
        type="info"
        show-icon
        :closable="false"
      />
      <template v-else>
        <div v-if="domains.length" class="perm-list">
          <el-checkbox-group v-model="permForm.domainIds">
            <el-checkbox v-for="d in domains" :key="d.id" :value="d.id" class="perm-item">
              <span class="perm-domain">{{ d.domain }}</span>
              <span v-if="d.remark" class="perm-remark">{{ d.remark }}</span>
            </el-checkbox>
          </el-checkbox-group>
        </div>
        <el-empty v-else description="域名列表为空，请先到「域名管理」添加域名" :image-size="70" />
      </template>
      <template #footer>
        <el-button @click="permVisible = false">关闭</el-button>
        <el-button v-if="permForm.role !== 'super_admin'" type="primary" :loading="saving" @click="onSavePermissions">
          保存
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resetVisible" title="重置密码" width="420px">
      <el-form label-position="top" @submit.prevent="onReset">
        <el-form-item label="管理员">
          <el-input :model-value="resetForm.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" required>
          <el-input v-model="resetForm.password" type="password" show-password maxlength="64" placeholder="8~64 位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onReset">重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  createAdmin,
  deleteAdmin,
  listAdmins,
  resetAdminPassword,
  setAdminDomains,
  updateAdmin,
} from '../../api/admin'
import { listManagedDomains } from '../../api/domain'
import type { AdminAccount, AdminRole, ManagedDomain, MenuKey } from '../../api/types'
import { useAuthStore } from '../../stores/auth'
import { formatTime } from '../../utils/format'
import { DEFAULT_ADMIN_MENUS, GRANTABLE_MENUS, menuLabel } from '../../utils/menus'

const auth = useAuthStore()
const list = ref<AdminAccount[]>([])
const domains = ref<ManagedDomain[]>([])
const loading = ref(false)
const saving = ref(false)

const formVisible = ref(false)
const formForm = reactive<{ id: number; username: string; role: AdminRole; password: string; menus: MenuKey[] }>({
  id: 0,
  username: '',
  role: 'admin',
  password: '',
  menus: [...DEFAULT_ADMIN_MENUS],
})

const permVisible = ref(false)
const permForm = reactive<{ id: number; username: string; role: AdminRole; domainIds: number[] }>({
  id: 0,
  username: '',
  role: 'admin',
  domainIds: [],
})

const resetVisible = ref(false)
const resetForm = reactive({ id: 0, username: '', password: '' })

async function load() {
  loading.value = true
  try {
    list.value = await listAdmins()
  } finally {
    loading.value = false
  }
}

async function loadDomains() {
  domains.value = await listManagedDomains()
}

function roleLabel(role: AdminRole) {
  return role === 'super_admin' ? '超级管理员' : '管理员'
}

function openCreate() {
  formForm.id = 0
  formForm.username = ''
  formForm.role = 'admin'
  formForm.password = ''
  formForm.menus = [...DEFAULT_ADMIN_MENUS]
  formVisible.value = true
}

function openEdit(row: AdminAccount) {
  formForm.id = row.id
  formForm.username = row.username
  formForm.role = row.role
  formForm.password = ''
  const menus = (row.menus ?? []).filter((menu) => menu !== 'admins')
  formForm.menus = menus.length ? menus : [...DEFAULT_ADMIN_MENUS]
  formVisible.value = true
}

async function onSubmit() {
  const username = formForm.username.trim()
  if (!username) {
    ElMessage.warning('请填写用户名')
    return
  }
  if (!formForm.id && (formForm.password.length < 8 || formForm.password.length > 64)) {
    ElMessage.warning('密码长度需为 8~64 位')
    return
  }
  if (formForm.role !== 'super_admin' && !formForm.menus.length) {
    ElMessage.warning('请至少选择一个可见菜单')
    return
  }
  saving.value = true
  try {
    const menus = formForm.role === 'super_admin' ? undefined : formForm.menus
    if (formForm.id) {
      await updateAdmin(formForm.id, username, formForm.role, menus)
      ElMessage.success('已保存')
    } else {
      await createAdmin(username, formForm.password, formForm.role, menus)
      ElMessage.success('已添加管理员')
    }
    formVisible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

async function openPermissions(row: AdminAccount) {
  permForm.id = row.id
  permForm.username = row.username
  permForm.role = row.role
  permForm.domainIds = [...row.domain_ids]
  if (!domains.value.length) {
    await loadDomains()
  }
  permVisible.value = true
}

function selectAll() {
  permForm.domainIds = domains.value.map((d) => d.id)
}

function clearAll() {
  permForm.domainIds = []
}

async function onSavePermissions() {
  saving.value = true
  try {
    await setAdminDomains(permForm.id, permForm.domainIds)
    permVisible.value = false
    ElMessage.success('域名权限已更新')
    await load()
  } finally {
    saving.value = false
  }
}

function openReset(row: AdminAccount) {
  resetForm.id = row.id
  resetForm.username = row.username
  resetForm.password = ''
  resetVisible.value = true
}

async function onReset() {
  if (resetForm.password.length < 8 || resetForm.password.length > 64) {
    ElMessage.warning('密码长度需为 8~64 位')
    return
  }
  saving.value = true
  try {
    await resetAdminPassword(resetForm.id, resetForm.password)
    resetVisible.value = false
    ElMessage.success('密码已重置')
  } finally {
    saving.value = false
  }
}

async function onDelete(row: AdminAccount) {
  await ElMessageBox.confirm(`删除管理员「${row.username}」？该账号将无法再登录后台。`, '确认删除', {
    type: 'warning',
  })
  await deleteAdmin(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(async () => {
  await Promise.all([load(), loadDomains()])
})
</script>

<style scoped>
.hint {
  font-size: 13px;
  color: var(--app-text-3);
}
.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}
.avatar {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: var(--app-primary);
  background: var(--app-primary-soft);
}
.avatar.super {
  color: #b45309;
  background: #fef3c7;
}
.name {
  font-weight: 600;
  color: var(--app-text);
}
.self {
  margin-left: 6px;
  padding: 1px 6px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  color: var(--app-primary);
  background: var(--app-primary-soft);
}
.meta {
  margin-top: 2px;
  font-size: 12px;
  color: var(--app-text-3);
  white-space: nowrap;
}
.domain-tag,
.menu-tag {
  margin: 2px 6px 2px 0;
}
.muted {
  color: var(--app-text-3);
}
.field-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--app-text-3);
}
.menu-options {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 16px;
}
.menu-options :deep(.el-checkbox) {
  margin-right: 0;
}
.perm-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 12px;
}
.perm-user {
  font-weight: 600;
}
.perm-sub {
  margin-top: 2px;
  font-size: 12px;
  color: var(--app-text-3);
}
.perm-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}
.perm-list {
  margin-top: 12px;
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius-sm);
  padding: 4px 12px;
}
.perm-item {
  display: flex;
  width: 100%;
  height: 38px;
  margin-right: 0;
}
.perm-domain {
  color: var(--app-text);
}
.perm-remark {
  margin-left: 8px;
  font-size: 12px;
  color: var(--app-text-3);
}
</style>
