# Release Checklist — Minisource Notifier

> Use this checklist before each production deployment.

## Pre-Release

- [ ] Version bumped in CHANGELOG.md
- [ ] CHANGELOG.md updated with release notes
- [ ] Swagger regenerated (`make swagger`)
- [ ] All tests pass (`make test`)
- [ ] All vet checks pass (`make vet`)
- [ ] Build passes (`make build`)
- [ ] Docker image builds (`make docker-build`)
- [ ] `.env.example` matches current config
- [ ] No secrets in committed files

## Database

- [ ] Database backup completed
- [ ] Migration dry-run on staging
- [ ] Pending migrations reviewed
- [ ] Rollback plan for each migration
- [ ] Connection pool settings reviewed

## Security

- [ ] AUTH_ENABLED=true in production config
- [ ] AUTH_JWT_SECRET is a strong random value
- [ ] CORS origins restricted to known domains
- [ ] Rate limiting enabled
- [ ] Security headers verified
- [ ] No provider secrets in config or code
- [ ] PII sanitization verified in audit logs

## Observability

- [ ] Health endpoint returns expected status
- [ ] Readiness endpoint passes
- [ ] Metrics endpoint accessible
- [ ] Logging level set to `info` or `warn`
- [ ] Logging output goes to stdout (for Docker)
- [ ] Distributed tracing verified if enabled

## Deployment

- [ ] Docker image pushed to registry
- [ ] Environment variables configured
- [ ] Database host/credentials correct
- [ ] Auth service URL correct
- [ ] Provider credentials configured in DB
- [ ] Worker config reviewed for expected load

## Post-Deployment

- [ ] Health check passes
- [ ] Dashboard overview returns real data
- [ ] Admin can list/retry notifications
- [ ] Service can create notifications
- [ ] Rate limits active (trigger 429 with high request rate)
- [ ] Error responses include requestId
- [ ] Monitor logs for errors
- [ ] Rollback procedure documented and tested

## Known Limitations (this release)

- Dashboard/observability metrics scan first 100 records (not full COUNT queries)
- No dynamic worker heartbeat tracking
- Rate limiter is in-memory (single instance only)
- No DB audit log persistence (structured log only)
- Provider test is always dry-run
