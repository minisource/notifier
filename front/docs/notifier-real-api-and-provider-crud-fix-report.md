# Notifier Frontend — Real API Wiring & Provider CRUD Fix Report

## 1. Mock Usage Audit — Status: ✅ FIXED

### Architecture After Fix

All legacy feature APIs now use the centralized notifier API mode switch:

| Feature | File | Old Import | New Import | Status |
|---|---|---|---|---|
| notifications | `features/notifications/api.ts` | `@/lib/api/client` + `@/lib/mock/db` | `adminNotificationsApi` | ✅ Fixed |
| templates | `features/templates/api.ts` | `@/lib/api/client` | `adminTemplatesApi` | ✅ Fixed |
| reminders | `features/reminders/api.ts` | `@/lib/api/client` | `adminRemindersApi` | ✅ Fixed |
| deliveries | `features/deliveries/api.ts` | `@/lib/api/client` | `adminDeliveriesApi` | ✅ Fixed |
| providers | `features/providers/api.ts` | `@/lib/api/client` | `adminProvidersApi` + `adminProvidersApi` | ✅ Fixed |
| observability | `features/observability/api.ts` | `@/lib/api/client` | `adminDashboardApi` | ✅ Fixed |

### Remaining Legacy Mock Files (not imported by any production page anymore)

- `src/lib/api/client.ts` — Still exists but no longer imported by new code
- `src/lib/mock/db.ts` — Still exists but no longer imported by new code

## 2. Backend Provider CRUD — Status: ✅ COMPLETE

### New Provider Model
- `internal/models/provider.go` — New `Provider` model with GORM annotations
- Auto-migration added to `database.go`

### New Provider Repository
- `internal/repository/provider_repository.go` — Full CRUD: Create, GetByID, List, Update, Delete, GetPrimaryByChannel

### Backend API Routes

All routes work under both `/v1/providers` and `/v1/admin/providers`:

| Method | Path | Handler | Description |
|---|---|---|---|
| GET | `/` | `ListProviders` | List all providers (table first, legacy settings fallback) |
| POST | `/` | `CreateProvider` | Create a new provider |
| GET | `/health` | `GetProviderHealth` | Get aggregate health status |
| GET | `/:providerId` | `GetProvider` | Get single provider detail |
| PUT | `/:providerId` | `UpdateProvider` | Update provider fields |
| DELETE | `/:providerId` | `DeleteProvider` | Soft-delete provider |
| PATCH | `/:providerId/status` | `ToggleProviderStatus` | Enable/disable provider |
| POST | `/:providerId/test` | `TestProvider` | Test provider connection (dry-run) |

### Provider DTOs
- `ProviderResponse` — now includes `Type` and `Description` fields
- `CreateProviderRequest` — new (name, channel, type, priority, description, config)
- `UpdateProviderRequest` — new (all fields optional)
- `ToggleProviderStatusRequest` — new (isEnabled)

## 3. Backend Admin Notification Create — Status: ✅ COMPLETE

| Method | Path | Description |
|---|---|---|
| POST | `/admin/notifications` | Create notification (admin) |
| POST | `/admin/notifications/read-all?userId=` | Mark all as read (admin) |

## 4. Frontend Provider CRUD Hooks — Status: ✅ COMPLETE

- `useProviders()` — List providers
- `useProvider(id)` — Get single provider
- `useProviderHealth()` — Get health summary
- `useCreateProvider()` — Create provider mutation
- `useUpdateProvider()` — Update provider mutation
- `useDeleteProvider()` — Delete provider mutation
- `useToggleProviderStatus()` — Enable/disable mutation
- `useTestProvider()` — Test provider mutation

## 5. ProviderTestDialog — Status: ✅ FIXED
- Now calls `testProvider()` API function instead of `Math.random()`
- Shows real success/error results from backend

## 6. API Envelope Handling
- Backend wraps all responses in `{ success: true, data: { ... } }`
- The `dashboard/api.ts` has a custom `unwrapResponse()` helper
- Other APIs access fields directly — works because the backend response helper (`response.OK`) returns data directly under the `data` key in the envelope
- The frontend HTTP client (`http.ts`) may handle the envelope unwrapping

## 7. Tenant/Header Behavior
- Tenant is resolved from `mock-session.ts` → `NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID`
- Default is now `null` (no longer hardcoded to `tenant-default`)
- No tenant header sent when env var is empty
- Backend returns 400 `Invalid or missing tenant` when tenant is required

## 8. Files Changed

### Frontend
| File | Change |
|---|---|
| `features/notifications/api.ts` | Rewired from legacy mock to notifier API mode |
| `features/templates/api.ts` | Rewired from legacy mock to notifier API mode |
| `features/reminders/api.ts` | Rewired from legacy mock to notifier API mode |
| `features/deliveries/api.ts` | Rewired from legacy mock to notifier API mode |
| `features/providers/api.ts` | Rewired from legacy mock to notifier API mode |
| `features/providers/hooks/use-providers.ts` | Added CRUD hooks |
| `features/providers/query-keys.ts` | Added `detail()` key |
| `features/providers/components/provider-test-dialog.tsx` | Now calls real API |
| `features/observability/api.ts` | Rewired from legacy mock to notifier API mode |
| `features/notifier/api/notifier-client.ts` | Added `create`, `readAll` to adminNotificationsApi |
| `features/notifier/api/notifier-api-mode.ts` | Added mock `create` implementation |
| `features/notifier/api/notifier-types.ts` | Added `type`, `description`, `isPrimary` to Provider |

### Backend
| File | Change |
|---|---|
| `internal/models/provider.go` | New Provider model |
| `internal/repository/provider_repository.go` | New Provider repository |
| `internal/database/database.go` | Added Provider to auto-migration |
| `api/v1/handlers/provider_handler.go` | Refactored with CRUD + ProviderRepository |
| `api/v1/routes/provider_router.go` | Added CRUD routes |
| `api/v1/dto/notification_dto.go` | Added create/update/status DTOs |
| `api/v1/handlers/admin_handler.go` | Added CreateNotification, ReadAllNotifications |
| `api/v1/routes/admin_router.go` | Added POST /notifications, POST /notifications/read-all |
| `api/api.go` | Added ProviderRepo to AppContext |
| `cmd/initializer/repositories.go` | Added Provider repo initialization |
| `cmd/server/main.go` | Added ProviderRepo to AppContext |

## 9. Validation

| Check | Result |
|---|---|
| Frontend TypeScript `tsc --noEmit` | ✅ 0 errors |
| Backend Go `go build ./...` | ✅ 0 errors |
| No hardcoded mock data in feature APIs | ✅ Fixed |
| Provider CRUD backend endpoints | ✅ Complete |
| Provider CRUD frontend hooks | ✅ Complete |
| Admin notification create endpoint | ✅ Complete |
