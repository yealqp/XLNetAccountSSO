import './style.css'

const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL || 'http://localhost:8080'
const CLIENT_ID = import.meta.env.VITE_CLIENT_ID || 'demo-web-client'
const STORAGE_KEY = 'demo-sso-tokens'
const OAUTH_KEY = 'demo-sso-oauth'

interface TokenResponse {
  access_token: string
  id_token?: string
  refresh_token: string
  token_type: string
  expires_in: number
  scope: string
}

interface AppState {
  tokens: TokenResponse | null
  profile: unknown | null
  status: string
}

const state: AppState = {
  tokens: readTokens(),
  profile: null,
  status: '等待发起授权流程。',
}

const root = document.querySelector<HTMLDivElement>('#app')

if (!root) {
  throw new Error('app root not found')
}

void bootstrap()

async function bootstrap() {
  await maybeHandleCallback()

  if (state.tokens?.access_token) {
    await loadUserInfo()
  }

  render()
}

function render() {
  root!.innerHTML = `
    <main class="shell">
      <section class="hero">
        <span class="muted">Public SPA client</span>
        <h1>OAuth2 PKCE Demo Client</h1>
        <p>
          这是一个最小业务系统示例，用来验证自建 SSO 授权中心的
          Authorization Code + PKCE 闭环。
        </p>
      </section>

      <section class="button-row">
        <button id="login-btn">发起授权登录</button>
        <button id="userinfo-btn" class="secondary">拉取 userinfo</button>
        <button id="refresh-btn" class="secondary">刷新令牌</button>
        <button id="revoke-btn" class="secondary">撤销 refresh token</button>
      </section>

      <section class="grid">
        <article class="panel">
          <h2>当前状态</h2>
          <div class="status-box">${escapeHtml(state.status)}</div>
        </article>

        <article class="panel">
          <h2>Client 配置</h2>
          <div class="status-box mono">
            client_id: ${escapeHtml(CLIENT_ID)}<br />
            redirect_uri: ${escapeHtml(getRedirectURI())}<br />
            authorize: ${escapeHtml(`${AUTH_BASE_URL}/oauth/authorize`)}<br />
            token: ${escapeHtml(`${AUTH_BASE_URL}/oauth/token`)}
          </div>
        </article>

        <article class="panel">
          <h2>令牌</h2>
          <div class="token-box mono">${escapeHtml(JSON.stringify(state.tokens, null, 2) || 'null')}</div>
        </article>

        <article class="panel">
          <h2>userinfo</h2>
          <div class="token-box mono">${escapeHtml(JSON.stringify(state.profile, null, 2) || 'null')}</div>
        </article>
      </section>
    </main>
  `

  attachEvents()
}

function attachEvents() {
  document.querySelector<HTMLButtonElement>('#login-btn')?.addEventListener('click', () => {
    void startAuthorization()
  })
  document.querySelector<HTMLButtonElement>('#userinfo-btn')?.addEventListener('click', () => {
    void loadUserInfo()
  })
  document.querySelector<HTMLButtonElement>('#refresh-btn')?.addEventListener('click', () => {
    void refreshToken()
  })
  document.querySelector<HTMLButtonElement>('#revoke-btn')?.addEventListener('click', () => {
    void revokeToken()
  })
}

async function startAuthorization() {
  const codeVerifier = randomBase64URL(48)
  const stateValue = randomBase64URL(18)
  const nonceValue = randomBase64URL(18)
  const codeChallenge = await sha256Base64URL(codeVerifier)

  sessionStorage.setItem(
    OAUTH_KEY,
    JSON.stringify({ codeVerifier, state: stateValue, nonce: nonceValue }),
  )

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: CLIENT_ID,
    redirect_uri: getRedirectURI(),
    scope: 'openid profile email offline_access',
    state: stateValue,
    nonce: nonceValue,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256',
  })

  window.location.href = `${AUTH_BASE_URL}/oauth/authorize?${params.toString()}`
}

async function maybeHandleCallback() {
  const url = new URL(window.location.href)
  if (url.pathname !== '/callback') {
    return
  }

  const code = url.searchParams.get('code')
  const returnedState = url.searchParams.get('state')
  const oauthSession = readOAuthSession()

  if (!code || !returnedState || !oauthSession || oauthSession.state !== returnedState) {
    state.status = '授权回调参数不完整或 state 校验失败。'
    render()
    return
  }

  const response = await fetch(`${AUTH_BASE_URL}/oauth/token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: CLIENT_ID,
      code,
      redirect_uri: getRedirectURI(),
      code_verifier: oauthSession.codeVerifier,
    }),
  })

  if (!response.ok) {
    state.status = `换取 token 失败：${await response.text()}`
    render()
    return
  }

  state.tokens = await response.json() as TokenResponse
  persistTokens(state.tokens)
  sessionStorage.removeItem(OAUTH_KEY)
  state.status = '已成功换取 access token、id token 和 refresh token。'
  window.history.replaceState({}, '', '/')
}

async function loadUserInfo() {
  if (!state.tokens?.access_token) {
    state.status = '请先完成授权登录。'
    render()
    return
  }

  const response = await fetch(`${AUTH_BASE_URL}/oauth/userinfo`, {
    headers: {
      Authorization: `Bearer ${state.tokens.access_token}`,
    },
  })

  if (!response.ok) {
    state.status = `userinfo 请求失败：${await response.text()}`
    render()
    return
  }

  state.profile = await response.json()
  state.status = 'userinfo 已刷新。'
  render()
}

async function refreshToken() {
  if (!state.tokens?.refresh_token) {
    state.status = '当前没有 refresh token。'
    render()
    return
  }

  const response = await fetch(`${AUTH_BASE_URL}/oauth/token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      grant_type: 'refresh_token',
      client_id: CLIENT_ID,
      refresh_token: state.tokens.refresh_token,
    }),
  })

  if (!response.ok) {
    state.status = `刷新 token 失败：${await response.text()}`
    render()
    return
  }

  state.tokens = await response.json() as TokenResponse
  persistTokens(state.tokens)
  state.status = 'refresh token 轮换成功。'
  render()
}

async function revokeToken() {
  if (!state.tokens?.refresh_token) {
    state.status = '当前没有可撤销的 refresh token。'
    render()
    return
  }

  const response = await fetch(`${AUTH_BASE_URL}/oauth/revoke`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      client_id: CLIENT_ID,
      token: state.tokens.refresh_token,
    }),
  })

  if (!response.ok) {
    state.status = `撤销 token 失败：${await response.text()}`
    render()
    return
  }

  state.status = 'refresh token 已撤销，本地令牌也已清空。'
  state.tokens = null
  state.profile = null
  persistTokens(null)
  render()
}

function persistTokens(tokens: TokenResponse | null) {
  if (!tokens) {
    localStorage.removeItem(STORAGE_KEY)
    return
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(tokens))
}

function readTokens() {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as TokenResponse
  }
  catch {
    return null
  }
}

function readOAuthSession() {
  const raw = sessionStorage.getItem(OAUTH_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as { codeVerifier: string; state: string; nonce: string }
  }
  catch {
    return null
  }
}

function getRedirectURI() {
  return `${window.location.origin}/callback`
}

function randomBase64URL(length: number) {
  const bytes = new Uint8Array(length)
  crypto.getRandomValues(bytes)
  return toBase64URL(bytes)
}

async function sha256Base64URL(value: string) {
  const encoded = new TextEncoder().encode(value)
  const digest = await crypto.subtle.digest('SHA-256', encoded)
  return toBase64URL(new Uint8Array(digest))
}

function toBase64URL(bytes: Uint8Array) {
  return btoa(String.fromCharCode(...bytes))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/g, '')
}

function escapeHtml(value: string) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}
