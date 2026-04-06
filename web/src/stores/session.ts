import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

import { fetchSession, login, logout } from '@/api/auth'
import type { UserSummary } from '@/types/api'
import { clearAuthToken, setAuthToken } from '@/utils/authToken'

export const useSessionStore = defineStore('session', () => {
  const user = shallowRef<UserSummary | null>(null)
  const ready = shallowRef(false)
  const authenticated = computed(() => user.value !== null)

  async function syncSession() {
    const session = await fetchSession()
    user.value = session.authenticated ? session.user ?? null : null
    if (!session.authenticated) {
      clearAuthToken()
    }
    ready.value = true
    return user.value
  }

  async function ensureSession() {
    if (!ready.value) {
      await syncSession()
    }
    return user.value
  }

  async function signIn(payload: { username: string; password: string }) {
    const session = await login(payload)
    setAuthToken(session.access_token)
    user.value = session.user ?? null
    ready.value = true
    return user.value
  }

  async function signOut() {
    try {
      await logout()
    }
    finally {
      clearAuthToken()
    }
    user.value = null
    ready.value = true
  }

  return {
    user,
    ready,
    authenticated,
    syncSession,
    ensureSession,
    signIn,
    signOut,
    setUser: (nextUser: UserSummary | null) => {
      user.value = nextUser
      ready.value = true
    },
    clear: () => {
      clearAuthToken()
      user.value = null
      ready.value = true
    },
  }
})
