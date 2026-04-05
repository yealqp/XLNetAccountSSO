<script setup lang="ts">
import { NLayout, NLayoutContent, NLayoutHeader, NSelect, NSpace, NText } from 'naive-ui'
import type { SelectOption } from 'naive-ui'
import { computed } from 'vue'
import { RouterView } from 'vue-router'

import { useThemeMode } from '@/composables/useThemeMode'

const { preference, effectiveMode, osTheme } = useThemeMode()

const themeOptions: SelectOption[] = [
  { label: '跟随系统', value: 'system' },
  { label: '浅色', value: 'light' },
  { label: '深色', value: 'dark' },
]

const themeHint = computed(() => {
  if (preference.value === 'system') {
    return `系统：${osTheme.value === 'dark' ? '深色' : '浅色'}`
  }

  return `当前：${effectiveMode.value === 'dark' ? '深色' : '浅色'}`
})
</script>

<template>
  <NLayout class="page-shell auth-layout">
    <NLayoutHeader bordered class="auth-header">
      <div>
        <div class="auth-title">XLNetAccount</div>
        <NText depth="3">统一账号与授权</NText>
      </div>

      <NSpace align="center" :wrap="true">
        <NText depth="3">{{ themeHint }}</NText>
        <NSelect v-model:value="preference" :options="themeOptions" size="small" class="theme-select" />
      </NSpace>
    </NLayoutHeader>

    <NLayoutContent class="auth-content">
      <div class="auth-container">
        <RouterView />
      </div>
    </NLayoutContent>
  </NLayout>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  min-height: 100dvh;
}

.auth-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
}

.auth-title {
  font-size: 22px;
  font-weight: 600;
}

.auth-content {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
}

.auth-container {
  width: 100%;
  max-width: 720px;
}

.theme-select {
  width: 120px;
}

@media (max-width: 720px) {
  .auth-header {
    align-items: flex-start;
  }

  .auth-content {
    padding: 16px;
  }
}
</style>
