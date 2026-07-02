<script setup lang="ts">
import { ConfigProvider } from '@arco-design/web-vue'
import { RouterView } from 'vue-router'
</script>

<template>
  <ConfigProvider>
    <div class="app-container">
      <RouterView />
    </div>
  </ConfigProvider>
</template>

<style scoped>
.app-container {
  display: flex;
  width: 100vw;
  height: 100vh;
  height: 100dvh;
  overflow: hidden;
}
</style>
