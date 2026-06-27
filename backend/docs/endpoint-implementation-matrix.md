# Endpoint Implementation Matrix

> Generated from `notifier/backend/api/api.go` route registrations.
> Last updated: Phase 6

## Legend

| Column | Description |
|--------|-------------|
| **Status** | ✅ implemented, 🔶 partial, ❌ 501, 🚫 deprecated |
| **Auth** | public, JWT, admin, service, legacy |
| **Role** | required role for access |

---

## Public / Health

| Method | Path | Status | Auth | Role | Purpose | Response DTO |
|--------|------|--------|------|------|---------|-------------|
| GET | `/api/v1/health` | ✅ | public | - | Service health check | HealthResponse |
| GET | `/api/v1/sms` | ✅ | public | - | Legacy SMS webhook | - |

## User /me (userId from JWT only)

| Method | Path | Status | Auth | Role | Purpose | Response DTO |
|--------|------|--------|------|------|---------|-------------|
| GET | `/api/v1/me/notifications` | ✅ | JWT | user | List my notifications | PaginatedNotificationResponse |
| GET | `/api/v1/me/notifications/unread` | ✅ | JWT | user | List unread | PaginatedNotificationResponse |
| GET | `/api/v1/me/notifications/unread-count` | ✅ | JWT | user | Unread count | UnreadCountResponse |
| POST | `/api/v1/me/notifications/read-all` | ✅ | JWT | user | Mark all read | ActionResponse |
| GET | `/api/v1/me/notifications/{id}` | ✅ | JWT | user,owner | Get notification | NotificationResponse |
| PUT | `/api/v1/me/notifications/{id}/read` | ✅ | JWT | user,owner | Mark read | ActionResponse |
| POST | `/api/v1/me/notifications/{id}/seen` | ✅ | JWT | user,owner | Mark seen | ActionResponse |
| POST | `/api/v1/me/notifications/{id}/click` | ✅ | JWT | user,owner | Mark clicked | ActionResponse |
| GET | `/api/v1/me/preferences` | ✅ | JWT | user | My preferences | []PreferenceResponse |
| PUT | `/api/v1/me/preferences` | ✅ | JWT | user | Update preferences | PreferenceResponse |
| PATCH | `/api/v1/me/preferences/channel/{channel}` | ✅ | JWT | user | Channel pref | PreferenceResponse |
| PATCH | `/api/v1/me/preferences/category/{category}` | ✅ | JWT | user | Category pref | PreferenceResponse |
| GET | `/api/v1/me/reminders` | ✅ | JWT | user | My reminders | PaginatedResponse |
| POST | `/api/v1/me/reminders` | ✅ | JWT | user | Create reminder | ReminderResponse |
| GET | `/api/v1/me/reminders/{id}` | ✅ | JWT | user,owner | Get reminder | ReminderResponse |
| PUT | `/api/v1/me/reminders/{id}` | ✅ | JWT | user,owner | Update reminder | ReminderResponse |
| POST | `/api/v1/me/reminders/{id}/cancel` | ✅ | JWT | user,owner | Cancel reminder | ReminderResponse |
| DELETE | `/api/v1/me/reminders/{id}` | ✅ | JWT | user,owner | Delete reminder | ActionResponse |

## Admin /admin (admin/super_admin role)

| Method | Path | Status | Auth | Role | Purpose | Response DTO |
|--------|------|--------|------|------|---------|-------------|
| GET | `/api/v1/admin/notifications` | ✅ | JWT | admin | List all notifications | PaginatedNotificationResponse |
| GET | `/api/v1/admin/notifications/{id}` | ✅ | JWT | admin | Get notification | NotificationResponse |
| POST | `/api/v1/admin/notifications/{id}/retry` | ✅ | JWT | admin | Retry notification | ActionResponse |
| POST | `/api/v1/admin/notifications/{id}/cancel` | ✅ | JWT | admin | Cancel notification | ActionResponse |
| GET | `/api/v1/admin/notifications/{id}/attempts` | ✅ | JWT | admin | Delivery attempts | []AttemptResponse |
| GET | `/api/v1/admin/notifications/{id}/deliveries` | ✅ | JWT | admin | Deliveries | []DeliveryResponse |
| PUT | `/api/v1/admin/notifications/{id}/read` | ✅ | JWT | admin | Mark read | ActionResponse |
| POST | `/api/v1/admin/notifications/{id}/seen` | ✅ | JWT | admin | Mark seen | ActionResponse |
| POST | `/api/v1/admin/notifications/{id}/click` | ✅ | JWT | admin | Mark clicked | ActionResponse |
| GET | `/api/v1/admin/templates` | ✅ | JWT | admin | List templates | PaginatedResponse |
| POST | `/api/v1/admin/templates` | ✅ | JWT | admin | Create template | TemplateResponse |
| GET | `/api/v1/admin/templates/key/{key}` | ✅ | JWT | admin | Get by key | TemplateResponse |
| POST | `/api/v1/admin/templates/render-preview` | ✅ | JWT | admin | Render preview | RenderPreviewResponse |
| GET | `/api/v1/admin/templates/{id}` | ✅ | JWT | admin | Get template | TemplateResponse |
| PUT | `/api/v1/admin/templates/{id}` | ✅ | JWT | admin | Update template | TemplateResponse |
| DELETE | `/api/v1/admin/templates/{id}` | ✅ | JWT | admin | Delete template | map[string]interface{} |
| POST | `/api/v1/admin/templates/{id}/render-preview` | ✅ | JWT | admin | Render by ID | RenderPreviewResponse |
| PATCH | `/api/v1/admin/templates/{id}/status` | ✅ | JWT | admin | Toggle active | TemplateResponse |
| GET | `/api/v1/admin/preferences/user/{userId}` | ✅ | JWT | admin | Get user prefs | []PreferenceResponse |
| PUT | `/api/v1/admin/preferences/user/{userId}` | ✅ | JWT | admin | Update pref | PreferenceResponse |
| PATCH | `/api/v1/admin/preferences/user/{userId}/channel/{channel}` | ✅ | JWT | admin | Channel pref | PreferenceResponse |
| PATCH | `/api/v1/admin/preferences/user/{userId}/category/{category}` | ✅ | JWT | admin | Category pref | PreferenceResponse |
| GET | `/api/v1/admin/reminders` | ✅ | JWT | admin | List reminders | PaginatedResponse |
| POST | `/api/v1/admin/reminders` | ✅ | JWT | admin | Create reminder | ReminderResponse |
| GET | `/api/v1/admin/reminders/{id}` | ✅ | JWT | admin | Get reminder | ReminderResponse |
| PUT | `/api/v1/admin/reminders/{id}` | ✅ | JWT | admin | Update reminder | ReminderResponse |
| POST | `/api/v1/admin/reminders/{id}/cancel` | ✅ | JWT | admin | Cancel reminder | ReminderResponse |
| DELETE | `/api/v1/admin/reminders/{id}` | ✅ | JWT | admin | Delete reminder | ActionResponse |
| GET | `/api/v1/admin/reminders/user/{userId}` | ✅ | JWT | admin | User reminders | PaginatedResponse |
| GET | `/api/v1/admin/providers` | ✅ | JWT | admin | List providers | []ProviderResponse |
| GET | `/api/v1/admin/providers/health` | ✅ | JWT | admin | Provider health | ProviderHealthResponse |
| POST | `/api/v1/admin/providers/{id}/test` | ✅ | JWT | admin | Test provider | ProviderTestResponse |
| GET | `/api/v1/admin/deliveries` | ✅ | JWT | admin | List deliveries | PaginatedResponse |
| GET | `/api/v1/admin/deliveries/{id}` | ✅ | JWT | admin | Get delivery | DeliveryResponse |
| POST | `/api/v1/admin/deliveries/{id}/retry` | ✅ | JWT | admin | Retry delivery | ActionResponse |
| GET | `/api/v1/admin/observability/health` | ✅ | JWT | admin | Detailed health | ObservabilityHealthResponse |
| GET | `/api/v1/admin/observability/readiness` | ✅ | JWT | admin | Readiness | ObservabilityReadinessResponse |
| GET | `/api/v1/admin/observability/metrics` | ✅ | JWT | admin | Metrics | ObservabilityMetricsResponse |
| GET | `/api/v1/admin/observability/queue` | ✅ | JWT | admin | Queue overview | QueueOverviewResponse |
| GET | `/api/v1/admin/observability/workers` | ✅ | JWT | admin | Workers overview | WorkerOverviewResponse |
| GET | `/api/v1/admin/dashboard/overview` | ✅ | JWT | admin | Dashboard | DashboardOverviewResponse |

## Internal / Service (service auth + scope)

| Method | Path | Status | Auth | Role | Purpose | Response DTO |
|--------|------|--------|------|------|---------|-------------|
| POST | `/api/v1/service/notifications` | ✅ | service | notifications:send | Create notification | NotificationResponse |
| POST | `/api/v1/service/notifications/batch` | ✅ | service | notifications:send | Batch create | map[string]interface{} |
| POST | `/api/v1/service/notifications/sync` | ✅ | service | notifications:send | Sync send | NotificationResponse |

## Legacy (backward compatibility)

| Method | Path | Status | Auth | Role | Purpose | Response DTO |
|--------|------|--------|------|------|---------|-------------|
| GET | `/api/v1/notifications` | ✅ | JWT | user | List (admin list) | PaginatedNotificationResponse |
| POST | `/api/v1/notifications` | ✅ | JWT | user | Create | NotificationResponse |
| POST | `/api/v1/notifications/batch` | ✅ | JWT | user | Batch create | map[string]interface{} |
| GET | `/api/v1/notifications/{id}` | ✅ | JWT | user | Get | NotificationResponse |
| PUT | `/api/v1/notifications/{id}/read` | ✅ | JWT | user | Mark read | ActionResponse |
| POST | `/api/v1/notifications/{id}/retry` | ✅ | JWT | user | Retry | ActionResponse |
| POST | `/api/v1/notifications/{id}/cancel` | ✅ | JWT | user | Cancel | ActionResponse |
| POST | `/api/v1/notifications/{id}/seen` | ✅ | JWT | user | Mark seen | ActionResponse |
| POST | `/api/v1/notifications/{id}/click` | ✅ | JWT | user | Mark clicked | ActionResponse |
| GET | `/api/v1/notifications/user/{userId}` | ✅ | JWT | user,self | User notifications | PaginatedNotificationResponse |
| GET | `/api/v1/notifications/user/{userId}/unread` | ✅ | JWT | user,self | Unread | PaginatedNotificationResponse |
| GET | `/api/v1/notifications/user/{userId}/unread-count` | ✅ | JWT | user,self | Unread count | UnreadCountResponse |
| POST | `/api/v1/notifications/user/{userId}/read-all` | ✅ | JWT | user,self | Mark all read | MarkAllAsReadResponse |
| GET | `/api/v1/notifications/{id}/attempts` | ✅ | JWT | user | Attempts | []AttemptResponse |
| GET | `/api/v1/notifications/{id}/deliveries` | ✅ | JWT | user | Deliveries | []DeliveryResponse |
| GET | `/api/v1/preferences/user/{userId}` | ✅ | JWT | user,self | Get prefs | []PreferenceResponse |
| PUT | `/api/v1/preferences/user/{userId}` | ✅ | JWT | user,self | Update pref | PreferenceResponse |
| PATCH | `/api/v1/preferences/user/{userId}/channel/{channel}` | ✅ | JWT | user,self | Channel pref | PreferenceResponse |
| PATCH | `/api/v1/preferences/user/{userId}/category/{category}` | ✅ | JWT | user,self | Category pref | PreferenceResponse |
| GET | `/api/v1/templates` | ✅ | JWT | admin | List | PaginatedResponse |
| POST | `/api/v1/templates` | ✅ | JWT | admin | Create | TemplateResponse |
| GET | `/api/v1/templates/{id}` | ✅ | JWT | admin | Get | TemplateResponse |
| PUT | `/api/v1/templates/{id}` | ✅ | JWT | admin | Update | TemplateResponse |
| DELETE | `/api/v1/templates/{id}` | ✅ | JWT | admin | Delete | map[string]interface{} |
| GET | `/api/v1/templates/key/{key}` | ✅ | JWT | admin | Get by key | TemplateResponse |
| POST | `/api/v1/templates/render-preview` | ✅ | JWT | admin | Render preview | RenderPreviewResponse |
| POST | `/api/v1/templates/{id}/render-preview` | ✅ | JWT | admin | Render by ID | RenderPreviewResponse |
| PATCH | `/api/v1/templates/{id}/status` | ✅ | JWT | admin | Toggle active | TemplateResponse |
| GET | `/api/v1/inapp` | ✅ | JWT | user | In-app list | PaginatedNotificationResponse |
| GET | `/api/v1/reminders` | ✅ | JWT | user | List | PaginatedResponse |
| POST | `/api/v1/reminders` | ✅ | JWT | user | Create | ReminderResponse |
| GET | `/api/v1/reminders/{id}` | ✅ | JWT | user,owner | Get | ReminderResponse |
| PUT | `/api/v1/reminders/{id}` | ✅ | JWT | user,owner | Update | ReminderResponse |
| POST | `/api/v1/reminders/{id}/cancel` | ✅ | JWT | user,owner | Cancel | ReminderResponse |
| DELETE | `/api/v1/reminders/{id}` | ✅ | JWT | user,owner | Delete | ActionResponse |
| GET | `/api/v1/reminders/user/{userId}` | ✅ | JWT | user,self | User reminders | PaginatedResponse |
| GET | `/api/v1/dashboard/overview` | ✅ | JWT | user | Dashboard | DashboardOverviewResponse |
| GET | `/api/v1/observability/health` | ✅ | public | - | Public health | ObservabilityHealthResponse |
| GET | `/api/v1/observability/readiness` | ✅ | public | - | Readiness | ObservabilityReadinessResponse |
| GET | `/api/v1/observability/metrics` | ✅ | JWT | user | JSON metrics | ObservabilityMetricsResponse |
| GET | `/api/v1/observability/queue` | ✅ | JWT | user | Queue | QueueOverviewResponse |
| GET | `/api/v1/observability/workers` | ✅ | JWT | user | Workers | WorkerOverviewResponse |
| GET | `/api/v1/deliveries` | ✅ | JWT | user | Deliveries | PaginatedResponse |
| GET | `/api/v1/deliveries/{id}` | ✅ | JWT | user | Get delivery | DeliveryResponse |
| POST | `/api/v1/deliveries/{id}/retry` | ✅ | JWT | user | Retry | ActionResponse |
| GET | `/api/v1/providers` | ✅ | JWT | user | List providers | []ProviderResponse |
| GET | `/api/v1/providers/health` | ✅ | JWT | user | Health | ProviderHealthResponse |
| POST | `/api/v1/providers/{id}/test` | ✅ | JWT | user | Test | ProviderTestResponse |

## WebSocket

| Method | Path | Status | Auth | Role | Purpose |
|--------|------|--------|------|------|---------|
| WS | `/ws` | ✅ | JWT/Service | user/service | Real-time notifications |

## Metrics & Swagger

| Method | Path | Status | Auth | Role | Purpose |
|--------|------|--------|------|------|---------|
| GET | `/metrics` | ✅ | public | - | Prometheus metrics |
| GET | `/swagger/*` | ✅ | public | - | Swagger UI |

## Summary

| Group | Total | ✅ Implemented | ❌ 501 | 🔶 Partial |
|-------|-------|---------------|--------|-----------|
| Public/Health | 2 | 2 | 0 | 0 |
| User /me | 18 | 18 | 0 | 0 |
| Admin /admin | 42 | 42 | 0 | 0 |
| Internal/Service | 3 | 3 | 0 | 0 |
| Legacy | 42 | 42 | 0 | 0 |
| WebSocket | 1 | 1 | 0 | 0 |
| Metrics/Swagger | 2 | 2 | 0 | 0 |
| **Total** | **110** | **110** | **0** | **0** |
