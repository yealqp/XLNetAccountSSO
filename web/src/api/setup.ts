import { request } from './http'

export interface SetupStatusResponse {
	initialized: boolean
}

export interface InitializeAdminPayload {
	username: string
	password: string
	email: string
}

export function fetchSetupStatus() {
	return request<SetupStatusResponse>('/api/setup/status')
}

export function initializeAdmin(payload: InitializeAdminPayload) {
	return request<SetupStatusResponse>('/api/setup/initialize', {
		method: 'POST',
		body: JSON.stringify(payload),
	})
}
