export interface UserSummary {
  id: number
  username: string
  email: string
  role: string
  status: string
  created_at: string
  updated_at: string
}

export interface SessionResponse {
  authenticated: boolean
  user?: UserSummary
}

export interface AuthTokenResponse extends SessionResponse {
  access_token: string
  token_type: 'Bearer'
  expires_in: number
}

export interface TOTPRequiredResponse {
  requires_totp: true
  totp_session_id: string
  user: UserSummary
}

export type LoginResponse = AuthTokenResponse | TOTPRequiredResponse

export interface PasskeyRecord {
	id: string
	name: string
	created_at: string
	updated_at: string
	last_used_at: string | null
}

export interface OverviewStats {
  users: number
  clients: number
  active_sessions: number
  access_tokens: number
}

export interface OAuthClientRecord {
  id: string
  name: string
  description: string
  icon_url: string
  client_id: string
  client_type: 'public' | 'confidential'
  redirect_uris: string[]
  scopes: string[]
  trusted: boolean
  created_by: number
  owner_username?: string
  created_at: string
  updated_at: string
  client_secret?: string
}

export interface UserRecord extends UserSummary {}

export interface TokenRecord {
  id: string
  token_kind: 'access' | 'refresh'
  client_id: string
  client_name: string
  user_id?: number
  owner_username?: string
  scope: string
  status: 'active' | 'revoked' | 'expired'
  expires_at: string
  revoked_at?: string | null
  created_at: string
  updated_at: string
  related_access_token_id?: string
}

export interface AuthorizationPreview {
  client: {
    id: string
    name: string
    description: string
    icon_url: string
    client_id: string
    redirect_uri: string
    client_type: string
    trusted: boolean
    requested_scopes: string[]
    requested_scope_details: Array<{
      key: string
      label: string
      description: string
    }>
  }
  user: {
    id: number
    username: string
  }
  request: {
    response_type: string
    client_id: string
    redirect_uri: string
    scope: string
    state: string
    nonce: string
    code_challenge: string
    code_challenge_method: string
  }
}

export interface AuthorizationDecisionPayload {
  response_type: string
  client_id: string
  redirect_uri: string
  scope: string
  state: string
  nonce: string
  code_challenge: string
  code_challenge_method: string
  approved: boolean
}
