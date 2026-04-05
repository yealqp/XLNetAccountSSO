import type { SessionResponse } from '@/types/api'

import { request } from './http'

export function fetchSession() {
  return request<SessionResponse>('/api/auth/session')
}

export function login(payload: { username: string; password: string }) {
  return request<SessionResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function logout() {
  return request<void>('/api/auth/logout', { method: 'POST' })
}
