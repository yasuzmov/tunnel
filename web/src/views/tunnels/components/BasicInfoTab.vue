<template>
  <el-descriptions :column="1" border>
    <el-descriptions-item label="名称">{{ detail.tunnel.name }}</el-descriptions-item>
    <el-descriptions-item label="备注">{{ detail.tunnel.remark || '-' }}</el-descriptions-item>
    <el-descriptions-item label="接入地址">
      <span class="mono">{{ detail.connect_url }}</span>
      <el-button link type="primary" @click="copyText(detail.connect_url)">复制</el-button>
    </el-descriptions-item>
    <el-descriptions-item label="接入 Key">
      <span class="mono">{{ detail.tunnel.key }}</span>
      <el-button link type="primary" @click="copyText(detail.tunnel.key)">复制</el-button>
      <el-button link type="danger" @click="onReset">重置</el-button>
    </el-descriptions-item>
    <el-descriptions-item label="最近在线">{{ formatTime(detail.tunnel.last_online_at) }}</el-descriptions-item>
  </el-descriptions>
  <el-button style="margin-top: 16px" type="primary" @click="copyBoth">一键复制地址 + Key</el-button>
</template>

<script setup lang="ts">
import { ElMessageBox } from 'element-plus'
import { resetTunnelKey } from '../../../api/tunnel'
import type { TunnelDetail } from '../../../api/types'
import { copyText } from '../../../utils/copy'
import { formatTime } from '../../../utils/format'

const props = defineProps<{ detail: TunnelDetail }>()
const emit = defineEmits<{ refresh: [] }>()

async function onReset() {
  await ElMessageBox.confirm('重置后旧客户端会立即掉线，确定继续？', '重置 Key')
  await resetTunnelKey(props.detail.tunnel.id)
  emit('refresh')
}

function copyBoth() {
  copyText(`地址: ${props.detail.connect_url}\nKey: ${props.detail.tunnel.key}`, '已复制地址和 Key')
}
</script>
