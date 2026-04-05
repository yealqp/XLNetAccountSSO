FROM node:22-alpine AS builder

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

ARG VITE_AUTH_BASE_URL=http://localhost:8080
ARG VITE_CLIENT_ID=demo-web-client

ENV VITE_AUTH_BASE_URL=$VITE_AUTH_BASE_URL
ENV VITE_CLIENT_ID=$VITE_CLIENT_ID

WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/package.json
COPY examples/client-demo/package.json examples/client-demo/package.json

RUN pnpm install --frozen-lockfile --filter @sso/client-demo...

COPY examples/client-demo ./examples/client-demo

RUN pnpm --filter @sso/client-demo build

FROM nginx:1.27-alpine

COPY examples/client-demo/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/examples/client-demo/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
