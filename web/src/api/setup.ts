import type { SetupStatusResponse, InitializeAdminPayload } from '@/types/api'

import { request } from './http'

export function fetchSetupStatus() {
	return request<SetupStatusResponse>('/api/setup/status')
}

export function initializeAdmin(payload: InitializeAdminPayload) {
	return request<SetupStatusResponse>('/api/setup/initialize', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}
