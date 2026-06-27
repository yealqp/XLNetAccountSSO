import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

import { resolveServerUrl } from '@/config/endpoints'
import { fetchPlatformSettings, fetchPublicSettings, updatePlatformSettings } from '@/api/settings'

const fallbackPlatformName = 'XLNetAccount'

export const usePlatformStore = defineStore('platform', () => {
	const platformName = shallowRef(fallbackPlatformName)
	const allowRegistration = shallowRef(false)
	const capApiEndpoint = shallowRef('')
	const capSiteKey = shallowRef('')
	const webIconURL = shallowRef('')
	const ready = shallowRef(false)

	const displayName = computed(() => platformName.value || fallbackPlatformName)
	const capWidgetEndpoint = computed(() => {
		const endpoint = capApiEndpoint.value.trim().replace(/\/+$/, '')
		const siteKey = capSiteKey.value.trim().replace(/^\/+|\/+$/g, '')
		if (!endpoint || !siteKey) {
			return ''
		}
		return `${endpoint}/${siteKey}/`
	})

	async function ensureLoaded() {
		if (ready.value) {
			return displayName.value
		}
		await loadPublicSettings()
		return displayName.value
	}

	async function loadPublicSettings() {
		const response = await fetchPublicSettings()
		applyPublicSettings(response)
		ready.value = true
		return displayName.value
	}

	async function loadAdminSettings() {
		const response = await fetchPlatformSettings()
		applyPublicSettings(response)
		ready.value = true
		return response
	}

	async function savePlatformSettings(payload: {
		platform_name: string
		allow_registration: boolean
		smtp_host: string
		smtp_user: string
		smtp_password: string
		smtp_port: string
		smtp_tls: boolean
		cap_api_endpoint: string
		cap_site_key: string
		cap_secret_key: string
		web_icon_url: string
	}) {
		const response = await updatePlatformSettings(payload)
		applyPublicSettings(response)
		ready.value = true
		return response
	}

	function applyPublicSettings(value: { platform_name: string; allow_registration?: boolean; cap_api_endpoint?: string; cap_site_key?: string; web_icon_url?: string }) {
		setPlatformName(value.platform_name)
		allowRegistration.value = Boolean(value.allow_registration)
		capApiEndpoint.value = value.cap_api_endpoint?.trim() || ''
		capSiteKey.value = value.cap_site_key?.trim() || ''
		setWebIcon(value.web_icon_url)
	}

	function setPlatformName(value: string) {
		platformName.value = value?.trim() || fallbackPlatformName
		if (typeof document !== 'undefined') {
			document.title = platformName.value
		}
	}

	function setWebIcon(value?: string) {
		webIconURL.value = value?.trim() || ''
		if (typeof document === 'undefined') {
			return
		}
		let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
		if (!link) {
			link = document.createElement('link')
			link.rel = 'icon'
			document.head.appendChild(link)
		}
		if (webIconURL.value) {
			link.href = resolveServerUrl(webIconURL.value)
		} else {
			link.removeAttribute('href')
		}
	}

	return {
		platformName,
		allowRegistration,
		capApiEndpoint,
		capSiteKey,
		capWidgetEndpoint,
		webIconURL,
		displayName,
		ready,
		ensureLoaded,
		loadPublicSettings,
		loadAdminSettings,
		savePlatformSettings,
		applyPublicSettings,
	}
})
