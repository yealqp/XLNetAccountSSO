<script setup lang="ts">
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NIcon,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NLayoutSider,
  NMenu,
  NSpin,
  NSpace,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { computed, h, shallowRef, watch } from 'vue'
import { RouterView, useRouter } from 'vue-router'
import {
  AppWindow,
  Boxes,
  LayoutDashboard,
  Link2,
  Settings,
  Shield,
  ShieldUser,
  UserCog,
  WalletCards,
} from 'lucide-vue-next'

import { useViewport } from '@/composables/useViewport'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'

const router = useRouter()
const message = useMessage()
const sessionStore = useSessionStore()
const platformStore = usePlatformStore()
const mobileMenuOpen = shallowRef(false)
const { isMobile, width } = useViewport()
void platformStore.ensureLoaded().catch(() => {})

const isAdmin = computed(() => sessionStore.user?.role === 'admin')

function renderLucideIcon(icon: typeof LayoutDashboard) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions = computed<MenuOption[]>(() => {
  const items: MenuOption[] = [
    {
      label: '导航',
      key: 'group-navigation',
      type: 'group',
      children: [
        { label: '概览', key: 'overview', icon: renderLucideIcon(LayoutDashboard) },
        { label: '应用', key: 'applications', icon: renderLucideIcon(AppWindow) },
        { label: '令牌', key: 'tokens', icon: renderLucideIcon(WalletCards) },
        { label: '普通设置', key: 'settings', icon: renderLucideIcon(Settings) },
        { label: '连接信息', key: 'connection-info', icon: renderLucideIcon(Link2) },
      ],
    },
  ]

  if (isAdmin.value) {
    items.push({
      label: '管理',
      key: 'group-management',
      type: 'group',
      children: [
        { label: '管理概览', key: 'manage-overview', icon: renderLucideIcon(LayoutDashboard) },
        { label: '用户管理', key: 'manage-users', icon: renderLucideIcon(UserCog) },
        { label: '应用管理', key: 'manage-applications', icon: renderLucideIcon(Boxes) },
        { label: '令牌管理', key: 'manage-tokens', icon: renderLucideIcon(ShieldUser) },
        { label: '系统设置', key: 'system-settings', icon: renderLucideIcon(Shield) },
      ],
    })
  }

  return items
})

const selectedKey = computed(() => {
  const currentRoute = router.currentRoute.value
	return String(currentRoute.name ?? 'overview')
})
const mobileDrawerWidth = computed(() => Math.min(280, Math.max(220, width.value - 24)))

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
  <NLayout class="page-shell admin-layout main-layout" :has-sider="!isMobile" position="absolute">
    <NLayoutSider v-if="!isMobile" bordered :width="220" class="admin-sider">
      <div class="sider-brand">
        <NText depth="3">{{ platformStore.displayName }}</NText>
      </div>
      <NMenu :value="selectedKey" :options="menuOptions" @update:value="(key) => handleNavigate(String(key))" />
    </NLayoutSider>

    <NLayout class="admin-content-layout content-layout">
      <NLayoutHeader bordered class="admin-header">
        <div class="header-spacer"></div>
        <NSpace align="center" :wrap="true" class="header-actions">
          <NButton v-if="isMobile" secondary @click="mobileMenuOpen = true">
            菜单
          </NButton>
          <NTag size="small" round type="info">{{ sessionStore.user?.role ?? 'user' }}</NTag>
          <NText depth="3">{{ sessionStore.user?.username }}</NText>
          <NButton tertiary type="error" @click="handleLogout">
            退出
          </NButton>
        </NSpace>
      </NLayoutHeader>

      <NLayoutContent class="content-body">
        <RouterView v-slot="{ Component, route: currentRoute }">
          <transition name="fade-slide" mode="out-in">
            <div v-if="Component" :key="currentRoute.path" class="route-container page-container">
              <Suspense>
                <template #default>
                  <component :is="Component" />
                </template>
                <template #fallback>
                  <div class="route-loading">
                    <NSpin size="medium" />
                  </div>
                </template>
              </Suspense>
            </div>
          </transition>
        </RouterView>
      </NLayoutContent>
    </NLayout>

    <NDrawer v-model:show="mobileMenuOpen" placement="left" :width="mobileDrawerWidth">
      <NDrawerContent :title="platformStore.displayName" body-content-style="padding: 0;" closable>
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

.admin-layout :deep(.n-layout-scroll-container) {
  min-height: 100%;
}

.admin-sider {
  height: 100vh;
  height: 100dvh;
}

.sider-brand {
  display: flex;
  align-items: center;
  height: 56px;
  padding: 0 20px;
  border-bottom: 1px solid #29292c;
}

.admin-content-layout {
  min-width: 0;
}

.content-layout {
  background: #101014;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  height: 56px;
  padding: 0 20px;
}

.header-spacer {
  flex: 1 1 auto;
}

.header-actions {
  justify-content: flex-end;
}

.content-body {
  height: calc(100vh - 56px);
  height: calc(100dvh - 56px);
  padding: 24px 28px 28px;
  overflow: auto;
}

.route-container {
  width: 100%;
  min-height: 100%;
}

.route-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
}

.fade-slide-enter-active {
  transition: all 0.3s ease-out;
}

.fade-slide-leave-active {
  transition: all 0.25s ease-in;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

@media (max-width: 900px) {
  .admin-header {
    align-items: flex-start;
    height: auto;
    min-height: 56px;
    padding: 12px 16px;
  }

  .content-body {
    height: calc(100vh - 56px);
    height: calc(100dvh - 56px);
    padding: 16px;
  }
}
</style>
