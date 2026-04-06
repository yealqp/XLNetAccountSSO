import type { OAuthClientRecord, OverviewStats, TokenRecord, UserRecord } from '@/types/api'

import { request } from './http'

export function fetchOverview() {
  return request<OverviewStats>('/api/overview')
}

export function fetchManagedOverview() {
  return request<OverviewStats>('/api/manage/overview')
}

export async function fetchClients() {
  const response = await request<{ items: OAuthClientRecord[] }>('/api/clients')
  return response.items
}

export async function fetchManagedClients() {
  const response = await request<{ items: OAuthClientRecord[] }>('/api/manage/clients')
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

export function updateManagedClient(id: string, payload: Record<string, unknown>) {
	return request<OAuthClientRecord>(`/api/manage/clients/${id}`, {
		method: 'PUT',
		body: JSON.stringify(payload),
	})
}

export function deleteClient(id: string) {
	return request<void>(`/api/clients/${id}`, {
		method: 'DELETE',
	})
}

export function deleteManagedClient(id: string) {
	return request<void>(`/api/manage/clients/${id}`, {
		method: 'DELETE',
	})
}

export function uploadClientIcon(file: File) {
	const formData = new FormData()
	formData.append('file', file)

	return request<{ icon_url: string }>('/api/client-icons/upload', {
		method: 'POST',
		body: formData,
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

export function updateUser(id: number, payload: Record<string, unknown>) {
  return request<UserRecord>(`/api/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function deleteUser(id: number) {
  return request<void>(`/api/users/${id}`, {
    method: 'DELETE',
  })
}

export async function fetchTokens() {
  const response = await request<{ items: TokenRecord[] }>('/api/tokens')
  return response.items
}

export async function fetchManagedTokens() {
  const response = await request<{ items: TokenRecord[] }>('/api/manage/tokens')
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
