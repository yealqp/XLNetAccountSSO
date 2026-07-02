import type { LoginResponse, SessionResponse } from '@/types/api'

import { request } from './http'

export function fetchSession() {
  return request<SessionResponse>('/api/auth/session')
}

export function login(payload: { username: string; password: string; captcha_token?: string }) {
	return request<LoginResponse>('/api/auth/login', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function register(payload: { username: string; email: string; password: string; code: string }) {
	return request<{ registered: boolean }>('/api/auth/register', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function updateProfile(payload: { username: string; password: string; code: string }) {
	return request<SessionResponse>('/api/me/profile', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function sendProfilePasswordCode() {
	return request<{ sent: boolean }>('/api/me/password/code/send', {
		method: 'POST',
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
