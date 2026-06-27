import { request } from './http'

export interface PlatformSettingsResponse {
	platform_name: string
	allow_registration: boolean
	web_icon_url?: string
	smtp_configured?: boolean
	cap_configured?: boolean
	cap_api_endpoint?: string
	cap_site_key?: string
}

export function fetchPublicSettings() {
	return request<PlatformSettingsResponse>('/api/settings/public')
}

export function fetchPlatformSettings() {
	return request<PlatformSettingsResponse>('/api/settings/platform')
}

export function updatePlatformSettings(payload: PlatformSettingsResponse) {
	return request<PlatformSettingsResponse>('/api/settings/platform', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function sendTestEmail(payload: { to: string }) {
	return request<{ sent: boolean }>('/api/settings/platform/test-email', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}

export function uploadWebIcon(file: File) {
	const formData = new FormData()
	formData.append('file', file)

	return request<{ web_icon_url: string }>('/api/settings/platform/icon/upload', {
		method: 'POST',
		body: formData,
	})
}
