# syntax=docker/dockerfile:1

# ── Builder ──────────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Copy source (vendor excluded via .dockerignore)
COPY backend/ ./

# Build with module download cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=mod -trimpath -ldflags="-s -w" -o /notifier-server ./cmd/server

# ── Runtime ──────────────────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl

# Non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

COPY --from=builder /notifier-server /app/notifier-server
COPY backend/config/ /app/config/
COPY backend/locales/ /app/locales/

RUN mkdir -p /app/logs && chown -R appuser:appgroup /app
USER appuser

WORKDIR /app
EXPOSE 9002

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -sf http://127.0.0.1:9002/health || exit 1

ENTRYPOINT ["/app/notifier-server"]
