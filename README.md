# OAuth2 SSO Platform

An OAuth2-compatible SSO starter built with Vue 3, Naive UI, TypeScript, Vue Router, Go Fiber, and MySQL.

## Projects

- `web/`: authorization UI and admin console
- `server/`: OAuth2-compatible authorization server

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
5. Run projects
   - `go run ./cmd/sso` inside `server/`
   - `pnpm --filter @sso/web dev`

## First-run setup

- The system no longer seeds a default admin account.
- On first boot, open the frontend and complete the initialization flow.
- The first administrator is created from `/api/setup/initialize` through the setup page.

## Docker deployment

The repository now includes a full Docker stack for MySQL, the Fiber server, and the Naive UI admin/auth app.

### Services

- `mysql`: MySQL 8.4 with a persisted volume
- `server`: Go Fiber OAuth2 authorization server on `http://localhost:8080`
- `web`: Naive UI dark-theme admin/auth frontend on `http://localhost:5173`

### Start the full stack

1. Build and start everything
   - `docker compose up --build -d`
2. Follow logs if needed
   - `docker compose logs -f server web`
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

- `web` uses Nginx to serve the built Vue app and reverse proxy `/api`, `/oauth`, and `/healthz` to the Fiber container.
- The Docker stack is configured for local development-style URLs, so `SERVER_BASE_URL` and `WEB_BASE_URL` already match the exposed host ports.
