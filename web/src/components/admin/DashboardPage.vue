<script setup lang="ts">
import { NAlert, NButton, NCard, NGrid, NGridItem, NSpace, NTag, useMessage } from 'naive-ui'
import { onMounted, shallowRef } from 'vue'
import { useRouter } from 'vue-router'

import { fetchOverview } from '@/api/admin'
import StatPanel from '@/components/admin/StatPanel.vue'
import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'
import type { OverviewStats } from '@/types/api'

const overview = shallowRef<OverviewStats | null>(null)
const sessionStore = useSessionStore()
const message = useMessage()
const loadError = shallowRef('')
const router = useRouter()

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
        <NSpace>
          <NTag round type="info">OAuth2</NTag>
          <NTag round>Opaque Token</NTag>
        </NSpace>
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
        <StatPanel label="用户数" :value="overview?.users ?? '--'" detail="当前已创建账号总数。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="客户端" :value="overview?.clients ?? '--'" detail="已注册 OAuth 应用数量。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="有效令牌" :value="overview?.access_tokens ?? '--'" detail="未撤销且未过期的 access token。" />
      </NGridItem>
    </NGrid>

    <NGrid cols="1 l:2" responsive="screen" :x-gap="16" :y-gap="16">
      <NGridItem>
        <NCard title="快捷入口">
          <NSpace>
            <NButton tertiary @click="router.push({ name: 'clients' })">客户端</NButton>
            <NButton tertiary @click="router.push({ name: 'tokens' })">令牌</NButton>
            <NButton tertiary @click="router.push({ name: 'connection-info' })">连接信息</NButton>
            <NButton v-if="sessionStore.user?.role === 'admin'" tertiary @click="router.push({ name: 'settings' })">设置</NButton>
          </NSpace>
        </NCard>
      </NGridItem>
      <NGridItem>
        <NCard title="提示">
          先确认客户端回调地址，再发起授权流程。
        </NCard>
      </NGridItem>
    </NGrid>
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
