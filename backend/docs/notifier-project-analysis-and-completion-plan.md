# Notifier Service — Repository Analysis & Backend Completion Plan

> **Date:** June 22, 2026  
> **Repository:** `github.com/minisource/notifier`  
> **Module:** `github.com/minisource/notifier`  
> **Language:** Go 1.26.0  
> **Framework:** Fiber v2  
> **Database:** PostgreSQL 16  
> **Cache:** Redis 7  
> **Message:** In-process channel (no external queue)

---

## Part 1 — Repository Overview

### 1.1 Repository Structure

```
notifier/backend/
├── cmd/
│   ├── server/            # Main entry point
│   │   └── main.go
│   ├── migrate/            # Migration CLI tool
│   │   └── main.go
│   ├── initializer/        # Application initializers (config, db, services)
│   │   ├── config.go
│   │   ├── database.go
│   │   ├── repositories.go
│   │   ├── servers.go
│   │   └── services.go
│   └── sms_probe/          # SMS test probe tool
│       └── main.go
├── api/
│   ├── api.go              # Fiber app setup & route registration
│   ├── middleware/
│   │   └── service_auth.go  # Service-to-service auth middleware
│   ├── grpc/
│   │   ├── server.go                    # gRPC server setup
│   │   ├── notification_handlers.go     # gRPC notification handlers
│   │   └── template_preference_handlers.go # gRPC template/preference handlers
│   └── v1/
│       ├── dto/
│       │   ├── notification_dto.go
│       │   └── sms_dto.go
│       ├── handlers/
│       │   ├── notification_handler.go
│       │   ├── preference_handler.go
│       │   └── template_handler.go
│       └── routes/
│           ├── health.go      # Exists
│           ├── notifications.go  # MISSING
│           ├── preferences.go    # MISSING
│           ├── sms.go           # MISSING
│           ├── templates.go     # MISSING
│           └── websocket.go     # MISSING
├── config/
│   └── config.go            # Environment-based config loading
├── constants/
│   └── constants.go         # Empty constants file
├── docs/
│   ├── docs.go              # Swagger docs
│   ├── I18N_USAGE.md
│   ├── Notifier.postman_collection.json
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── database/
│   │   └── database.go      # DB connection, GORM auto-migrate, seed data
│   ├── models/
│   │   ├── notification.go
│   │   ├── notification_log.go
│   │   ├── notification_preference.go
│   │   ├── notification_template.go
│   │   ├── service_client.go
│   │   ├── setting.go
│   │   └── sms_template.go
│   ├── notification/         # Legacy/empty package (moved to service)
│   │   ├── service.go
│   │   ├── sms_service.go
│   │   ├── email_service.go
│   │   └── push.go
│   ├── platform/
│   │   ├── email/
│   │   │   ├── email.go      # SMTP client implementation
│   │   │   └── smtp_test.go.old
│   │   ├── push/
│   │   │   └── push.go       # FCM client, Mock client
│   │   └── sms/
│   │       ├── sms.go        # SMS client factory + config parsing
│   │       ├── kavenegar_test.go
│   │       ├── mock_provider_test.go
│   │       └── platforms/
│   │           ├── base.go       # SmsClient interface
│   │           ├── kavenegar.go  # Kavenegar SMS
│   │           ├── twilio.go     # Twilio SMS
│   │           ├── mock.go       # Mock provider
│   │           ├── aliyun.go     # Aliyun SMS
│   │           ├── aws.go        # AWS SMS
│   │           ├── azure.go      # Azure SMS
│   │           ├── baidubce.go   # Baidu SMS
│   │           ├── gccpay.go     # GCCPay SMS
│   │           ├── huawei.go     # Huawei SMS
│   │           ├── huyi.go       # Huyi SMS
│   │           ├── infobip.go    # Infobip SMS
│   │           ├── msg91.go      # MSG91 SMS
│   │           ├── netgsm.go     # NetGSM SMS
│   │           ├── oson.go       # Oson SMS
│   │           ├── smsbao.go     # SMSBao
│   │           ├── submail.go    # Submail SMS
│   │           ├── tencent.go    # Tencent SMS
│   │           ├── ucloud.go     # UCloud SMS
│   │           ├── unisms.go     # Uni SMS
│   │           └── volcengine.go # Volcengine SMS
│   ├── repository/
│   │   ├── notification_repository.go
│   │   ├── notification_log_repository.go
│   │   ├── notification_preference_repository.go
│   │   ├── notification_template_repository.go
│   │   ├── setting_repository.go
│   │   └── sms_template_repository.go
│   ├── service/
│   │   ├── notification_service.go  # Core notification service
│   │   ├── preference_service.go    # Preference management
│   │   ├── template_service.go      # Template management
│   │   ├── admin_service.go         # Admin operations
│   │   ├── errors.go                # Error codes & constructors
│   │   └── handlers.go              # SMS/Email/Push handler adapters
│   ├── websocket/
│   │   └── hub.go           # WebSocket hub implementation
│   └── worker/
│       ├── notification_worker.go   # Worker pool + retry logic
│       └── errors.go                # Worker-specific errors
├── migrations/
│   ├── 000001_initial_schema.up.sql
│   ├── 000001_initial_schema.down.sql
│   ├── 000002_add_multi_tenancy.up.sql
│   ├── 000002_add_multi_tenancy.down.sql
│   ├── 000002_sms_templates.up.sql  (exists as file, not yet read)
│   ├── 000002_sms_templates.down.sql
│   ├── 000003_add_audit_security.up.sql
│   ├── 000003_add_audit_security.down.sql
│   ├── 000004_optimize_database.up.sql
│   └── 000004_optimize_database.down.sql
├── pkg/
│   ├── metrics/
│   │   ├── counters.go
│   │   └── histograms.go
│   ├── service_errors/
│   │   ├── error_code.go
│   │   └── service_error.go
│   └── tracing/
│       └── tracer.go
├── proto/
│   └── notifier/v1/
│       └── notifier.proto
├── scripts/
│   ├── create_sms_templates.go.backup
│   ├── ensure_sms_templates.go.backup
│   ├── init-db.sql
│   └── setup.ps1
├── tests/
│   ├── e2e/
│   │   ├── notifier_api_test.go
│   │   ├── notifier_batch_test.go
│   │   ├── notifier_crud_test.go
│   │   ├── notifier_templates_test.go
│   │   ├── notifier_websocket_test.go
│   │   └── swagger_routes_test.go
│   └── load-test/
│       ├── notifier-send.js     # k6 script
│       └── README.md
├── .env.example
├── docker-compose.yml
├── docker-compose.dev.yml
├── docker-compose.prod.yml
├── Dockerfile
├── go.mod
├── Makefile
├── README.md
├── SECURITY.md
├── Taskfile.yml
├── migrate.exe                # Pre-built migration binary
├── package.json               # Empty
└── test_proto.go              # Proto test helper
```

### 1.2 Main Language & Framework

- **Language:** Go 1.26.0 (module: `go 1.26.0`)
- **Framework:** [Fiber v2](https://github.com/gofiber/fiber) (`github.com/gofiber/fiber/v2 v2.52.11`)
- **ORM:** GORM v2 (`gorm.io/gorm v1.25.12`)
- **gRPC:** `google.golang.org/grpc v1.78.0`
- **Logging:** `github.com/minisource/go-common/logging` (supports zap & zerolog)
- **Validation:** `github.com/go-playground/validator/v10` (in go-common)
- **Auth:** `github.com/minisource/go-sdk/auth`
- **Prometheus:** `github.com/prometheus/client_golang v1.20.5`
- **OpenTelemetry:** `go.opentelemetry.io/otel v1.40.0`
- **Swagger:** `github.com/gofiber/swagger`
- **Tracing:** Jaeger exporter (`go.opentelemetry.io/otel/exporters/jaeger`)
- **I18n:** In-house `github.com/minisource/go-common/i18n`
- **Rate Limiting:** `github.com/didip/tollbooth/v7` (in go-common)

### 1.3 Go Version

- **Module:** `go 1.26.0` (very latest)
- **Dockerfile:** Uses `golang:1.23.4-alpine` (outdated vs module requirement)

### 1.4 Entry Points

| Entry Point | Path | Purpose |
|---|---|---|
| Main server | `cmd/server/main.go` | Starts HTTP (Fiber), gRPC, WebSocket, workers |
| Migration CLI | `cmd/migrate/main.go` | Standalone migration runner |
| SMS probe | `cmd/sms_probe/main.go` | Debug tool to test SMS via SDK |

### 1.5 Module Name

```
github.com/minisource/notifier
```

### 1.6 Package Layout

```
cmd/            → Entry points
api/            → Transport layer (HTTP handlers, gRPC, middleware, routes)
config/         → Configuration loading
constants/      → Global constants (currently empty)
internal/
  database/     → DB connection, auto-migrate, seed data
  models/       → Domain models (GORM structs)
  notification/ → Legacy/empty package
  platform/     → Provider integrations (SMS/Email/Push)
  repository/   → Data access layer (interfaces + implementations)
  service/      → Business logic layer
  websocket/    → WebSocket hub
  worker/       → Async worker pool & retry
migrations/     → SQL migration files
pkg/
  metrics/      → Prometheus metric definitions
  service_errors/ → Service error types
  tracing/       → OpenTelemetry/Jaeger setup
proto/          → Protobuf definitions
tests/
  e2e/          → End-to-end tests
  load-test/    → k6 load test scripts
```

### 1.7 Configuration Files

| File | Status |
|---|---|
| `config/config.go` | **Implemented** — env-based config with defaults |
| `.env.example` | **Implemented** — full reference |
| `.env` | **Exists** (gitignored) |
| `.env.dev` | **Exists** |
| `.env.prod` | **Exists** |

Note: The README mentions YAML config files (`config-development.yml`, etc.) but these don't exist. Config is purely env-based.

### 1.8 Docker Files

| File | Status |
|---|---|
| `Dockerfile` | **Implemented** — multi-stage alpine build |
| `docker-compose.yml` | **Implemented** — alias to prod |
| `docker-compose.dev.yml` | **Implemented** — PostgreSQL, Redis, optional MailHog, Adminer |
| `docker-compose.prod.yml` | **Implemented** — full production with health checks, resource limits, networks |

**Issues:**
- Dockerfile uses `golang:1.23.4-alpine` but module requires `go 1.26.0`
- Dockerfile copies `./config` directory but config is env-based, not YAML
- Dev compose has commented-out network sections

### 1.9 CI/CD Files

| File | Status |
|---|---|
| `.github/` | **Unknown** — not inspected |
| `Makefile` | **Implemented** — comprehensive build/test/migrate/docker commands |
| `Taskfile.yml` | **Implemented** — task runner alternative |

### 1.10 Existing Documentation

| File | Status |
|---|---|
| `README.md` | **Implemented** — good overview, API examples, config examples |
| `SECURITY.md` | **Implemented** — generic security policy |
| `docs/swagger.json` | **Implemented** — Swagger specification |
| `docs/swagger.yaml` | **Implemented** — Swagger specification |
| `docs/Notifier.postman_collection.json` | **Implemented** — Postman collection |
| `docs/I18N_USAGE.md` | **Implemented** — i18n usage guide |

### 1.11 Existing Tests

| File | Type | Status |
|---|---|---|
| `tests/e2e/notifier_api_test.go` | E2E | **Partial** — uses `go-common/testing/e2e`, requires running services + auth |
| `tests/e2e/notifier_batch_test.go` | E2E | **Unknown content** |
| `tests/e2e/notifier_crud_test.go` | E2E | **Unknown content** |
| `tests/e2e/notifier_templates_test.go` | E2E | **Unknown content** |
| `tests/e2e/notifier_websocket_test.go` | E2E | **Unknown content** |
| `tests/e2e/swagger_routes_test.go` | E2E | **Unknown content** |
| `tests/load-test/notifier-send.js` | k6 | **Implemented** — realistic load test |
| `internal/platform/sms/kavenegar_test.go` | Unit | **Partial** |
| `internal/platform/sms/mock_provider_test.go` | Unit | **Partial** |
| `internal/platform/email/smtp_test.go.old` | Unit | **Old/outdated** (`.old` suffix) |

**Assessment:** No pure unit tests for the service layer. E2E tests exist but require auth service + running infrastructure. No repository tests, no handler tests.

### 1.12 Existing Migrations

| Migration | Status |
|---|---|
| `000001_initial_schema` | **Implemented** — creates all core tables |
| `000002_add_multi_tenancy` | **Implemented** — adds tenant_id columns, tenants table, row-level security |
| `000002_sms_templates` | **Exists** (file exists) |
| `000003_add_audit_security` | **Implemented** — audit_logs table, RLS policies, triggers |
| `000004_optimize_database` | **Implemented** — advanced indexes, pg_trgm, autovacuum config |

**Issues:**
- `000002_sms_templates` migration may conflict with `000002_add_multi_tenancy` (same version number)
- Migration runner uses `golang-migrate/migrate` for SQL files, but GORM `AutoMigrate` also runs at startup — **dual migration system** (conflict risk)
- The GORM `AutoMigrate` in `internal/database/database.go` runs independently of SQL migrations
- Status: `READ`

### 1.13 Existing API Definitions

| Location | Protocol | Status |
|---|---|---|
| `api/api.go` | HTTP REST | **Partial** — routes registered but route files missing |
| `api/grpc/server.go` | gRPC | **Implemented** — full server setup with auth interceptors |
| `proto/notifier/v1/notifier.proto` | Protobuf | **Implemented** — comprehensive service definitions |
| `go-sdk/notifier/client.go` | Go SDK (gRPC) | **Implemented** — full client with all methods |
| `docs/swagger.json` | OpenAPI | **Implemented** — auto-generated |

---

## Part 2 — Current Architecture Analysis

### 2.1 Application Bootstrap

**Status: Implemented**
- File: `cmd/server/main.go`
- Flow: `InitConfig()` → `InitLogger()` → `InitMetrics()` → `InitTracing()` → `InitTranslator()` → `InitDatabase()` → `InitRepositories()` → `InitServices()` → `InitGRPCServer()` → `InitServer()` (Fiber)

### 2.2 Dependency Injection Style

**Status: Implemented**  
**Pattern:** Manual constructor injection  
**Files:** `cmd/initializer/*.go`
- `InitRepositories()` creates all repository structs as a `Repositories` struct
- `InitServices()` creates all services, passing repos explicitly
- The worker is created, then notification service is re-created with the worker reference (two-phase init)

### 2.3 HTTP Server Setup

**Status: Implemented**
- File: `api/api.go`
- Framework: Fiber v2
- Port: `SERVER_INTERNAL_PORT` (default: `9002`)
- JSON encoder/decoder: Standard `json.Marshal`/`json.Unmarshal`

### 2.4 gRPC Server Setup

**Status: Implemented**
- Files: `api/grpc/server.go`, `api/grpc/notification_handlers.go`, `api/grpc/template_preference_handlers.go`
- Port: `GRPC_PORT` (default: `9003`)
- Auth: Uses `go-common/grpc` interceptors with JWT scope validation
- Registers 3 services: `NotificationService`, `TemplateService`, `PreferenceService`
- **Note:** `StreamNotifications` returns `codes.Unimplemented`

### 2.5 Routing

**Status: Partial**

**Implemented routes (in `api/api.go`):**
```
/api/v1/health/                          → GET    (health handler)
/api/v1/sms                              → ???   (SMS handler)
(if auth enabled)
  /api/v1/notifications                  → ???   (jwt auth)
  /api/v1/preferences                    → ???   (jwt auth)
  /api/v1/templates                      → ???   (jwt auth + admin roles)
  /api/v1/service/notifications          → ???   (service auth)
  /ws                                    → WebSocket (token auth)
(if auth disabled)
  /api/v1/notifications                  → ???   (public)
  /api/v1/preferences                    → ???
  /api/v1/templates                      → ???
```

**CRITICAL ISSUE:** Route files are **missing**:
- `api/v1/routes/notifications.go` — **DOES NOT EXIST**
- `api/v1/routes/preferences.go` — **DOES NOT EXIST**
- `api/v1/routes/templates.go` — **DOES NOT EXIST**
- `api/v1/routes/sms.go` — **DOES NOT EXIST**
- `api/v1/routes/websocket.go` — **DOES NOT EXIST**

Only `health.go` exists. The code in `api/api.go` references these route files, so the project will **not compile** without them.

However, handler files DO exist:
- `api/v1/handlers/notification_handler.go` — **EXISTS** (with Create, Batch, Get, Unread, MarkAsRead)
- `api/v1/handlers/preference_handler.go` — **EXISTS** (GetUser, Update)
- `api/v1/handlers/template_handler.go` — **EXISTS** (CRUD)

### 2.6 Middleware

**Status: Implemented**  
**Files:**
- `api/api.go` — order: Logger, SecurityHeaders, RequestValidation, Prometheus, Tracing, CORS, Recover, Tenant, WebSocket upgrade
- `api/middleware/service_auth.go` — wraps `go-common/http/middleware.RemoteServiceAuthMiddleware`

### 2.7 Config Loading

**Status: Implemented**  
**File:** `config/config.go`
- Uses `godotenv` to load `.env`
- All config fields have defaults
- Singleton pattern (`sync.Once`)
- Supports `PORT` env override for external port

### 2.8 Logging

**Status: Implemented**
- Files: Uses `go-common/logging`
- Supports zap and zerolog
- JSON encoding, configurable level
- Middleware: `middleware.DefaultStructuredLogger`
- Log categories: General, Postgres, Internal, Validation

### 2.9 Error Handling

**Status: Partial**
- Files: `internal/service/errors.go`, `pkg/service_errors/*`
- Service errors have codes but no i18n message integration
- Uses `go-common/response` for HTTP responses
- No centralized error handler at the HTTP layer

### 2.10 Validation

**Status: Partial**
- Uses `go-common/http/middleware.RequestValidation` middleware
- DTO structs have `validate` tags but no explicit validation middleware
- `go-playground/validator/v10` is in go.sum but not explicitly used in handlers

### 2.11 Database Access

**Status: Implemented**
- **ORM:** GORM v2
- **Driver:** `gorm.io/driver/postgres`
- **Connection pool:** Configurable MaxIdleConns, MaxOpenConns, ConnMaxLifetime
- **Dual migration system:** GORM `AutoMigrate` + `golang-migrate/migrate` SQL files

### 2.12 Repository Pattern

**Status: Implemented**

Repositories (all with interface + struct pattern):
| Repository | File |
|---|---|
| `NotificationRepository` | `internal/repository/notification_repository.go` |
| `NotificationLogRepository` | `internal/repository/notification_log_repository.go` |
| `NotificationPreferenceRepository` | `internal/repository/notification_preference_repository.go` |
| `NotificationTemplateRepository` | `internal/repository/notification_template_repository.go` |
| `SettingRepository` | `internal/repository/setting_repository.go` |
| `SMSTemplateRepository` | `internal/repository/sms_template_repository.go` |

### 2.13 Service/Usecase Layer

**Status: Implemented**

| Service | File | Status |
|---|---|---|
| `NotificationService` | `internal/service/notification_service.go` | **Implemented** — Create, CreateSync, Batch, Get, GetUnread, MarkAsRead |
| `TemplateService` | `internal/service/template_service.go` | **Implemented** — CRUD |
| `PreferenceService` | `internal/service/preference_service.go` | **Implemented** — Get, Update, BatchUpdate, Reset |
| `AdminService` | `internal/service/admin_service.go` | **Partial** — stats/delivery stats have structure but no DB queries |

### 2.14 Provider Integrations

**Status: Implemented (SMS), Partial (Email/Push)**

| Provider Type | Status | Files |
|---|---|---|
| SMS | **Over-engineered** — 16+ providers | `internal/platform/sms/*`, `internal/platform/sms/platforms/*` |
| Email | **Partial** — SMTP only | `internal/platform/email/email.go` |
| Push | **Partial** — FCM legacy only | `internal/platform/push/push.go` |

### 2.15 Queue/Event System

**Status: Implemented (in-process only)**
- File: `internal/worker/notification_worker.go`
- In-memory channel-based queue (`chan *NotificationJob`)
- Configurable: `NumWorkers`, `QueueSize`
- No external message broker
- No persistence — messages are lost on restart
- **No dead-letter queue**
- **No delayed/scheduled message support**

### 2.16 Scheduler/Worker System

**Status: Implemented**
- Worker pool with goroutine workers
- Periodic retry processor (30s interval)
- Exponential backoff retry
- Sync send capability

### 2.17 Health Checks

**Status: Implemented**
- Endpoint: `/api/v1/health/` (GET)
- File: `api/v1/routes/health.go`, handler via `NewHealthHandler()`
- Ready endpoint: not yet implemented (referenced as `/ready` in tenant config skip paths)

### 2.18 Metrics/Tracing

**Status: Partial**
- **Metrics:** Prometheus counters and histograms registered but not wired to handlers
- **Tracing:** Jaeger exporter implemented, middleware uses `go-common/http/middleware.Tracing`
- `/metrics` endpoint registered with `promhttp.Handler()`

---

## Part 3 — Existing Notification Features

### 3.1 Email

| Aspect | Status |
|---|---|
| Current status | **Partial** — SMTP implementation exists |
| Code files | `internal/platform/email/email.go` |
| Missing pieces | SendGrid, Mailgun, SES, Resend — no implementations |
| Risks | SMTP PlainAuth is insecure without TLS |
| Tasks | Add more providers, implement HTML template rendering, add attachments |

### 3.2 SMS

| Aspect | Status |
|---|---|
| Current status | **Over-engineered** — 16+ providers implemented |
| Code files | `internal/platform/sms/*`, `internal/platform/sms/platforms/*` |
| Key providers | Kavenegar (Iran), Twilio, Tencent, Huawei, Infobip, and many more |
| Risks | Many providers untested in production; unstructured error handling |
| Tasks | Consolidate to 2-3 strategic providers, add circuit breaker, test Kavenegar + Twilio |

### 3.3 Push Notifications

| Aspect | Status |
|---|---|
| Current status | **Partial** — FCM legacy HTTP API only |
| Code files | `internal/platform/push/push.go` |
| Missing pieces | FCM v1 HTTP API, APNs, OneSignal, Expo |
| Tasks | Add FCM v1, add APNs, add provider fallback |

### 3.4 In-App Notification

| Aspect | Status |
|---|---|
| Current status | **Implemented** — store in DB + WebSocket delivery |
| Code files | `internal/models/notification.go`, `internal/websocket/hub.go` |
| Missing pieces | Read/seen status tracking exists (read_at field), but no seen tracking, no pull-to-refresh endpoint for offline users |
| Tasks | Add seen_at field, add notification count endpoint, ensure WebSocket reconnection works |

### 3.5 Templates

| Aspect | Status |
|---|---|
| Current status | **Implemented** |
| Code files | `internal/models/notification_template.go`, `internal/repository/notification_template_repository.go` |
| Missing pieces | Template variable rendering with Go templates (only raw storage), provider-specific template mapping |
| Tasks | Add Go template rendering with `text/template`, add preview/render endpoint, add versioning |

### 3.6 Notification Preferences

| Aspect | Status |
|---|---|
| Current status | **Implemented** |
| Code files | `internal/models/notification_preference.go`, `internal/repository/notification_preference_repository.go` |
| Missing pieces | Preference-checking in notification creation (exists but basic), category-level preferences partially done |
| Tasks | Add quiet hours enforcement, rate limiting per user per channel |

### 3.7 Notification Categories

| Aspect | Status |
|---|---|
| Current status | **Missing** — no category entity |
| Code files | None |
| Missing pieces | No notification category model, no grouping |
| Tasks | Add NotificationCategory model, add category_id to notifications |

### 3.8 Notification Events

| Aspect | Status |
|---|---|
| Current status | **Partial** — log audit trail exists |
| Code files | `internal/models/notification_log.go` |
| Missing pieces | No webhook callbacks for notification events (sent/failed/read) |
| Tasks | Add webhook event system |

### 3.9 Delivery Attempts

| Aspect | Status |
|---|---|
| Current status | **Implemented** |
| Code files | `internal/models/notification_log.go`, `internal/worker/notification_worker.go` |
| Notes | Each attempt creates a NotificationLog entry |

### 3.10 Retry

| Aspect | Status |
|---|---|
| Current status | **Implemented** |
| Code files | `internal/worker/notification_worker.go` |
| Strategy | Exponential backoff: `baseDelay * 2^retryCount` (capped at maxDelay) |
| Defaults | MaxRetries=3, BaseDelay=5s, MaxDelay=300s |
| Retry processor | Polls every 30s for notifications in `retrying` status |

### 3.11 Rate Limiting

| Aspect | Status |
|---|---|
| Current status | **Missing** — no rate limiting implemented |
| Code files | None specific to notifier |
| Note | `go-common/limiter` has IP-based limiter; `go-common/http/middleware` has rate limiter middleware |
| Tasks | Add per-provider rate limiting, per-tenant rate limiting, per-user rate limiting |

### 3.12 Scheduled Notifications

| Aspect | Status |
|---|---|
| Current status | **Partial** — scheduled_at field exists, pending processor checks it |
| Code files | `internal/repository/notification_repository.go` (GetPendingNotifications) |
| Missing pieces | No cron-style scheduling, no recurring reminders, no integration with scheduler service |
| Tasks | Add cron expressions, integrate with Minisource scheduler or own scheduling |

### 3.13 Webhooks/Callbacks

| Aspect | Status |
|---|---|
| Current status | **Missing** |
| Code files | None |
| Tasks | Add webhook event system, delivery status callbacks, read receipts |

### 3.14 Read/Seen Status

| Aspect | Status |
|---|---|
| Current status | **Partial** — read_at field exists |
| Code files | `internal/models/notification.go`, `internal/repository/notification_repository.go` (MarkAsRead) |
| Missing pieces | No "seen" vs "read" distinction, no bulk read-all, no read receipts for senders |
| Tasks | Add seen_at, implement read receipts, add read-all endpoint |

---

## Part 4 — API Surface Analysis

### 4.1 Current HTTP Endpoints

Based on `api/api.go` and handler files:

| Method | Path | Auth | Handler | Status |
|---|---|---|---|---|
| GET | `/api/v1/health/` | Public | Health handler | **Implemented** |
| ??? | `/api/v1/sms` | Public | SMS handler | **Missing route file** |
| POST | `/api/v1/notifications` | JWT/Service | CreateNotification | **Handler exists, route file missing** |
| POST | `/api/v1/notifications/batch` | JWT/Service | CreateBatchNotifications | **Handler exists, route file missing** |
| GET | `/api/v1/notifications/user/:userId` | JWT/Service | GetUserNotifications | **Handler exists, route file missing** |
| GET | `/api/v1/notifications/user/:userId/unread` | JWT/Service | GetUnreadNotifications | **Handler exists, route file missing** |
| PUT | `/api/v1/notifications/:notificationId/read` | JWT/Service | MarkAsRead | **Handler exists, route file missing** |
| GET | `/api/v1/preferences/user/:userId` | JWT/Service | GetUserPreferences | **Handler exists, route file missing** |
| PUT | `/api/v1/preferences/user/:userId` | JWT/Service | UpdatePreference | **Handler exists, route file missing** |
| POST | `/api/v1/templates` | JWT + Admin | CreateTemplate | **Handler exists, route file missing** |
| GET | `/api/v1/templates` | JWT + Admin | GetAllTemplates | **Handler exists, route file missing** |
| GET | `/api/v1/templates/:templateId` | JWT + Admin | GetTemplate | **Handler exists, route file missing** |
| PUT | `/api/v1/templates/:templateId` | JWT + Admin | UpdateTemplate | **Handler exists, route file missing** |
| DELETE | `/api/v1/templates/:templateId` | JWT + Admin | DeleteTemplate | **Handler exists, route file missing** |
| WS | `/ws` | Token | WebSocket hub | **Handler exists, route file missing** |
| GET | `/metrics` | Public | Prometheus | **Implemented** |
| GET | `/swagger/*` | Public | Swagger | **Implemented** |

### 4.2 gRPC Services (from proto)

**NotificationService:**
| Method | Request | Response |
|---|---|---|
| `CreateNotification` | `CreateNotificationRequest` | `CreateNotificationResponse` |
| `CreateBatchNotifications` | `CreateBatchNotificationsRequest` | `CreateBatchNotificationsResponse` |
| `GetUserNotifications` | `GetUserNotificationsRequest` | `GetUserNotificationsResponse` |
| `GetUnreadNotifications` | `GetUnreadNotificationsRequest` | `GetUnreadNotificationsResponse` |
| `MarkAsRead` | `MarkAsReadRequest` | `MarkAsReadResponse` |
| `GetNotification` | `GetNotificationRequest` | `GetNotificationResponse` |
| `SendSMS` | `SendSMSRequest` | `SendSMSResponse` |
| `SendEmail` | `SendEmailRequest` | `SendEmailResponse` |
| `StreamNotifications` | `StreamNotificationsRequest` | `stream Notification` (unimplemented) |

**TemplateService:**
| Method | Request | Response |
|---|---|---|
| `CreateTemplate` | `CreateTemplateRequest` | `CreateTemplateResponse` |
| `GetTemplate` | `GetTemplateRequest` | `GetTemplateResponse` |
| `GetAllTemplates` | `GetAllTemplatesRequest` | `GetAllTemplatesResponse` |
| `UpdateTemplate` | `UpdateTemplateRequest` | `UpdateTemplateResponse` |
| `DeleteTemplate` | `DeleteTemplateRequest` | `DeleteTemplateResponse` |

**PreferenceService:**
| Method | Request | Response |
|---|---|---|
| `GetUserPreferences` | `GetUserPreferencesRequest` | `GetUserPreferencesResponse` |
| `UpdatePreference` | `UpdatePreferenceRequest` | `UpdatePreferenceResponse` |

**AdminService (gRPC only, defined in proto but NOT registered in server.go):**
- `GetNotificationLogs`, `GetStatistics`, `GetDeliveryStats`
- `RetryFailedNotifications`, `RetryNotification`, `CancelNotification`
- `BulkDeleteNotifications`
- `GetServiceClient`, `ListServiceClients`, `CreateServiceClient`, `UpdateServiceClient`, `DeleteServiceClient`
- `GetFailedNotifications`

### 4.3 Missing Endpoints (Recommended)

**Suggested REST API design (to be added in route files):**
```
POST   /api/v1/notifications/send            → Implements existing handler
POST   /api/v1/notifications/bulk            → Already has handler
GET    /api/v1/notifications/{id}            → Missing handler
GET    /api/v1/notifications                 → List with filters (missing)
POST   /api/v1/notifications/{id}/read       → Already has handler
POST   /api/v1/notifications/read-all        → Missing

POST   /api/v1/templates                     → Has handler
GET    /api/v1/templates                     → Has handler
GET    /api/v1/templates/{id}                → Has handler
PUT    /api/v1/templates/{id}                → Has handler
DELETE /api/v1/templates/{id}                → Has handler

POST   /api/v1/preferences                   → Missing
GET    /api/v1/preferences/{user_id}         → Has handler (different path)
PUT    /api/v1/preferences/{user_id}         → Has handler

POST   /api/v1/reminders                     → Missing (entirely)
GET    /api/v1/reminders/{id}                → Missing
PUT    /api/v1/reminders/{id}                → Missing
DELETE /api/v1/reminders/{id}                → Missing

GET    /healthz                              → Health check
GET    /readyz                               → Readiness check
```

---

## Part 5 — Required Notifier Domain Model

### 5.1 Existing Domain Models

The following models exist in `internal/models/`:

| Model | File | Status |
|---|---|---|
| `Notification` | `notification.go` | **Implemented** |
| `NotificationLog` | `notification_log.go` | **Implemented** |
| `NotificationPreference` | `notification_preference.go` | **Implemented** |
| `NotificationTemplate` | `notification_template.go` | **Implemented** |
| `ServiceClient` | `service_client.go` | **Implemented** |
| `Setting` | `setting.go` | **Implemented** |
| `SMSTemplate` | `sms_template.go` | **Implemented** |

### 5.2 Recommended Complete Domain Models

**Missing models (to be added):**
- `NotificationRecipient` — separate from Notification for multi-recipient
- `NotificationChannel` — channel configuration & health
- `NotificationDelivery` — per-recipient delivery tracking
- `NotificationAttempt` — per-attempt detailed tracking (replaces NotificationLog)
- `Reminder` — scheduled reminder entity
- `ProviderCredential` — separate from settings for security
- `WebhookEvent` — callback configuration
- `NotificationCategory` — categorization

### 5.3 Recommended Enum Values

```go
// NotificationType
type NotificationType string
const (
    NotificationTypeSMS    NotificationType = "sms"
    NotificationTypeEmail  NotificationType = "email"
    NotificationTypePush   NotificationType = "push"
    NotificationTypeInApp  NotificationType = "in_app"
)

// NotificationStatus
type NotificationStatus string
const (
    NotificationStatusPending   NotificationStatus = "pending"
    NotificationStatusQueued    NotificationStatus = "queued"
    NotificationStatusProcessing NotificationStatus = "processing"
    NotificationStatusSent      NotificationStatus = "sent"
    NotificationStatusFailed    NotificationStatus = "failed"
    NotificationStatusCancelled NotificationStatus = "cancelled"
    NotificationStatusDelivered NotificationStatus = "delivered"
    NotificationStatusRead      NotificationStatus = "read"
    NotificationStatusClicked   NotificationStatus = "clicked"
)

// DeliveryStatus (for notification_recipients)
type DeliveryStatus string
const (
    DeliveryStatusPending    DeliveryStatus = "pending"
    DeliveryStatusSent       DeliveryStatus = "sent"
    DeliveryStatusFailed     DeliveryStatus = "failed"
    DeliveryStatusRetrying   DeliveryStatus = "retrying"
    DeliveryStatusDelivered  DeliveryStatus = "delivered"
    DeliveryStatusRead       DeliveryStatus = "read"
    DeliveryStatusClicked    DeliveryStatus = "clicked"
)

// ChannelType
type ChannelType string
const (
    ChannelTypeEmail   ChannelType = "email"
    ChannelTypeSMS     ChannelType = "sms"
    ChannelTypePush    ChannelType = "push"
    ChannelTypeInApp   ChannelType = "in_app"
    ChannelTypeWebhook ChannelType = "webhook"
)
```

---

## Part 6 — Database Schema & Migrations

### 6.1 Current Tables (from SQL migrations)

| Table | Purpose | Status |
|---|---|---|
| `notifications` | Core notification records | **Implemented** |
| `notification_templates` | Reusable templates | **Implemented** |
| `notification_preferences` | Per-user preferences | **Implemented** |
| `notification_logs` | Delivery audit trail | **Implemented** |
| `settings` | Dynamic configuration | **Implemented** |
| `service_clients` | Service-to-service auth | **Implemented** |
| `sms_templates` | Provider-specific SMS template mappings | **Implemented** |
| `tenants` | Multi-tenancy | **Implemented** |
| `audit_logs` | Change audit trail | **Implemented** |

### 6.2 Schema Gaps

| Missing Table | Reason |
|---|---|
| `notification_recipients` | Current Notification has single recipient fields; multi-recipient needs separate table |
| `notification_deliveries` | Per-recipient delivery tracking with status transitions |
| `notification_attempts` | Granular per-attempt tracking (replaces less structured logs) |
| `reminders` | Scheduled/cron-based notification triggers |
| `provider_credentials` | Credentials stored in settings table — should be separate for security |
| `webhook_events` | Callback configuration for notification status changes |
| `notification_categories` | Category grouping (marketing, alerts, updates, etc.) |

### 6.3 Key Design Questions

1. **Is multi-tenancy needed?** YES — already partially implemented (tenant_id columns, RLS policies, tenants table)
2. **Should it support project/app isolation?** YES — tenant_id provides this
3. **Should it store provider responses?** YES — notification_logs has `provider_response` JSONB field
4. **Should it store template variables?** YES — notification_templates has `variables` JSONB field
5. **Should it store read/seen status?** YES — notifications has `read_at` field, but no `seen_at`

---

## Part 7 — Provider Integrations

### 7.1 Current Provider Architecture

**SMS Provider Interface:**
```go
// internal/platform/sms/platforms/base.go
type SmsClient interface {
    SendMessage(param map[string]string, targetPhoneNumber ...string) error
}
```

**Email Provider Interface:**
```go
// internal/platform/email/email.go
type EmailClient interface {
    SendEmail(to, subject, body string, isHTML bool) error
}
```

**Push Provider Interface:**
```go
// internal/platform/push/push.go
type PushClient interface {
    SendPush(deviceToken, title, body string, data map[string]string) error
}
```

### 7.2 Recommended Unified Provider Interface

```go
// Recommended unified provider interface

package provider

import "context"

type Channel string

const (
    ChannelEmail Channel = "email"
    ChannelSMS   Channel = "sms"
    ChannelPush  Channel = "push"
    ChannelInApp Channel = "in_app"
)

type Message struct {
    ID          string
    Channel     Channel
    Recipient   string   // email, phone, device token, or user ID
    Subject     string
    Body        string
    HTMLContent string   // for email
    Metadata    map[string]string
    TemplateID  string
    Variables   map[string]interface{}
}

type SendResult struct {
    ProviderMsgID string
    ProviderName  string
    Status        string   // sent, failed, queued
    Error         error
    RawResponse   string   // provider raw response
}

type Provider interface {
    Name() string
    Channel() Channel
    Send(ctx context.Context, msg *Message) (*SendResult, error)
    IsHealthy(ctx context.Context) bool
}

type ProviderManager interface {
    GetProvider(channel Channel, preferredName ...string) (Provider, error)
    GetDefaultProvider(channel Channel) (Provider, error)
    GetFallbackProvider(channel Channel) (Provider, error)
    RegisterProvider(provider Provider)
}
```

### 7.3 Provider Selection Strategy

```
For each channel:
  1. Check if preferred provider is specified
  2. Use default provider for that channel
  3. If default fails (circuit breaker open, rate limited, error):
     → Fallback to secondary provider
     → If all fail, mark as failed with retry

Health monitoring:
  - Periodically ping each provider
  - Track success/failure rate
  - Circuit breaker pattern (via go-common/common/circuit_breaker.go)

Rate limits:
  - Per-provider rate limit
  - Per-tenant rate limit
  - Sliding window counter in Redis
```

### 7.4 Immediate Recommendation

Given DiviPay's requirements (Iranian market), focus on:
- **SMS:** Kavenegar (primary for Iran), Twilio (fallback for international)
- **Email:** SMTP (for now), add SendGrid/Resend later
- **Push:** FCM v1 (migrate from legacy FCM)
- **In-App:** Already implemented via WebSocket

---

## Part 8 — Queue, Retry, Scheduler & Workers

### 8.1 Current Queue System

**Status:** In-memory channel-based queue
- No persistence
- No delayed delivery
- No dead-letter handling
- No external message broker integration
- Worker pool with configurable concurrency

### 8.2 Recommended Architecture

**Option A: PostgreSQL-based queue (recommended for MVP)**
- Use `LISTEN/NOTIFY` or periodic polling
- Persist notifications with status in DB
- Worker picks up `pending`/`retrying` notifications
- Simplest path, no additional infrastructure

**Option B: Redis queue (recommended for production)**
- Use Redis lists or streams for queuing
- Faster than PostgreSQL polling
- Already have Redis in docker-compose
- Use `github.com/redis/go-redis/v9` (already in dependencies)

**Option C: Dedicated message broker (future)**
- RabbitMQ or NATS for high volume
- More complex operations
- Not recommended for initial DiviPay needs

### 8.3 Scheduled Reminders Strategy

**Recommendation: notifier owns internal scheduling first, integrates with scheduler service later**

1. **Phase 1:** Add `reminders` table + in-process scheduler using `time.Ticker` or `robfig/cron`
2. **Phase 2:** Optionally integrate with Minisource scheduler service via gRPC/webhook
3. **Phase 3:** Support both (scheduler for complex schedules, internal for simple ones)

### 8.4 Retry & Dead-Letter Queue

```
Flow:
  1. Worker picks notification from queue
  2. Attempts delivery via provider
  3. If successful: mark as sent, log success
  4. If failed:
     a. Increment retry_count
     b. If retry_count < max_retries: schedule retry with backoff
     c. If retry_count >= max_retries: move to dead-letter queue (DLQ)
  5. DLQ notifications: manual admin review, auto-retry after configurable delay, or discard after TTL

Dead-letter handling:
  - Mark as "dead" status
  - Store in notification_logs with error details
  - Admin endpoint: retry from DLQ
  - TTL: auto-delete after 30 days (configurable)
```

### 8.5 Idempotency

```
Recommendation: Add idempotency_key to notifications
  - Caller generates unique idempotency_key
  - Notifier checks for duplicate before creating
  - If duplicate: return existing notification ID
  - TTL: 24 hours (configurable)
  - Storage: Redis SET with TTL, or DB unique index
```

---

## Part 9 — Security & Authentication

### 9.1 Current Auth

| Aspect | Status |
|---|---|
| Service-to-service auth | **Implemented** — JWT + scope-based via `go-sdk/auth` |
| JWT validation | **Partial** — `AuthMiddleware` in go-common |
| Auth service integration | **Partial** — auth client configured, but service may not be ready |
| API key / internal token | **Partial** — ServiceClient model exists |
| Tenant isolation | **Implemented** — RLS policies, tenant middleware |
| Rate limiting | **Missing** |
| Input validation | **Partial** — validation middleware present, DTOs have tags |
| Secrets management | **Missing** — API keys in env vars, no secret manager |
| Webhook signature | **Missing** — no webhooks yet |
| PII handling | **Missing** — no PII redaction in logs |

### 9.2 PII Handling Recommendations

Notifier processes: phone numbers, emails, device tokens, message contents.

**Required measures:**
1. **Log redaction:** Never log full phone numbers or email addresses in production logs
2. **Database encryption:** Encrypt sensitive columns (provider credentials, PII)
3. **Data retention:** Configurable retention policy for notification content
4. **GDPR compliance:** Support for user data deletion (export/delete notifications)
5. **Audit trail:** All access to PII should be logged

```go
// Example log redaction helper
func redactPhone(phone string) string {
    if len(phone) > 4 {
        return phone[:4] + "****"
    }
    return "****"
}
```

---

## Part 10 — Configuration

### 10.1 Current Config Structure

**Status:** Good foundation, needs extension

**File:** `config/config.go`

**Missing config fields:**
```
REDIS_URL               → For Redis queue
JWT_PUBLIC_KEY           → For JWT validation without auth service
INTERNAL_API_KEY         → Simple service-to-service auth fallback
SMTP_HOST / PORT / ...   → Currently in database — should also support env
SMS_PROVIDER             → Primary SMS provider name
SMS_API_KEY              → Primary SMS API key
FCM_CREDENTIALS_JSON     → FCM service account JSON
DEFAULT_FROM_EMAIL       → Default sender email
DEFAULT_SMS_SENDER       → Default SMS sender ID/number
LOG_LEVEL                → Already exists
TRACING_ENDPOINT         → Already exists
```

### 10.2 Recommended Config Validation

Add `Validate()` method to `Config` that checks:
- At least one email provider configured
- At least one SMS provider configured
- Database URL reachable
- Redis URL reachable (if queue enabled)

---

## Part 11 — Observability

### 11.1 Current Status

| Aspect | Status |
|---|---|
| Structured logging | **Implemented** — via go-common/logging |
| Request ID | **Implemented** — via go-common middleware |
| Correlation ID | **Implemented** — tracing context propagation |
| Delivery tracking logs | **Implemented** — notification_logs |
| Prometheus metrics | **Partial** — counters/histograms defined but not wired |
| OpenTelemetry tracing | **Implemented** — Jaeger exporter |
| Health checks | **Partial** — `/health` exists, no `/readyz` |
| Readiness checks | **Missing** |
| Provider failure metrics | **Missing** |
| Queue depth metrics | **Missing** |

### 11.2 Recommended Metrics

```go
// Prometheus metrics to add
notifications_sent_total{channel, provider, status}
notifications_failed_total{channel, provider, error_type}
notification_delivery_duration_seconds{channel, provider}
notification_retry_total{channel, provider}
notification_queue_depth
provider_error_total{provider, error_type}
worker_processing_duration_seconds{worker_id}
```

### 11.3 Integration with Minisource `log` Service

The `go-sdk/log` SDK exists with full client implementation. However:
- The notifier does not currently integrate with it
- Recommendation: send structured log entries to log service for centralized logging
- This is optional for Phase 0/1, recommended for Phase 11

---

## Part 12 — Testing Strategy

### 12.1 Current Test Status

- **E2E tests:** Exist but require running auth service + DB
- **Unit tests:** Minimal (mock SMS + old SMTP test)
- **Load tests:** k6 script exists

### 12.2 Recommended Tests

| Test Type | Files to Add | Priority |
|---|---|---|
| Unit: notification service | `internal/service/notification_service_test.go` | **High** |
| Unit: preference service | `internal/service/preference_service_test.go` | **High** |
| Unit: template service | `internal/service/template_service_test.go` | **High** |
| Unit: repository tests (with mock DB or test DB) | `internal/repository/*_test.go` | **Medium** |
| Unit: provider mock tests | `internal/platform/sms/*_test.go` | **Medium** |
| Unit: worker tests | `internal/worker/*_test.go` | **Medium** |
| Unit: WebSocket hub tests | `internal/websocket/hub_test.go` | **Medium** |
| Unit: template rendering tests | `internal/service/template_render_test.go` | **Low** |
| Unit: preference filtering tests | `internal/service/preference_filter_test.go` | **Low** |
| Integration: HTTP handler tests | `api/v1/handlers/*_test.go` | **High** |
| Integration: migration tests | `migrations/*_test.go` | **Medium** |
| Integration: retry tests | `internal/worker/retry_test.go` | **Medium** |
| Integration: idempotency tests | `internal/service/idempotency_test.go` | **Low** |

---

## Part 13 — SDK & Integration Requirements

### 13.1 Current Go SDK

**File:** `go-sdk/notifier/client.go`

The Go SDK has a comprehensive gRPC-based client with:
```go
SendSMS(ctx, userID, phone, body) → (string, error)  // Deprecated
SendSMSWithData(ctx, *SMSRequest) → (string, error)   // Preferred
SendEmail(ctx, userID, email, subject, body) → (string, error)
SendPush(ctx, userID, body) → (string, error)
SendInApp(ctx, userID, body) → (string, error)
CreateNotification(ctx, *CreateNotificationRequest) → (string, error)
CreateBatchNotifications(ctx, []*CreateNotificationRequest) → ([]string, error)
GetUserNotifications(ctx, userID, page, pageSize) → ([]*Notification, int64, error)
GetUnreadNotifications(ctx, userID, page, pageSize) → ([]*Notification, int64, error)
MarkAsRead(ctx, notificationID) → error
GetNotification(ctx, notificationID) → (*Notification, error)
StreamNotifications(ctx, userID, handler) → error
CreateTemplate(ctx, *CreateTemplateRequest) → (string, error)
GetTemplate(ctx, templateID) → (*Template, error)
GetAllTemplates(ctx, page, pageSize) → ([]*Template, int64, error)
GetUserPreferences(ctx, userID) → ([]*Preference, error)
UpdatePreference(ctx, *UpdatePreferenceRequest) → error
```

### 13.2 Missing SDK Methods

```go
// To be added
SendSMSByTemplate(ctx, phone, template, data map[string]string) → (string, error)
SendEmailByTemplate(ctx, email, template, data map[string]interface{}) → (string, error)
MarkAllAsRead(ctx, userID) → error
GetNotificationCount(ctx, userID, unreadOnly bool) → (int64, error)
CreateReminder(ctx, *ReminderRequest) → (string, error)
CancelReminder(ctx, reminderID) → error
GetUserPreferencesByChannel(ctx, userID, channel) → (*Preference, error)
```

### 13.3 C# SDK

No C# SDK exists for notifier yet. Recommended methods:
```csharp
Task<string> SendSmsAsync(string phone, string template, Dictionary<string, string> data);
Task<string> SendEmailAsync(string email, string subject, string body);
Task<string> SendPushAsync(string userId, string title, string body);
Task<string> SendInAppAsync(string userId, string body);
Task<List<Notification>> GetUserNotificationsAsync(string userId, int page, int pageSize);
Task MarkAsReadAsync(string notificationId);
Task<int> GetUnreadCountAsync(string userId);
```

### 13.4 API Contract Generation

- gRPC proto definitions exist and are comprehensive
- OpenAPI (Swagger) docs exist but need regeneration after route files are created
- Recommendation: Add `buf` generation or `protoc` in CI pipeline, generate OpenAPI from proto

---

## Part 14 — DiviPay Integration Requirements

### 14.1 Use Case Mapping

| # | Use Case | Trigger | Channels | Template | Variables | Recipient | Priority | Delivery |
|---|---|---|---|---|---|---|---|---|
| 1 | OTP / Login | Auth service | SMS + Email | `otp_verification` | `{code, expiry, appName}` | User's phone/email | **Urgent** | Immediate (sync) |
| 2 | Debt reminder | DiviPay backend | Push + In-App (+ optional SMS) | `debt_reminder` | `{debtorName, amount, dueDate, paymentLink}` | User | **High** | Scheduled |
| 3 | Settlement reminder | DiviPay backend | Push + In-App | `settlement_reminder` | `{groupName, amount, date}` | User | **High** | Scheduled |
| 4 | Payment status | DiviPay backend | Push + In-App + Email | `payment_status` | `{status, amount, transactionId, date}` | User | **Normal** | Immediate |
| 5 | Group invitation | DiviPay backend | Push + In-App + Email | `group_invitation` | `{groupName, inviterName, joinLink}` | User | **Normal** | Immediate |
| 6 | Public link sharing | DiviPay backend | SMS + Push | `public_link` | `{linkName, url, senderName}` | Recipient | **Low** | Immediate |
| 7 | Account/security alert | Auth service | Email + In-App | `security_alert` | `{alertType, deviceInfo, time, location}` | User | **Urgent** | Immediate |
| 8 | Wallet charge/withdraw | DiviPay backend | Push + In-App + Email | `wallet_transaction` | `{amount, type, balance, date}` | User | **High** | Immediate |

### 14.2 Integration Requirements

1. **OTP (Use Case 1):** Requires `CreateNotificationSync` for immediate delivery with error feedback
2. **Debt/Settlement reminders (Use Cases 2, 3):** Requires scheduler integration
3. **Payment notifications (Use Case 4):** Standard async notification, template-based
4. **Group invitations (Use Case 5):** Needs sender info + receiver preferences
5. **Public links (Use Case 6):** Can target non-users (email/SMS), needs PII handling
6. **Security alerts (Use Case 7):** Bypass preference checks, always deliver
7. **Wallet transactions (Use Case 8):** High volume, needs rate limiting

---

## Part 15 — Completion Roadmap

### Phase 0 — Stabilize Project Structure

**Objective:** Fix what's broken so the project compiles and runs

| Task | Files | Complexity |
|---|---|---|
| Create missing route files | `api/v1/routes/notifications.go`, `preferences.go`, `templates.go`, `sms.go`, `websocket.go` | **Low** |
| Fix Dockerfile Go version | `Dockerfile` (1.23.4 → 1.26.0) | **Low** |
| Fix migration version conflict | `migrations/000002_*` | **Low** |
| Remove stale binary | `bin/`, `migrate.exe`, `__debug_bin*.exe` | **Low** |
| Clean up `internal/notification/` legacy package | Move to `internal/service/` or delete | **Low** |

**Acceptance Criteria:** `go build ./...` succeeds, `go run cmd/server/main.go` starts

**Dependencies:** None

---

### Phase 1 — Config, Logging, Health Checks

**Objective:** Production-ready config validation, structured logging, readiness checks

| Task | Files | Complexity |
|---|---|---|
| Add Config.Validate() method | `config/config.go` | **Low** |
| Add /readyz endpoint | `api/v1/routes/health.go` | **Low** |
| Add DB readiness check | `internal/database/database.go` | **Low** |
| Add Redis readiness check | New `internal/database/redis.go` | **Low** |
| Add provider health check | `internal/platform/` new health methods | **Medium** |

**Acceptance Criteria:** `/healthz` and `/readyz` return proper status, config validation runs at startup

**Dependencies:** Phase 0

---

### Phase 2 — Domain Models & Database Migrations

**Objective:** Complete domain model, add missing tables

| Task | Files | Complexity |
|---|---|---|
| Create `Reminder` model | `internal/models/reminder.go` | **Medium** |
| Create `WebhookEvent` model | `internal/models/webhook_event.go` | **Medium** |
| Create `ProviderCredential` model | `internal/models/provider_credential.go` | **Medium** |
| Create `NotificationCategory` model | `internal/models/notification_category.go` | **Low** |
| Add migration for reminder tables | `migrations/000005_add_reminders.up.sql` | **Medium** |
| Add migration for webhook events | `migrations/000006_add_webhooks.up.sql` | **Medium** |
| Add migration for provider credentials | `migrations/000007_provider_credentials.up.sql` | **Medium** |
| Add migration for notification_categories | `migrations/000008_categories.up.sql` | **Low** |
| Fix dual migration system (choose one) | `internal/database/database.go` | **Medium** |

**Acceptance Criteria:** All new tables created via migration, models map correctly

**Dependencies:** Phase 1

---

### Phase 3 — REST/gRPC API Contracts

**Objective:** Complete API surface, fix missing route files

| Task | Files | Complexity |
|---|---|---|
| Create `routes/notifications.go` | `api/v1/routes/notifications.go` | **Low** |
| Create `routes/preferences.go` | `api/v1/routes/preferences.go` | **Low** |
| Create `routes/templates.go` | `api/v1/routes/templates.go` | **Low** |
| Create `routes/sms.go` | `api/v1/routes/sms.go` | **Low** |
| Create `routes/websocket.go` | `api/v1/routes/websocket.go` | **Low** |
| Add missing GET /v1/notifications/{id} endpoint | `api/v1/handlers/notification_handler.go` | **Low** |
| Add POST /v1/notifications/read-all | `api/v1/handlers/notification_handler.go` | **Low** |
| Add AdminService gRPC registration | `api/grpc/server.go` | **Medium** |
| Implement gRPC StreamNotifications | `api/grpc/notification_handlers.go` | **Medium** |
| Regenerate Swagger docs | `docs/swagger.json` | **Low** |

**Acceptance Criteria:** All endpoints respond correctly, Swagger docs accurate, gRPC methods register

**Dependencies:** Phase 0, Phase 2

---

### Phase 4 — Provider Abstraction

**Objective:** Unified provider interface

| Task | Files | Complexity |
|---|---|---|
| Create unified Provider interface | `internal/provider/provider.go` (new dir) | **Medium** |
| Create ProviderManager with fallback | `internal/provider/manager.go` | **Medium** |
| Add circuit breaker wrapper | `internal/provider/circuit_breaker.go` | **Medium** |
| Add provider health check | `internal/provider/health.go` | **Medium** |
| Create SMS provider adapter (existing → unified) | `internal/provider/sms_adapter.go` | **Medium** |
| Create Email provider adapter (existing → unified) | `internal/provider/email_adapter.go` | **Medium** |
| Create Push provider adapter (existing → unified) | `internal/provider/push_adapter.go` | **Medium** |

**Acceptance Criteria:** ProviderManager works, fallback works, health checks pass

**Dependencies:** Phase 2

---

### Phase 5 — Email/SMS/Push Implementations

**Objective:** Production-ready provider implementations

| Task | Files | Complexity |
|---|---|---|
| Add SendGrid email provider | `internal/platform/email/sendgrid.go` | **Medium** |
| Add FCM v1 HTTP API push provider | `internal/platform/push/fcm_v1.go` | **Medium** |
| Add APNs push provider | `internal/platform/push/apns.go` | **High** |
| Test and harden Kavenegar SMS | `internal/platform/sms/platforms/kavenegar.go` | **Low** |
| Test and harden Twilio SMS | `internal/platform/sms/platforms/twilio.go` | **Low** |
| Add HTML email template rendering | `internal/platform/email/renderer.go` | **Medium** |
| Add email attachment support | `internal/platform/email/attachment.go` | **Low** |
| Remove unused SMS provider code | Multiple platform files | **Low** |

**Acceptance Criteria:** SMS (Kavenegar, Twilio), Email (SMTP, SendGrid), Push (FCM v1) all functional

**Dependencies:** Phase 4

---

### Phase 6 — In-App Notifications & Read Status

**Objective:** Complete in-app delivery + read/seen tracking

| Task | Files | Complexity |
|---|---|---|
| Add `seen_at` field | Migration + model update | **Low** |
| Add unread count endpoint | `api/v1/handlers/notification_handler.go` | **Low** |
| Add read-all endpoint | `api/v1/handlers/notification_handler.go` | **Low** |
| Implement offline notification sync (last seen ID) | `internal/service/notification_service.go` | **Medium** |
| Add WebSocket reconnection handling | `internal/websocket/hub.go` | **Medium** |

**Acceptance Criteria:** Unread count accurate, read-all works, offline sync works

**Dependencies:** Phase 3

---

### Phase 7 — Queue/Retry/Worker

**Objective:** Production-grade async processing

| Task | Files | Complexity |
|---|---|---|
| Add Redis queue implementation | `internal/worker/redis_queue.go` | **High** |
| Add dead-letter queue handling | `internal/worker/dead_letter.go` | **Medium** |
| Add idempotency support | `internal/service/idempotency.go` | **Medium** |
| Add priority queue support | `internal/worker/priority_queue.go` | **Medium** |
| Add worker metrics (queue depth, processing time) | `internal/worker/metrics.go` | **Low** |
| Add graceful shutdown improvements | `internal/worker/notification_worker.go` | **Low** |

**Acceptance Criteria:** Redis queue works, DLQ functional, idempotency prevents duplicates

**Dependencies:** Phase 1 (Redis connectivity)

---

### Phase 8 — Reminders/Scheduling

**Objective:** Scheduled and recurring notification support

| Task | Files | Complexity |
|---|---|---|
| Implement Reminder CRUD service | `internal/service/reminder_service.go` | **Medium** |
| Implement Reminder repository | `internal/repository/reminder_repository.go` | **Medium** |
| Implement in-process scheduler | `internal/scheduler/scheduler.go` (new dir) | **High** |
| Add reminder API endpoints | `api/v1/handlers/reminder_handler.go` | **Medium** |
| Add reminder integration with scheduler service (optional) | `internal/scheduler/external.go` | **High** |
| Add recurring reminder support (cron expressions) | `internal/scheduler/cron.go` | **High** |

**Acceptance Criteria:** Reminders created, trigger at correct time, recurring works

**Dependencies:** Phase 3, Phase 7

---

### Phase 9 — Preferences/Templates

**Objective:** Complete preference filtering, template rendering

| Task | Files | Complexity |
|---|---|---|
| Add Go template variable rendering | `internal/service/template_renderer.go` | **Medium** |
| Add template preview endpoint | `api/v1/handlers/template_handler.go` | **Low** |
| Add quiet hours enforcement in preference check | `internal/service/notification_service.go` | **Medium** |
| Add digest email generation | `internal/service/digest_service.go` | **High** |
| Add category-level preference filtering | `internal/service/notification_service.go` | **Medium** |
| Add preference-based channel routing | `internal/service/notification_service.go` | **Medium** |

**Acceptance Criteria:** Templates render correctly, quiet hours respected, digests generated

**Dependencies:** Phase 3, Phase 5

---

### Phase 10 — Security/Auth/Rate Limiting

**Objective:** Production security hardening

| Task | Files | Complexity |
|---|---|---|
| Add rate limiting middleware | `api/middleware/ratelimit.go` | **Medium** |
| Add per-provider rate limiting | `internal/provider/ratelimit.go` | **Medium** |
| Add PII log redaction | `internal/middleware/pii_redaction.go` | **Low** |
| Add secrets management integration (env → vault) | `config/config.go` | **Medium** |
| Add webhook signature verification | `internal/service/webhook.go` | **Medium** |
| Add data retention policy enforcement | `internal/service/retention.go` | **Medium** |
| Add GDPR data export/delete | `internal/service/gdpr.go` | **Medium** |

**Acceptance Criteria:** Rate limits enforced, PII not in logs, secrets not in config files

**Dependencies:** Phase 1, Phase 3

---

### Phase 11 — Tests

**Objective:** Comprehensive test coverage

| Task | Files | Complexity |
|---|---|---|
| Add notification service unit tests | `internal/service/*_test.go` | **Medium** |
| Add HTTP handler integration tests | `api/v1/handlers/*_test.go` | **Medium** |
| Add repository tests (testcontainers) | `internal/repository/*_test.go` | **High** |
| Add worker/retry tests | `internal/worker/*_test.go` | **Medium** |
| Add provider mock tests | `internal/provider/*_test.go` | **Medium** |
| Add migration tests | `migrations/*_test.go` | **Medium** |
| Add E2E tests (with test containers) | `tests/e2e/*_test.go` | **High** |

**Acceptance Criteria:** >70% code coverage, all critical paths tested

**Dependencies:** All previous phases

---

### Phase 12 — Docker/CI/Deployment

**Objective:** Production deployment ready

| Task | Files | Complexity |
|---|---|---|
| Fix Dockerfile Go version | `Dockerfile` | **Low** |
| Add CI pipeline (GitHub Actions) | `.github/workflows/ci.yml` | **Medium** |
| Add deployment manifests (K8s/Docker Swarm) | `deployments/` (new dir) | **High** |
| Add database migration CI step | `.github/workflows/migrate.yml` | **Medium** |
| Add health check monitoring | `deployments/monitoring/` (new dir) | **Medium** |
| Add log aggregation config | `deployments/logging/` (new dir) | **Medium** |

**Acceptance Criteria:** CI passes, deployment works, monitoring active

**Dependencies:** All previous phases

---

### Phase 13 — SDK Integration

**Objective:** Complete client SDKs

| Task | Files | Complexity |
|---|---|---|
| Add missing Go SDK methods | `go-sdk/notifier/client.go` | **Medium** |
| Create C# SDK for notifier | New C# project | **High** |
| Add OpenAPI spec generation in CI | New CI step | **Low** |
| Add SDK examples and documentation | README updates | **Low** |

**Acceptance Criteria:** Go SDK has all methods, C# SDK exists, OpenAPI spec auto-generated

**Dependencies:** Phase 3

---

### Phase 14 — DiviPay Integration

**Objective:** DiviPay-specific notification scenarios

| Task | Files | Complexity |
|---|---|---|
| Create DiviPay-specific templates | Seed data updates | **Low** |
| Add DiviPay template grouping | New seed data | **Low** |
| Test OTP flow end-to-end | Integration tests | **Medium** |
| Test debt reminder flow | Integration tests | **Medium** |
| Test payment notification flow | Integration tests | **Medium** |
| Performance testing for DiviPay load profile | k6 script updates | **Medium** |

**Acceptance Criteria:** All 8 DiviPay use cases pass, OTP under 5s, batch reminders under 30s

**Dependencies:** Phases 0-13

---

## Part 16 — Recommended Final Architecture

### Package Layout

```
notifier/
├── cmd/notifier/main.go              # Entry point
├── internal/
│   ├── config/                        # Configuration loading + validation
│   ├── domain/                        # Domain models (entities, value objects)
│   │   ├── notification.go
│   │   ├── recipient.go
│   │   ├── template.go
│   │   ├── preference.go
│   │   ├── reminder.go
│   │   ├── provider.go
│   │   ├── webhook.go
│   │   └── errors.go                  # Domain errors
│   ├── repository/                    # Data access (interfaces in domain, implementations here)
│   │   ├── interfaces.go             # Repository interfaces
│   │   ├── postgres/                  # PostgreSQL implementations
│   │   └── redis/                    # Redis implementations
│   ├── service/                       # Business logic / use cases
│   │   ├── notification_service.go
│   │   ├── template_service.go
│   │   ├── preference_service.go
│   │   ├── reminder_service.go
│   │   ├── admin_service.go
│   │   └── provider_service.go        # Provider selection & routing
│   ├── provider/                      # Provider abstraction
│   │   ├── provider.go               # Provider interface + types
│   │   ├── manager.go                 # Provider lifecycle & routing
│   │   ├── sms/                       # SMS implementations
│   │   ├── email/                     # Email implementations
│   │   └── push/                      # Push implementations
│   ├── worker/                        # Async processing
│   │   ├── worker.go                  # Worker pool
│   │   ├── queue.go                   # Queue abstraction
│   │   ├── redis_queue.go             # Redis implementation
│   │   ├── retry.go                   # Retry logic
│   │   └── dead_letter.go            # Dead-letter handling
│   ├── scheduler/                     # Reminder scheduling
│   │   ├── scheduler.go              # Internal scheduler
│   │   └── cron.go                    # Cron expression support
│   ├── websocket/                     # WebSocket hub
│   ├── middleware/                    # HTTP/gRPC middleware
│   └── observability/                 # Metrics, tracing, logging
├── transport/                         # API layer
│   ├── http/                          # HTTP handlers + routes
│   │   ├── handler/                   # Request handlers
│   │   ├── middleware/                # HTTP-specific middleware
│   │   ├── dto/                       # Request/response DTOs
│   │   └── router.go                  # Route registration
│   └── grpc/                          # gRPC handlers
│       ├── server.go
│       └── handler/
├── migrations/                        # SQL migrations
├── api/                               # API specifications
│   ├── proto/                         # Protobuf definitions
│   └── openapi/                       # OpenAPI specs
├── pkg/                               # Shared packages
│   └── client/                        # Public client SDK
├── deploy/                            # Deployment files
└── docs/                              # Documentation
```

### Dependency Direction

```
transport/http  →  service  →  repository  →  database (PostgreSQL/Redis)
transport/grpc  →  service  →  provider     →  external APIs (SMS/Email/Push)
                    service  →  worker       →  queue (Redis)
                    service  →  websocket    →  browser clients
                    service  →  scheduler    →  internal timing
```

### Key Architecture Decisions

1. **Clean Architecture with domain-driven package layout**
2. **Repository interfaces in domain package**, implementations separated
3. **Unified Provider interface** with Manager for routing and fallback
4. **Redis as the queue backend** (simplest production path with existing infra)
5. **Internal scheduler** for reminders (integration with scheduler service later)
6. **Observability as a cross-cutting concern** (metrics, tracing, structured logging)

---

## Part 17 — Risks & Open Questions

### Open Questions

| # | Question | Implication |
|---|---|---|
| 1 | **Should notifier own notification preferences or should product services own them?** | If product services own preferences, notifier needs to fetch preferences via API (latency). If notifier owns them, need sync mechanism when users change preferences. |
| 2 | **Should reminders be scheduled by notifier or scheduler service?** | Notifier scheduling = simpler but less powerful. Scheduler service = more scalable but cross-service dependency. |
| 3 | **Should read status be stored in notifier or product DB?** | Notifier storage = unified view. Product DB = product knows immediately. Recommendation: notifier storage with webhook callbacks. |
| 4 | **Should provider credentials be per tenant/project?** | If yes, need tenant-aware provider selection. If no, simpler but less flexible. |
| 5 | **Which SMS provider for Iranian numbers?** | Kavenegar is already integrated. Need to verify production readiness and cost model. |
| 6 | **Should OTP be handled by auth directly or notifier?** | Auth should trigger OTP via notifier (separate concerns). But OTP needs synchronous response for user experience. |
| 7 | **What retention policy is required for message content?** | PII laws (GDPR, Iran PDPL) require data minimization. Default: 90 days for content, 1 year for metadata. |
| 8 | **Should notifier support delivery receipts (read/clicked)?** | Email: read receipts via tracking pixel. SMS: delivery receipts (varies by provider). Push: delivery receipts via FCM/APNs. |
| 9 | **Should we consolidate the 16 SMS providers or keep them all?** | Risk: untested code, maintenance burden. Recommendation: keep Kavenegar + Twilio, remove the rest or move to separate repo. |
| 10 | **How to handle the auth service dependency when auth doesn't exist yet?** | Notifier currently handles this gracefully (AUTH_ENABLED=false), but production needs auth. |

### Risks

| Risk | Severity | Mitigation |
|---|---|---|
| **Dual migration system** (GORM AutoMigrate + golang-migrate) | **High** | Choose one system. Recommend: golang-migrate for SQL, disable AutoMigrate in production. |
| **Missing route files** | **High** | Create them in Phase 0 before any other work. |
| **Auth service dependency** | **Medium** | Implement internal API key fallback for service-to-service auth. |
| **In-process queue (no persistence)** | **High** | Add Redis queue in Phase 7. Until then, notifications may be lost on restart. |
| **Go version mismatch** (go.mod 1.26, Dockerfile 1.23.4) | **Medium** | Fix Dockerfile Go version. |
| **Stale binaries in repo** (bin/*.exe) | **Low** | Add to .gitignore, clean up. |
| **21 SMS providers untested** | **Medium** | Remove unused providers, test Kavenegar + Twilio only. |
| **No rate limiting** | **High** | Add rate limiting before production deployment. |
| **PII in logs** | **High** | Add log redaction middleware immediately. |

---

## Part 18 — Immediate Next Actions

Here are the **exact 10 tasks** to start implementing immediately:

1. **Create missing route files** — `api/v1/routes/notifications.go`, `preferences.go`, `templates.go`, `sms.go`, `websocket.go` (these are just route registration files that call existing handlers)

2. **Fix Dockerfile Go version** — Change `FROM golang:1.23.4-alpine` to `FROM golang:1.26.0-alpine`

3. **Fix migration version conflict** — Rename `000002_sms_templates` to `000003` (or merge with `000002_add_multi_tenancy`)

4. **Choose migration system** — Disable GORM `AutoMigrate` in production config, use only `golang-migrate` for SQL migrations. Keep AutoMigrate only for development.

5. **Clean up repo artifacts** — Remove `*.exe` files, old test files (`.old`), stale binaries

6. **Add `Config.Validate()` method** — Validate required config at startup

7. **Add `/readyz` and `/healthz` endpoints** — With proper DB and Redis health checks

8. **Add PII log redaction** — Before any production deployment, add phone/email redaction in logs

9. **Configure Redis connection** — Add Redis config and connectivity in the init flow (needed for queue)

10. **Run `go build ./...` and `go vet ./...`** — Verify the project compiles and fix any compilation errors

---

> **End of Analysis**  
> *This document was generated by analyzing the actual source code in the repository. It represents the factual current state of the project with specific recommendations for completion.*
