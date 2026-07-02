<script setup lang="ts">
import {
  Button,
  Drawer,
  Layout,
  LayoutContent,
  LayoutHeader,
  LayoutSider,
  Menu,
  MenuItem,
  MenuItemGroup,
  Message,
  Space,
  Spin,
  Tag,
  TypographyText as Text,
} from '@arco-design/web-vue'
import { computed, shallowRef, watch } from 'vue'

type MenuOption = {
  label: string
  key: string
  type?: 'group'
  icon?: object
  children?: MenuOption[]
}
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
} from '@lucide/vue'

import { useViewport } from '@/composables/useViewport'
import { usePlatformStore } from '@/stores/platform'
import { useSessionStore } from '@/stores/session'

const router = useRouter()
const sessionStore = useSessionStore()
const platformStore = usePlatformStore()
const mobileMenuOpen = shallowRef(false)
const { isMobile, width } = useViewport()
void platformStore.ensureLoaded().catch(() => {})

const isAdmin = computed(() => sessionStore.user?.role === 'admin')


const menuOptions = computed<MenuOption[]>(() => {
  const items: MenuOption[] = [
    {
      label: '导航',
      key: 'group-navigation',
      type: 'group',
      children: [
        { label: '概览', key: 'overview', icon: LayoutDashboard },
        { label: '应用', key: 'applications', icon: AppWindow },
        { label: '令牌', key: 'tokens', icon: WalletCards },
        { label: '普通设置', key: 'settings', icon: Settings },
        { label: '连接信息', key: 'connection-info', icon: Link2 },
      ],
    },
  ]

  if (isAdmin.value) {
    items.push({
      label: '管理',
      key: 'group-management',
      type: 'group',
      children: [
        { label: '管理概览', key: 'manage-overview', icon: LayoutDashboard },
        { label: '用户管理', key: 'manage-users', icon: UserCog },
        { label: '应用管理', key: 'manage-applications', icon: Boxes },
        { label: '令牌管理', key: 'manage-tokens', icon: ShieldUser },
        { label: '系统设置', key: 'system-settings', icon: Shield },
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
  Message.success('已退出登录')
  await router.push({ name: 'login' })
}

async function handleNavigate(key: string) {
  mobileMenuOpen.value = false
  await router.push({ name: key })
}
</script>

<template>
  <Layout class="page-shell admin-layout main-layout">
    <LayoutSider v-if="!isMobile" bordered :width="220" class="admin-sider">
      <div class="sider-brand">
        <Text type="secondary">{{ platformStore.displayName }}</Text>
      </div>
      <Menu :selected-keys="[selectedKey]" @menu-item-click="(key: any) => handleNavigate(String(key))">
        <template v-for="item in menuOptions" :key="item.key">
          <MenuItemGroup v-if="item.type === 'group'" :title="item.label">
            <MenuItem v-for="child in item.children" :key="child.key">
              <template #icon>
                <component :is="child.icon" :size="16" />
              </template>
              {{ child.label }}
            </MenuItem>
          </MenuItemGroup>
          <MenuItem v-else :key="item.key">
            <template #icon>
              <component :is="item.icon" :size="16" />
            </template>
            {{ item.label }}
          </MenuItem>
        </template>
      </Menu>
    </LayoutSider>

    <Layout class="admin-content-layout content-layout">
      <LayoutHeader bordered class="admin-header">
        <div class="header-spacer"></div>
        <Space align="center" wrap class="header-actions">
          <Button v-if="isMobile" type="secondary" @click="mobileMenuOpen = true">
            菜单
          </Button>
          <Tag size="small" round color="blue">{{ sessionStore.user?.role === 'admin' ? '管理员' : '普通用户' }}</Tag>
          <Text type="secondary">{{ sessionStore.user?.username }}</Text>
          <Button type="text" status="danger" @click="handleLogout">
            退出
          </Button>
        </Space>
      </LayoutHeader>

      <LayoutContent class="content-body">
        <RouterView v-slot="{ Component, route: currentRoute }">
          <transition name="fade-slide" mode="out-in">
            <div v-if="Component" :key="currentRoute.path" class="route-container page-container">
              <Suspense>
                <template #default>
                  <component :is="Component" />
                </template>
                <template #fallback>
                  <div class="route-loading">
                    <Spin />
                  </div>
                </template>
              </Suspense>
            </div>
          </transition>
        </RouterView>
      </LayoutContent>
    </Layout>

    <Drawer v-model:visible="mobileMenuOpen" placement="left" :width="mobileDrawerWidth" :title="platformStore.displayName" closable>
        <Menu :selected-keys="[selectedKey]" @menu-item-click="(key: any) => handleNavigate(String(key))">
          <template v-for="item in menuOptions" :key="item.key">
            <MenuItemGroup v-if="item.type === 'group'" :title="item.label">
              <MenuItem v-for="child in item.children" :key="child.key">
                <template #icon>
                  <component :is="child.icon" :size="16" />
                </template>
                {{ child.label }}
              </MenuItem>
            </MenuItemGroup>
            <MenuItem v-else :key="item.key">
              <template #icon>
                <component :is="item.icon" :size="16" />
              </template>
              {{ item.label }}
            </MenuItem>
          </template>
        </Menu>
          </Drawer>
  </Layout>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
  min-height: 100dvh;
}

.admin-layout :deep(.arco-layout-content) {
  min-height: 100%;
}

.admin-sider {
  height: 100vh;
  height: 100dvh;
}

.sider-brand {
  display: flex;
  align-items: center;
  height: 48px;
  padding: 0 16px;
  border-bottom: 1px solid var(--color-hairline);
}

.admin-content-layout {
  min-width: 0;
}

.content-layout {
  background: var(--color-canvas);
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  height: 48px;
  padding: 0 16px;
}

.header-spacer {
  flex: 1 1 auto;
}

.header-actions {
  justify-content: flex-end;
}

.content-body {
  height: calc(100vh - 48px);
  height: calc(100dvh - 48px);
  padding: 16px 20px 20px;
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
    min-height: 48px;
    padding: 10px 12px;
  }

  .content-body {
    height: calc(100vh - 48px);
    height: calc(100dvh - 48px);
    padding: 12px;
  }
}
</style>
