# Minisource Notifier

Notification microservice for the Minisource platform. Supports SMS, Email, Push, and In-App notifications with a full admin panel.

## Architecture

```
notifier/
├── backend/          # Go API server (Fiber + PostgreSQL + Redis)
├── front/            # Next.js admin panel
├── proto/            # Protobuf definitions (gRPC)
├── .env              # Local development config
└── docker-compose.yml
```

## Services

| Service | Stack | Port | Description |
|---------|-------|------|-------------|
| **Backend** | Go + Fiber v2 | 9002 (HTTP), 9003 (gRPC) | Notification API, workers, providers |
| **Frontend** | Next.js | 3001 | Admin dashboard, notification management |

## Quick Start

### Backend

```bash
cd backend

# Copy env
cp .env.example .env

# Start dependencies
docker compose -f docker-compose.dev.yml up -d

# Run
go run ./cmd/server
```

Server: `http://127.0.0.1:9002`
Swagger: `http://127.0.0.1:9002/swagger/index.html`

### Frontend

```bash
cd front

# Install
npm install

# Run
npm run dev
```

Frontend: `http://127.0.0.1:3004`

## Features

### Backend
- Multi-channel: SMS, Email, Push, In-App
- Async workers with configurable pool and retry
- Templates with variable substitution
- Per-user channel preferences
- Scheduled reminders
- Delivery tracking with attempt history
- Idempotency keys for duplicate prevention
- WebSocket for real-time in-app notifications
- Rate limiting, security headers, PII sanitization
- JWT + service token authentication
- Prometheus metrics, health/readiness probes

### Frontend
- Dashboard with notification metrics
- Notification management (list, create, retry, cancel)
- Template CRUD
- User preference management
- Delivery logs
- Provider health monitoring

## API Groups

| Group | Base Path | Auth | Purpose |
|-------|-----------|------|---------|
| Public | `/v1/health` | None | Health checks |
| User | `/v1/me/*` | JWT | User's own notifications/preferences |
| Admin | `/v1/admin/*` | JWT + admin | Full admin operations |
| Service | `/v1/service/*` | Service token | Internal notification creation |
| WebSocket | `/ws` | JWT/Service | Real-time notifications |
| Metrics | `/metrics` | None | Prometheus metrics |

## Configuration

Key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_INTERNAL_PORT` | `9002` | HTTP port |
| `GRPC_PORT` | `9003` | gRPC port |
| `AUTH_ENABLED` | `true` | Enable JWT auth |
| `AUTH_JWKS_URL` | | JWKS URL for token validation |
| `POSTGRES_HOST` | `localhost` | DB host |
| `REDIS_HOST` | `localhost` | Redis host |
| `WORKER_NUM_WORKERS` | `5` | Async worker count |
| `WORKER_QUEUE_SIZE` | `100` | Worker queue size |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |

## Docker

```bash
# Build
docker build -t minisource/notifier ./backend

# Run (production)
docker compose -f docker-compose.prod.yml up -d

# Run (development)
docker compose -f docker-compose.dev.yml up -d
```

### Docker Hub

| Image | Description |
|-------|-------------|
| `minisource/notifier` | Backend API |
| `minisource/notifier-front` | Frontend (admin panel) |

Images are built and pushed automatically on push to `main`.

## Auth Integration

The notifier validates JWT tokens from the Auth service:

1. **JWKS** (recommended) — Fetch public keys from Auth's `/.well-known/jwks.json`
2. **Token validation** — Call `GET /v1/tokens/validate` on Auth service
3. **Introspection** — Call `POST /v1/auth/introspect` on Auth service

### Service Client

| Field | Value |
|-------|-------|
| Client ID | `notifier-service` |
| Scopes | `tokens:validate` |

## Development

```bash
# Backend
cd backend
go build ./...
go vet ./...
go test ./...
swag init -g cmd/server/main.go

# Frontend
cd front
npm run lint
npm run type-check
npm run build
```

## Documentation

| Doc | Path |
|-----|------|
| Endpoint Matrix | `backend/docs/endpoint-implementation-matrix.md` |
| Integration Scenarios | `backend/docs/integration-scenarios.md` |
| Error Codes | `backend/docs/error-codes.md` |
| Configuration | `backend/docs/configuration.md` |
| Database Schema | `backend/docs/database.md` |
| Production Checklist | `backend/docs/production-readiness-checklist.md` |
| Production Readiness Report | `backend/docs/final-production-readiness-report.md` |

## License

MIT
