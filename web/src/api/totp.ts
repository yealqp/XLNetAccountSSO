import type { AuthTokenResponse, BooleanResponse } from '@/types/api'

import { request } from './http'

export function fetchTOTPStatus() {
  return request<BooleanResponse>('/api/me/totp/status')
}

export function startTOTPSetup() {
  return request<{ secret: string; uri: string; qr_data_uri: string; setup_session_id: string }>('/api/me/totp/setup/start', {
    method: 'POST',
  })
}

export function verifyTOTPSetup(payload: { setup_session_id: string; code: string }) {
  return request<BooleanResponse>('/api/me/totp/setup/verify', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function disableTOTP() {
  return request<BooleanResponse>('/api/me/totp/disable', {
    method: 'POST',
  })
}

export function verifyTOTPLogin(payload: { totp_session_id: string; code: string }) {
  return request<AuthTokenResponse>('/api/auth/totp/login/verify', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
