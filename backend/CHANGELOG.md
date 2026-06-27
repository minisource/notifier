# Changelog

All notable changes to the Notifier service will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

#### Phase 6 — Final QA, Integration, SDK, CI/CD, Release Preparation
- Endpoint implementation matrix documenting all 110 registered routes
- Integration scenarios documentation (10 scenarios covering all API groups)
- Error codes reference with examples
- API client/SDK generation guide (TypeScript, Go, cURL)
- Comprehensive configuration documentation for all env vars
- Database documentation (tables, indexes, migration strategy)
- Production readiness report with full posture analysis
- Release checklist for deployment prep
- HTTP API examples (REST Client / Postman)
- Smoke test script covering health, admin, user, and service flows
- GitHub Actions CI workflow (test, vet, build, Swagger, Docker)
- CHANGELOG.md with all phase summaries
- Updated `.env.example` with rate limit, CORS extended fields
- Updated `Makefile` with vet, swagger, validate targets
- Updated `README.md` with complete API group overview, Docker usage, and production notes

#### Phase 5 — Dashboard, Observability, Metrics, Production Hardening
- Request ID middleware (X-Request-Id generation, propagation, logging)
- Rate limiting middleware (configurable sliding-window per-IP)
- Centralized PII/sanitization helpers (email/phone masking, provider response sanitization)
- Config validation at startup (warns in dev, fails in production)
- Production readiness checklist (17 categories, 80+ checks)
- Enhanced dashboard with channel/status breakdown, daily trend, provider health
- Enhanced observability with dependency health checks, typed readiness, enhanced metrics

#### Phase 4 — Deliveries, Providers, Queue, Retry, Dead-letter
- Provider list/health/test endpoints (DB-backed config)
- Delivery list/detail/retry endpoints (notification as delivery unit)
- Queue overview (status counts, oldest pending, next retry)
- Worker overview (configured pool info)
- Notification attempts/deliveries from NotificationLog
- Admin routes for providers, deliveries, observability under /admin

#### Phase 3.5 — User/Admin/Internal API Separation + RBAC
- /me API group (18 endpoints, userId from JWT only)
- /admin API group (28 admin endpoints)
- Auth context helpers (GetCurrentUserID, IsAdmin, IsService)
- Self-or-admin access control middleware
- Service-only notification creation with scope validation

#### Phase 3 — Templates, Preferences, Reminders
- Template CRUD with variable substitution and preview
- User notification preferences with channel/category settings
- Reminder scheduling and lifecycle management

#### Phase 2 — Notifications Core
- Notification creation with idempotency
- Batch notification support
- In-app notification support with WebSocket broadcast
- Worker pool for async processing
- Retry mechanism with exponential backoff

#### Phase 0/1 — API Contract, Swagger, DTO
- Standard ErrorResponse and PaginatedResponse DTOs
- Swagger documentation for all endpoints
- Standardized error codes

### Changed
- Upgraded from basic REST to full API access model (/me + /admin separation)
- Enhanced from basic middleware to production-grade (rate limiting, request ID, sanitization)
- Dashboard from partial snapshot to comprehensive aggregation

### Fixed
- Build errors: unused imports, type mismatches, nonexistent fields
- Auth context: config struct fields, import aliases
- Provider handler: non-existent config refs → DB-backed lookup

### Security
- PII masking (email, phone, metadata, provider responses)
- Rate limiting protection
- Request ID propagation for audit trail
- Security headers (X-Content-Type-Options, X-Frame-Options, Referrer-Policy)
- No secrets in logs, Swagger, or API responses

### Documentation
- API endpoint implementation matrix (110 routes)
- Integration scenarios (10 documented flows)
- Error codes reference with examples
- Configuration reference for all 60+ env vars
- Database schema documentation
- Production readiness checklist
- API client generation guide
- HTTP API examples (cURL / REST Client)
- Production readiness report

### Known Limitations
- Dashboard/metrics scan first 100 records (not full COUNT queries)
- No dynamic worker heartbeat tracking
- Rate limiter is in-memory (single instance only)
- Provider test always dry-run
- No DB audit log model
- PII masking not retroactive to existing DB data
