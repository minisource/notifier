# Notifier Service

A high-performance, scalable notification microservice supporting SMS, Email, Push, and In-App notifications.

## Features

- **Multi-Channel**: SMS, Email, Push, In-App notifications
- **API Access Model**: `/me` (user), `/admin` (operator), `/service` (internal) separation
- **RBAC**: JWT-based auth with admin, user, and service roles
- **Idempotency**: Duplicate prevention via idempotency keys
- **Templates**: Reusable templates with variable substitution
- **Preferences**: Per-user channel and category settings
- **Reminders**: Scheduled notification reminders
- **Workers**: Async processing with configurable pool and retry
- **Deliveries**: Delivery tracking with attempt history
- **Observability**: Dashboard, health, readiness, metrics, queue, workers
- **Production Hardening**: Rate limiting, request ID, PII sanitization, security headers

## Architecture

| Layer | Technology | Purpose |
|-------|-----------|---------|
| API | Fiber (Go) | RESTful HTTP + WebSocket |
| Auth | go-common middleware | JWT + service token validation |
| Service | Go service layer | Business logic, orchestration |
| Repository | GORM (PostgreSQL) | Database operations |
| Worker | Go goroutines | Async notification processing |
| WebSocket | Gorilla/websocket | Real-time in-app notifications |

## API Groups

| Group | Base Path | Auth | Purpose |
|-------|-----------|------|---------|
| Public | `/api/v1/health` | None | Service health |
| User | `/api/v1/me/*` | JWT (userId from token) | User's own notifications/prefs/reminders |
| Admin | `/api/v1/admin/*` | JWT + admin role | Full admin operations |
| Service | `/api/v1/service/*` | Service token | Internal notification creation |
| Legacy | `/api/v1/notifications/*` | JWT | Backward compatible user routes |
| WebSocket | `/ws` | JWT/Service | Real-time notifications |
| Metrics | `/metrics` | None | Prometheus metrics |
| Swagger | `/swagger/*` | None | API documentation |

## Quick Start

### Prerequisites
- Go 1.23+
- PostgreSQL 16+
- Docker (optional)

### Local Development

```bash
# 1. Clone and enter directory
git clone https://github.com/minisource/notifier.git
cd notifier/backend

# 2. Copy env config
cp .env.example .env

# 3. Start PostgreSQL
docker compose -f docker-compose.dev.yml up -d

# 4. Run migrations (if SQL migrations exist)
go run ./cmd/migrate up

# 5. Start server
go run ./cmd/server
```

Server starts on port **9002** (HTTP) and **9003** (gRPC).

### Verify

```bash
# Health check
curl http://localhost:9002/api/v1/health/

# Swagger UI
open http://localhost:9002/swagger/index.html
```

## Docker

```bash
# Build
docker build -t minisource-notifier .

# Run with Docker Compose (production)
docker compose -f docker-compose.prod.yml up -d

# Run with Docker Compose (development)
docker compose -f docker-compose.dev.yml up -d
```

## Configuration

All configuration is via environment variables. Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_INTERNAL_PORT` | `9002` | HTTP port |
| `AUTH_ENABLED` | `true` | Enable JWT auth |
| `POSTGRES_HOST` | `localhost` | DB host |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiter |

See [Configuration Documentation](docs/configuration.md) and `.env.example` for complete reference.

## Key Endpoints

### User (/me)
```http
GET    /api/v1/me/notifications
GET    /api/v1/me/notifications/unread
POST   /api/v1/me/notifications/read-all
GET    /api/v1/me/preferences
PUT    /api/v1/me/preferences
GET    /api/v1/me/reminders
POST   /api/v1/me/reminders
```

### Admin (/admin)
```http
GET    /api/v1/admin/dashboard/overview
GET    /api/v1/admin/notifications
POST   /api/v1/admin/notifications/{id}/retry
GET    /api/v1/admin/providers
GET    /api/v1/admin/providers/health
GET    /api/v1/admin/deliveries
GET    /api/v1/admin/observability/metrics
GET    /api/v1/admin/observability/queue
```

### Service (/service)
```http
POST   /api/v1/service/notifications
```

## Development Commands

```bash
make build        # Build binary
make test         # Run tests
make vet          # Run go vet
make swagger      # Regenerate Swagger
make validate     # Run all validations
make docker-build # Build Docker image
make run          # Build and run
```

## Documentation

| Doc | Path |
|-----|------|
| Endpoint Matrix | `docs/endpoint-implementation-matrix.md` |
| Integration Scenarios | `docs/integration-scenarios.md` |
| Error Codes | `docs/error-codes.md` |
| Configuration | `docs/configuration.md` |
| Database Schema | `docs/database.md` |
| API Client Generation | `docs/api-client-generation.md` |
| Production Checklist | `docs/production-readiness-checklist.md` |
| Release Checklist | `docs/release-checklist.md` |
| Production Readiness Report | `docs/final-production-readiness-report.md` |
| HTTP Examples | `docs/http/notifier.http` |

## Production Notes

1. **Auth**: `AUTH_ENABLED=true` with strong `AUTH_JWT_SECRET`
2. **Database**: Use SQL migrations (not GORM AutoMigrate) for schema changes
3. **CORS**: Restrict `CORS_ALLOW_ORIGINS` to known frontend domains
4. **Rate Limiting**: Enable and configure per-route limits
5. **Providers**: Configure SMS/Email/Push providers in DB `settings` table
6. **Logging**: Set `LOGGER_LEVEL=info` and `LOGGER_ENCODING=json`
7. **Workers**: Tune `WORKER_NUM_WORKERS` and `WORKER_QUEUE_SIZE` for expected load

## License

MIT License

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_STYLE.md](CODE_STYLE.md).
