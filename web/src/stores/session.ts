import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

import { fetchSession, login, logout } from '@/api/auth'
import { verifyTOTPLogin } from '@/api/totp'
import { finishPasskeyLogin } from '@/api/passkeys'
import type { AuthTokenResponse, LoginResponse, TOTPRequiredResponse, UserSummary } from '@/types/api'
import type { AuthenticationCredentialJSON } from '@/types/webauthn'
import { clearAuthToken, setAuthToken } from '@/utils/authToken'

export type SignInResult =
  | { type: 'full' }
  | { type: 'totp'; sessionId: string; user: UserSummary }

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

  async function signIn(payload: { username: string; password: string; captcha_token?: string }): Promise<SignInResult> {
    const response = await login(payload)
    if (isTOTPRequired(response)) {
      return { type: 'totp', sessionId: response.totp_session_id, user: response.user }
    }
    applyAuthSession(response)
    return { type: 'full' }
  }

  async function signInWithTOTP(payload: { totp_session_id: string; code: string }) {
    const session = await verifyTOTPLogin(payload)
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

  function isTOTPRequired(response: LoginResponse): response is TOTPRequiredResponse {
    return (response as TOTPRequiredResponse).requires_totp === true
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
    signInWithTOTP,
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
