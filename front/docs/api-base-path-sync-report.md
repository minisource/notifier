# API Base Path Sync Report

## 1. Problem

Notifier API routes were returning 404 because of a mismatch between the backend route prefix and the frontend base URL.

## 2. Current State (Before Fix)

| Component | Path | Status |
|---|---|---|
| Backend route registration | `/api` → `/v1` → routes | ❌ `/api` prefix present |
| Swagger `@BasePath` | `/api` | ❌ Wrong |
| Frontend HTTP client default | `http://localhost:9002/api/v1` | ❌ `/api` in URL |
| Frontend env default | `http://localhost:9002/api/v1` | ❌ `/api` in URL |
| Frontend docs/examples | `http://localhost:9002/api/v1` | ❌ `/api` in URL |
| Legacy axios client | `http://localhost:9002/api` | ❌ Wrong |

## 3. Chosen Canonical API Base Path

```
/v1
```

This was chosen because:
- `/api` prefix is unnecessary — the API is already distinguishable by its port and path structure
- Matches common microservice convention (`/v1/resource`)
- Cleaner, shorter URLs
- Consistent with other Minisource services

## 4. Backend Route Registration Changes

### Before
```go
api := app.Group("/api")
v1 := api.Group("/v1")
// routes under /api/v1/...
```

### After
```go
// Legacy rewrite middleware for /api/v1/* → /v1/*
app.Use("/api", func(c *fiber.Ctx) error {
    path := c.Path()
    if len(path) >= 8 && path[:8] == "/api/v1/" {
        c.Path("/v1/" + path[8:])
    } else if path == "/api/v1" {
        c.Path("/v1")
    }
    return c.Next()
})

// Canonical v1 group (no /api prefix)
v1 := app.Group("/v1")
// routes under /v1/...
```

### Legacy Compatibility
The `/api/v1` → `/v1` rewrite middleware silently rewrites old-style paths to new canonical paths. This means old clients/tools that call `/api/v1/...` will still work without changes.

## 5. Auth SkipPaths Update

Updated from `/api/v1/health` → `/v1/health` in both places:
- `AuthConfig.SkipPaths` (for JWT auth)
- `ServiceAuthConfig.SkipPaths` (for service token auth)

## 6. Swagger Changes

| Field | Before | After |
|---|---|---|
| `@BasePath` | `/api` | `/v1` |
| Generated swagger.json basePath | `/api` | `/v1` |

Swagger was regenerated with: `swag init -g cmd/server/main.go`

## 7. Frontend Changes

| File | Before | After |
|---|---|---|
| `src/shared/api/http-client.ts` default | `http://localhost:9002/api/v1` | `http://localhost:9002/v1` |
| `src/lib/config/env.ts` default | `http://localhost:9002/api/v1` | `http://localhost:9002/v1` |
| `src/features/notifier/config/notifier-config.ts` default | `http://localhost:9002/api/v1` | `http://localhost:9002/v1` |
| `src/config/index.ts` default | `http://localhost:9002/api` | `http://localhost:9002/v1` |
| `src/api/client.ts` default | `http://localhost:9002/api` | `http://localhost:9002/v1` |
| `.env` | `http://localhost:9002/api/v1` | `http://localhost:9002/v1` |
| `.env.example` | `http://localhost:9002/api/v1` | `http://localhost:9002/v1` |

Frontend endpoint definitions in `notifier-client.ts` and `me-client.ts` did NOT change — they use relative paths like `/admin/dashboard/overview` which combine with the base URL:
- Before: `http://localhost:9002/api/v1` + `/admin/dashboard/overview` = `http://localhost:9002/api/v1/admin/dashboard/overview`
- After: `http://localhost:9002/v1` + `/admin/dashboard/overview` = `http://localhost:9002/v1/admin/dashboard/overview`

## 8. Documentation Updates

| File | Updated |
|---|---|
| `docs/notifier-frontend-readme.md` | ✅ URL updated |
| `docs/notifier-frontend-real-backend-smoke-test.md` | ✅ URL and curl examples updated |
| `docs/notifier-frontend-api-matrix.md` | ✅ All 54 endpoint paths updated |

## 9. Validation Results

| Check | Result |
|---|---|
| Backend build (`go build ./api/...`) | ✅ Passed |
| Frontend TypeScript (`tsc --noEmit`) | ✅ 0 errors |
| Frontend Lint (`npm run lint`) | ✅ 0 errors (pre-existing warnings) |
| Swagger regeneration | ✅ `@BasePath /v1` confirmed |
| Legacy `/api/v1` routes | ↔️ Rewritten via middleware (no 404) |
| Canonical `/v1` routes | ✅ New canonical prefix |

## 10. Remaining Limitations

1. **Swagger UI** may still reference old paths in its embedded server URL if not manually configured
2. **Legacy clients** calling `/api/v1/...` will work via rewrite middleware but may not work with future API versions
3. The `/api/v1` rewrite middleware adds minimal overhead for every `/api/*` request
4. Some integration test files across the Minisource project may still reference `/api/v1/...`
