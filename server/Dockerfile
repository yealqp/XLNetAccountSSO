FROM node:22-alpine AS web-builder

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

WORKDIR /workspace

RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/package.json

RUN pnpm install --frozen-lockfile --filter @sso/web...

COPY web ./web

RUN pnpm --filter @sso/web build

FROM golang:1.25-alpine AS builder

WORKDIR /src/server

RUN apk add --no-cache ca-certificates git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server ./
COPY --from=web-builder /workspace/server/internal/app/web-dist ./internal/app/web-dist

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/sso ./cmd/sso

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget && adduser -D -h /app appuser

WORKDIR /app

COPY --from=builder /out/sso /app/sso

USER appuser

EXPOSE 8080

CMD ["./sso"]
