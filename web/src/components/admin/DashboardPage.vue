<script setup lang="ts">
import { NAlert, NButton, NCard, NDescriptions, NDescriptionsItem, NGrid, NGridItem, NIcon, NSpace, NTag, useMessage } from 'naive-ui'
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AppWindow, Link2, Settings, Shield, WalletCards } from 'lucide-vue-next'

import { fetchManagedOverview, fetchOverview } from '@/api/admin'
import StatPanel from '@/components/admin/StatPanel.vue'
import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'
import type { OverviewStats } from '@/types/api'

const overview = shallowRef<OverviewStats | null>(null)
const sessionStore = useSessionStore()
const message = useMessage()
const loadError = shallowRef('')
const router = useRouter()
const route = useRoute()
const manageAll = computed(() => route.meta.manageScope === 'all')
const userProfileItems = computed(() => [
  {
    label: '用户 ID',
    value: String(sessionStore.user?.id ?? '-'),
  },
  {
    label: '用户名',
    value: sessionStore.user?.username ?? '-',
  },
  {
    label: '绑定邮箱',
    value: sessionStore.user?.email || '-',
  },
])

onMounted(async () => {
  await loadOverview()
})

watch(() => route.fullPath, () => {
  void loadOverview()
})

async function loadOverview() {
  loadError.value = ''

  try {
    overview.value = await (manageAll.value ? fetchManagedOverview() : fetchOverview())
  }
  catch (error) {
    loadError.value = error instanceof ApiError ? error.message : '加载概览失败'
    message.error(loadError.value)
  }
}
</script>

<template>
  <section class="page-stack">
    <NCard v-if="manageAll">
      <div class="welcome-header">
        <div>
          <h1 class="page-title">欢迎回来，{{ sessionStore.user?.username }}</h1>
          <p class="page-subtitle">{{ manageAll ? '查看平台整体资源概览。' : '查看您当前账号的资源概览。' }}</p>
        </div>
        <NSpace>
          <NTag round type="info">OAuth2</NTag>
          <NTag round>Opaque Token</NTag>
        </NSpace>
      </div>
    </NCard>

    <NCard v-else>
      <div class="welcome-header">
        <div>
          <h1 class="page-title">用户概览</h1>
          <p class="page-subtitle">查看当前账号信息与您拥有的资源数量。</p>
        </div>
        <NTag round type="info">{{ sessionStore.user?.role ?? 'user' }}</NTag>
      </div>

      <NDescriptions label-placement="top" :column="3" bordered>
        <NDescriptionsItem v-for="item in userProfileItems" :key="item.label" :label="item.label">
          {{ item.value }}
        </NDescriptionsItem>
      </NDescriptions>
    </NCard>

    <NAlert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <NButton size="small" tertiary @click="loadOverview">重试</NButton>
      </div>
    </NAlert>

    <NGrid :cols="manageAll ? '1 s:2 l:3' : '1 s:2'" responsive="screen" :x-gap="16" :y-gap="16">
      <NGridItem v-if="manageAll">
        <StatPanel label="用户数" :value="overview?.users ?? '--'" detail="当前已创建账号总数。" />
      </NGridItem>
      <NGridItem>
        <StatPanel :label="manageAll ? '客户端' : '应用数量'" :value="overview?.clients ?? '--'" :detail="manageAll ? '已注册 OAuth 应用数量。' : '当前账号拥有的应用数量。'" />
      </NGridItem>
      <NGridItem>
        <StatPanel :label="manageAll ? '有效令牌' : '令牌数量'" :value="overview?.access_tokens ?? '--'" :detail="manageAll ? '未撤销且未过期的 access token。' : '当前账号拥有的有效令牌数量。'" />
      </NGridItem>
    </NGrid>

    <NGrid cols="1 l:2" responsive="screen" :x-gap="16" :y-gap="16">
      <NGridItem>
        <NCard title="快捷入口">
          <NSpace>
            <NButton tertiary @click="router.push({ name: manageAll ? 'manage-applications' : 'applications' })">
              <template #icon><NIcon><AppWindow /></NIcon></template>
              {{ manageAll ? '应用管理' : '应用' }}
            </NButton>
            <NButton tertiary @click="router.push({ name: manageAll ? 'manage-tokens' : 'tokens' })">
              <template #icon><NIcon><WalletCards /></NIcon></template>
              {{ manageAll ? '令牌管理' : '令牌' }}
            </NButton>
            <NButton tertiary @click="router.push({ name: 'settings' })">
              <template #icon><NIcon><Settings /></NIcon></template>
              普通设置
            </NButton>
            <NButton tertiary @click="router.push({ name: 'connection-info' })">
              <template #icon><NIcon><Link2 /></NIcon></template>
              连接信息
            </NButton>
            <NButton v-if="sessionStore.user?.role === 'admin'" tertiary @click="router.push({ name: 'system-settings' })">
              <template #icon><NIcon><Shield /></NIcon></template>
              系统设置
            </NButton>
          </NSpace>
        </NCard>
      </NGridItem>
      <NGridItem>
        <NCard :title="manageAll ? '管理提示' : '使用提示'">
          {{ manageAll ? '在管理页面可以查看全量资源，并识别资源所属用户。' : '您只能查看和维护自己拥有的应用与令牌。' }}
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
