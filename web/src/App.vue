<script setup lang="ts">
import {
  NConfigProvider,
  NDialogProvider,
  NGlobalStyle,
  NMessageProvider,
  NNotificationProvider,
  darkTheme,
} from 'naive-ui'
import { computed } from 'vue'
import { RouterView } from 'vue-router'

import { useThemeMode } from '@/composables/useThemeMode'

const { effectiveMode } = useThemeMode()

const theme = computed(() => (effectiveMode.value === 'dark' ? darkTheme : null))
</script>

<template>
  <NConfigProvider :theme="theme">
    <NDialogProvider>
      <NNotificationProvider>
        <NMessageProvider>
          <NGlobalStyle />
          <RouterView />
        </NMessageProvider>
      </NNotificationProvider>
    </NDialogProvider>
  </NConfigProvider>
</template>
