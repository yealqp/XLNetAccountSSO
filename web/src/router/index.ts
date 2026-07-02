import { createRouter, createWebHistory } from 'vue-router'

import AuthLayout from '@/layouts/AuthLayout.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'
import { usePlatformStore } from '@/stores/platform'
import { useSetupStore } from '@/stores/setup'
import { useSessionStore } from '@/stores/session'
import { pinia } from '@/stores/pinia'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'landing',
      component: () => import('@/components/landing/LandingPage.vue'),
    },
    {
      path: '/auth',
      component: AuthLayout,
      children: [
        {
          path: 'login',
          name: 'login',
          component: () => import('@/components/auth/LoginPage.vue'),
        },
        {
          path: 'register',
          name: 'register',
          component: () => import('@/components/auth/RegisterPage.vue'),
        },
        {
          path: 'setup',
          name: 'setup',
          component: () => import('@/components/auth/SetupPage.vue'),
        },
        {
          path: 'authorize',
          name: 'authorize',
          component: () => import('@/components/auth/AuthorizePage.vue'),
        },
      ],
    },
    {
      path: '/dashboard',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'overview',
          component: () => import('@/components/admin/DashboardPage.vue'),
        },
        {
          path: 'admin/overview',
          name: 'manage-overview',
          meta: { requiresAdmin: true, manageScope: 'all' },
          component: () => import('@/components/admin/DashboardPage.vue'),
        },
        {
          path: 'applications',
          name: 'applications',
          component: () => import('@/components/admin/ClientsPage.vue'),
        },
        {
          path: 'tokens',
          name: 'tokens',
          component: () => import('@/components/admin/TokensPage.vue'),
        },
        {
          path: 'admin/users',
          name: 'manage-users',
          meta: { requiresAdmin: true },
          component: () => import('@/components/admin/UsersPage.vue'),
        },
        {
          path: 'admin/applications',
          name: 'manage-applications',
          meta: { requiresAdmin: true, manageScope: 'all' },
          component: () => import('@/components/admin/ClientsPage.vue'),
        },
        {
          path: 'admin/tokens',
          name: 'manage-tokens',
          meta: { requiresAdmin: true, manageScope: 'all' },
          component: () => import('@/components/admin/TokensPage.vue'),
        },
        {
          path: 'admin/system-settings',
          name: 'system-settings',
          meta: { requiresAdmin: true },
          component: () => import('@/components/admin/SystemSettingsPage.vue'),
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/components/admin/SettingsPage.vue'),
        },
        {
          path: 'connection-info',
          name: 'connection-info',
          component: () => import('@/components/admin/ConnectionInfoPage.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const sessionStore = useSessionStore(pinia)
  const setupStore = useSetupStore(pinia)
  const platformStore = usePlatformStore(pinia)

  // Landing page is fully public — render instantly without blocking on API calls
  if (to.name === 'landing') {
    setupStore.ensureStatus().catch(() => {})
    return true
  }

  const initialized = await setupStore.ensureStatus()

  if (!initialized && to.name !== 'setup') {
    return {
      name: 'setup',
      query: { next: to.fullPath },
    }
  }

  if (initialized && to.name === 'setup') {
    if (sessionStore.authenticated || (await sessionStore.ensureSession())) {
      return { name: 'overview' }
    }
    return { name: 'login' }
  }

  if (initialized && to.name === 'register') {
    await platformStore.ensureLoaded().catch(() => {})
    if (!platformStore.allowRegistration) {
      return {
        name: 'login',
        query: to.query.next === undefined ? undefined : { next: to.query.next },
      }
    }
  } else if (initialized && (to.name === 'login' || to.name === 'authorize')) {
    platformStore.ensureLoaded().catch(() => {})
  }

  if (to.meta.requiresAuth) {
    await sessionStore.ensureSession()
    if (!sessionStore.authenticated) {
      return {
        name: 'login',
        query: { next: to.fullPath },
      }
    }
  }

  if (to.meta.requiresAdmin && sessionStore.user?.role !== 'admin') {
    return { name: 'overview' }
  }

  if (to.name === 'login' && !to.query.next) {
    await sessionStore.ensureSession()
    if (sessionStore.authenticated) {
      return { name: 'overview' }
    }
  }

  return true
})

export default router
