# Minisource Notifier — Production-Ready Completion Roadmap

> **Date:** June 22, 2026  
> **Repository:** `github.com/minisource/notifier`  
> **Module:** `github.com/minisource/notifier`  
> **Go Version:** 1.26.0 (local)  
> **Build Status:** `go build ./cmd/server` ✅ passes  
> **Test Status:** `go test ./...` ✅ passes (1 package with tests)  
> **Go Vet:** ✅ passes  
> **Go Mod Tidy:** ❌ fails (expected — uses Go workspace with local modules)

---

## 1. Current Repository Status

### 1.1 Build/Package Status

| Check | Status | Details |
|---|---|---|
| `go version` | ✅ 1.26.0 | Installed locally |
| `go build ./cmd/server` | ✅ Passes | Server binary builds |
| `go build ./cmd/migrate` | ✅ Passes | Migration binary builds |
| `go vet ./...` | ✅ Passes | No issues found |
| `go test ./...` | ✅ Passes | Only `internal/platform/sms` has tests (1.6s) |
| `go mod tidy` | ❌ Fails | `github.com/minisource/go-common@v0.0.4-0.20250402190339-caa3304676a9: invalid version: unknown revision caa3304676a9` — expected in workspace mode |
| `go.work` | ✅ Exists | Root-level Go workspace includes all services |

### 1.2 Directory Structure

```
notifier/backend/
├── cmd/server/main.go            # Entry point
├── cmd/migrate/main.go           # Migration CLI
├── cmd/initializer/              # Bootstrap logic (5 files)
├── cmd/sms_probe/main.go         # Debug SMS tool
├── api/
│   ├── api.go                    # Fiber + route registration
│   ├── middleware/service_auth.go
│   ├── grpc/                     # gRPC server + handlers (3 files)
│   └── v1/
│       ├── dto/                  # Request/response DTOs (2 files)
│       ├── handlers/             # HTTP handlers (3 files)
│       └── routes/               # Route registration ✅ ALL EXIST
│           ├── health.go
│           ├── notification_router.go
│           ├── preference_router.go
│           ├── template_router.go
│           ├── sms_router.go
│           └── websocket_router.go
├── config/config.go              # Env-based config
├── internal/
│   ├── database/database.go      # DB connection + AutoMigrate + seed
│   ├── models/ (7 files)         # GORM models
│   ├── platform/
│   │   ├── email/email.go        # SMTP only
│   │   ├── push/push.go          # FCM legacy HTTP API
│   │   └── sms/                  # 21 SMS providers + factory
│   ├── repository/ (6 files)     # Repository interfaces + implementations
│   ├── service/ (6 files)        # Business logic
│   ├── websocket/hub.go          # WebSocket hub
│   └── worker/ (2 files)         # In-memory queue + retry
├── migrations/ (10 files)        # SQL migrations
├── pkg/
│   ├── metrics/ (2 files)        # Prometheus counters/histograms
│   ├── service_errors/ (2 files)
│   └── tracing/tracer.go         # Jaeger exporter
├── proto/notifier/v1/notifier.proto  # Protobuf definitions
├── tests/
│   ├── e2e/ (6 files)            # E2E tests (build-tagged)
│   └── load-test/                # k6 scripts
├── docs/                         # Swagger, Postman, documentation
├── Dockerfile                    # Multi-stage build
├── docker-compose*.yml           # Dev + Prod compose files
├── Makefile                      # Build/test/migrate commands
└── Taskfile.yml                  # Alternative task runner
```

### 1.3 Configuration Files

| File | Status | Notes |
|---|---|---|
| `config/config.go` | ✅ Implemented | Env-based with defaults via `godotenv` |
| `.env.example` | ✅ Implemented | Comprehensive reference |
| `.env` | ✅ Exists | Gitignored |
| `docker-compose.dev.yml` | ✅ Implemented | PostgreSQL 16 + Redis 7 + MailHog + Adminer |
| `docker-compose.prod.yml` | ✅ Implemented | Full production setup |

---

## 2. Current Build/Test Status

### 2.1 Build Results

```
$ go build ./cmd/server
→ SUCCESS (no output, exit 0)

$ go build ./cmd/migrate  
→ SUCCESS (no output, exit 0)

$ go vet ./...
→ SUCCESS (no output, exit 0)
```

### 2.2 Test Results

```
$ go test ./...
?       github.com/minisource/notifier/api        [no test files]
?       github.com/minisource/notifier/api/grpc   [no test files]
... (many no test files)
ok      github.com/minisource/notifier/internal/platform/sms   1.633s
```

**Critical:** Only 1 package has unit tests: `internal/platform/sms`.  
All other packages have **zero test files**.

### 2.3 E2E Tests

The `tests/e2e/` directory has 6 test files, but they are build-tagged with `//go:build e2e` and require:
- Running auth service on `localhost:9001`
- Running notifier on `localhost:9002`
- Database with seed data
- Service tokens

These are **not runnable** in normal `go test ./...`.

---

## 3. Critical Blockers

### Blocker 1: Migration Version Conflict ⚠️

**Files:** `migrations/000002_add_multi_tenancy.up.sql` and `migrations/000002_sms_templates.up.sql`

Both have version `000002`. When `golang-migrate` runs, only one of them will be applied.

**Fix:** Rename `000002_sms_templates` to `000003` and renumber subsequent migrations.

### Blocker 2: Dockerfile Go Version Mismatch ⚠️

**Dockerfile:** `FROM golang:1.23.4-alpine`  
**go.mod:** `go 1.26.0`

Docker build will fail or produce incorrect binary.

**Fix:** Update Dockerfile to `golang:1.26.0-alpine` (or the latest available Alpine Go image).

### Blocker 3: Dual Migration System ⚠️

**GORM AutoMigrate** runs via `internal/database/database.go:RunMigrations()`  
**SQL migrations** run via `cmd/migrate/main.go` using `golang-migrate/migrate`

These two systems can conflict:
- `RunMigrations` is called at server startup when `DB_RUN_MIGRATIONS=true` (default)
- SQL migrations are run separately via `go run cmd/migrate/main.go up`

**Risk:** Schema drift, missed columns, conflicting changes.

**Fix:** Default `DB_RUN_MIGRATIONS` to `false` in production. Use SQL migrations as source of truth.

### Blocker 4: `go mod tidy` Fails

`go mod tidy` fails because the `go-common` dependency has an invalid pseudo-version.  
This is a **workspace-level issue** — the local module is not reachable from the module proxy.

This does NOT block builds (workspace resolution works), but it blocks CI pipelines that run `go mod tidy`.

### Blocker 5: No CI/CD Pipeline

No `.github/workflows/ci.yml` or CI configuration exists in `notifier/backend`.

---

## 4. Architecture Assessment

### 4.1 Strengths

1. **Well-structured package layout** — Clean separation of concerns
2. **Repository pattern** — Proper interfaces + implementations
3. **Service layer** — Business logic isolated from transport
4. **gRPC + HTTP** — Dual protocol support
5. **Comprehensive proto definitions** — All major services defined
6. **WebSocket hub** — Real-time notification delivery
7. **Worker pool** — Configurable async processing
8. **Retry logic** — Exponential backoff with configurable limits
9. **Multi-tenancy prepared** — tenant_id columns, RLS, tenants table
10. **Seed data mechanism** — Default templates and settings

### 4.2 Weaknesses

1. **In-memory queue only** — No persistence; notifications lost on restart
2. **Dual migration system** — GORM AutoMigrate + SQL migrations
3. **21 SMS providers** — Over-engineered, most untested in production
4. **Only SMTP for email** — No SendGrid, Mailgun, SES
5. **FCM legacy HTTP API only** — No FCM v1, no APNs
6. **No rate limiting** — Any service can spam any endpoint
7. **PII in logs** — Phone numbers and emails logged in plain text
8. **No unit tests** — Business logic untested
9. **AdminService gRPC not registered** — Defined in proto but not wired
10. **StreamNotifications unimplemented** — Returns `codes.Unimplemented`

### 4.3 Technology Stack

| Component | Technology | Status |
|---|---|---|
| HTTP Framework | Fiber v2 | ✅ Production-grade |
| ORM | GORM v2 | ✅ Production-grade |
| Database | PostgreSQL 16 | ✅ Configured |
| Cache/Queue | Redis 7 | ✅ In compose, not used for queue |
| gRPC | google.golang.org/grpc | ✅ Configured |
| Auth | go-sdk/auth + JWT | ✅ Partial |
| Logging | go-common/logging | ✅ Structured |
| Metrics | Prometheus | ⚠️ Partial (defined, not wired) |
| Tracing | Jaeger + OTel | ✅ Configured |
| Validation | go-playground/validator | ⚠️ Partial |
| I18n | go-common/i18n | ✅ Configured |
| Migration | golang-migrate + GORM | ⚠️ Dual system |

---

## 5. Missing Production Features

| Feature | Status | Priority |
|---|---|---|
| Persistent queue | ❌ Missing | **Critical** |
| Rate limiting | ❌ Missing | **Critical** |
| PII-safe logging | ❌ Missing | **Critical** |
| Unit tests | ❌ Missing | **High** |
| CI/CD pipeline | ❌ Missing | **High** |
| Email providers (SendGrid, etc.) | ❌ Missing | **High** |
| Reminder/scheduler | ❌ Missing | **High** |
| Read/Seen/Click tracking | ⚠️ Partial | **Medium** |
| Provider health checks | ❌ Missing | **Medium** |
| Circuit breaker for providers | ❌ Missing | **Medium** |
| Idempotency keys | ❌ Missing | **Medium** |
| Dead-letter queue | ❌ Missing | **Medium** |
| Webhook events | ❌ Missing | **Medium** |
| Admin APIs | ⚠️ Partial (gRPC only) | **Medium** |
| Health/Readiness endpoints | ⚠️ Partial (health exists) | **Low** |
| OpenAPI spec auto-generation | ❌ Missing | **Low** |

---

## 6. Security Risks

| Risk | Severity | Current State |
|---|---|---|
| PII leaked in logs | **High** | Phone numbers and emails logged in plain text in handler files |
| No rate limiting | **High** | Any client can call send endpoints without throttling |
| Auth disabled fallback | **Medium** | When `AUTH_ENABLED=false`, all endpoints are public |
| Secrets in env files | **Medium** | API keys stored in `.env` files, no vault integration |
| No input sanitization | **Medium** | DTOs have `validate` tags but explicit validation is inconsistent |
| No webhook verification | **Medium** | No webhook signatures (no webhooks yet) |
| Provider keys in DB config | **Medium** | `settings` table stores provider credentials in plaintext |
| SMTP PlainAuth without TLS | **Low** | PlainAuth used in `smtp.PlainAuth` without TLS guarantee |

---

## 7. Migration Risks

| Risk | Severity | Details |
|---|---|---|
| Version conflict | **High** | Two migrations with version `000002` |
| GORM AutoMigrate drift | **Medium** | AutoMigrate can add columns that SQL migrations don't expect |
| Seed data duplication | **Low** | `database.go:seedDefaultData()` creates templates AND migration `000001` has INSERT statements |
| Data loss on down-migration | **Low** | Down migrations exist but may not restore all state |
| Missing down migrations | **Low** | Down migrations for new tables may not exist |

---

## 8. API Gaps

### 8.1 Existing HTTP Endpoints

| Method | Path | Status |
|---|---|---|
| GET | `/api/v1/health/` | ✅ Implemented |
| POST | `/api/v1/notifications` | ✅ Handler + Route |
| POST | `/api/v1/notifications/batch` | ✅ Handler + Route |
| GET | `/api/v1/notifications/user/:userId` | ✅ Handler + Route |
| GET | `/api/v1/notifications/user/:userId/unread` | ✅ Handler + Route |
| PUT | `/api/v1/notifications/:notificationId/read` | ✅ Handler + Route |
| GET | `/api/v1/preferences/user/:userId` | ✅ Handler + Route |
| PUT | `/api/v1/preferences/user/:userId` | ✅ Handler + Route |
| POST | `/api/v1/templates` | ✅ Handler + Route |
| GET | `/api/v1/templates` | ✅ Handler + Route |
| GET | `/api/v1/templates/:templateId` | ✅ Handler + Route |
| PUT | `/api/v1/templates/:templateId` | ✅ Handler + Route |
| DELETE | `/api/v1/templates/:templateId` | ✅ Handler + Route |
| WS | `/ws` | ✅ Handler + Route |
| GET | `/metrics` | ✅ Implemented |

### 8.2 Missing Endpoints

| Endpoint | Purpose | Priority |
|---|---|---|
| `GET /api/v1/notifications/:id` | Get single notification by ID | **High** |
| `GET /api/v1/notifications/user/:userId/unread-count` | Unread count without full list | **High** |
| `POST /api/v1/notifications/user/:userId/read-all` | Mark all notifications as read | **High** |
| `POST /api/v1/reminders` | Create scheduled reminder | **High** |
| `GET /api/v1/reminders/:id` | Get reminder | **High** |
| `PUT /api/v1/reminders/:id` | Update reminder | **High** |
| `DELETE /api/v1/reminders/:id` | Delete reminder | **High** |
| `GET /healthz` | Health check (k8s-style) | **Low** |
| `GET /readyz` | Readiness check | **Low** |

### 8.3 gRPC Gaps

| Method | Status |
|---|---|
| `NotificationService/StreamNotifications` | ❌ Unimplemented (returns `codes.Unimplemented`) |
| `AdminService` (all methods) | ❌ Not registered in server.go |

---

## 9. Provider Gaps

| Provider Type | Status | Missing |
|---|---|---|
| SMS: Kavenegar | ✅ Implemented | Production hardening, tests |
| SMS: Twilio | ✅ Implemented | Production hardening, tests |
| SMS: 19 other providers | ⚠️ Implemented (untested) | Remove or test |
| Email: SMTP | ✅ Implemented | TLS hardening, HTML templates |
| Email: SendGrid | ❌ Missing | Needs implementation |
| Email: Mailgun | ❌ Missing | Needs implementation |
| Push: FCM (legacy) | ✅ Implemented | Deprecated API |
| Push: FCM v1 | ❌ Missing | Needs implementation |
| Push: APNs | ❌ Missing | Needs implementation |
| In-App: Database | ✅ Implemented | Works with WebSocket |

---

## 10. Queue/Retry/Scheduler Gaps

| Feature | Status | Details |
|---|---|---|
| Queue persistence | ❌ Missing | In-memory channel only |
| Dead-letter queue | ❌ Missing | No DLQ state |
| Idempotency | ❌ Missing | No idempotency_key |
| Priority queue | ❌ Missing | Uses `priority` field but no priority-based scheduling |
| Scheduled delivery | ⚠️ Partial | `scheduled_at` field exists, pending processor checks it |
| Cron-based reminders | ❌ Missing | No reminder model or scheduler |
| Graceful shutdown | ⚠️ Partial | Context-based cancel exists, but in-flight jobs may be lost |

---

## 11. Observability Gaps

| Feature | Status | Details |
|---|---|---|
| Structured logging | ✅ Implemented | go-common/logging with zap/zerolog |
| Request ID | ✅ Implemented | go-common middleware |
| Correlation ID | ✅ Implemented | Tracing context propagation |
| Health endpoint | ✅ Implemented | `/api/v1/health/` |
| Readiness endpoint | ❌ Missing | No `/readyz` |
| Prometheus metrics | ⚠️ Partial | Counters/histograms defined but not wired to handlers |
| Provider metrics | ❌ Missing | No per-provider success/failure tracking |
| Queue depth metrics | ❌ Missing | No queue metrics |
| Retry metrics | ❌ Missing | No retry count/duration tracking |

---

## 12. Testing Gaps

| Test Category | Status | Action Needed |
|---|---|---|
| Unit: Notification service | ❌ Missing | Create `internal/service/*_test.go` |
| Unit: Template service | ❌ Missing | Create `internal/service/*_test.go` |
| Unit: Preference service | ❌ Missing | Create `internal/service/*_test.go` |
| Unit: Repository (mock DB) | ❌ Missing | Create `internal/repository/*_test.go` |
| Unit: Provider mock | ✅ Partial | SMS-only |
| Unit: Worker/Retry | ❌ Missing | Create `internal/worker/*_test.go` |
| Unit: WebSocket hub | ❌ Missing | Create `internal/websocket/hub_test.go` |
| Integration: HTTP handlers | ❌ Missing | Create `api/v1/handlers/*_test.go` |
| Integration: Migrations | ❌ Missing | Create `migrations/*_test.go` |
| E2E | ⚠️ Partial | Build-tagged, requires running services |
| Load test | ✅ Implemented | k6 script in `tests/load-test/` |

---

## 13. Deployment Gaps

| Feature | Status | Details |
|---|---|---|
| Dockerfile | ⚠️ Partial | Go version mismatch (1.23.4 vs 1.26.0) |
| Docker Compose dev | ✅ Implemented | PostgreSQL + Redis |
| Docker Compose prod | ✅ Implemented | Full production with resource limits |
| GitHub Actions CI | ❌ Missing | No `.github/workflows/` |
| Migration in CI | ❌ Missing | No migration step in pipeline |
| Health check in deployment | ✅ Partial | Docker health check configured |
| Deployment manifests | ❌ Missing | No K8s/Docker Swarm manifests |

---

## 14. Production-Ready Target Architecture

### Final Package Layout

```
notifier/backend/
├── cmd/                          # Entry points
├── internal/
│   ├── config/                   # Configuration + validation
│   ├── domain/                   # Domain models
│   ├── repository/               # Data access (interfaces + implementations)
│   │   ├── postgres/             # PostgreSQL implementations
│   │   └── redis/                # Redis implementations
│   ├── service/                  # Business logic
│   ├── provider/                 # Provider abstraction
│   │   ├── sms/                  # SMS implementations (Kavenegar, Twilio)
│   │   ├── email/                # Email implementations (SMTP, SendGrid)
│   │   └── push/                 # Push implementations (FCM v1)
│   ├── worker/                   # Queue + retry + dead-letter
│   ├── scheduler/                # Reminder scheduling
│   ├── websocket/                # WebSocket hub
│   ├── middleware/               # HTTP/gRPC middleware
│   └── observability/            # Metrics, tracing
├── transport/                    # API layer
│   ├── http/                     # HTTP handlers + routes
│   └── grpc/                     # gRPC handlers
├── migrations/                   # SQL migrations (single source of truth)
├── api/                          # Proto + OpenAPI specs
├── pkg/client/                   # Public client SDK
├── deploy/                       # Deployment manifests
└── docs/                         # Documentation
```

### Dependency Flow

```
HTTP/gRPC → Service → Repository → PostgreSQL
                  → Provider → External APIs
                  → Worker → Queue (Redis)
                  → WebSocket → Clients
                  → Scheduler → Internal timing
```

---

## 15. Implementation Phases

### Phase 0 — Stabilize Project Structure [CURRENT]

**Objective:** Fix critical structural issues before adding features.

| Task | Files | Complexity |
|---|---|---|
| Fix migration version conflict | Rename `000002_sms_templates` → `000003`, renumber others | **Low** |
| Fix Dockerfile Go version | `Dockerfile`: 1.23.4 → 1.26.0 | **Low** |
| Add `DB_AUTO_MIGRATE` default false | `config/config.go`, `internal/database/database.go` | **Low** |
| Update `.env.example` | Add `DB_AUTO_MIGRATE`, improve documentation | **Low** |
| Add `.gitignore` for binaries | Add `*.exe`, `bin/` to gitignore | **Low** |
| Remove stale binaries | `bin/*.exe`, `migrate.exe`, `__debug_bin*.exe` | **Low** |
| Remove legacy `internal/notification/` | Clean up empty/legacy package | **Low** |

**Validation:**
```bash
go build ./cmd/server
go vet ./...
```

**Rollback:** Simple file renames and config changes; easily reversible.

---

### Phase 1 — Core Notification REST/gRPC API

**Objective:** Complete stable notification API endpoints.

| Task | Files | Complexity |
|---|---|---|
| Add `GET /v1/notifications/:id` | Handler + route | **Low** |
| Add `GET /v1/notifications/user/:userId/unread-count` | Handler + route | **Low** |
| Add `POST /v1/notifications/user/:userId/read-all` | Handler + service method | **Low** |
| Add `POST /v1/notifications/send` alias | Route (delegates to create) | **Low** |
| Register AdminService gRPC | `api/grpc/server.go` | **Medium** |
| Implement gRPC StreamNotifications | `api/grpc/notification_handlers.go` | **Medium** |
| Standardize error response format | Middleware + response helpers | **Medium** |
| Add request validation middleware | `api/middleware/validation.go` | **Low** |

**Validation:**
```bash
go build ./cmd/server
go vet ./...
# Manual: curl http://localhost:9002/api/v1/notifications/:id
```

---

### Phase 2 — Templates & Rendering

**Objective:** Production-safe notification templates.

| Task | Files | Complexity |
|---|---|---|
| Add Go template rendering with `text/template` | `internal/service/template_renderer.go` | **Medium** |
| Add template preview endpoint | `api/v1/handlers/template_handler.go` | **Low** |
| Add template key lookup | `internal/repository/notification_template_repository.go` | **Low** |
| Add locale support (fa/en) | Template model update + rendering | **Medium** |
| Add safe missing variable handling | Template renderer | **Low** |
| Add default system templates seed | `internal/database/database.go` | **Low** |
| Add template validation | `internal/service/template_service.go` | **Low** |

**Validation:**
```bash
go test ./internal/service/ -run Template
go build ./cmd/server
```

---

### Phase 3 — Preferences

**Objective:** Full notification preference control.

| Task | Files | Complexity |
|---|---|---|
| Add preference filtering in send flow | `internal/service/notification_service.go` | **Medium** |
| Add category-level preferences | Model + repository update | **Medium** |
| Add quiet hours enforcement | Service logic | **Medium** |
| Add per-channel preference endpoints | `api/v1/handlers/preference_handler.go` | **Low** |
| Add preference-based channel routing | Service logic | **Medium** |

**Validation:**
```bash
go test ./internal/service/ -run Preference
```

---

### Phase 4 — Provider Abstraction

**Objective:** Unified, reliable provider system.

| Task | Files | Complexity |
|---|---|---|
| Create unified `Provider` interface | `internal/provider/provider.go` | **Medium** |
| Create `ProviderManager` with fallback | `internal/provider/manager.go` | **Medium** |
| Add circuit breaker wrapper | `internal/provider/circuit_breaker.go` | **Medium** |
| Add provider health check | `internal/provider/health.go` | **Medium** |
| Add provider error classification | `internal/provider/errors.go` | **Low** |
| Adapt existing SMS providers to unified interface | `internal/provider/sms_adapter.go` | **Medium** |
| Adapt existing Email provider | `internal/provider/email_adapter.go` | **Medium** |
| Add SendGrid email provider | `internal/platform/email/sendgrid.go` | **Medium** |
| Remove unused SMS providers | Cleanup 16+ provider files | **Low** |

**Validation:**
```bash
go test ./internal/provider/...
```

---

### Phase 5 — Persistent Queue, Retry, Worker

**Objective:** Reliable async notification delivery.

| Task | Files | Complexity |
|---|---|---|
| Add DB-backed pending queue | `internal/worker/db_queue.go` | **High** |
| Add dead-letter queue state | Model + repository | **Medium** |
| Add idempotency key support | Service + repository | **Medium** |
| Add priority queue support | Worker logic | **Medium** |
| Add worker metrics | `internal/worker/metrics.go` | **Low** |
| Improve graceful shutdown | Worker context handling | **Low** |
| Add `WORKER_ENABLED`, `WORKER_CONCURRENCY` config | `config/config.go` | **Low** |

**Validation:**
```bash
go test ./internal/worker/...
```

---

### Phase 6 — Reminders / Scheduled Notifications

**Objective:** Scheduled notification support.

| Task | Files | Complexity |
|---|---|---|
| Create Reminder model | `internal/models/reminder.go` | **Medium** |
| Create Reminder repository | `internal/repository/reminder_repository.go` | **Medium** |
| Create Reminder service | `internal/service/reminder_service.go` | **Medium** |
| Create reminder HTTP handlers | `api/v1/handlers/reminder_handler.go` | **Medium** |
| Create reminder routes | `api/v1/routes/reminder_router.go` | **Low** |
| Implement due-reminder processor | `internal/scheduler/scheduler.go` | **High** |
| Add migration for reminders table | `migrations/000006_add_reminders.up.sql` | **Low** |

**Validation:**
```bash
go test ./internal/service/ -run Reminder
```

---

### Phase 7 — In-App, WebSocket, Read/Seen/Click

**Objective:** Complete user notification inbox.

| Task | Files | Complexity |
|---|---|---|
| Add `seen_at`, `delivered_at`, `clicked_at` fields | Model + migration | **Low** |
| Add unread count endpoint | Handler | **Low** |
| Add read-all endpoint | Handler | **Low** |
| Implement offline notification sync | Service | **Medium** |
| Improve WebSocket reconnection | Hub | **Medium** |
| Add WebSocket auth documentation | README | **Low** |

**Validation:**
```bash
go build ./cmd/server
# Manual: WebSocket connection test
```

---

### Phase 8 — Security, Multi-Tenancy, Rate Limiting

**Objective:** Production security hardening.

| Task | Files | Complexity |
|---|---|---|
| Add rate limiting middleware | `api/middleware/ratelimit.go` | **Medium** |
| Add per-provider rate limiting | `internal/provider/ratelimit.go` | **Medium** |
| Add PII log redaction | Logging helper | **Low** |
| Add request ID validation | Middleware | **Low** |
| Add input validation for all DTOs | Validation middleware | **Medium** |
| Add API key fallback auth | `api/middleware/service_auth.go` | **Medium** |
| Add data retention policy | Service logic | **Medium** |

**Validation:**
```bash
go test ./...
```

---

### Phase 9 — Observability & Admin Operations

**Objective:** Make operations visible.

| Task | Files | Complexity |
|---|---|---|
| Add `/readyz` endpoint | Health handler | **Low** |
| Wire Prometheus metrics to handlers | `api/api.go` + middleware | **Medium** |
| Add provider health metrics | Provider layer | **Medium** |
| Add worker/queue depth metrics | Worker | **Low** |
| Add notification send/fail metrics | Service | **Low** |
| Add admin notification stats endpoint | Handler | **Medium** |
| Register AdminService gRPC | `api/grpc/server.go` | **Low** |

**Validation:**
```bash
# Manual: curl /healthz, /readyz, /metrics
```

---

### Phase 10 — SDK & API Contract

**Objective:** Easy consumption by other services.

| Task | Files | Complexity |
|---|---|---|
| Add missing Go SDK methods | `go-sdk/notifier/client.go` | **Medium** |
| Create C# SDK plan | `csharp-sdk/` | **High** |
| Update OpenAPI spec | `docs/swagger.json` | **Low** |
| Add integration examples to README | `README.md` | **Low** |

**Validation:** Documentation review.

---

### Phase 11 — CI/CD, Docker, Deployment

**Objective:** Deployable service.

| Task | Files | Complexity |
|---|---|---|
| Create GitHub Actions CI | `.github/workflows/ci.yml` | **Medium** |
| Add Docker build to CI | CI config | **Medium** |
| Add migration step to CI | CI config | **Medium** |
| Create production deployment docs | `docs/deployment.md` | **Low** |
| Add production readiness checklist | README | **Low** |

**Validation:**
```bash
docker build -t minisource-notifier .
```

---

### Phase 12 — Tests & Hardening

**Objective:** Comprehensive test coverage.

| Task | Files | Complexity |
|---|---|---|
| Add notification service unit tests | `internal/service/*_test.go` | **Medium** |
| Add HTTP handler tests | `api/v1/handlers/*_test.go` | **Medium** |
| Add repository tests | `internal/repository/*_test.go` | **High** |
| Add worker/retry tests | `internal/worker/*_test.go` | **Medium** |
| Add migration tests | `migrations/*_test.go` | **Medium** |
| Run race detector | `go test -race ./...` | **Low** |

**Validation:**
```bash
go test -race ./...
```

---

## 16. Acceptance Criteria Per Phase

| Phase | Key Acceptance Criteria |
|---|---|
| Phase 0 | `go build ./cmd/server` passes, migrations renumbered, Dockerfile Go version fixed |
| Phase 1 | All core notification endpoints respond correctly, error format consistent |
| Phase 2 | Template CRUD works, rendering works with variables, fa/en supported |
| Phase 3 | Preferences stored, send flow respects preferences, tests pass |
| Phase 4 | Provider abstraction works, mock provider works, real providers disabled by default |
| Phase 5 | Queued notifications survive restart, retry works, idempotency prevents duplicates |
| Phase 6 | Reminder CRUD works, due reminders picked up, cancel works |
| Phase 7 | In-app inbox works, unread count accurate, seen/read/click distinct |
| Phase 8 | Unauthorized requests rejected, rate limits work, PII not in logs |
| Phase 9 | Health/readiness work, metrics exposed, admin endpoints functional |
| Phase 10 | OpenAPI updated, SDK plan documented, README has examples |
| Phase 11 | CI passes, Docker build works, deployment docs exist |
| Phase 12 | >50% code coverage, race detector clean, all critical paths tested |

---

## 17. Validation Commands

```bash
# After each phase
cd notifier/backend

# Build checks
go build ./cmd/server
go build ./cmd/migrate

# Code quality
go vet ./...

# Tests
go test ./... 2>&1

# If available
go test -race ./...

# Docker
docker build -t minisource-notifier .

# Health checks (when server runs)
curl http://localhost:9002/api/v1/health/
```

---

## 18. Rollback/Risk Notes

| Phase | Risk | Rollback |
|---|---|---|
| Phase 0: Migration rename | DB migration state inconsistency | Restore original filenames, `migrate force` to fix version |
| Phase 0: Dockerfile change | None (build-time only) | Revert version |
| Phase 0: AutoMigrate disable | Dev environments may need AutoMigrate | Set `DB_AUTO_MIGRATE=true` |
| Phase 1: New endpoints | No risk (additive) | Remove routes |
| Phase 2: Template rendering | Template syntax errors | Fix template, no DB changes |
| Phase 5: DB queue | New table/index creation | Down migration |
| Phase 6: Reminders | New table | Down migration |
| Phase 8: Rate limiting | May break legitimate high-volume senders | Disable via config |

---

## 19. Final Production-Readiness Checklist

### Build & Runtime
- [x] Project builds (`go build ./cmd/server` ✅)
- [ ] Tests pass (`go test ./...`)
- [ ] Server starts locally
- [ ] Docker build works
- [ ] Migrations run from clean DB
- [ ] Health and readiness endpoints work

### Core Features
- [ ] Send notification works
- [ ] Batch send works
- [ ] SMS works with mock and real provider
- [ ] Email works with mock and SMTP
- [ ] In-app notification persisted
- [ ] User notification list works
- [ ] Unread count works
- [ ] Mark read/read-all works
- [ ] Templates work
- [ ] Preferences work
- [ ] Reminders work
- [ ] Worker retry works
- [ ] Failed/dead-letter state exists

### Reliability
- [ ] Queue is persistent (DB-backed)
- [ ] Retry is implemented
- [ ] Idempotency exists
- [ ] No queued notification lost on restart
- [ ] Provider errors tracked
- [ ] Worker shuts down gracefully

### Security
- [ ] Service-to-service auth exists
- [ ] Dangerous endpoints protected
- [ ] Rate limiting exists
- [ ] PII redacted in logs
- [ ] Tenant/project isolation enforced
- [ ] Secrets come from env

### Observability
- [ ] Structured logs exist
- [ ] Request/correlation ID exists
- [ ] Metrics exposed
- [ ] Delivery status queryable
- [ ] Provider health visible

### Documentation
- [ ] README updated
- [ ] API docs exist
- [ ] Deployment docs exist
- [ ] Provider docs exist
- [ ] Template docs exist
- [ ] Reminder docs exist

### Reusability
- [ ] No DiviPay-specific hardcoded logic
- [ ] Generic template keys/categories supported
- [ ] Service usable by auth, DiviPay, and future projects
- [ ] API contracts stable enough for SDKs

---

## Appendix: Dependency Map

```
notifier
├── go-common (local workspace)
│   ├── logging/
│   ├── http/middleware/
│   ├── grpc/
│   ├── response/
│   ├── i18n/
│   ├── common/
│   ├── metrics/
│   └── testing/
├── go-sdk (local workspace)
│   ├── auth/
│   ├── notifier/
│   └── log/
├── scheduler (external - optional)
└── log (external - optional)
```

---

> **End of Roadmap**  
> *This document reflects the actual state of the repository as of June 22, 2026.*
