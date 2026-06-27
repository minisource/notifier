# Notifier Frontend — Phase 0/1 Plan

## Audit, API Integration Foundation, and Notifier Dashboard Shell

---

## 1. Current Frontend Structure Audit

The Notifier frontend (`notifier/front/`) already has a sophisticated codebase:

**Stack:** Next.js 15 (App Router) + TypeScript + Tailwind + shadcn/ui + TanStack Query + next-intl + Zustand + Zod + React Hook Form + Sonner + Lucide

**Existing structure:**
```
src/
  app/[locale]/                    # Route pages under locale
    dashboard/
    notifications/ + [id]/, new/
    templates/ + [id]/, new/
    reminders/ + [id]/, new/
    deliveries/ + [id]/
    providers/
    preferences/
    observability/
    settings/
    tenants/
  features/                        # Feature-based modules
    dashboard/  (api, hooks, components, types, query-keys)
    notifications/ (api, hooks, components, types, query-keys, schemas)
    deliveries/  (api, hooks, types, query-keys)
    providers/   (api, hooks, types, query-keys)
    templates/   (api, hooks, types, query-keys, schemas)
    reminders/   (api, hooks, types, query-keys, schemas)
    preferences/ (api, hooks, types, query-keys, schemas)
    observability/ (api, hooks, types, query-keys)
    tenants/     (api, hooks, types, query-keys)
  components/
    layout/      (app-shell, sidebar, topbar, header, mobile-nav, etc.)
    shared/      (status-badge, channel-badge, metric-card, page-header, etc.)
    ui/          (shadcn/ui components)
  lib/
    api/         (mock client + types)
    mock/        (db, session, deliveries, etc.)
    utils/       (date, direction, format, ids)
  stores/        (auth.store, ui.store)
  types/         (notifier.ts, index.ts)
  messages/      (en.json, fa.json)
```

**Current state:**

| Aspect | Status |
|--------|--------|
| App Shell (sidebar + topbar) | ✅ Complete with Persian/English nav |
| Dashboard (overview) | ✅ Real metrics, provider health, channel breakdown, recent notifications |
| Notifications (list + detail) | ✅ Complete with filters, pagination, timeline, attempts, actions |
| Send Notification (form) | ✅ Complete form with template selector, variables, scheduling |
| Templates (list) | ⚠️ Skeleton — shows "no templates" |
| Templates (detail/edit) | ⚠️ Skeleton — shows ID only |
| Templates (create) | ⚠️ Skeleton — shows title only |
| Reminders (list) | ⚠️ Skeleton — shows "no reminders" |
| Reminders (detail) | ⚠️ Skeleton — shows ID only |
| Reminders (create) | ⚠️ Skeleton — shows title only |
| Deliveries (list) | ⚠️ Skeleton — shows "no deliveries" |
| Providers (list) | ⚠️ Skeleton — shows "no providers" |
| Preferences (list) | ⚠️ Skeleton — shows title only |
| Observability | ✅ Real health, metrics, copy diagnostics |
| Settings | ✅ Theme, language, API mode display |
| Tenants | ✅ List with mock data |

## 2. Current Routing Structure

```
/[locale]/                          → Redirects to /dashboard
/[locale]/dashboard                 → Dashboard overview
/[locale]/notifications             → Notification list
/[locale]/notifications/new         → Send notification form
/[locale]/notifications/[id]        → Notification detail + timeline + attempts
/[locale]/templates                 → Template list (skeleton)
/[locale]/templates/new             → Create template (skeleton)
/[locale]/templates/[id]            → Template detail/edit (skeleton)
/[locale]/reminders                 → Reminder list (skeleton)
/[locale]/reminders/new             → Create reminder (skeleton)
/[locale]/reminders/[id]            → Reminder detail (skeleton)
/[locale]/deliveries                → Delivery list (skeleton)
/[locale]/deliveries/[id]           → Delivery detail (skeleton)
/[locale]/providers                 → Provider list (skeleton)
/[locale]/preferences               → Preferences (skeleton)
/[locale]/observability             → Observability/metrics
/[locale]/settings                  → Settings/theme/language
/[locale]/tenants                   → Tenant list
```

All pages are under `/[locale]/` directly — no `/admin/` or `/user/me/` prefix separation in routing.

## 3. Current UI / Component System

**shadcn/ui installed:**
alert-dialog, avatar, badge, button, card, command, dialog, dropdown-menu, input, label, popover, scroll-area, select, separator, sheet, skeleton, sonner, switch, table, tabs, textarea, tooltip

**Shared components:**
PageHeader, PageContainer, SectionCard, MetricCard, StatusBadge, ChannelBadge, MiniProgress, SearchInput, Pagination, ConfirmDialog, EmptyState, ErrorState, LoadingState (TableSkeleton)

## 4. Current API / Client Setup

**Mock-only mode.** All data flows through `src/lib/api/client.ts` which returns mock data with simulated delay. No real HTTP client exists.

**Architecture:**
```
lib/api/client.ts          → Mock functions returning in-memory data
features/**/api.ts         → Wraps mock functions, adds transformation
features/**/hooks/*.ts     → TanStack Query hooks calling api.ts
```

Mock data lives in `src/lib/mock/db.ts` with ~20 notifications, 12 templates, 8 reminders, 8 providers, etc.

## 5. Current i18n / RTL Status

**i18n:** next-intl with `fa` and `en` locales. Full message files for both languages covering all existing pages. The user's prompt heavily emphasizes cleaning up auth, API client, and frontend structure.

**RTL:** Supported via `getDirection()` utility. AppShell sets `dir={isRtl ? 'rtl' : 'ltr'}`. CSS variables handle RTL spacing.

## 6. Current Auth Status

**Minimal.** `src/stores/auth.store.ts` stores user/token in localStorage via Zustand persist.
**No auth verification.** `src/lib/mock/session.ts` hardcodes a mock admin session.
**No auth middleware** in frontend routes.

## 7. Required Backend Endpoints

See full backend API at `docs/endpoint-implementation-matrix.md` in the backend project.

**Admin endpoints needed:**
```
GET    /admin/dashboard/overview
GET    /admin/notifications
GET    /admin/notifications/{id}
POST   /admin/notifications/{id}/retry
POST   /admin/notifications/{id}/cancel
GET    /admin/notifications/{id}/attempts
GET    /admin/notifications/{id}/deliveries
GET    /admin/deliveries
GET    /admin/deliveries/{id}
POST   /admin/deliveries/{id}/retry
GET    /admin/providers
GET    /admin/providers/health
POST   /admin/providers/{id}/test
GET    /admin/templates
POST   /admin/templates
GET    /admin/templates/{id}
PUT    /admin/templates/{id}
DELETE /admin/templates/{id}
POST   /admin/templates/render-preview
PATCH  /admin/templates/{id}/status
GET    /admin/reminders
POST   /admin/reminders
GET    /admin/reminders/{id}
PUT    /admin/reminders/{id}
DELETE /admin/reminders/{id}
POST   /admin/reminders/{id}/cancel
GET    /admin/preferences/user/{userId}
PUT    /admin/preferences/user/{userId}
GET    /admin/observability/health
GET    /admin/observability/readiness
GET    /admin/observability/metrics
GET    /admin/observability/queue
GET    /admin/observability/workers
```

**User /me endpoints needed:**
```
GET    /me/notifications
GET    /me/notifications/unread
GET    /me/notifications/unread-count
GET    /me/notifications/{id}
POST   /me/notifications/{id}/read
POST   /me/notifications/{id}/seen
POST   /me/notifications/{id}/click
POST   /me/notifications/read-all
GET    /me/preferences
PUT    /me/preferences
PATCH  /me/preferences/channel/{channel}
PATCH  /me/preferences/category/{category}
GET    /me/reminders
POST   /me/reminders
GET    /me/reminders/{id}
PUT    /me/reminders/{id}
POST   /me/reminders/{id}/cancel
DELETE /me/reminders/{id}
```

## 8. Admin Pages Required

| Page | Route | Current Status | Target |
|------|-------|---------------|--------|
| Overview/Dashboard | /dashboard | ✅ Real | Keep |
| Notifications | /notifications | ✅ Complete | Keep |
| Notification Detail | /notifications/[id] | ✅ Complete | Keep |
| Send Notification | /notifications/new | ✅ Complete | Keep |
| Templates | /templates | ⚠️ Skeleton | Fill with real data |
| Template Detail | /templates/[id] | ⚠️ Skeleton | Fill with real data |
| Create Template | /templates/new | ⚠️ Skeleton | Fill with real data |
| Reminders | /reminders | ⚠️ Skeleton | Fill with real data |
| Reminder Detail | /reminders/[id] | ⚠️ Skeleton | Fill with real data |
| Create Reminder | /reminders/new | ⚠️ Skeleton | Fill with real data |
| Deliveries | /deliveries | ⚠️ Skeleton | Fill with real data |
| Delivery Detail | /deliveries/[id] | ⚠️ Skeleton | Fill with real data |
| Providers | /providers | ⚠️ Skeleton | Fill with real data |
| Preferences | /preferences | ⚠️ Skeleton | Fill with real data |
| Observability | /observability | ✅ Real | Keep |
| Settings | /settings | ✅ Real | Keep |

## 9. User-Facing Pages Required

| Page | Route | Note |
|------|-------|------|
| My Notifications | /me/notifications | Future phase |
| My Preferences | /me/preferences | Future phase |
| My Reminders | /me/reminders | Future phase |

These are **out of scope for Phase 0/1** since we're focused on admin dashboard first.

## 10. Mock Auth Strategy

**Current:** Hardcoded session in `src/lib/mock/session.ts`.

**Target:**
- Create `src/shared/auth/auth-types.ts` — `AuthSession` interface
- Create `src/shared/auth/auth-adapter.ts` — reads from env → localStorage → default
- Create `src/shared/auth/mock-session.ts` — default mock values
- Support `NEXT_PUBLIC_NOTIFIER_MOCK_*` env vars
- Support localStorage override for development
- No real auth implementation (no login/register)

## 11. Feature-Based Structure Proposal

Keep the existing `features/` structure. Add:
```
src/
  features/notifier/
    api/
      notifier-types.ts       # Backend-aligned TypeScript types
      notifier-client.ts      # Admin API client (real HTTP)
      me-client.ts            # User /me API client (real HTTP)
      notifier-queries.ts     # TanStack Query hooks
      notifier-mutations.ts   # TanStack Query mutations
      notifier-mocks.ts       # Mock fallback data
```

## 12. Component Inventory

**Already existing (usable as-is):**
- AppShell, Sidebar, Topbar, MobileNav
- PageHeader, PageContainer, SectionCard
- StatusBadge, ChannelBadge, MetricCard
- SearchInput, Pagination, ConfirmDialog
- EmptyState, ErrorState, LoadingState (TableSkeleton)
- NotificationSummaryCard, NotificationTimeline, NotificationAttemptsList
- NotificationActionMenu, NotificationFilters, NotificationTable
- All shadcn/ui components

**Needed (from requirements):**
- DataTable component (can use shadcn Table + TanStack Table)
- DateRangePicker component (can use shadcn Popover + Calendar)
- StatusCard component (can use MetricCard)

## 13. Data Fetching Strategy

**Phase 0/1 (this phase):**
- All pages use existing mock data via `features/*/api.ts`
- Real HTTP client created but NOT wired to all pages yet
- Mock fallback adapter created
- Auth adapter created

**Phase 2 (future):**
- Swap mock data for real API calls via the HTTP client
- Add env-flag-based mock/real switching

## 14. Error / Loading / Empty State Strategy

**Already implemented consistently across pages:**
- Loading: `LoadingState` with skeletons or `TableSkeleton`
- Error: `ErrorState` with retry button
- Empty: `EmptyState` with descriptive message
- Used in: dashboard, notifications (list + detail), observability

**Apply to skeleton pages:**
- Templates (list, detail, create)
- Reminders (list, detail, create)
- Deliveries (list, detail)
- Providers (list)
- Preferences

## 15. Implementation Phases

### Step 1 — Planning ✅
- [x] Audit current structure
- [x] Document what exists
- [x] Document what needs to change
- [x] This planning document

### Step 2 — Auth Foundation
- Create `src/shared/auth/auth-types.ts`
- Create `src/shared/auth/mock-session.ts`
- Create `src/shared/auth/auth-adapter.ts`
- Update `.env.example`

### Step 3 — HTTP Client Foundation
- Create `src/shared/api/http-client.ts`
- Create `src/shared/api/api-error.ts`
- Create `src/features/notifier/api/notifier-types.ts`

### Step 4 — API Client + Hooks
- Create `notifier-client.ts` (admin endpoints)
- Create `me-client.ts` (user /me endpoints)
- Create `notifier-queries.ts` (query hooks)
- Create `notifier-mutations.ts` (mutation hooks)
- Create `notifier-mocks.ts` (mock fallback)

### Step 5 — Fill Skeleton Pages
- Templates (list, detail, create) — bind to existing mock data
- Reminders (list, detail, create) — bind to existing mock data
- Deliveries (list, detail) — bind to existing mock data
- Providers (list) — bind to existing mock data
- Preferences — bind to existing mock data

### Step 6 — Validation
- Run `npm run lint`
- Run `npm run type-check`
- Run `npm run build`

---

## Phase 2 — Deep Page Implementation and Real Admin UX

### Scope
Transform the Notifier frontend from skeleton/basic pages into a professional, fully usable admin console. Enhance every page with real UX: advanced tables, filters, forms, detail views, dialogs, actions, and proper error/loading/empty states.

### Current State (After Phase 0/1)

| Page | Status After Phase 0/1 |
|------|----------------------|
| Dashboard/Overview | ✅ Good — metrics, provider health, channel breakdown, recent notifications |
| Notifications List | ✅ Good — filters, pagination, table, actions |
| Notification Detail | ✅ Good — summary, timeline, attempts, metadata, actions |
| Send Notification | ✅ Good — form with template selector, variables, scheduling |
| Templates List | ⚠️ Filled — table with filters, channel/locale |
| Template Create | ⚠️ Filled — basic form with type/locale/variables |
| Template Detail | ⚠️ Filled — basic info display |
| Reminders List | ⚠️ Filled — table with status filter |
| Reminder Create | ⚠️ Filled — basic form |
| Reminder Detail | ⚠️ Filled — basic info with cancel |
| Deliveries List | ⚠️ Filled — table with filters |
| Delivery Detail | ⚠️ Filled — basic info with retry |
| Providers | ⚠️ Filled — cards with health, success rate |
| Preferences | ⚠️ Filled — channel toggles |
| Observability | ✅ Good — health, metrics, refresh |

### Remaining Limitations (After Phase 2)
- No real API integration — all data is mock-only
- No WebSocket/realtime updates
- No charts library — trend shown as bars/cards
- Auth still mock-only
- No E2E tests
- New API hooks (notifier-queries.ts) not wired to pages yet
- No `/notifier/` route prefix

---

## Phase 3 — Polish, Realtime, Notification Center, PWA, Tests, Accessibility

### Scope
Transform the Notifier frontend from usable admin console into professional production-quality web app: realtime polling, notification center, PWA, onboarding tours, accessibility, responsive/RTL hardening, performance tuning, unit tests.

### Remaining Limitations (After Phase 3)
- No real API integration — all data is mock-only
- No WebSocket/SSE realtime — using polling fallback
- No charts library — trend shown as CSS bars
- Auth still mock-only — no real login/register
- Drive.js not installed — tours documented, not implemented
- No E2E tests
- New API hooks (notifier-queries.ts) still not wired to pages

---

## Phase 4 — Final API Wiring, Contract Validation, E2E, Release Readiness

### Scope
Finalize the Notifier frontend for real backend integration and release readiness. Document all API contracts, types, auth adapter, security notes, and create complete documentation set.

### Current API Integration Status

| Aspect | Status |
|--------|--------|
| Admin API client (`notifier-client.ts`) | ✅ Fully typed with all endpoints |
| Me API client (`me-client.ts`) | ✅ Fully typed with all endpoints |
| HTTP client (`http-client.ts`) | ✅ Handles auth, tenant, project, request-id headers |
| TanStack Query hooks (`notifier-queries.ts`) | ✅ All endpoints covered with keys and mutations |
| Mock fallback (`notifier-mocks.ts`) | ✅ Full mock data for all endpoints |
| Page integration with API hooks | ⚠️ All pages use old feature hooks, not the new API hooks |
| OpenAPI/Swagger contract validation | ❌ Not done |
| Type mapper/normalizer functions | ❌ Not created |
| Mock mode hardening | ⚠️ Mock banner exists in topbar but no MockModeBanner component |
| Auth adapter completeness | ⚠️ hasRole exists, but hasAnyRole, isAdminLike, isUser missing |
| Query keys centralized | ⚠️ Keys in notifier-queries.ts, no standalone file |
| URL query params for filters | ❌ Not implemented (local state only) |
| E2E tests | ❌ Not created |
| Documentation | ❌ API matrix, contract notes, security notes, release report missing |

### Files to Create

**Documentation:**
- `docs/notifier-frontend-api-matrix.md` — Complete API endpoint audit
- `docs/notifier-frontend-openapi-contract-notes.md` — Swagger contract validation
- `docs/auth-adapter-before-real-auth.md` — Auth adapter guide for future replacement
- `docs/notifier-frontend-security-notes.md` — Security audit
- `docs/notifier-frontend-real-backend-smoke-test.md` — Smoke test guide
- `docs/notifier-frontend-manual-test-scenarios.md` — E2E manual scenarios
- `docs/notifier-frontend-readme.md` — Frontend README/doc
- `docs/notifier-frontend-release-report.md` — Final release report

**Code:**
- `src/features/notifier/api/notifier-mappers.ts` — Type mapper/normalizer functions
- `src/features/notifier/api/notifier-query-keys.ts` — Centralized query keys (extract)
- `src/shared/components/mock-mode-banner.tsx` — Visible mock mode indicator

**Tests:**
- `src/shared/api/__tests__/api-error.test.ts` — ApiError parser tests

### Files to Modify

- `src/shared/auth/auth-types.ts` — Add hasAnyRole, isAdminLike, isUser
- `src/shared/auth/auth-adapter.ts` — Implement new methods
- `docs/notifier-frontend-qa-checklist.md` — Add Phase 4 final QA section
- `.env.example` — Finalize with all vars
- `src/messages/en.json` — Add any missing i18n keys
- `src/messages/fa.json` — Add any missing i18n keys

### Release-Readiness Checklist

- [ ] API matrix document created
- [ ] OpenAPI contract notes created
- [ ] Auth adapter documentation created
- [ ] Frontend README created
- [ ] Release report created
- [ ] Security notes created
- [ ] Smoke test guide created
- [ ] Manual test scenarios documented
- [ ] Type mappers created for null safety
- [ ] MockModeBanner component created
- [ ] Auth adapter refined with hasAnyRole, isAdminLike, isUser
- [ ] Query keys extracted to standalone file
- [ ] QA checklist updated with final QA section
- [ ] .env.example finalized
- [ ] All new i18n strings in fa/en
- [ ] TypeScript passes
- [ ] Lint passes
- [ ] Build passes

### Remaining Limitations Before Auth Service
- No real login/register — mock auth adapter only
- /admin and /me API calls work in mock mode only
- Real backend integration may fail if backend requires real JWT tokens
- Auth adapter is replaceable but untested against real auth service
- No E2E tests with real backend
- URL query params for filters not implemented
- No `/notifier/` route prefix

### Recommended Next Step After Notifier
1. Notifier Full End-to-End Backend + Frontend Smoke/QA (if backend is running)
2. Start Auth Backend (if Notifier frontend needs real auth before full E2E)
