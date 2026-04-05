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
      redirect: '/admin',
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
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          meta: { requiresAdmin: true },
          component: () => import('@/components/admin/DashboardPage.vue'),
        },
        {
          path: 'clients',
          name: 'clients',
          meta: { requiresAdmin: true },
          component: () => import('@/components/admin/ClientsPage.vue'),
        },
        {
          path: 'users',
          name: 'users',
          meta: { requiresAdmin: true },
          component: () => import('@/components/admin/UsersPage.vue'),
        },
        {
          path: 'settings',
          name: 'settings',
          meta: { requiresAdmin: true },
          component: () => import('@/components/admin/SettingsPage.vue'),
        },
        {
          path: 'tokens',
          name: 'tokens',
          component: () => import('@/components/admin/TokensPage.vue'),
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

  const initialized = await setupStore.ensureStatus()

  if (!initialized && to.name !== 'setup') {
    return {
      name: 'setup',
      query: { next: to.fullPath },
    }
  }

  if (initialized && to.name === 'setup') {
    if (sessionStore.authenticated || (await sessionStore.ensureSession())) {
      return { name: sessionStore.user?.role === 'admin' ? 'dashboard' : 'tokens' }
    }
    return { name: 'login' }
  }

  if (initialized && (to.name === 'login' || to.name === 'register' || to.name === 'authorize')) {
		await platformStore.ensureLoaded().catch(() => {})
	}

  if (to.name === 'register' && !platformStore.allowRegistration) {
		return { name: 'login' }
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
    return { name: 'tokens' }
  }

  if (to.name === 'login' && !to.query.next) {
    await sessionStore.ensureSession()
    if (sessionStore.authenticated) {
      return { name: sessionStore.user?.role === 'admin' ? 'dashboard' : 'tokens' }
    }
  }

  return true
})

export default router
