# ── Go API builder ──────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go

# ── SvelteKit builder ────────────────────────────────────────────────────────────
FROM node:22-alpine AS node-builder
WORKDIR /app
RUN npm install -g pnpm
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
# BETTER_AUTH_SECRET must exist at build time because SvelteKit runs server code
# during vite build. This placeholder is never used at runtime.
ARG BETTER_AUTH_SECRET=build-time-placeholder-not-used-at-runtime
RUN BETTER_AUTH_SECRET=$BETTER_AUTH_SECRET pnpm build

# ── Runtime ───────────────────────────────────────────────────────────────────────
FROM node:22-alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates bash
RUN npm install -g pnpm

# Go API
COPY --from=go-builder /app/api ./api
RUN touch .env

# SvelteKit app
COPY --from=node-builder /app/build ./build
COPY --from=node-builder /app/package.json /app/pnpm-lock.yaml ./
# Full install so drizzle-kit is available for migrations at startup
RUN pnpm install --frozen-lockfile --ignore-scripts
# Files drizzle-kit needs to push the schema
COPY --from=node-builder /app/drizzle.config.ts ./drizzle.config.ts
COPY --from=node-builder /app/src/lib/server/db ./src/lib/server/db

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENV NODE_ENV=production
EXPOSE 3000
ENTRYPOINT ["docker-entrypoint.sh"]
