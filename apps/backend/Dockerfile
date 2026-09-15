# ─────────────────────────────────────────────
# Stage 1 — Builder
# ─────────────────────────────────────────────
FROM golang:1.27.1-alpine AS builder

# Install git and ca-certs (needed for go mod download over HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go module files first for layer-cache efficiency
COPY apps/backend/go.mod apps/backend/go.sum ./
RUN go mod download

# Copy source
COPY apps/backend/ .

# Build arguments for version metadata
ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

# Build the API binary (HTTP server + embedded Asynq worker)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w \
      -X main.Version=${VERSION} \
      -X main.GitSHA=${GIT_SHA} \
      -X main.BuildTime=${BUILD_TIME}" \
    -o /dist/moneyflow-api \
    ./cmd/api

# Build the migrate binary (standalone migration runner)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w \
      -X main.Version=${VERSION} \
      -X main.GitSHA=${GIT_SHA} \
      -X main.BuildTime=${BUILD_TIME}" \
    -o /dist/moneyflow-migrate \
    ./cmd/migrate

# ─────────────────────────────────────────────
# Stage 2 — Runtime
# ─────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

# Distroless nonroot already uses uid=65532, no extra USER needed
LABEL org.opencontainers.image.title="moneyflow-api"
LABEL org.opencontainers.image.description="MoneyFlow Go backend (API + migrate)"
LABEL org.opencontainers.image.source="https://github.com/omar-shahieen/moneyflow"
LABEL org.opencontainers.image.name="omarmohamed04/moneyflow"

# Copy timezone data so TIMESTAMPTZ works correctly
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# Copy CA certs for outbound HTTPS (Clerk, Resend, New Relic, etc.)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy both binaries
COPY --from=builder /dist/moneyflow-api     /moneyflow-api
COPY --from=builder /dist/moneyflow-migrate /moneyflow-migrate

# Backwards-compatible symlink: /moneyflow → /moneyflow-api
# (K8s deployments that override command use the explicit binary names)
COPY --from=builder /dist/moneyflow-api /moneyflow

EXPOSE 8080

# Graceful shutdown via SIGTERM (Asynq / http.Server both respect it)
STOPSIGNAL SIGTERM

# Default: start the API server
ENTRYPOINT ["/moneyflow-api"]
