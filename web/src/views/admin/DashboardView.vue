<script setup lang="ts">
import { NCard, NGrid, NGridItem, NTag } from 'naive-ui'
import { onMounted, shallowRef } from 'vue'

import { fetchOverview } from '@/api/admin'
import StatPanel from '@/components/admin/StatPanel.vue'
import type { OverviewStats } from '@/types/api'

const overview = shallowRef<OverviewStats | null>(null)

onMounted(async () => {
  overview.value = await fetchOverview()
})
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <h1 class="page-title">系统总览</h1>
        <p class="page-subtitle">用户、客户端、会话与令牌概览。</p>
      </div>
      <NTag round type="info">Opaque Token</NTag>
    </header>

    <NGrid cols="1 s:2 l:4" responsive="screen" :x-gap="16" :y-gap="16">
      <NGridItem>
        <StatPanel label="用户数" :value="overview?.users ?? '--'" detail="管理所有登录账户与角色。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="客户端" :value="overview?.clients ?? '--'" detail="已注册的 OAuth2 应用数量。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="活跃会话" :value="overview?.active_sessions ?? '--'" detail="当前仍在有效期内的登录会话。" />
      </NGridItem>
      <NGridItem>
        <StatPanel label="有效 Access Token" :value="overview?.access_tokens ?? '--'" detail="尚未撤销且未过期的令牌。" />
      </NGridItem>
    </NGrid>

    <NCard title="联调提示" class="dashboard-card">
      先确认 demo client 的回调地址，再发起授权流程。
    </NCard>
  </section>
</template>

<style scoped>
.dashboard-card {
  margin-top: 24px;
}
</style>
