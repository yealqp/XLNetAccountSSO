import type { AuthorizationDecisionPayload, AuthorizationPreview } from '@/types/api'

import { request } from './http'

export function previewAuthorization(params: URLSearchParams) {
  return request<AuthorizationPreview>(`/api/oauth/requests/preview?${params.toString()}`)
}

export function decideAuthorization(payload: AuthorizationDecisionPayload) {
  return request<{ redirect_to: string }>('/api/oauth/requests/decision', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
