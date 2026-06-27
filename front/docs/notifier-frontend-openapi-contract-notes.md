# Notifier Frontend — OpenAPI Contract Notes

> Last updated: June 2024
> Source: Backend planning documents + frontend DTOs

## OpenAPI Source

The backend Swagger/OpenAPI specification is expected at one of:
- `{backend_url}/swagger/doc.json` — Go Echo swagger
- `{backend_url}/openapi.json` — if configured
- `docs/swagger.json` — in backend project root

The frontend types were hand-authored based on the backend planning documents and endpoint implementation matrix.

## Endpoints Validated

### Admin Dashboard

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/admin/dashboard/overview` | GET | ✅ Aligned — `DashboardOverview` | Complex DTO with nested breakdowns |
| `/admin/observability/health` | GET | ✅ Aligned — `ObservabilityHealth` | Includes `dependencies` array |
| `/admin/observability/readiness` | GET | ✅ Aligned — `ReadinessResult` | Simple checks array |
| `/admin/observability/metrics` | GET | ✅ Aligned — `ObservabilityMetrics` | Large DTO with queue+workers nested |
| `/admin/observability/queue` | GET | ✅ Aligned — `QueueOverview` | |
| `/admin/observability/workers` | GET | ✅ Aligned — `WorkerOverview` | |

### Admin Notifications

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/admin/notifications` | GET | ✅ Aligned — `PaginatedResponse<Notification>` | Pagination shape assumed |
| `/admin/notifications/{id}` | GET | ✅ Aligned — `Notification` | Large DTO |
| `/admin/notifications/{id}/retry` | POST | ✅ Aligned — returns `Notification` | |
| `/admin/notifications/{id}/cancel` | POST | ✅ Aligned — returns void | |
| `/admin/notifications/{id}/read` | POST | ✅ Aligned | |
| `/admin/notifications/{id}/seen` | POST | ✅ Aligned | |
| `/admin/notifications/{id}/click` | POST | ✅ Aligned | |
| `/admin/notifications/{id}/attempts` | GET | ✅ Aligned — `DeliveryAttempt[]` | |
| `/admin/notifications/{id}/deliveries` | GET | ✅ Aligned — `NotificationDelivery[]` | |

### Admin Deliveries

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/admin/deliveries` | GET | ✅ Aligned — `PaginatedResponse<NotificationDelivery>` | |
| `/admin/deliveries/{id}` | GET | ✅ Aligned — `NotificationDelivery` | |
| `/admin/deliveries/{id}/retry` | POST | ✅ Aligned — returns `Notification` | Confirmed with backend |

### Admin Providers

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/admin/providers` | GET | ✅ Aligned — `Provider[]` | |
| `/admin/providers/health` | GET | ✅ Aligned — `ProviderHealth[]` | |
| `/admin/providers/{id}/test` | POST | ✅ Aligned — `ProviderTestResult` | |

### Admin Templates

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/admin/templates` | GET | ✅ Aligned — `Template[]` (array, not paginated) | Verify backend response shape |
| `/admin/templates/{id}` | GET | ✅ Aligned — `Template` | |
| `/admin/templates` | POST | ✅ Aligned — `CreateTemplateInput` → `Template` | |
| `/admin/templates/{id}` | PUT | ✅ Aligned — `UpdateTemplateInput` → `Template` | |
| `/admin/templates/{id}` | DELETE | ✅ Aligned — void | |
| `/admin/templates/render-preview` | POST | ✅ Aligned — `RenderPreviewInput` → `RenderPreviewResult` | |
| `/admin/templates/{id}/render-preview` | POST | ✅ Aligned | Alternative with templateId in path |
| `/admin/templates/{id}/status` | PATCH | ✅ Aligned — `Template` | |

### Admin Reminders

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/admin/reminders` | GET | ✅ Aligned — `PaginatedResponse<Reminder>` | |
| `/admin/reminders/{id}` | GET | ✅ Aligned — `Reminder` | |
| `/admin/reminders` | POST | ✅ Aligned — `CreateReminderInput` → `Reminder` | |
| `/admin/reminders/{id}` | PUT | ✅ Aligned — `UpdateReminderInput` → `Reminder` | |
| `/admin/reminders/{id}/cancel` | POST | ✅ Aligned — returns `Reminder` | |
| `/admin/reminders/{id}` | DELETE | ✅ Aligned — void | |

### Me Notifications

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/me/notifications` | GET | ✅ Aligned — `PaginatedResponse<Notification>` | |
| `/me/notifications/unread` | GET | ✅ Aligned — `Notification[]` | |
| `/me/notifications/unread-count` | GET | ✅ Aligned — `{ count: number }` | |
| `/me/notifications/{id}` | GET | ✅ Aligned — `Notification` | |
| `/me/notifications/{id}/read` | POST | ✅ Aligned — void | |
| `/me/notifications/{id}/seen` | POST | ✅ Aligned — void | |
| `/me/notifications/{id}/click` | POST | ✅ Aligned — void | |
| `/me/notifications/read-all` | POST | ✅ Aligned — void | |

### Me Preferences

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/me/preferences` | GET | ✅ Aligned — `UserPreference` | |
| `/me/preferences` | PUT | ✅ Aligned — `UpdatePreferenceInput` → `UserPreference` | |
| `/me/preferences/channel/{channel}` | PATCH | ✅ Aligned | |
| `/me/preferences/category/{category}` | PATCH | ✅ Aligned | |

### Me Reminders

| Endpoint | Method | Frontend Type Status | Notes |
|----------|--------|---------------------|-------|
| `/me/reminders` | GET | ✅ Aligned — `Reminder[]` (array) | Verify single vs paginated |
| `/me/reminders/{id}` | GET | ✅ Aligned — `Reminder` | |
| `/me/reminders` | POST | ✅ Aligned | |
| `/me/reminders/{id}` | PUT | ✅ Aligned | |
| `/me/reminders/{id}/cancel` | POST | ✅ Aligned | |
| `/me/reminders/{id}` | DELETE | ✅ Aligned — void | |

## DTO Mismatches & Potential Issues

### 1. Notification `mockNotifications` uses `any[]`
The mock data in `notifier-mocks.ts` uses `const mockNotifications: any[]` — this bypasses TypeScript type checking. Should use `Notification[]` with partial mock data. This is a type safety gap.

### 2. Pagination shape for `/admin/templates`
Frontend expects `Template[]` (array) from `/admin/templates`. If backend returns paginated response, the type will mismatch. Verify backend response shape.

### 3. Pagination shape for `/me/reminders`
Frontend expects `Reminder[]` (array) from `/me/reminders`. Verify backend.

### 4. Error response shape
The frontend `ApiError` parser expects:
```json
{ "error": { "code": "...", "message": "..." }, "requestId": "..." }
```
If backend uses a different error shape (e.g., flat `{ "code": "...", "message": "..." }` without wrapping `error` key), the parser will fail.

### 5. Status enum compatibility
Frontend supports both `cancelled` and `canceled` in `NotificationStatus` type. Backend may use only one.

### 6. Optional timestamp fields
Many timestamp fields (`sentAt`, `deliveredAt`, `seenAt`, `readAt`, `clickedAt`, `failedAt`) are optional in frontend types. Backend must tolerate null/undefined for these.

## Recommended Fixes

1. **Type mock data properly** — Change `mockNotifications` from `any[]` to `Notification[]` or a partial type
2. **Verify template list shape** — Check if backend returns array or paginated for `/admin/templates`
3. **Verify me reminders shape** — Check if backend returns array or paginated for `/me/reminders`
4. **Test error parsing** — Verify `ApiError.fromResponse()` works with actual backend error responses
5. **Add tolerance for both paginated and array responses** — Create a normalization function if backend is inconsistent

## Backend Auth Assumptions

The frontend HTTP client sends:
- `Authorization: Bearer <token>` (from mock session)
- `X-Tenant-Id` header
- `X-Project-Id` header
- `X-Request-Id` header (per-request UUID)

Backend must accept these headers for auth to work. If backend uses a different auth mechanism (e.g., API key header instead of Bearer token), the HTTP client needs adjustment.
