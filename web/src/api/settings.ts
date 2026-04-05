import { request } from './http'

export interface PlatformSettingsResponse {
	platform_name: string
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
