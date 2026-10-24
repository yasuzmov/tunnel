<!-- 穿透记录详情：四个标签页由子组件承担展示。 -->
<template>
  <div class="page-card" v-loading="!detail">
    <el-page-header @back="router.push('/tunnels')" :content="detail?.tunnel.name || '穿透详情'" />
    <el-tabs v-model="tab" style="margin-top: 16px">
      <el-tab-pane label="基本信息" name="basic">
        <BasicInfoTab v-if="detail" :detail="detail" @refresh="load" />
      </el-tab-pane>
      <el-tab-pane label="客户端联通记录" name="sessions">
        <SessionsTab v-if="detail" :tunnel-id="detail.tunnel.id" @refresh="load" />
      </el-tab-pane>
      <el-tab-pane label="域名配置" name="domains">
        <DomainsTab v-if="detail" :tunnel="detail.tunnel" @refresh="load" />
      </el-tab-pane>
      <el-tab-pane label="请求日志" name="logs">
        <LogsTab v-if="detail" :tunnel-id="detail.tunnel.id" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTunnel } from '../../api/tunnel'
import type { TunnelDetail } from '../../api/types'
import BasicInfoTab from './components/BasicInfoTab.vue'
import SessionsTab from './components/SessionsTab.vue'
import DomainsTab from './components/DomainsTab.vue'
import LogsTab from './components/LogsTab.vue'

const route = useRoute()
const router = useRouter()
const detail = ref<TunnelDetail | null>(null)
const tab = ref('basic')

async function load() {
  detail.value = await getTunnel(Number(route.params.id))
}

onMounted(load)
</script>
