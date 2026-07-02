<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { Alert, Button, Card, Grid, GridItem, Message, Tag } from '@arco-design/web-vue'
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AppWindow, Cog, Link2, Shield, User, WalletCards } from '@lucide/vue'

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
    { name: manageAll.value ? 'manage-applications' : 'applications', icon: AppWindow, label: manageAll.value ? '应用管理' : '应用', desc: manageAll.value ? '管理全部 OAuth 应用' : '查看和管理我的应用' },
    { name: manageAll.value ? 'manage-tokens' : 'tokens', icon: WalletCards, label: manageAll.value ? '令牌管理' : '令牌', desc: manageAll.value ? '查看和吊销访问令牌' : '查看和吊销我的令牌' },
    { name: 'connection-info', icon: Link2, label: '连接信息', desc: 'OAuth2 / OIDC 端点详情' },
    { name: 'settings', icon: Cog, label: '普通设置', desc: '修改账号密码和安全设置' },
  ]
  if (sessionStore.user?.role === 'admin') {
    links.push({ name: 'system-settings', icon: Shield, label: '系统设置', desc: '平台名称、注册开关等' })
  }
  return links
})

const statsConfig = computed(() => {
  if (manageAll.value) {
    return [
      { label: '用户数', value: overview.value?.users ?? '--', detail: '已创建账号总数' },
      { label: '应用', value: overview.value?.clients ?? '--', detail: '已注册 OAuth 应用' },
      { label: '活跃会话', value: overview.value?.active_sessions ?? '--', detail: '当前有效登录会话' },
      { label: '有效令牌', value: overview.value?.access_tokens ?? '--', detail: '未撤销的 access token' },
    ]
  }
  return [
    { label: '应用', value: overview.value?.clients ?? '--', detail: '我创建的 OAuth 应用' },
    { label: '有效令牌', value: overview.value?.access_tokens ?? '--', detail: '我的活跃 access token' },
    { label: '活跃会话', value: overview.value?.active_sessions ?? '--', detail: '当前有效登录会话' },
  ]
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
    <div class="dash-hero">
      <div class="dash-hero-left">
        <div class="dash-title-row">
          <h1 class="page-title">{{ welcomeTitle }}，{{ sessionStore.user?.username }}</h1>
          <Tag round color="blue">OAuth2</Tag>
        </div>
        <p class="page-subtitle">{{ welcomeSubtitle }}</p>
        <div v-if="!manageAll" class="dash-user-meta">
          <span class="dash-user-tag">UID {{ sessionStore.user?.id }}</span>
          <span class="dash-user-tag">{{ sessionStore.user?.email }}</span>
        </div>
      </div>
    </div>

    <Alert v-if="loadError" type="error" :show-icon="false">
      <div class="page-alert">
        <span>{{ loadError }}</span>
        <Button size="small" type="text" @click="loadOverview">重试</Button>
      </div>
    </Alert>

    <Grid :cols="{xs:1,sm:3,lg:manageAll ? 4 : 3}" :col-gap="14" :row-gap="14">
      <GridItem v-for="stat in statsConfig" :key="stat.label">
        <StatPanel :label="stat.label" :value="stat.value" :detail="stat.detail" />
      </GridItem>
    </Grid>

    <Card size="small">
      <template #title>
        <span style="font-size:14px;">快捷入口</span>
      </template>
      <div class="dash-links">
        <button
          v-for="link in quickLinks"
          :key="link.name"
          class="dash-link-btn"
          @click="router.push({ name: link.name })"
        >
          <component :is="link.icon" :size="20" />
          <div class="dash-link-text">
            <span class="dash-link-label">{{ link.label }}</span>
            <span class="dash-link-desc">{{ link.desc }}</span>
          </div>
        </button>
      </div>
    </Card>
  </section>
</template>

<style scoped>
.dash-hero {
  margin-bottom: 4px;
}

.dash-hero-left {
  min-width: 0;
}

.dash-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.dash-user-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
}

.dash-user-tag {
  font-size: 12px;
  color: var(--color-charcoal);
}

.dash-user-tag + .dash-user-tag::before {
  content: '·';
  margin-right: 12px;
}

.page-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.dash-links {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 8px;
}

.dash-link-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid var(--color-hairline);
  border-radius: var(--radius-md, 8px);
  background: rgba(0, 0, 0, 0.02);
  color: var(--color-body);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
  text-align: left;
}

.dash-link-btn:hover {
  border-color: var(--color-accent-blue);
  background: rgba(59, 158, 255, 0.06);
  color: var(--color-ink);
}

.dash-link-text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.dash-link-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-ink);
}

.dash-link-desc {
  font-size: 11px;
  color: var(--color-charcoal);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 820px) {
  .page-alert {
    align-items: flex-start;
    flex-direction: column;
  }

  .dash-links {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 480px) {
  .dash-links {
    grid-template-columns: 1fr;
  }
}
</style>
