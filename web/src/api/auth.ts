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

export function register(payload: { email: string; password: string; display_name: string; code: string }) {
	return request<{ registered: boolean }>('/api/auth/register', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function sendRegisterCode(payload: { email: string; captcha_token: string }) {
	return request<{ sent: boolean }>('/api/auth/register/code/send', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function logout() {
  return request<void>('/api/auth/logout', { method: 'POST' })
}
