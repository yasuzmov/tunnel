<template>
  <LogTable :tunnel-id="tunnelId" />
</template>

<script setup lang="ts">
import LogTable from '../../logs/components/LogTable.vue'

defineProps<{ tunnelId: number }>()
</script>
