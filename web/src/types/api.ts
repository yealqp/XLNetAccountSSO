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

import type {
  PublicKeyCredentialCreationOptionsEnvelopeJSON,
  PublicKeyCredentialRequestOptionsEnvelopeJSON,
} from '@/types/webauthn'

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

export interface ListResponse<T> { items: T[] }

export interface BooleanResponse { enabled: boolean }

export interface SentResponse { sent: boolean }

export interface PlatformSettingsResponse {
  platform_name: string
  allow_registration: boolean
  web_icon_url?: string
  smtp_configured?: boolean
  cap_configured?: boolean
  cap_api_endpoint?: string
  cap_site_key?: string
}

export interface SetupStatusResponse {
  initialized: boolean
}

export interface InitializeAdminPayload {
  username: string
  password: string
  email: string
}

export interface OAuthBindingRecord {
  id: string
  user_id: number
  provider: string
  email: string
  name: string
  created_at: string
}

export interface PasskeyLoginStartResponse {
  session_id: string
  options: PublicKeyCredentialRequestOptionsEnvelopeJSON
}

export interface PasskeyRegistrationStartResponse {
  session_id: string
  options: PublicKeyCredentialCreationOptionsEnvelopeJSON
}

export interface PasskeyRegistrationFinishResponse {
  credential: PasskeyRecord
}

export interface CreateClientInput {
  name: string
  description: string
  icon_url: string
  client_id: string
  client_type: string
  redirect_uris: string[]
  scopes: string[]
  trusted: boolean
}

export interface UpdateClientInput extends Partial<CreateClientInput> {}

export interface CreateUserInput {
  username: string
  password: string
  email: string
}

export interface UpdateUserInput extends Partial<CreateUserInput> {}
