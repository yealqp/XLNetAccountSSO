import type { AuthTokenResponse, ListResponse, PasskeyLoginStartResponse, PasskeyRecord, PasskeyRegistrationFinishResponse, PasskeyRegistrationStartResponse } from '@/types/api'
import type {
	AuthenticationCredentialJSON,
	RegistrationCredentialJSON,
} from '@/types/webauthn'

import { request } from './http'

export function startPasskeyLogin() {
	return request<PasskeyLoginStartResponse>('/api/auth/passkeys/login/start', {
		method: 'POST',
	})
}

export function finishPasskeyLogin(payload: { session_id: string; credential: AuthenticationCredentialJSON }) {
	return request<AuthTokenResponse>('/api/auth/passkeys/login/finish', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function fetchPasskeys() {
	return request<ListResponse<PasskeyRecord>>('/api/me/passkeys')
}

export function startPasskeyRegistration() {
	return request<PasskeyRegistrationStartResponse>('/api/me/passkeys/register/start', {
		method: 'POST',
	})
}

export function finishPasskeyRegistration(payload: { session_id: string; name?: string; credential: RegistrationCredentialJSON }) {
	return request<PasskeyRegistrationFinishResponse>('/api/me/passkeys/register/finish', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function deletePasskey(passkeyId: string) {
	return request<void>(`/api/me/passkeys/${encodeURIComponent(passkeyId)}/delete`, {
		method: 'POST',
	})
}
