import type { AuthTokenResponse, PasskeyRecord } from '@/types/api'
import type {
	AuthenticationCredentialJSON,
	PublicKeyCredentialCreationOptionsEnvelopeJSON,
	PublicKeyCredentialRequestOptionsEnvelopeJSON,
	RegistrationCredentialJSON,
} from '@/types/webauthn'

import { request } from './http'

interface PasskeyListResponse {
	items: PasskeyRecord[]
}

interface PasskeyLoginStartResponse {
	session_id: string
	options: PublicKeyCredentialRequestOptionsEnvelopeJSON
}

interface PasskeyRegistrationStartResponse {
	session_id: string
	options: PublicKeyCredentialCreationOptionsEnvelopeJSON
}

interface PasskeyRegistrationFinishResponse {
	credential: PasskeyRecord
}

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
	return request<PasskeyListResponse>('/api/me/passkeys')
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
