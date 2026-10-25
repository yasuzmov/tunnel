<!--
  登录页左侧品牌区：纯色背景 + 产品要点。纯展示，无 Props / Emits。
  窄屏下整块隐藏。
-->
<template>
  <section class="brand">
    <div class="inner">
      <div class="logo">
        <el-icon :size="22"><Connection /></el-icon>
      </div>
      <h1>内网穿透管理平台</h1>
      <p class="sub">把本地服务安全地暴露到公网，请求全链路可追踪。</p>
      <ul class="features">
        <li v-for="f in features" :key="f.title">
          <span class="dot"><el-icon><component :is="f.icon" /></el-icon></span>
          <div>
            <div class="f-title">{{ f.title }}</div>
            <div class="f-desc">{{ f.desc }}</div>
          </div>
        </li>
      </ul>
    </div>
  </section>
</template>

<script setup lang="ts">
import { Connection, Document, Lightning, Lock } from '@element-plus/icons-vue'

const features = [
  { title: 'WebSocket 多路复用隧道', desc: 'HTTP / WebSocket 流式转发，断线自动重连', icon: Lightning },
  { title: '域名与证书自动化', desc: '自动生成 nginx 配置，可选 certbot 签发 HTTPS', icon: Lock },
  { title: '请求日志全记录', desc: 'Header、参数、Body 与耗时一目了然', icon: Document },
]
</script>

<style scoped>
.brand {
  position: relative;
  flex: 1.1;
  overflow: hidden;
  color: #fff;
  background: #1f2937;
  display: flex;
  align-items: center;
  padding: 64px;
}
.inner {
  position: relative;
  max-width: 460px;
  animation: rise 0.8s var(--app-ease) both;
}
.logo {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(8px);
}
h1 {
  margin: 28px 0 12px;
  font-size: 32px;
  font-weight: 700;
  letter-spacing: 1px;
}
.sub {
  margin: 0 0 40px;
  color: rgba(255, 255, 255, 0.75);
  line-height: 1.7;
}
.features {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.features li {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}
.dot {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, 0.14);
}
.f-title {
  font-weight: 600;
}
.f-desc {
  margin-top: 4px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.65);
}
@keyframes rise {
  from { transform: translate3d(0, 20px, 0); opacity: 0; }
  to { transform: none; opacity: 1; }
}
@media (max-width: 900px) {
  .brand { display: none; }
}
</style>
