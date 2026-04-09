import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

import { fetchSession, login, logout } from '@/api/auth'
import { finishPasskeyLogin } from '@/api/passkeys'
import type { AuthTokenResponse, UserSummary } from '@/types/api'
import type { AuthenticationCredentialJSON } from '@/types/webauthn'
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
		return applyAuthSession(session)
	}

	async function signInWithPasskey(payload: { session_id: string; credential: AuthenticationCredentialJSON }) {
		const session = await finishPasskeyLogin(payload)
		return applyAuthSession(session)
	}

	function applyAuthSession(session: AuthTokenResponse) {
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
		signInWithPasskey,
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
