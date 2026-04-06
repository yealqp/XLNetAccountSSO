# OAuth2 SSO Platform

An OAuth2-compatible SSO starter built with Vue 3, Naive UI, TypeScript, Vue Router, Go Fiber, and MySQL.

## Projects

- `web/`: authorization UI and admin console
- `server/`: OAuth2-compatible authorization server, also serves the built frontend

## Local development

1. Start MySQL
   - `docker compose up -d mysql`
2. Copy env files
   - `copy server\.env.example server\.env`
   - `copy web\.env.example web\.env`
3. Install frontend dependencies
   - `pnpm install`
4. Fetch Go dependencies
   - `go mod tidy` inside `server/`
5. Build frontend assets for embedding
   - `pnpm --filter @sso/web build`
6. Run projects
   - `go run ./cmd/sso` inside `server/`
   - Optional separate frontend dev server: `pnpm --filter @sso/web dev`

## First-run setup

- The system no longer seeds a default admin account.
- On first boot, open the frontend and complete the initialization flow.
- The first administrator is created from `/api/setup/initialize` through the setup page.

## Docker deployment

The repository now includes a full Docker stack for MySQL, the Fiber server, and the Naive UI admin/auth app.

### Services

- `mysql`: MySQL 8.4 with a persisted volume
- `server`: Go Fiber OAuth2 authorization server with embedded frontend on `http://localhost:8080`

### Start the full stack

1. Build and start everything
   - `docker compose up --build -d`
2. Follow logs if needed
   - `docker compose logs -f server`
3. Stop everything
   - `docker compose down`
4. Stop and remove MySQL data
   - `docker compose down -v`

### Docker defaults

- MySQL database: `sso_platform`
- MySQL app user: `sso_app / sso_app`
- First admin: created manually on first login/setup

### Optional Docker environment file

If you want to change host URLs, database credentials, or seeded account values for Docker, copy the root env template first:

1. `copy .env.example .env`
2. Edit `.env`
3. Rebuild with `docker compose up --build -d`

### Notes

- `pnpm --filter @sso/web build` writes the frontend bundle into `server/internal/app/web-dist/`, and the Go server embeds that directory into the final binary.
- The Docker server image builds the frontend first, then compiles a single backend binary with embedded assets.
- The Docker stack is configured for local development-style URLs and OIDC can be enabled through the OIDC-related environment variables in the example files.
