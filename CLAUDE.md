# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Backend (Go)
cd server && go run ./cmd/sso       # Start dev server (port :8080)
cd server && go build ./cmd/sso     # Build binary
cd server && go test ./...          # Run all tests
cd server && go mod tidy            # Tidy dependencies

# Frontend (Vue 3)
pnpm install                        # Install deps (root)
pnpm dev:web                        # Vite dev server (:5173, API proxy to :8080)
pnpm build:web                      # Production build

node scripts/test-oidc.js           # Test OIDC flow
```

## Project Structure

```
├── server/                         # Go Fiber OAuth2 authorization server
│   ├── cmd/sso/main.go             # Entrypoint (DB connection, auto-migrate, start Fiber)
│   └── internal/
│       ├── config/config.go        # Env-based config (DB, WebAuthn, OIDC, SMTP, CAPTCHA)
│       ├── app/app.go              # Wires Store → Services → Handlers → Fiber routes
│       ├── model/models.go         # GORM models (User, OAuthClient, tokens, sessions, etc.)
│       ├── repository/             # Data access layer (CRUD for all entities)
│       ├── service/                # Business logic (auth, OAuth2/OIDC, passkey, TOTP, admin)
│       ├── http/
│       │   ├── handlers/           # Fiber handlers (auth, OAuth, admin, passkey)
│       │   └── middleware/auth.go  # Session resolver, RequireSession, RequireAdmin
│       └── pkg/
│           ├── logger/             # Structured logging setup
│           └── security/           # OIDC token signing, password hashing, CSPRNG
├── web/                            # Vue 3 SPA (Arco Design Vue, Pinia, Vue Router)
│   └── src/
│       ├── api/                    # HTTP client (request<T>() + per-domain API modules)
│       ├── stores/                 # Pinia stores (session, setup, platform)
│       ├── components/             # Pages organized by domain (auth, admin, landing, ui)
│       ├── layouts/                # AuthLayout / AdminLayout (router-based)
│       ├── router/index.ts         # Routes with beforeEach guards (setup check, auth, admin)
│       ├── utils/                  # authToken, authExpiry, authNext, webauthn
│       ├── config/                 # endpoints (API URL resolution), theme (Naive UI overrides)
│       └── types/                  # API and WebAuthn response types
├── scripts/test-oidc.js           # OIDC conformance test helper
└── DESIGN.md                       # Visual design system (Resend-inspired dark theme)
```

## Architecture

### Backend Layered Architecture
- **Handler** (HTTP layer, validates input, calls service, returns JSON envelope: `{code, data, message}`)
- **Service** (Business logic, can call other services and the store)
- **Repository/Store** (GORM-based CRUD, single `Store` struct with all DB methods)
- **Config** loaded from env vars at startup

Routes are organized in `app.go`:
- `/.well-known/*` — OIDC discovery and JWKS
- `/oauth/*` — OAuth2 authorize/token/revoke/introspect/userinfo
- `/api/setup/*` — First-run admin initialization
- `/api/auth/*` — Login, register, passkey, TOTP
- `/api/*` (secured) — User profile, clients, tokens
- `/api/admin/*` — Admin-only: users, system settings, client management

### OAuth2 / OIDC Support
- Authorization code flow with PKCE (S256)
- Token exchange (authorization code → access+refresh tokens)
- Token introspection, revocation
- OpenID Connect discovery (`/.well-known/openid-configuration`) and JWKS
- UserInfo endpoint
- OIDC private key JWTs for ID tokens

### Authentication Methods
1. **Password** (bcrypt) + optional TOTP 2FA
2. **Passkeys** (WebAuthn) — login, registration, management
3. **Email verification codes** for registration and profile changes

### Frontend State & Guards
- `useSetupStore` — checks `/api/setup/status`, redirects to setup page if uninitialized
- `useSessionStore` — manages auth state, syncs with backend via `/api/auth/session`
- `usePlatformStore` — loads public settings (registration toggle etc.)
- Router `beforeEach` guard handles: setup check → auth check → admin role check

### Key Patterns
- Token hashing: sensitive tokens (session, auth code, access/refresh) are SHA-256 hashed before storing
- Session cookie: token stored in `localStorage` (managed by `authToken.ts`), sent as `Authorization: Bearer <token>`
- API envelope: all responses wrapped in `{code, data, message}` — typed via `ApiEnvelope<T>` on frontend
- Assets: binary data (icons, platform logos) stored as base64 in the `assets` DB table, served at `/api/assets/*`
