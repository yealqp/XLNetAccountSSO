<script setup lang="ts">
import {
  NConfigProvider,
  NDialogProvider,
  NGlobalStyle,
  NMessageProvider,
  NNotificationProvider,
  darkTheme,
} from 'naive-ui'
import { RouterView } from 'vue-router'

import { darkThemeOverrides } from '@/config/theme'
</script>

<template>
  <NConfigProvider :theme="darkTheme" :theme-overrides="darkThemeOverrides">
    <NDialogProvider>
      <NNotificationProvider>
        <NMessageProvider>
          <NGlobalStyle />
          <div class="app-container">
            <RouterView />
          </div>
        </NMessageProvider>
      </NNotificationProvider>
    </NDialogProvider>
  </NConfigProvider>
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
