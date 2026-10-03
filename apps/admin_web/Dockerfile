# syntax=docker/dockerfile:1.7
#
# Admin console image: the Vue bundle is built once, then served as static
# files by an unprivileged nginx. The runtime has no Node.js or build tooling.

# The type-check step runs the native `typescript-native-bridge` (tsgo)
# binary, which links against glibc and fails on musl-based Alpine.
ARG NODE_IMAGE=node:24-bookworm-slim
ARG NGINX_IMAGE=nginxinc/nginx-unprivileged:1.29-alpine

FROM ${NODE_IMAGE} AS build

WORKDIR /src

# Vite inlines VITE_* values at build time, so the API origin is fixed per
# image. Build one image per environment, or route through the same origin.
ARG VITE_API_BASE_URL=
ARG VITE_API_TIMEOUT_MS=15000
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_API_TIMEOUT_MS=${VITE_API_TIMEOUT_MS}

# Install dependencies first so source edits do not invalidate the npm layer.
COPY apps/admin_web/package.json apps/admin_web/package-lock.json ./apps/admin_web/
WORKDIR /src/apps/admin_web
RUN npm ci

# The console imports generated TypeScript contracts from packages/. Keep the
# repository-relative layout so the existing import path works unchanged.
WORKDIR /src
COPY apps/admin_web ./apps/admin_web
COPY packages/contracts/generated/typescript ./packages/contracts/generated/typescript
COPY VERSION ./VERSION
WORKDIR /src/apps/admin_web
RUN npm run build

FROM ${NGINX_IMAGE}

# Vite emits origin-relative asset URLs, so the bundle works behind any host.
COPY --from=build /src/apps/admin_web/dist /usr/share/nginx/html
COPY apps/admin_web/docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -q -T 3 -O /dev/null http://127.0.0.1:8080/healthz || exit 1
