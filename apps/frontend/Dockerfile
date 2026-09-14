# ─────────────────────────────────────────────
# Stage 1 — Builder
# ─────────────────────────────────────────────
FROM oven/bun:1.2-alpine AS builder

WORKDIR /build

# Copy workspace manifests for efficient caching
COPY package.json bun.lock nx.json ./
COPY apps/frontend/package.json apps/frontend/project.json* ./apps/frontend/
COPY packages/openapi/package.json packages/openapi/project.json* ./packages/openapi/
COPY packages/zod/package.json packages/zod/project.json* ./packages/zod/
COPY packages/emails/package.json packages/emails/project.json* ./packages/emails/

# Install all workspace dependencies
RUN bun install --frozen-lockfile

# Copy source files
COPY packages/ ./packages/
COPY apps/frontend/ ./apps/frontend/

# Build frontend (Vite produces /build/apps/frontend/dist)
RUN bun nx run frontend:build

# ─────────────────────────────────────────────
# Stage 2 — Runtime (Nginx)
# ─────────────────────────────────────────────
FROM nginx:1.27-alpine AS runtime

LABEL org.opencontainers.image.title="moneyflow-frontend"
LABEL org.opencontainers.image.description="MoneyFlow React/Vite SPA"
LABEL org.opencontainers.image.source="https://github.com/omar-shahieen/moneyflow"
LABEL org.opencontainers.image.name="omarmohamed04/moneyflow"

# Remove default nginx config and replace with our SPA config
RUN rm /etc/nginx/conf.d/default.conf
COPY deploy/nginx/spa.conf /etc/nginx/conf.d/spa.conf

# Copy built SPA assets
COPY --from=builder /build/apps/frontend/dist /usr/share/nginx/html

EXPOSE 80

STOPSIGNAL SIGTERM

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
