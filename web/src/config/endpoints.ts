export function getServerBaseUrl() {
  const configured = import.meta.env.VITE_API_BASE_URL?.trim()
  if (configured) {
    return configured.replace(/\/$/, '')
  }

  if (typeof window !== 'undefined') {
    return window.location.origin.replace(/\/$/, '')
  }

  return 'http://localhost'
}

export function resolveServerUrl(raw?: string) {
  const value = raw?.trim() || ''
  if (!value) {
    return ''
  }

  if (/^https?:\/\//i.test(value) || value.startsWith('blob:') || value.startsWith('data:')) {
    return value
  }

  const baseUrl = getServerBaseUrl()
  return new URL(value.replace(/^\/+/, '/'), `${baseUrl}/`).toString()
}

export function getConnectionInfo() {
  const baseUrl = getServerBaseUrl()

  return {
    baseUrl,
    issuer: baseUrl,
    authorizeUrl: `${baseUrl}/oauth/authorize`,
    tokenUrl: `${baseUrl}/oauth/token`,
    userInfoUrl: `${baseUrl}/oauth/userinfo`,
    introspectUrl: `${baseUrl}/oauth/introspect`,
    revokeUrl: `${baseUrl}/oauth/revoke`,
    openidConfigurationUrl: `${baseUrl}/.well-known/openid-configuration`,
    jwksUrl: `${baseUrl}/.well-known/jwks.json`,
  }
}
