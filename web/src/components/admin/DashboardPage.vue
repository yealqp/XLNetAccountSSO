<script setup lang="ts">
import { NAlert, NButton, NCard, NGrid, NGridItem, NTag, useMessage } from 'naive-ui'
import { onMounted, shallowRef } from 'vue'

import { fetchOverview } from '@/api/admin'
import StatPanel from '@/components/admin/StatPanel.vue'
import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'
import type { OverviewStats } from '@/types/api'

const overview = shallowRef<OverviewStats | null>(null)
const sessionStore = useSessionStore()
const message = useMessage()
const loadError = shallowRef('')

onMounted(async () => {
  await loadOverview()
})

async function loadOverview() {
  loadError.value = ''

  try {
    overview.value = await fetchOverview()
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '加载概览失败'
    message.error(loadError.value)
  }
}
</script>

<template>
  <section class="page-stack">
    <NCard>
      <div class="welcome-header">
        <div>
          <h1 class="page-title">欢迎回来，{{ sessionStore.user?.display_name ?? sessionStore.user?.username }}</h1>
          <p class="page-subtitle">用户、客户端与令牌概览。</p>
        </div>
        <NTag round type="info">Dark</NTag>
      </div>
    </NCard>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <NButton size="small" tertiary @click="loadOverview">重试</NButton>
      </div>
    </NAlert>

    <NGrid cols="1 s:2 l:3" responsive="screen" :x-gap="16" :y-gap="16">
      <NGridItem>
        <StatPanel label="用户数" :value="overview?.users ?? '--'" detail="管理所有登录账户与角色。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="客户端" :value="overview?.clients ?? '--'" detail="已注册的 OAuth2 应用数量。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="有效 Access Token" :value="overview?.access_tokens ?? '--'" detail="尚未撤销且未过期的令牌。" />
      </NGridItem>
    </NGrid>

    <NCard title="联调提示">
      先确认 demo client 的回调地址，再发起授权流程。
    </NCard>
  </section>
</template>

<style scoped>
.welcome-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.page-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

@media (max-width: 820px) {
  .welcome-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .page-alert {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
