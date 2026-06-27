# API URL & CORS Debug Report

## 1. Root Cause of `/v1/v1`

The double `/v1/v1` in API URLs was caused by the prefix being present in **two places simultaneously**:

| Layer | Before Fix | After Fix |
|---|---|---|
| Backend route mount | `/v1` (correct) | `/v1` (unchanged) |
| Swagger `@BasePath` | `/v1` (correct) | `/v1` (unchanged) |
| Swagger `@Router` annotations | `/v1/admin/...` ❌ | `/admin/...` ✅ |
| Frontend `baseURL` | `http://localhost:9002/v1` | `http://localhost:9002/v1` (unchanged) |
| Frontend endpoint paths | `/v1/admin/...` ❌ | `/admin/...` ✅ |
| Legacy axios `baseURL` | `http://localhost:9002/v1` | `http://localhost:9002/v1` (unchanged) |
| Legacy axios endpoint paths | `/v1/admin/...` ❌ | `/admin/...` ✅ |

**The rule violated:** Only `BasePath`/`baseURL` should include `/v1`. All `@Router` annotations and endpoint path definitions must be relative (no `/v1` prefix).

## 2. Files Changed

| File | Change |
|---|---|
| 14 handler files in `api/v1/handlers/*.go` | Removed `/v1` prefix from 95 `@Router` annotations |
| `src/api/services/admin.ts` | Endpoint paths: `/v1/admin/...` → `/admin/...` |
| `src/api/services/templates.ts` | `super('/v1/templates')` → `super('/templates')` |
| `src/api/services/notifications.ts` | `super('/v1/notifications')` → `super('/notifications')` |
| `src/api/services/preferences.ts` | `super('/v1/preferences')` → `super('/preferences')` |
| `notifier/backend/docs/` | Swagger regenerated |

## 3. Current Backend Route Structure

```
/v1/health
/v1/ready
/v1/sms/...
/v1/notifications/...
/v1/preferences/...
/v1/templates/...
/v1/inapp/...
/v1/reminders/...
/v1/providers/...
/v1/dashboard/...
/v1/observability/...
/v1/deliveries/...
/v1/me/notifications/...
/v1/me/preferences/...
/v1/me/reminders/...
/v1/admin/notifications/...
/v1/admin/templates/...
/v1/admin/reminders/...
/v1/admin/preferences/...
/v1/admin/providers/...
/v1/admin/deliveries/...
/v1/admin/observability/...
/v1/admin/dashboard/overview
/v1/service/notifications/...   (service-to-service)
/ws
/metrics
/swagger/*
```

## 4. CORS Config

| Setting | Value |
|---|---|
| `CORS_ALLOW_ORIGINS` | `*` (wildcard — allow all in dev) |
| `CORS_ALLOW_METHODS` | `GET,POST,PUT,PATCH,DELETE,OPTIONS` |
| `CORS_ALLOW_HEADERS` | `Origin,Content-Type,Accept,Authorization,X-Request-Id,X-Tenant-Id` |
| CORS middleware position | Before auth middleware ✅ (runs at app level, auth is inside route groups) |

### CORS Middleware Order (api.go)

```
1. RequestID
2. StructuredLogger
3. SecurityHeaders
4. RequestValidation
5. Prometheus
6. Tracing
7. CORS          ← before auth
8. RateLimiter
9. Recover
10. Tenant
11. WebSocket upgrade
12. RegisterRoutes (auth middleware is inside here)
```

CORS runs before auth middleware, so `OPTIONS` preflight requests will be handled before JWT validation.

## 5. Swagger Validation

- `@BasePath /v1` ✅
- `@Router /admin/...` (no `/v1` prefix) ✅ — verified 0 remaining instances
- Generated swagger.json: `0` occurrences of `/v1/v1` ✅

## 6. Frontend Validation

- `NEXT_PUBLIC_NOTIFIER_API_BASE_URL` default: `http://localhost:9002/v1` ✅
- Endpoint paths: `/admin/dashboard/overview`, `/me/notifications` (no `/v1` prefix) ✅
- Legacy axios services: no `/v1` in endpoint paths ✅
- TypeScript check: 0 errors ✅
- Backend build: passed ✅
