import type { ListResponse, OAuthBindingRecord } from '@/types/api'
import { getServerBaseUrl } from '@/config/endpoints'
import { getAuthToken } from '@/utils/authToken'

import { request } from './http'

export function fetchOAuthBindings() {
  return request<ListResponse<OAuthBindingRecord>>('/api/me/oauth/accounts')
}

export function unlinkOAuthBinding(id: string) {
  return request<void>(`/api/me/oauth/${id}/unlink`, { method: 'POST' })
}

export function getOAuthBindUrl(provider: string) {
  return getServerBaseUrl() + `/api/me/oauth/${provider}/link?token=${encodeURIComponent(getAuthToken())}`
}
