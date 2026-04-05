import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

import { fetchSession, login, logout } from '@/api/auth'
import type { UserSummary } from '@/types/api'

export const useSessionStore = defineStore('session', () => {
  const user = shallowRef<UserSummary | null>(null)
  const ready = shallowRef(false)
  const authenticated = computed(() => user.value !== null)

  async function syncSession() {
    const session = await fetchSession()
    user.value = session.authenticated ? session.user ?? null : null
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
    user.value = session.user ?? null
    ready.value = true
    return user.value
  }

  async function signOut() {
    await logout()
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
  }
})
