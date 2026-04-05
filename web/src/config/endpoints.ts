export function getServerBaseUrl() {
  const configured = import.meta.env.VITE_API_BASE_URL?.trim()
  if (configured) {
    return configured.replace(/\/$/, '')
  }

  if (typeof window !== 'undefined') {
    return `${window.location.protocol}//${window.location.hostname}:8080`
  }

  return 'http://localhost:8080'
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
