import { request } from './http'

export interface PlatformSettingsResponse {
	platform_name: string
	allow_registration: boolean
	smtp_host?: string
	smtp_user?: string
	smtp_password?: string
	smtp_port?: string
	smtp_tls?: boolean
	cap_api_endpoint?: string
	cap_secret_key?: string
}

export function fetchPublicSettings() {
	return request<PlatformSettingsResponse>('/api/settings/public')
}

export function fetchPlatformSettings() {
	return request<PlatformSettingsResponse>('/api/settings/platform')
}

export function updatePlatformSettings(payload: PlatformSettingsResponse) {
	return request<PlatformSettingsResponse>('/api/settings/platform', {
		method: 'PUT',
		body: JSON.stringify(payload),
	})
}
