# Notifier Frontend — API Endpoint Matrix

> Generated: June 2024
> Source: Frontend API clients + planning docs

## Legend

| Column | Meaning |
|--------|---------|
| API Group | `admin`, `me`, `legacy`, `internal` |
| Status | `wired` = connected to pages via hooks, `partial` = hooks exist but not wired to pages, `mock-only` = only mock fallback available |
| Auth Type | `session` = uses auth adapter token/headers |

---

## Dashboard

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| Overview | /dashboard | GET | `/admin/dashboard/overview` | `/admin/dashboard/overview` | admin | session | admin/operator | mock-only | Hook: `useAdminDashboardOverview`, not wired to page |
| Health | /dashboard | GET | `/admin/observability/health` | `/admin/observability/health` | admin | session | admin/operator | mock-only | Hook: `useAdminHealth`, not wired |
| Queue | /dashboard | GET | `/admin/observability/queue` | `/admin/observability/queue` | admin | session | admin/operator | mock-only | Hook: `useAdminQueueOverview` |
| Workers | /dashboard | GET | `/admin/observability/workers` | `/admin/observability/workers` | admin | session | admin/operator | mock-only | Hook: `useAdminWorkersOverview` |
| Provider Health | /dashboard | GET | `/admin/providers/health` | `/admin/providers/health` | admin | session | admin/operator | mock-only | Hook: `useAdminProviderHealth` |

## Notifications (Admin)

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| List | /notifications | GET | `/admin/notifications` | `/admin/notifications` | admin | session | admin/operator | mock-only | Hook: `useAdminNotifications` |
| Detail | /notifications/[id] | GET | `/admin/notifications/{id}` | `/admin/notifications/{id}` | admin | session | admin/operator | mock-only | Hook: `useAdminNotification` |
| Retry | /notifications/[id] | POST | `/admin/notifications/{id}/retry` | `/admin/notifications/{id}/retry` | admin | session | admin/operator | hook exists | Mutation: `useRetryNotification` |
| Cancel | /notifications/[id] | POST | `/admin/notifications/{id}/cancel` | `/admin/notifications/{id}/cancel` | admin | session | admin/operator | hook exists | Mutation: `useCancelNotification` |
| Mark Read | - | POST | `/admin/notifications/{id}/read` | `/admin/notifications/{id}/read` | admin | session | admin/operator | client exists | In `notifier-client.ts` |
| Mark Seen | - | POST | `/admin/notifications/{id}/seen` | `/admin/notifications/{id}/seen` | admin | session | admin/operator | client exists | In `notifier-client.ts` |
| Mark Clicked | - | POST | `/admin/notifications/{id}/click` | `/admin/notifications/{id}/click` | admin | session | admin/operator | client exists | In `notifier-client.ts` |
| Attempts | /notifications/[id] | GET | `/admin/notifications/{id}/attempts` | `/admin/notifications/{id}/attempts` | admin | session | admin/operator | hook exists | `useAdminNotificationAttempts` |
| Deliveries | /notifications/[id] | GET | `/admin/notifications/{id}/deliveries` | `/admin/notifications/{id}/deliveries` | admin | session | admin/operator | hook exists | `useAdminNotificationDeliveries` |

## Deliveries (Admin)

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| List | /deliveries | GET | `/admin/deliveries` | `/admin/deliveries` | admin | session | admin/operator | mock-only | Hook: `useAdminDeliveries` |
| Detail | /deliveries/[id] | GET | `/admin/deliveries/{id}` | `/admin/deliveries/{id}` | admin | session | admin/operator | mock-only | Hook: `useAdminDelivery` |
| Retry | /deliveries/[id] | POST | `/admin/deliveries/{id}/retry` | `/admin/deliveries/{id}/retry` | admin | session | admin/operator | hook exists | Mutation: `useRetryDelivery` |

## Providers (Admin)

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| List | /providers | GET | `/admin/providers` | `/admin/providers` | admin | session | admin/operator | mock-only | Hook: `useAdminProviders` |
| Health | /providers | GET | `/admin/providers/health` | `/admin/providers/health` | admin | session | admin/operator | mock-only | Hook: `useAdminProviderHealth` |
| Test | /providers | POST | `/admin/providers/{id}/test` | `/admin/providers/{id}/test` | admin | session | admin/operator | hook exists | Mutation: `useTestProvider` |

## Templates (Admin)

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| List | /templates | GET | `/admin/templates` | `/admin/templates` | admin | session | admin/operator | mock-only | Hook: `useAdminTemplates` |
| Detail | /templates/[id] | GET | `/admin/templates/{id}` | `/admin/templates/{id}` | admin | session | admin/operator | mock-only | Hook: `useAdminTemplate` |
| Create | /templates/new | POST | `/admin/templates` | `/admin/templates` | admin | session | admin/operator | hook exists | Mutation: `useCreateTemplate` |
| Update | /templates/[id] | PUT | `/admin/templates/{id}` | `/admin/templates/{id}` | admin | session | admin/operator | hook exists | Mutation: `useUpdateTemplate` |
| Delete | /templates | DELETE | `/admin/templates/{id}` | `/admin/templates/{id}` | admin | session | admin/operator | hook exists | Mutation: `useDeleteTemplate` |
| Render Preview | /templates/[id] | POST | `/admin/templates/render-preview` | `/admin/templates/render-preview` | admin | session | admin/operator | hook exists | Mutation: `useRenderTemplatePreview` |
| Status Patch | /templates/[id] | PATCH | `/admin/templates/{id}/status` | `/admin/templates/{id}/status` | admin | session | admin/operator | client exists | In `adminTemplatesApi` |

## Reminders (Admin)

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| List | /reminders | GET | `/admin/reminders` | `/admin/reminders` | admin | session | admin/operator | mock-only | Hook: `useAdminReminders` |
| Detail | /reminders/[id] | GET | `/admin/reminders/{id}` | `/admin/reminders/{id}` | admin | session | admin/operator | mock-only | Hook: `useAdminReminder` |
| Create | /reminders/new | POST | `/admin/reminders` | `/admin/reminders` | admin | session | admin/operator | hook exists | Mutation: `useCreateReminder` |
| Update | /reminders/[id] | PUT | `/admin/reminders/{id}` | `/admin/reminders/{id}` | admin | session | admin/operator | hook exists | Mutation: `useUpdateReminder` |
| Cancel | /reminders/[id] | POST | `/admin/reminders/{id}/cancel` | `/admin/reminders/{id}/cancel` | admin | session | admin/operator | hook exists | Mutation: `useCancelReminder` |
| Delete | /reminders | DELETE | `/admin/reminders/{id}` | `/admin/reminders/{id}` | admin | session | admin/operator | hook exists | Mutation: `useDeleteReminder` |

## Observability (Admin)

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| Health | /observability | GET | `/admin/observability/health` | `/admin/observability/health` | admin | session | admin/operator | mock-only | |
| Readiness | /observability | GET | `/admin/observability/readiness` | `/admin/observability/readiness` | admin | session | admin/operator | mock-only | |
| Metrics | /observability | GET | `/admin/observability/metrics` | `/admin/observability/metrics` | admin | session | admin/operator | mock-only | |
| Queue | /observability | GET | `/admin/observability/queue` | `/admin/observability/queue` | admin | session | admin/operator | mock-only | |
| Workers | /observability | GET | `/admin/observability/workers` | `/admin/observability/workers` | admin | session | admin/operator | mock-only | |

## Me — Notifications (User)

| Feature | Component | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|-----------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| List | NotificationCenter | GET | `/me/notifications` | `/me/notifications` | me | session | any | hook exists | Hook: `useMeNotifications` |
| Unread Count | NotificationCenter | GET | `/me/notifications/unread-count` | `/me/notifications/unread-count` | me | session | any | hook exists | Hook: `useMeUnreadCount` |
| Detail | NotificationCenter | GET | `/me/notifications/{id}` | `/me/notifications/{id}` | me | session | any | hook exists | Hook: `useMeNotification` |
| Mark Read | NotificationCenter | POST | `/me/notifications/{id}/read` | `/me/notifications/{id}/read` | me | session | any | wired | In `notification-center-wrapper.tsx` |
| Mark Seen | NotificationCenter | POST | `/me/notifications/{id}/seen` | `/me/notifications/{id}/seen` | me | session | any | client exists | Not wired |
| Mark Clicked | NotificationCenter | POST | `/me/notifications/{id}/click` | `/me/notifications/{id}/click` | me | session | any | client exists | Not wired |
| Read All | NotificationCenter | POST | `/me/notifications/read-all` | `/me/notifications/read-all` | me | session | any | wired | In `notification-center-wrapper.tsx` |

## Me — Preferences

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| Get | /preferences | GET | `/me/preferences` | `/me/preferences` | me | session | any | hook exists | Hook: `useMePreferences` |
| Update | /preferences | PUT | `/me/preferences` | `/me/preferences` | me | session | any | hook exists | Mutation: `useMeUpdatePreference` |
| Update Channel | /preferences | PATCH | `/me/preferences/channel/{channel}` | `/me/preferences/channel/{channel}` | me | session | any | client exists | In `mePreferencesApi` |
| Update Category | /preferences | PATCH | `/me/preferences/category/{category}` | `/me/preferences/category/{category}` | me | session | any | client exists | In `mePreferencesApi` |

## Me — Reminders

| Feature | Page | Method | Frontend Path | Backend Endpoint | API Group | Auth | Role | Status | Notes |
|---------|------|--------|---------------|-----------------|-----------|------|------|--------|-------|
| List | - | GET | `/me/reminders` | `/me/reminders` | me | session | any | hook exists | Hook: `useMeReminders` |
| Detail | - | GET | `/me/reminders/{id}` | `/me/reminders/{id}` | me | session | any | hook exists | Hook: `useMeReminder` |
| Create | - | POST | `/me/reminders` | `/me/reminders` | me | session | any | hook exists | Mutation: `useMeCreateReminder` |
| Cancel | - | POST | `/me/reminders/{id}/cancel` | `/me/reminders/{id}/cancel` | me | session | any | hook exists | Mutation: `useMeCancelReminder` |

---

## Summary

| API Group | Endpoints Defined | Hooks/Mutations Created | Wired to Pages | Mock-Only |
|-----------|------------------|------------------------|----------------|-----------|
| Admin Dashboard | 6 | 6 | 0 | 6 |
| Admin Notifications | 9 | 7 | 0 | 2 |
| Admin Deliveries | 3 | 3 | 0 | 2 |
| Admin Providers | 3 | 3 | 0 | 2 |
| Admin Templates | 7 | 6 | 0 | 2 |
| Admin Reminders | 6 | 6 | 0 | 2 |
| Admin Observability | 5 | 5 | 0 | 5 |
| Me Notifications | 7 | 4 | 3 | 0 |
| Me Preferences | 4 | 2 | 0 | 0 |
| Me Reminders | 4 | 4 | 0 | 0 |

**Key finding:** All 54 endpoints are defined in API clients, but **zero admin pages use the new API hooks**. All pages still use legacy feature hooks (`features/*/hooks/*`) backed by mock data. The hooks exist and are ready — the wiring was deferred to a future phase.
