<!--
  修改密码弹窗。
  Props:
    modelValue (boolean) 是否显示，配合 v-model。
  Emits:
    update:modelValue  关闭弹窗。
    success            修改成功；由父组件负责退出登录，因为旧 JWT 在服务端仍有效直到过期。
-->
<template>
  <el-dialog
    :model-value="modelValue"
    title="修改密码"
    width="420px"
    :close-on-click-modal="false"
    @update:model-value="emit('update:modelValue', $event)"
    @closed="reset"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="onSubmit">
      <el-form-item label="原密码" prop="oldPassword">
        <el-input v-model="form.oldPassword" type="password" show-password autocomplete="current-password" />
      </el-form-item>
      <el-form-item label="新密码" prop="newPassword">
        <el-input v-model="form.newPassword" type="password" show-password autocomplete="new-password" />
        <div class="hint">8~64 位，建议包含字母、数字与符号</div>
      </el-form-item>
      <el-form-item label="确认新密码" prop="confirm">
        <el-input v-model="form.confirm" type="password" show-password autocomplete="new-password" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="onSubmit">确认修改</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { changePassword } from '../../api/auth'

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ oldPassword: '', newPassword: '', confirm: '' })

/** 前端规则与服务端 validateNewPassword 保持一致，提前拦截明显错误。 */
const rules: FormRules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, max: 64, message: '长度需为 8~64 位', trigger: 'blur' },
    {
      validator: (_r, v: string, cb) => (v && v === form.oldPassword ? cb(new Error('不能与原密码相同')) : cb()),
      trigger: 'blur',
    },
  ],
  confirm: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_r, v: string, cb) => (v !== form.newPassword ? cb(new Error('两次输入不一致')) : cb()),
      trigger: 'blur',
    },
  ],
}

async function onSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await changePassword(form.oldPassword, form.newPassword)
    emit('update:modelValue', false)
    emit('success')
  } finally {
    saving.value = false
  }
}

function reset() {
  form.oldPassword = ''
  form.newPassword = ''
  form.confirm = ''
  formRef.value?.clearValidate()
}
</script>

<style scoped>
.hint {
  font-size: 12px;
  color: var(--app-text-3);
  margin-top: 4px;
  line-height: 1.4;
}
</style>
