<template>
  <el-drawer :model-value="modelValue" title="请求详情" size="52%" @close="emit('update:modelValue', false)">
    <div v-if="item">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="Request ID">{{ item.request_id }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ item.status }}</el-descriptions-item>
        <el-descriptions-item label="方法">{{ item.method }}</el-descriptions-item>
        <el-descriptions-item label="URL">{{ item.scheme }}://{{ item.host }}{{ item.path }}</el-descriptions-item>
        <el-descriptions-item label="客户端 IP">{{ item.remote_ip }}</el-descriptions-item>
        <el-descriptions-item label="本地目标">{{ item.local_target || '-' }}</el-descriptions-item>
        <el-descriptions-item label="命中规则">{{ item.matched_rule || '-' }}</el-descriptions-item>
        <el-descriptions-item label="错误">{{ item.error_message || '-' }}</el-descriptions-item>
      </el-descriptions>
      <h4>时间线</h4>
      <p>接收 {{ formatTime(item.received_at) }} / 转发 {{ formatTime(item.forwarded_at) }} / 首字节 {{ formatTime(item.first_byte_at) }} / 完成 {{ formatTime(item.completed_at) }}</p>
      <p>总耗时 {{ item.duration_ms }} ms，上游 {{ item.upstream_ms }} ms</p>
      <el-button type="primary" @click="copyText(toCurl(item), '已复制 cURL')">复制为 cURL</el-button>
      <h4>Query</h4>
      <pre class="mono">{{ prettyJSON(item.query_params) }}</pre>
      <h4>请求头</h4>
      <pre class="mono">{{ prettyJSON(item.request_headers) }}</pre>
      <h4>请求体 {{ truncated(item.body?.request_truncated) }}</h4>
      <pre class="mono">{{ decodeBody(item.body?.request_body) }}</pre>
      <h4>响应头</h4>
      <pre class="mono">{{ prettyJSON(item.response_headers) }}</pre>
      <h4>响应体 {{ truncated(item.body?.response_truncated) }}</h4>
      <pre class="mono">{{ decodeBody(item.body?.response_body) }}</pre>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getLog } from '../../../api/log'
import type { RequestLog } from '../../../api/types'
import { copyText } from '../../../utils/copy'
import { toCurl } from '../../../utils/curl'
import { decodeBody, formatTime, prettyJSON } from '../../../utils/format'

const props = defineProps<{ modelValue: boolean; logId: number }>()
const emit = defineEmits<{ 'update:modelValue': [boolean] }>()
const item = ref<RequestLog | null>(null)

watch(
  () => [props.modelValue, props.logId] as const,
  async ([open, id]) => {
    if (open && id) {
      item.value = await getLog(id)
    }
  },
)

function truncated(v?: boolean) {
  return v ? '(已截断)' : ''
}
</script>

<style scoped>
pre { white-space: pre-wrap; background: #f5f7fa; padding: 8px; max-height: 240px; overflow: auto; }
h4 { margin: 16px 0 8px; }
</style>
