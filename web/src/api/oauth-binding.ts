import { request } from './http'
import { getServerBaseUrl } from '@/config/endpoints'

export interface OAuthBindingRecord {
  id: string
  user_id: number
  provider: string
  email: string
  name: string
  created_at: string
}

export function fetchOAuthBindings() {
  return request<{ items: OAuthBindingRecord[] }>('/api/me/oauth/accounts')
}

export function unlinkOAuthBinding(id: string) {
  return request<void>(`/api/me/oauth/${id}/unlink`, { method: 'POST' })
}

export function getOAuthBindUrl(provider: string) {
  const token = typeof window !== 'undefined' ? window.localStorage.getItem('xlnetaccount-access-token') : ''
  return getServerBaseUrl() + `/api/me/oauth/${provider}/link?token=${encodeURIComponent(token || '')}`
}
