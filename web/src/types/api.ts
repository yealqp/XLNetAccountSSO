export interface UserSummary {
  id: string
  username: string
  display_name: string
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
  created_by: string
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
    id: string
    username: string
    display_name: string
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
