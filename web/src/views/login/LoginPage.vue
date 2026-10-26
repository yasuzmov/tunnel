<!-- 管理后台登录页：左侧品牌展示，右侧登录表单。 -->
<template>
  <div class="login">
    <LoginBrandPanel />
    <section class="form-side">
      <div class="form-card">
        <h2>欢迎回来</h2>
        <p class="tip">请使用管理员账号登录</p>
        <el-form ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="onSubmit">
          <el-form-item prop="username">
            <el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" autocomplete="username" />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="密码"
              :prefix-icon="Lock"
              show-password
              autocomplete="current-password"
            />
          </el-form-item>
          <el-button class="submit" type="primary" native-type="submit" :loading="loading">登 录</el-button>
        </el-form>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'
import { useAuthStore } from '../../stores/auth'
import LoginBrandPanel from './components/LoginBrandPanel.vue'

const formRef = ref<FormInstance>()
const form = reactive({ username: 'admin', password: '' })
const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()

async function onSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid || loading.value) return
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    await router.push(auth.homePath)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login {
  height: 100%;
  display: flex;
  background: var(--app-bg);
}
.form-side {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
}
.form-card {
  width: 100%;
  max-width: 380px;
  animation: slide-in 0.7s var(--app-ease) 0.1s both;
}
h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
}
.tip {
  margin: 8px 0 32px;
  color: var(--app-text-3);
}
.submit {
  width: 100%;
  height: 44px;
  margin-top: 8px;
  font-size: 15px;
  border-radius: 10px;
  box-shadow: 0 8px 20px rgba(58, 111, 247, 0.3);
  transition: transform 0.2s var(--app-ease), box-shadow 0.2s ease;
}
.submit:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(58, 111, 247, 0.35);
}
.foot {
  margin-top: 28px;
  font-size: 12px;
  color: var(--app-text-3);
  line-height: 1.6;
}
:deep(.el-input__wrapper) {
  border-radius: 10px;
}
@keyframes slide-in {
  from { transform: translate3d(32px, 0, 0); opacity: 0; }
  to { transform: none; opacity: 1; }
}
</style>
