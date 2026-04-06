import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

import { fetchSetupStatus, initializeAdmin } from '@/api/setup'

export const useSetupStore = defineStore('setup', () => {
	const initialized = shallowRef(true)
	const ready = shallowRef(false)

	const needsSetup = computed(() => !initialized.value)

	async function syncStatus() {
		const response = await fetchSetupStatus()
		initialized.value = response.initialized
		ready.value = true
		return initialized.value
	}

	async function ensureStatus() {
		if (!ready.value) {
			await syncStatus()
		}
		return initialized.value
	}

	async function completeSetup(payload: {
		username: string
		password: string
		email: string
	}) {
		await initializeAdmin(payload)
		initialized.value = true
		ready.value = true
	}

	return {
		initialized,
		ready,
		needsSetup,
		syncStatus,
		ensureStatus,
		completeSetup,
	}
})
