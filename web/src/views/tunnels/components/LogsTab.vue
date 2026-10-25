<!--
  单条穿透记录的请求日志。
  Props:
    tunnelId (number) 记录 ID。
    enabled (boolean) 当前是否记录转发日志。
  Emits:
    refresh 开关变更后通知父页面重新拉取详情。
-->
<template>
  <div class="switch-bar">
    <div>
      <div class="title">记录转发日志</div>
      <div class="desc">关闭后新的请求仍然转发，但不再写入日志。已有日志可以继续查看。</div>
    </div>
    <el-switch :model-value="enabled" :loading="saving" @change="onToggle" />
  </div>
  <el-alert
    v-if="!enabled"
    class="hint"
    type="info"
    :closable="false"
    show-icon
    title="当前未记录新的转发日志"
  />
  <LogTable :tunnel-id="tunnelId" />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { setTunnelLogging } from '../../../api/tunnel'
import LogTable from '../../logs/components/LogTable.vue'

const props = defineProps<{ tunnelId: number; enabled: boolean }>()
const emit = defineEmits<{ refresh: [] }>()
const saving = ref(false)

async function onToggle(value: boolean | string | number) {
  const enabled = Boolean(value)
  saving.value = true
  try {
    await setTunnelLogging(props.tunnelId, enabled)
    ElMessage.success(enabled ? '已开始记录转发日志' : '已停止记录转发日志')
    emit('refresh')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.switch-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
  padding: 12px 14px;
  border: 1px solid var(--app-border);
  border-radius: 10px;
  background: #f7f8fc;
}
.title {
  font-weight: 600;
  font-size: 14px;
}
.desc {
  margin-top: 4px;
  font-size: 12px;
  color: var(--app-text-3);
}
.hint {
  margin-bottom: 14px;
}
</style>
