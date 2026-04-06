const authTokenKey = 'xlnetaccount-access-token'

export function getAuthToken() {
	if (typeof window === 'undefined') {
		return ''
	}
	return window.localStorage.getItem(authTokenKey) ?? ''
}

export function setAuthToken(token: string) {
	if (typeof window === 'undefined') {
		return
	}
	window.localStorage.setItem(authTokenKey, token)
}

export function clearAuthToken() {
	if (typeof window === 'undefined') {
		return
	}
	window.localStorage.removeItem(authTokenKey)
}
