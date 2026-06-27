# Backend Route 404 & CORS Fix Report

## 1. Evidence

### Failing curl output
```
GET /v1/admin/dashboard/overview

HTTP/1.1 404 Not Found
X-Request-Id: 1ca5d56d-ec19-4ee5-bedf-64a852ba4bd8
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With
Access-Control-Allow-Methods: POST, GET, OPTIONS, PUT, DELETE, UPDATE

Cannot GET /v1/admin/dashboard/overview
```

## 2. Root Cause — Route 404

**Dashboard routes were registered as a separate top-level group under `/v1`, NOT nested under `/admin`.**

In `api.go` (`RegisterRoutes`), the dashboard was registered as:
```go
// Separate group — produces /v1/dashboard/overview
dashboard := v1.Group("/dashboard")
routers.Dashboard(dashboard, dashboardHandler)
```

But the frontend calls `/admin/dashboard/overview` (via baseURL `/v1` + endpoint `/admin/dashboard/overview`).

The correct path is `/v1/admin/dashboard/overview`, which means the dashboard endpoint must be registered **under the `/admin` group**.

## 3. Root Cause — CORS Error

Two issues:

1. **`Access-Control-Allow-Origin: *` with `Access-Control-Allow-Credentials: true`** — This combination is invalid per CORS spec. Browsers reject it.

2. **Missing required headers** — `X-Tenant-Id`, `X-Project-Id`, `X-Request-Id`, `Idempotency-Key` were not in allowed headers. `PATCH` method was missing. `UPDATE` was present (non-standard).

3. **Comma-separated origins in single header** — The custom middleware set the raw config string (e.g., `http://localhost:3000,http://127.0.0.1:3000`) directly as `Access-Control-Allow-Origin`, which browsers reject. Must echo back the specific request origin.

## 4. Backend Route Mount Structure (After Fix)

```
/v1
├── /health              → public
├── /sms                 → public (legacy)
├── /notifications       → JWT required
├── /preferences         → JWT required
├── /templates           → JWT + admin role
├── /inapp               → JWT required
├── /reminders           → JWT required
├── /providers           → JWT required
├── /dashboard           → JWT required (/overview)
├── /observability       → JWT required
├── /deliveries          → JWT required
├── /me                  → JWT required (userId from token)
├── /service             → Service token required
│   └── /notifications   → Scope: notifications:send
├── /admin               → JWT + admin/super_admin role
│   ├── /notifications
│   ├── /templates
│   ├── /preferences
│   ├── /reminders
│   ├── /providers
│   ├── /deliveries
│   ├── /observability
│   └── /dashboard       → NEW: /overview
└── /ws                  → WebSocket
```

## 5. Files Changed

| File | Change |
|---|---|
| `go-common/http/middleware/cors.go` | Rewritten CORS middleware: origin echoing, no `*`+credentials, correct headers/methods |
| `notifier/backend/api/api.go` | Added dashboard overview under `/admin` group (both auth branches) + route dump on startup |
| `notifier/backend/config/config.go` | Default CORS origins changed from `*` to explicit dev origins |

## 6. Final CORS Config

**Backend defaults (config.go):**
```
CORS_ALLOW_ORIGINS=http://localhost:3000,http://127.0.0.1:3000,http://localhost:3001,http://127.0.0.1:3001
CORS_ALLOW_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOW_HEADERS=Authorization,Content-Type,Accept,Origin,X-Tenant-Id,X-Project-Id,X-Request-Id,Idempotency-Key
CORS_EXPOSE_HEADERS=X-Request-Id
CORS_ALLOW_CREDENTIALS=false
```

**Middleware behavior:**
- If `allowOrigins` is `*` → sets `Access-Control-Allow-Origin: *`, no credentials header
- If `allowOrigins` is a list → echoes back the request's `Origin` header when it matches the allowed list, with `Access-Control-Allow-Credentials: true`
- Non-matching origins → no `Access-Control-Allow-Origin` header → browser blocks request
- OPTIONS preflight → returns 204 with all CORS headers set

**Middleware order:** CORS runs as app-level middleware (before route-level auth middleware), so OPTIONS preflight is never blocked by auth.

## 7. Swagger Changes

- Swagger `@BasePath /v1` (unchanged, already correct)
- Dashboard handler annotation already uses `@Router /admin/dashboard/overview [get]` (no `/v1` prefix — correct)
- Swagger regenerated successfully

## 8. Validation Results

| Check | Result |
|---|---|
| Backend build (`go build ./cmd/server`) | ✅ Passed |
| Swagger regeneration (`swag init`) | ✅ Passed — basePath: /v1 |
| Frontend TypeScript (`tsc --noEmit`) | ✅ 0 errors |
| Frontend lint (`npm run lint`) | ✅ 0 errors |
| CORS middleware no `*`+credentials bug | ✅ Fixed |
| CORS middleware no origin-fallback bug | ✅ Fixed |
| Dashboard route under `/admin` group | ✅ Added |
| Route dump on startup | ✅ Added (development mode only) |

## 9. Remaining Limitations

1. **curl validation not performed** — Requires starting the backend with the new code and running curl commands. This was not done because the backend requires a database connection and full service initialization.
2. **Frontend Network tab validation not done** — Requires both backend and frontend running simultaneously.
3. **Route dump output not captured** — The route dump runs at startup; output will appear in the backend logs when the server is started.
