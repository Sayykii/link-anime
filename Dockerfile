# ============================================================
# Stage 1: Build frontend with Bun
# ============================================================
FROM oven/bun:1 AS frontend-builder

WORKDIR /build/frontend
COPY frontend/package.json frontend/bun.lock* ./
RUN bun install --frozen-lockfile || bun install

COPY frontend/ ./
RUN bun run build-only

# ============================================================
# Stage 2: Build Go backend
# ============================================================
FROM golang:1.24-alpine AS backend-builder

RUN apk add --no-cache git

WORKDIR /build

# Copy go module files and download deps
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Copy frontend dist into cmd/server/dist so embed can find it
COPY --from=frontend-builder /build/frontend/dist ./cmd/server/dist/

# Tidy modules (resolves all imports and creates go.sum)
RUN go mod tidy

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /link-anime ./cmd/server

# ============================================================
# Stage 3: Runtime
# ============================================================
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata shadow su-exec

# Create app directory
RUN mkdir -p /app/data

# Copy binary
COPY --from=backend-builder /link-anime /app/link-anime

# Copy entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Default env vars
ENV LA_PORT=8787 \
    LA_DATA_DIR=/app/data \
    LA_PASSWORD=changeme \
    LA_DOWNLOAD_DIR=/data/downloads/complete/anime \
    LA_MEDIA_DIR=/data/media/anime \
    LA_MOVIES_DIR=/data/media/anime-movies \
    PUID=977 \
    PGID=988 \
    TZ=Europe/Sofia

EXPOSE 8787

VOLUME ["/app/data"]

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/link-anime"]
