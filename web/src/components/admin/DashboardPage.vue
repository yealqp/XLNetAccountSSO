<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Button, Card, Descriptions, DescriptionsItem, Grid, GridItem, Message, Space, Tag } from '@arco-design/web-vue'
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AppWindow, Link2, Settings, Shield, WalletCards } from 'lucide-vue-next'

useHead({ title: '概览 — XLNetAccount' })

import { fetchManagedOverview, fetchOverview } from '@/api/admin'
import StatPanel from '@/components/admin/StatPanel.vue'
import { ApiError } from '@/api/http'
import { useSessionStore } from '@/stores/session'
import type { OverviewStats } from '@/types/api'

const overview = shallowRef<OverviewStats | null>(null)
const sessionStore = useSessionStore()
const loadError = shallowRef('')
const router = useRouter()
const route = useRoute()
const manageAll = computed(() => route.meta.manageScope === 'all')

const welcomeTitle = computed(() => manageAll.value ? '欢迎回来' : '用户概览')
const welcomeSubtitle = computed(() => manageAll.value ? '平台整体资源概览。' : '当前账号信息与资源概览。')

const quickLinks = computed(() => {
  const links = [
    { name: manageAll.value ? 'manage-applications' : 'applications', icon: AppWindow, label: manageAll.value ? '应用管理' : '应用' },
    { name: manageAll.value ? 'manage-tokens' : 'tokens', icon: WalletCards, label: manageAll.value ? '令牌管理' : '令牌' },
    { name: 'connection-info', icon: Link2, label: '连接信息' },
    { name: 'settings', icon: Settings, label: '普通设置' },
  ]
  if (sessionStore.user?.role === 'admin') {
    links.push({ name: 'system-settings', icon: Shield, label: '系统设置' })
  }
  return links
})

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
    Message.error(loadError.value)
  }
}
</script>

<template>
  <section class="page-stack">
    <div class="overview-hero">
      <div class="overview-hero-main">
        <div class="overview-hero-text">
          <h1 class="page-title">{{ welcomeTitle }}，{{ sessionStore.user?.username }}</h1>
          <p class="page-subtitle">{{ welcomeSubtitle }}</p>
        </div>
        <Space wrap>
          <Tag round>OAuth2</Tag>
        </Space>
      </div>
      <Descriptions v-if="!manageAll" layout="vertical" :column="3" size="small" class="overview-profile">
        <DescriptionsItem label="用户 ID">
          {{ sessionStore.user?.id ?? '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="用户名">
          {{ sessionStore.user?.username ?? '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="绑定邮箱">
          {{ sessionStore.user?.email || '-' }}
        </DescriptionsItem>
      </Descriptions>
    </div>

    <Alert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <Button size="small" type="text" @click="loadOverview">重试</Button>
      </div>
    </Alert>

    <Grid :cols="manageAll ? {xs:1,sm:2,lg:4} : {xs:1,sm:2}" :col-gap="12" :row-gap="12">
      <GridItem v-if="manageAll">
        <StatPanel label="用户数" :value="overview?.users ?? '--'" detail="已创建账号总数" />
      </GridItem>
      <GridItem>
        <StatPanel
          :label="manageAll ? '客户端' : '应用数量'"
          :value="overview?.clients ?? '--'"
          :detail="manageAll ? '已注册 OAuth 应用' : '当前账号拥有的应用'"
        />
      </GridItem>
      <GridItem>
        <StatPanel
          :label="manageAll ? '有效令牌' : '令牌数量'"
          :value="overview?.access_tokens ?? '--'"
          :detail="manageAll ? '未撤销且未过期的 access token' : '当前账号的有效令牌'"
        />
      </GridItem>
    </Grid>

    <Card size="small">
      <div class="quick-links">
        <button
          v-for="link in quickLinks"
          :key="link.name"
          class="quick-link-btn"
          @click="router.push({ name: link.name })"
        >
          <component :is="link.icon" :size="18" style="vertical-align: middle;" />
          <span>{{ link.label }}</span>
        </button>
      </div>
    </Card>
  </section>
</template>

<style scoped>
.overview-hero {
  display: grid;
  gap: 12px;
}

.overview-hero-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.overview-hero-text {
  min-width: 0;
}

.overview-profile {
  margin-top: 0;
}

.page-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.quick-links {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.quick-link-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--color-hairline-strong);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-body);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.quick-link-btn:hover {
  background: rgba(0, 0, 0, 0.04);
  border-color: rgba(0, 0, 0, 0.2);
}

@media (max-width: 820px) {
  .overview-hero-main {
    flex-direction: column;
  }

  .page-alert {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
