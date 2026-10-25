<!-- 仪表盘「接入指引」卡片：按顺序展示从建记录到绑定域名的步骤。纯展示，无 Props / Emits。 -->
<template>
  <div class="page-card">
    <div class="head">
      <h3>接入指引</h3>
      <el-button v-if="auth.can('tunnels')" type="primary" link @click="router.push('/tunnels')">
        前往穿透记录
      </el-button>
    </div>
    <ol class="steps">
      <li v-for="(s, i) in steps" :key="s.title">
        <span class="num">{{ i + 1 }}</span>
        <div>
          <div class="s-title">{{ s.title }}</div>
          <div class="s-desc">{{ s.desc }}</div>
        </div>
      </li>
    </ol>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const steps = [
  { title: '新建穿透记录', desc: '系统生成唯一的接入 Key 与客户端接入地址' },
  { title: '配置客户端', desc: '在 Windows 客户端填写地址与 Key，测试连通后连接' },
  { title: '设置转发规则', desc: '按端口或 hosts 域名把请求转给本地服务' },
  { title: '绑定公网域名', desc: '自动写入 nginx 配置，可选签发 HTTPS 证书' },
]
</script>

<style scoped>
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
h3 {
  margin: 0;
  font-size: 16px;
}
.steps {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}
.steps li {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-radius: 10px;
  background: #f7f9fe;
}
.num {
  width: 26px;
  height: 26px;
  flex-shrink: 0;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-size: 13px;
  font-weight: 700;
  color: #fff;
  background: var(--app-primary);
}
.s-title {
  font-weight: 600;
  font-size: 14px;
}
.s-desc {
  margin-top: 4px;
  font-size: 12px;
  color: var(--app-text-2);
  line-height: 1.5;
}
</style>
