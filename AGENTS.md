# AGENTS.md

Repository guide for coding agents working in this project.

## Project Overview

- Monorepo root contains two active projects:
  - `server/`: Go 1.25 + Fiber + GORM + MySQL OAuth2/OIDC-compatible auth server
  - `web/`: Vue 3 + TypeScript + Vite + Naive UI admin/auth frontend
- Package manager: `pnpm`
- Default frontend app: `@sso/web`
- Current app name is configurable from the admin Settings page and persisted in DB.
- First boot no longer auto-creates an admin; the system is initialized through the setup flow.

## Rules Files

- No `.cursorrules` file found.
- No `.cursor/rules/` directory found.
- No `.github/copilot-instructions.md` file found.
- If these files are added later, update this document and follow them as higher-priority repository guidance.

## Working Directories

- Repo root: run workspace-level `pnpm` and Docker commands here.
- Backend workdir: `server/`
- Frontend workdir: `web/`

## Build / Run Commands

### Root

- Install workspace deps: `pnpm install`
- Run frontend dev server: `pnpm --filter @sso/web dev`
- Build frontend: `pnpm --filter @sso/web build`
- Start all Docker services: `docker compose up --build -d`
- Start only backend container: `docker compose up --build -d server`
- Start only frontend container: `docker compose up --build -d web`

### Backend (`server/`)

- Run server locally: `go run ./cmd/sso`
- Format Go code: `gofmt -w cmd internal`
- Run all backend tests: `go test ./...`
- Run one package tests: `go test ./internal/service`
- Run one named test: `go test ./internal/service -run TestName`
- Run one test with verbose output: `go test -v ./internal/service -run TestName`

### Frontend (`web/`)

- Dev server: `pnpm dev`
- Production build: `pnpm build`
- Preview built app: `pnpm preview`
- Type-check is part of build via `vue-tsc --noEmit`

## Lint / Format Status

- There is currently no dedicated lint script in root, `web/`, or `server/`.
- For Go changes, always run `gofmt -w cmd internal`.
- For frontend changes, rely on existing style conventions and run `pnpm --filter @sso/web build` to catch TS/SFC issues.

## Required Verification Before Finishing

- Backend-only changes: `gofmt -w cmd internal && go test ./...`
- Frontend-only changes: `pnpm --filter @sso/web build`
- Cross-stack changes: run both commands above.
- If you touch auth, OAuth, tokens, setup, or settings flows, prefer verifying both backend and frontend.

## Architecture Notes

- Backend is layered as config -> repository -> service -> HTTP handlers.
- Keep business rules in `internal/service/`.
- Keep DB access in `internal/repository/`.
- Keep request/response translation in `internal/http/handlers/`.
- Frontend uses route-level SFC pages in `web/src/components/auth/` and `web/src/components/admin/`.
- Layout shells live in `web/src/layouts/`.
- Shared API wrappers live in `web/src/api/`.
- Shared Pinia state lives in `web/src/stores/`.

## Backend Style Guide

- Use `gofmt`; do not manually fight its formatting.
- Keep packages small and focused.
- Prefer explicit service methods over embedding business logic in handlers.
- Return sentinel errors from `internal/service/errors.go` when the handler needs to map status codes.
- Wrap user-facing validation failures with context, e.g. `fmt.Errorf("%w: 中文提示", ErrInvalidInput)`.
- Use repository helpers rather than ad-hoc SQL in handlers or services.
- Keep HTTP handlers thin: parse input, call service, map response/error.
- Use `fiber.Map` for simple JSON responses and helper serializers like `publicUser` for repeated structures.
- Use `context.Background()` consistently in current codepaths unless there is an existing request context pattern nearby.
- IDs and opaque tokens are generated via `internal/pkg/security` helpers.
- Passwords are always hashed; never store or log plaintext secrets.
- Remove related auth artifacts when deleting principals/resources where required by current service logic.

## Backend Naming Conventions

- Exported structs and methods: PascalCase.
- Unexported helpers: camelCase.
- Prefer clear nouns for models (`OAuthClient`, `AccessToken`) and verbs for service methods (`CreateClient`, `DeleteUser`).
- Keep JSON field names snake_case to match existing API responses.

## Frontend Style Guide

- Use Vue 3 SFCs with `<script setup lang="ts">`.
- Prefer Composition API primitives (`computed`, `shallowRef`, `reactive`, `watch`).
- Imports are grouped as:
  1. framework/library imports
  2. local app imports
  3. type-only imports where helpful
- Use `@/` alias for app-local imports.
- Match existing style: no semicolons, single quotes, concise computed/state names.
- Keep route pages in SFC components; avoid recreating a `views/` tree.
- Favor Naive UI components instead of custom HTML/CSS when the library already provides the pattern.
- Keep custom CSS scoped and minimal.
- Use `shallowRef` for async-loaded objects and arrays unless deep reactivity is required.
- Surface API failures through `ApiError` and Naive UI feedback (`useMessage`, inline `NAlert`, retry buttons).

## Frontend Naming Conventions

- Components: PascalCase file names, e.g. `ClientsPage.vue`.
- Stores: `useXxxStore`.
- Composables: `useXxx.ts`.
- API wrappers: verb-based functions like `fetchClients`, `updatePlatformSettings`.
- Use descriptive booleans such as `isSubmitting`, `isUploadingIcon`, `loadError`.

## Auth / OAuth Behavior to Preserve

- System supports first-run initialization via `/api/setup/status` and `/api/setup/initialize`.
- Login must not assume a seeded admin exists.
- OAuth authorize flow may be entered through `/auth/authorize` on the frontend or `/oauth/authorize` on the backend.
- `userinfo` responses are scope-sensitive.
- Refresh-token rotation must revoke the old refresh token and the old access token.
- Client deletion should also delete related authorization codes, access tokens, refresh tokens, and icon files.
- User deletion should also delete related authorization codes, tokens, and sessions.

## Settings / Branding Notes

- Platform name is persisted in DB, not hardcoded only in frontend.
- Public settings are exposed through `/api/settings/public`.
- Admin settings are managed through `/api/settings/platform`.
- When changing branding, update both frontend usage and backend defaults/fallbacks if relevant.

## Client Icon Notes

- `icon_url` may be a remote URL or an uploaded file.
- Remote icons are downloaded by the backend and re-served from `/client-icons/...`.
- Uploaded icons are stored by the backend before being attached to a client.
- Do not bypass the backend by making the UI depend on third-party remote image URLs after save.

## Documentation Expectations

- Update `README.md` when startup flow, initialization flow, or exposed endpoints change.
- Update this `AGENTS.md` when scripts, conventions, or repository rules change.

## Safe Agent Behavior

- Do not assume demo data exists; it has been removed.
- Do not reintroduce default admin auto-fill or seed-only login hints unless explicitly requested.
- Prefer additive schema changes over destructive migrations unless the user asks for cleanup.
- If you add a new endpoint or admin page, wire both frontend route/navigation and backend handler/service layers consistently.
