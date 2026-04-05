import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

import { fetchPlatformSettings, fetchPublicSettings, updatePlatformSettings } from '@/api/settings'

const fallbackPlatformName = 'XLNetAccount'

export const usePlatformStore = defineStore('platform', () => {
	const platformName = shallowRef(fallbackPlatformName)
	const ready = shallowRef(false)

	const displayName = computed(() => platformName.value || fallbackPlatformName)

	async function ensureLoaded() {
		if (ready.value) {
			return displayName.value
		}
		await loadPublicSettings()
		return displayName.value
	}

	async function loadPublicSettings() {
		const response = await fetchPublicSettings()
		setPlatformName(response.platform_name)
		ready.value = true
		return displayName.value
	}

	async function loadAdminSettings() {
		const response = await fetchPlatformSettings()
		setPlatformName(response.platform_name)
		ready.value = true
		return displayName.value
	}

	async function savePlatformName(nextName: string) {
		const response = await updatePlatformSettings({ platform_name: nextName })
		setPlatformName(response.platform_name)
		ready.value = true
		return displayName.value
	}

	function setPlatformName(value: string) {
		platformName.value = value?.trim() || fallbackPlatformName
		if (typeof document !== 'undefined') {
			document.title = platformName.value
		}
	}

	return {
		platformName,
		displayName,
		ready,
		ensureLoaded,
		loadPublicSettings,
		loadAdminSettings,
		savePlatformName,
	}
})
