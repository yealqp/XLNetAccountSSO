# OAuth2 SSO Platform

An OAuth2-compatible SSO starter built with Vue 3, Naive UI, TypeScript, Vue Router, Go Fiber, and MySQL.

## Projects

- `web/`: authorization UI and admin console
- `server/`: OAuth2-compatible authorization server
- `examples/client-demo/`: demo public client using Authorization Code + PKCE

## Local development

1. Start MySQL
   - `docker compose up -d mysql`
2. Copy env files
   - `copy server\.env.example server\.env`
   - `copy web\.env.example web\.env`
   - `copy examples\client-demo\.env.example examples\client-demo\.env`
3. Install frontend dependencies
   - `pnpm install`
4. Fetch Go dependencies
   - `go mod tidy` inside `server/`
5. Run projects
   - `go run ./cmd/sso` inside `server/`
   - `pnpm --filter @sso/web dev`
   - `pnpm --filter @sso/client-demo dev`

## Default seed data

- Admin user: `admin / Admin123!`
- Demo client ID: `demo-web-client`
- Demo client callback: `http://localhost:4173/callback`

## OAuth flow

1. Demo client redirects browser to `http://localhost:8080/oauth/authorize`
2. Backend forwards browser to the Vue authorization page
3. User logs in and approves the request
4. Vue app asks backend to issue an authorization code
5. Demo client exchanges the code for opaque tokens at `POST /oauth/token`

## Docker deployment

The repository now includes a full Docker stack for MySQL, the Fiber server, the Naive UI admin/auth app, and the demo public client.

### Services

- `mysql`: MySQL 8.4 with a persisted volume
- `server`: Go Fiber OAuth2 authorization server on `http://localhost:8080`
- `web`: Naive UI dark-theme admin/auth frontend on `http://localhost:5173`
- `client-demo`: PKCE demo client on `http://localhost:4173`

### Start the full stack

1. Build and start everything
   - `docker compose up --build -d`
2. Follow logs if needed
   - `docker compose logs -f server web client-demo`
3. Stop everything
   - `docker compose down`
4. Stop and remove MySQL data
   - `docker compose down -v`

### Docker defaults

- MySQL database: `sso_platform`
- MySQL app user: `sso_app / sso_app`
- Admin user: `admin / Admin123!`
- Demo client ID: `demo-web-client`
- Demo callback: `http://localhost:4173/callback`

### Optional Docker environment file

If you want to change host URLs, database credentials, or seeded account values for Docker, copy the root env template first:

1. `copy .env.example .env`
2. Edit `.env`
3. Rebuild with `docker compose up --build -d`

### Notes

- `web` uses Nginx to serve the built Vue app and reverse proxy `/api`, `/oauth`, and `/healthz` to the Fiber container.
- `client-demo` is served as a static SPA so `/callback` works correctly after OAuth redirection.
- The Docker stack is configured for local development-style URLs, so `SERVER_BASE_URL`, `WEB_BASE_URL`, and the demo callback already match the exposed host ports.
