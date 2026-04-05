import type { OAuthClientRecord, OverviewStats, SessionRecord, TokenRecord, UserRecord } from '@/types/api'

import { request } from './http'

export function fetchOverview() {
  return request<OverviewStats>('/api/overview')
}

export async function fetchClients() {
  const response = await request<{ items: OAuthClientRecord[] }>('/api/clients')
  return response.items
}

export function createClient(payload: Record<string, unknown>) {
  return request<OAuthClientRecord>('/api/clients', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateClient(id: string, payload: Record<string, unknown>) {
  return request<OAuthClientRecord>(`/api/clients/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export async function fetchUsers() {
  const response = await request<{ items: UserRecord[] }>('/api/users')
  return response.items
}

export function createUser(payload: Record<string, unknown>) {
  return request<UserRecord>('/api/users', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateUser(id: string, payload: Record<string, unknown>) {
  return request<UserRecord>(`/api/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export async function fetchSessions() {
  const response = await request<{ items: SessionRecord[] }>('/api/sessions')
  return response.items
}

export function revokeSession(id: string) {
  return request<void>(`/api/sessions/${id}`, { method: 'DELETE' })
}

export async function fetchTokens() {
  const response = await request<{ items: TokenRecord[] }>('/api/tokens')
  return response.items
}

export function revokeAccessToken(id: string) {
  return request<void>(`/api/tokens/access/${id}`, { method: 'DELETE' })
}

export function revokeRefreshToken(id: string) {
  return request<void>(`/api/tokens/refresh/${id}`, { method: 'DELETE' })
}

export function revokeClientTokens(clientId: string) {
  return request<void>(`/api/tokens/client/${encodeURIComponent(clientId)}`, { method: 'DELETE' })
}
