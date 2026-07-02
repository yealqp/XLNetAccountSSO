import type { PlatformSettingsResponse, SentResponse } from '@/types/api'

import { request } from './http'

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
	return request<SentResponse>('/api/settings/platform/test-email', {
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
