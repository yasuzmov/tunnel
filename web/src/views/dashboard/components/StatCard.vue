<!--
  仪表盘统计卡片。
  Props:
    label (string)            指标名称。
    value (string | number)   指标值。
    icon (Component)          左侧图标组件。
    color (string, 默认 '#3a6ff7') 主题色，用于图标底色与装饰光晕。
    hint (string, 默认 '')    底部辅助说明。
-->
<template>
  <div class="stat" :style="{ '--c': color }">
    <div class="glow" />
    <div class="icon"><el-icon :size="20"><component :is="icon" /></el-icon></div>
    <div class="label">{{ label }}</div>
    <div class="value">{{ value }}</div>
    <div v-if="hint" class="hint">{{ hint }}</div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

withDefaults(
  defineProps<{ label: string; value: string | number; icon: Component; color?: string; hint?: string }>(),
  { color: '#3a6ff7', hint: '' },
)
</script>

<style scoped>
.stat {
  position: relative;
  overflow: hidden;
  padding: 20px;
  border-radius: var(--app-radius);
  background: var(--app-surface);
  border: 1px solid var(--app-border);
  box-shadow: var(--app-shadow);
  transition: transform 0.3s var(--app-ease), box-shadow 0.3s ease;
}
.stat:hover {
  transform: translateY(-3px);
  box-shadow: var(--app-shadow-lg);
}
.glow {
  position: absolute;
  right: -40px;
  top: -40px;
  width: 140px;
  height: 140px;
  border-radius: 50%;
  background: var(--c);
  opacity: 0.08;
}
.icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  color: var(--c);
  background: color-mix(in srgb, var(--c) 12%, transparent);
}
.label {
  margin-top: 16px;
  font-size: 13px;
  color: var(--app-text-2);
}
.value {
  margin-top: 6px;
  font-size: 30px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--app-text-3);
}
</style>
