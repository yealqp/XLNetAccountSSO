# OAuth2 SSO Platform

An OAuth2-compatible SSO platform built with Vue 3 (Naive UI) + Go Fiber + MySQL.

## Projects

- `web/`: Authorization UI and admin console (独立部署)
- `server/`: OAuth2-compatible authorization server + admin API

## Local development

### Backend

1. Start MySQL: `docker compose up -d mysql`
2. Copy env: `copy server\.env.example server\.env`
3. Fetch deps: `go mod tidy` inside `server/`
4. Run: `go run ./cmd/sso` inside `server/`

### Frontend

1. `pnpm install`
2. `pnpm dev:web` (Vite dev server on `http://localhost:5173`, API proxy to backend)

## Passkeys / WebAuthn

- Passkey login and binding require a secure context (`https://...`) or `localhost`.
- `WEBAUTHN_RP_ID` should match the login page domain only (no scheme or port).
- `WEBAUTHN_RP_ORIGINS` accepts a comma-separated list of allowed frontend origins.

## First-run setup

- The system no longer seeds a default admin account.
- On first boot, open the frontend and complete the initialization flow.
- The first administrator is created from `/api/setup/initialize` through the setup page.

## Docker deployment

### Services

- `mysql`: MySQL 8.0 with a persisted volume
- `server`: Go Fiber OAuth2 authorization server on `http://localhost:8080`

### Start

- Full stack: `docker compose up --build -d`
- Just server: `docker compose up --build -d server`

### Docker defaults

- MySQL database: `sso_platform`
- MySQL app user: `sso_app / sso_app`
