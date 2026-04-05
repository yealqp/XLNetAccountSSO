import { createRouter, createWebHistory } from 'vue-router'

import AuthLayout from '@/layouts/AuthLayout.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'
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
          component: () => import('@/views/auth/LoginView.vue'),
        },
        {
          path: 'authorize',
          name: 'authorize',
          component: () => import('@/views/auth/AuthorizeView.vue'),
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
          component: () => import('@/views/admin/DashboardView.vue'),
        },
        {
          path: 'clients',
          name: 'clients',
          meta: { requiresAdmin: true },
          component: () => import('@/views/admin/ClientsView.vue'),
        },
        {
          path: 'users',
          name: 'users',
          meta: { requiresAdmin: true },
          component: () => import('@/views/admin/UsersView.vue'),
        },
        {
          path: 'sessions',
          name: 'sessions',
          component: () => import('@/views/admin/SessionsView.vue'),
        },
        {
          path: 'tokens',
          name: 'tokens',
          component: () => import('@/views/admin/TokensView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const sessionStore = useSessionStore(pinia)

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
