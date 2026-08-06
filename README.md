# Minisource Notifier Microservice & Admin Dashboard

Notification microservice and administration panel for the Minisource platform. Multi-channel support for SMS, Email, Push, and In-App notifications with async worker pools, delivery tracking, and Next.js admin UI.

## Docker Images

Published to Docker Hub:
- **Backend**: `minisource/notifier-backend:v1.0.1`
- **Frontend**: `minisource/notifier-frontend:v1.0.1`

```bash
docker pull minisource/notifier-backend:v1.0.1
docker pull minisource/notifier-frontend:v1.0.1
```

## Repository Architecture

```
notifier/
├── backend/          # Go API server (Fiber + PostgreSQL + Redis + gRPC)
│   ├── cmd/
│   ├── internal/
│   └── Dockerfile
├── front/            # Next.js Admin Panel (Next.js 14/15 + TailwindCSS)
│   ├── src/
│   └── Dockerfile
├── docker-compose.yml
└── .github/workflows/ # GitHub Actions CI/CD workflows
```

## Services

| Service | Stack | Port | Description |
|---------|-------|------|-------------|
| **Backend** | Go + Fiber v2 | 9002 (HTTP), 9003 (gRPC) | Notification API, async worker pool, providers |
| **Frontend** | Next.js + React | 3001 / 3004 | Admin dashboard, template editor, delivery logs |

## Features

### Backend
- **Multi-channel**: SMS, Email, Push, In-App notifications
- **Async Workers**: Configurable worker pool with automated retries and dead-letter queues
- **Templating Engine**: Dynamic template substitution with variable validation
- **Channel Preferences**: Per-user and per-tenant notification settings
- **Delivery Tracking**: Full attempt history, status updates, and delivery logs
- **Real-Time In-App**: WebSockets for instant delivery of in-app notifications
- **Security & Multi-Tenancy**: PII sanitization, idempotency key enforcement, rate limiting, and JWT/Service token validation

### Frontend (`front/`)
- **Metrics Dashboard**: Delivery statistics, success/failure ratios, and active channel metrics
- **Notification Manager**: Send, cancel, inspect, and retry notifications
- **Template Manager**: Rich template editor with variable placeholders
- **Delivery Logs & Health**: Provider health checks and detailed delivery logs

## Quick Start

### Running with Docker Compose

```bash
cd notifier
docker compose up -d
```

### Local Development

#### Backend
```bash
cd notifier/backend
cp .env.example .env
docker compose -f docker-compose.dev.yml up -d
go run ./cmd/server
```
Server: `http://127.0.0.1:9002`  
Swagger: `http://127.0.0.1:9002/swagger/index.html`

#### Frontend
```bash
cd notifier/front
npm install --legacy-peer-deps
npm run dev
```
Frontend: `http://127.0.0.1:3004`

## Configuration

Key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_INTERNAL_PORT` | `9002` | HTTP API port |
| `GRPC_PORT` | `9003` | gRPC service port |
| `AUTH_ENABLED` | `true` | Enable JWT authentication |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `REDIS_HOST` | `localhost` | Redis host |
| `NEXT_PUBLIC_API_URL` | `http://localhost:9002` | Frontend API endpoint |

## CI/CD Pipeline

Automated via GitHub Actions (`.github/workflows/ci.yml`):
- **Go Backend**: `go vet` and test execution
- **Frontend**: ESLint, TypeScript compilation check, Vitest suite
- **Docker Push**: Automated multi-stage build & push to `minisource/notifier-backend` and `minisource/notifier-frontend` on release tag push (`v*`).
