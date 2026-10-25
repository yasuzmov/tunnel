<!-- 根组件：登录页与后台框架之间做整屏过渡。 -->
<template>
  <div class="app-root">
    <router-view v-slot="{ Component, route }">
      <transition name="screen">
        <component :is="Component" :key="route.path === '/login' ? 'login' : 'main'" />
      </transition>
    </router-view>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.app-root {
  position: relative;
  height: 100%;
  overflow: hidden;
}
</style>
