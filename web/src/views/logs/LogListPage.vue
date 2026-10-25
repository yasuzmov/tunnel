<!-- 全局请求日志页：只负责外层卡片，表格与筛选在 LogTable 中。 -->
<template>
  <div class="page-card">
    <LogTable />
  </div>
</template>

<script setup lang="ts">
import LogTable from './components/LogTable.vue'
</script>
