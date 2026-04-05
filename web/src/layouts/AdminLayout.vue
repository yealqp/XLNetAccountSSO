<script setup lang="ts">
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NLayoutSider,
  NMenu,
  NSelect,
  NSpace,
  NText,
  useMessage,
} from 'naive-ui'
import type { MenuOption, SelectOption } from 'naive-ui'
import { computed, shallowRef, watch } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'

import { useThemeMode } from '@/composables/useThemeMode'
import { useViewport } from '@/composables/useViewport'
import { useSessionStore } from '@/stores/session'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const sessionStore = useSessionStore()
const mobileMenuOpen = shallowRef(false)
const { isMobile, width } = useViewport()
const { preference, effectiveMode, osTheme } = useThemeMode()

const routeCopy: Record<string, { title: string; subtitle: string }> = {
  dashboard: { title: '系统总览', subtitle: '查看运行状态。' },
  clients: { title: '客户端管理', subtitle: '维护应用与 scope。' },
  users: { title: '用户管理', subtitle: '管理账号与角色。' },
  sessions: { title: '会话管理', subtitle: '管理登录会话。' },
  tokens: { title: '令牌控制', subtitle: '查看与吊销令牌。' },
}

const isAdmin = computed(() => sessionStore.user?.role === 'admin')

const menuOptions = computed<MenuOption[]>(() => {
  const items: MenuOption[] = []

  if (isAdmin.value) {
    items.push(
      { label: '总览', key: 'dashboard' },
      { label: '客户端', key: 'clients' },
      { label: '用户', key: 'users' },
    )
  }

  items.push(
    { label: '会话', key: 'sessions' },
    { label: '令牌', key: 'tokens' },
  )

  return items
})

const selectedKey = computed(() => String(route.name ?? 'dashboard'))
const currentPage = computed(() => routeCopy[selectedKey.value] ?? routeCopy.dashboard)
const mobileDrawerWidth = computed(() => Math.min(280, Math.max(220, width.value - 24)))

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

watch(isMobile, (nextIsMobile) => {
  if (!nextIsMobile) {
    mobileMenuOpen.value = false
  }
})

async function handleLogout() {
  await sessionStore.signOut()
  message.success('已退出登录')
  await router.push({ name: 'login' })
}

async function handleNavigate(key: string) {
  mobileMenuOpen.value = false
  await router.push({ name: key })
}
</script>

<template>
  <NLayout class="page-shell admin-layout" :has-sider="!isMobile">
    <NLayoutSider v-if="!isMobile" bordered :width="220">
      <NMenu :value="selectedKey" :options="menuOptions" @update:value="(key) => handleNavigate(String(key))" />
    </NLayoutSider>

    <NLayout>
      <NLayoutHeader bordered class="admin-header">
        <div>
          <NText depth="3">XLNetAccount</NText>
          <div class="admin-title">{{ currentPage.title }}</div>
          <div class="admin-subtitle">{{ currentPage.subtitle }}</div>
        </div>

        <NSpace align="center" :wrap="true">
          <NText depth="3">{{ themeHint }}</NText>
          <NSelect v-model:value="preference" :options="themeOptions" size="small" class="theme-select" />
          <NButton v-if="isMobile" secondary @click="mobileMenuOpen = true">
            菜单
          </NButton>
          <NText depth="3">{{ sessionStore.user?.display_name ?? sessionStore.user?.username }}</NText>
          <NButton tertiary type="error" @click="handleLogout">
            退出
          </NButton>
        </NSpace>
      </NLayoutHeader>

      <NLayoutContent content-style="padding: 16px;">
        <RouterView />
      </NLayoutContent>
    </NLayout>

    <NDrawer v-model:show="mobileMenuOpen" placement="left" :width="mobileDrawerWidth">
      <NDrawerContent title="XLNetAccount" body-content-style="padding: 0;" closable>
        <NMenu :value="selectedKey" :options="menuOptions" @update:value="(key) => handleNavigate(String(key))" />
      </NDrawerContent>
    </NDrawer>
  </NLayout>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
  min-height: 100dvh;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
}

.admin-title {
  margin-top: 4px;
  font-size: 22px;
  font-weight: 600;
}

.admin-subtitle {
  margin-top: 4px;
  opacity: 0.72;
}

.theme-select {
  width: 120px;
}

@media (max-width: 820px) {
  .admin-header {
    align-items: flex-start;
  }
}
</style>
