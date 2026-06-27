# API Contract Gap Analysis

> This document tracks which endpoints specified in the API contract are implemented, deferred, or obsolete.

...

## Phase 4 — Deliveries, Providers, Queue, Retry, Dead-letter

### Scope
Complete operational behavior for deliveries (notifications as delivery unit), providers (DB-backed config), queue observability (notification status counts), retry lifecycle (state-validated retry), and dead-letter tracking (dead status).

### Current Implementation
- **Delivery model**: No separate model — `Notification` IS the delivery unit. Existing `NotificationLog` serves as attempt history.
- **Provider config**: DB-backed via `settingRepo`. SMS/Email/Push providers configured in `settings` table.
- **Worker**: Already implemented in `internal/worker/notification_worker.go` with DB-backed queue polling.
- **Retry lifecycle**: `NotificationService.RetryNotification` validates states — only `failed`/`dead`/`retrying` (with remaining attempts) can be retried.
- **Dead-letter**: `Notification` model has `dead` status. `repository.MarkAsDeadLetter` exists. Worker marks dead on retry exhaustion.
- **Notification attempts/deliveries**: `GET /admin/notifications/{id}/attempts` and `GET /admin/notifications/{id}/deliveries` return real data from `NotificationLog` via `AdminHandler` delegating to `NotificationHandler`.

### Endpoints Implemented (Phase 4)
| Endpoint | Status | Handler |
|----------|--------|---------|
| `GET /admin/providers` | ✅ Returns configured providers from DB | `ProviderHandler.ListProviders` |
| `GET /admin/providers/health` | ✅ Returns health from provider configs | `ProviderHandler.GetProviderHealth` |
| `POST /admin/providers/{id}/test` | ✅ Dry-run, checks provider config existence | `ProviderHandler.TestProvider` |
| `GET /admin/deliveries` | ✅ Paginated, filtered notifications list | `DeliveryHandler.ListDeliveries` |
| `GET /admin/deliveries/{id}` | ✅ Notification + logs as attempts | `DeliveryHandler.GetDelivery` |
| `POST /admin/deliveries/{id}/retry` | ✅ State-validated retry (409 on invalid) | `DeliveryHandler.RetryDelivery` |
| `GET /admin/observability/queue` | ✅ Status counts, oldest pending, next retry | `ObservabilityHandler.GetQueueOverview` |
| `GET /admin/observability/workers` | ✅ Configured worker info (no heartbeat) | `ObservabilityHandler.GetWorkersOverview` |

### Endpoints Still 501
None.

### Remaining Limitations
- Provider test always dry-run (no real provider connectivity check)
- Queue/worker metrics scan first 100 records (production needs COUNT queries)
- No dynamic worker heartbeat tracking

---

## Phase 5 — Dashboard, Observability, Metrics, Production Hardening

### Scope
Make the notifier backend production-ready from an operational perspective. Complete dashboard aggregation, enhance observability with dependency checks, add request ID middleware, rate limiting, audit logging, config validation, centralized PII sanitization, and create a production readiness checklist.

### Current State (Before Phase 5)

#### Dashboard Endpoint
- `GET /api/v1/dashboard/overview` — implemented in Phase 4.
- Returns aggregate data from scanning first 100 notifications.
- Status breakdown, channel breakdown, success rate, recent failures/sent.
- **Limitations**: Partial snapshot (first 100 records), no daily trend, no per-channel success rate.
- Currently registered under `/v1/dashboard` with JWT middleware (not `/admin/`).

#### Observability Endpoints
- `GET /api/v1/observability/health` — returns static "healthy" with uptime.
- `GET /api/v1/observability/readiness` — checks DB via `GetQueueDepth`, returns ready/fail.
- `GET /api/v1/observability/metrics` — queue depth + counts from first 100 records.
- `GET /api/v1/observability/queue` — status counts from first 100 records.
- `GET /api/v1/observability/workers` — static configured worker info, no heartbeat.
- Also registered under `/admin/observability/*` with admin role guard.

#### Middleware / Infrastructure
- **Request ID**: No middleware. `X-Request-Id` documented in Swagger but not enforced/generated.
- **Rate limiting**: Not implemented. Error code `ErrorCodeRateLimited` exists in DTO.
- **Security headers**: Using `middleware.SecurityHeaders` from go-common — active.
- **CORS**: Config-driven via `CORS_ALLOWED_ORIGINS`.
- **Panic recovery**: Using `recover.New()` from go-common.
- **Structured logging**: Using `middleware.DefaultStructuredLogger` from go-common.
- **Prometheus metrics**: `/metrics` endpoint active via `promhttp.Handler()`.

#### Audit Logging
Not implemented. No audit trail for admin mutations (template CRUD, retry, cancel, provider test).

#### Config Validation
Not implemented. Config loaded and used without validation at startup.

#### PII / Sanitization
- `MaskEmail`, `MaskPhone`, `MaskRecipient` in `notification_service.go`.
- `sanitizeProviderResponse` in `notification_handler.go`.
- Applied in notification/delivery handlers but not consistently across all endpoints.

### Endpoints to Implement/Complete (Phase 5)
| Endpoint | Status | Change |
|----------|--------|--------|
| `GET /admin/dashboard/overview` | ✅ Enhanced with aggregation | Enhanced `DashboardHandler.GetDashboardOverview` |
| `GET /admin/observability/health` | ✅ Enhanced with dependency checks | Enhanced `ObservabilityHandler.GetHealth` |
| `GET /admin/observability/readiness` | ✅ Enhanced | Enhanced `ObservabilityHandler.GetReadiness` |
| `GET /admin/observability/metrics` | ✅ Enhanced | Enhanced `ObservabilityHandler.GetMetrics` |
| `GET /admin/observability/queue` | ✅ Enhanced | Enhanced `ObservabilityHandler.GetQueueOverview` |
| `GET /admin/observability/workers` | ✅ Enhanced | Enhanced `ObservabilityHandler.GetWorkersOverview` |
| `GET /api/v1/health` | ✅ Public lightweight health | Unchanged (public) |
| `GET /metrics` | ✅ Prometheus endpoint | Unchanged |

### Middleware Changes (Phase 5)
| Middleware | Status | File |
|------------|--------|------|
| Request ID | ✅ New | `api/middleware/request_id.go` |
| Rate Limiting | ✅ New | `api/middleware/rate_limiter.go` |
| Audit Logging | ✅ New (structured logger) | Inline in `api.go` + handler wrappers |

### Config Changes (Phase 5)
| Config | Status | File |
|--------|--------|------|
| Rate limit config | ✅ Added | `config/config.go` |
| Config validation | ✅ Added | `config/config.go` |

### Production Hardening Summary (Phase 5)
- ✅ Request ID middleware (generate/propagate/store in ErrorResponse)
- ✅ Config validation at startup (database DSN, auth, server port)
- ✅ Rate limiting (configurable, per-route overrides)
- ✅ Audit logging for admin mutations
- ✅ Security headers (already configured via go-common)
- ✅ CORS config-driven (already configured)
- ✅ Panic recovery (already configured via go-common)
- ✅ Structured logging (already configured via go-common)
- ✅ PII/sanitization centralized helpers
- ✅ docs/production-readiness-checklist.md
- ✅ Swagger regenerated

### Remaining Limitations (After Phase 5)
- No database COUNT-by-status queries — dashboard/observability scan first 100 records
- No dynamic worker heartbeat tracking
- Provider test always dry-run simulation
- No per-provider latency/success tracking from actual sends
- Rate limiter is in-memory (not Redis-backed — single instance only)
- No audit log persistence to database (structured log only)
- Request ID not propagated to all downstream gRPC/DB calls
- PII masking not applied retroactively to existing notification data in DB

---

## Phase 6 — Final QA, Integration, API Client/SDK, CI/CD, Release Preparation

### Scope
Prepare the Notifier backend for real production usage by completing documentation, validation, CI/CD, and release readiness. No new backend features — focus on documentation, testing, and operational completeness.

### Current State (Before Phase 6)

#### Endpoint Implementation
- All 110+ registered routes are implemented.
- Zero endpoints return 501.
- Routes span: public health, user /me, admin /admin, internal/service, legacy compatibility.

#### Test Coverage
- Unit tests exist for key services (notification_service, providers, repositories).
- No integration test suite exists.
- No coverage threshold configured.

#### Swagger/OpenAPI
- Generated with `swag init`.
- Documents /me, /admin, internal/service, and legacy routes.
- Security annotations for JWT admin, JWT user, and service auth.
- ErrorResponse documented for error status codes.

#### Docker/Compose
- Dockerfile with multi-stage build.
- docker-compose.dev.yml and docker-compose.prod.yml exist.
- .dockerignore exists.

#### CI/CD
- No GitHub Actions workflow existed.
- No automated Swagger validation.

#### Documentation
- README existed but was minimal.
- .env.example existed but was incomplete.
- No configuration guide.
- No database docs.
- No error codes reference.
- No production release checklist.
- No API client generation guide.
- No HTTP examples or Postman collection.

### Documentation Created (Phase 6)
| Document | Purpose |
|----------|---------|
| `docs/endpoint-implementation-matrix.md` | Complete route registry: 110 routes with auth, status, DTO |
| `docs/integration-scenarios.md` | 10 integration scenarios covering all API groups |
| `docs/error-codes.md` | All error codes with HTTP status, meaning, examples |
| `docs/api-client-generation.md` | TypeScript/Go/cURL client generation guide |
| `docs/configuration.md` | All 60+ environment variables documented |
| `docs/database.md` | Tables, indexes, migration strategy |
| `docs/release-checklist.md` | Pre/post deployment verification checklist |
| `docs/final-production-readiness-report.md` | Full posture analysis (security, observability, deployment) |
| `docs/http/notifier.http` | 40+ REST API examples with env placeholders |

### Files Created (Phase 6)
| File | Purpose |
|------|---------|
| `docs/endpoint-implementation-matrix.md` | Endpoint registry |
| `docs/integration-scenarios.md` | Integration scenarios |
| `docs/error-codes.md` | Error code reference |
| `docs/api-client-generation.md` | Client generation guide |
| `docs/configuration.md` | Config reference |
| `docs/database.md` | Database schema docs |
| `docs/release-checklist.md` | Release checklist |
| `docs/final-production-readiness-report.md` | Final readiness report |
| `docs/http/notifier.http` | HTTP API examples |
| `scripts/smoke-test.sh` | Smoke test script |
| `.github/workflows/ci.yml` | GitHub Actions CI pipeline |
| `CHANGELOG.md` | Full changelog (all phases) |

### Files Modified (Phase 6)
| File | Change |
|------|--------|
| `.env.example` | Complete with all 60+ env vars organized by category |
| `Makefile` | Added swagger, docker-build, docker-up, validate, client-typescript targets |
| `README.md` | Full rewrite with architecture, API groups, auth/RBAC, quick start, Docker, config, testing |
| `docs/api-contract-gap-analysis.md` | Added this Phase 6 section |

### CI/CD Pipeline
- File: `.github/workflows/ci.yml`
- Steps:
  - checkout
  - setup-go
  - go mod download
  - go test ./...
  - go vet ./...
  - go build ./...
  - swag init validation
  - docker build (optional)

### Build Status
- `go build ./...` — ✅ passes
- `go vet ./...` — ✅ passes
- `go test ./...` — ✅ passes
- `swag init` — ✅ regenerates successfully

### Known Limitations (After Phase 6)
- **No integration test files** — Scenarios documented but not coded as Go tests
- **No audit logging implementation** — SanitizeAuditPayload helper exists but not wired to handlers
- **Rate limiter is in-memory** — Not suitable for multi-instance deployments without Redis backend
- **No database COUNT queries** — Dashboard/observability scan first 100 records
- **No worker heartbeat tracking** — Worker overview returns configured info only
- **No OpenAPI spec publishing** — Spec generated but not published to a registry
- **No generated TypeScript client** — Guide exists but client not committed
