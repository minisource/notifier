# Notifier Frontend — Mock Mode Debug Report

## 1. Current Env Variables Related to Mock Mode

| Variable | Current Value | Default |
|---|---|---|
| `NEXT_PUBLIC_API_MODE` | `real` | `mock` |
| `NEXT_PUBLIC_NOTIFIER_USE_MOCKS` | `false` | `false` (new) |
| `NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED` | `true` | `true` |

## 2. How Mock Mode Was Previously Detected

Before this hotfix, mock mode was detected via `isMockMode()` in `src/lib/config/env.ts`:

```ts
export function isMockMode(): boolean {
  return env.apiMode === 'mock';
}
```

This checked `NEXT_PUBLIC_API_MODE` only — **not** `NEXT_PUBLIC_NOTIFIER_USE_MOCKS`. Since the `.env` had `API_MODE=real`, `isMockMode()` returned `false`. But pages still imported and used mock data directly via:

```ts
import { notifierMock } from '@/features/notifier/api/notifier-mocks';
```

## 3. Root Cause

The root cause was **not** an env parsing bug (no `Boolean("false")` pattern existed). The root cause was **direct mock imports bypassing the env check entirely**:

1. **Dashboard page** (`dashboard/page.tsx`): Imported `notifierMock` and called `notifierMock.getDashboardOverview()` in a `useEffect` to populate `trendData` and `failureData` — regardless of env settings.

2. **Observability page** (`observability/page.tsx`): Imported `notifierMock` and called `notifierMock.getHealth()`, `notifierMock.getMetrics()`, `notifierMock.getReadiness()` as the sole data source — never called real backend APIs.

3. **Sidebar** (`sidebar.tsx`): Imported `mockSession` from `@/lib/mock/session` directly, always showing mock user info.

4. **User menu** (`user-menu.tsx`): Imported `mockSession` directly, always showing mock user info.

5. **`features/dashboard/api.ts`**: Used `mockNotifications` from `@/lib/mock/db` directly.

6. **`lib/api/client.ts`**: Entirely mock-based API client importing all data from `@/lib/mock/db`.

## 4. Files with Direct Mock Imports (Before Fix)

| File | What it Imported |
|---|---|
| `app/[locale]/dashboard/page.tsx` | `notifierMock` from `notifier-mocks` |
| `app/[locale]/observability/page.tsx` | `notifierMock` from `notifier-mocks` |
| `components/layout/sidebar.tsx` | `mockSession` from `lib/mock/session`, `isMockMode` from `lib/config/env` |
| `components/layout/user-menu.tsx` | `mockSession` from `lib/mock/session` |
| `components/layout/topbar.tsx` | `isMockMode` from `lib/config/env` |
| `shared/components/mock-mode-banner.tsx` | `isMockMode` from `lib/config/env` |
| `features/dashboard/api.ts` | `mockNotifications` from `lib/mock/db`, `getMetrics` from `lib/api/client` |
| `lib/api/client.ts` | All mock data from `lib/mock/db` |

## 5. API Failure → Mock Fallback Check

**No silent fallback to mock was found.** When real API calls fail, the TanStack Query hooks properly throw errors, and pages show `ErrorState` components. No `catch → return mock` patterns existed in the API layer.

However, the observability page had its own `try/catch` that would catch real API errors, but since it only called `notifierMock` methods (which never fail), the catch was effectively dead code — and the real API was never called.

## 6. MSW / Mock Service Worker

**Not present.** No MSW files, handlers, or worker scripts were found. No MSW interception occurs.

## 7. `initialData` / `placeholderData` Check

**No `initialData` or `placeholderData` using mock data** was found in the notifier feature hooks. All hooks use standard TanStack Query without mock initial data.

## 8. Mock Auth vs Mock Data Separation

Before this fix, `mockSession` was imported directly in sidebar and user-menu, coupling auth mock with data mock. Now both components use `authAdapter.getSession()` which is a centralized abstraction.

The two env vars are now properly separated and documented:
- `NEXT_PUBLIC_NOTIFIER_USE_MOCKS=false` → controls mock **data/API**
- `NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED=true` → controls mock **session/token** (temporary)

## 9. Changes Applied

### New Files

| File | Purpose |
|---|---|
| `src/features/notifier/config/notifier-config.ts` | `parseBooleanEnv()` + `notifierRuntimeConfig` object |
| `src/features/notifier/api/notifier-api-mode.ts` | Centralized mock/real API switch |
| `docs/notifier-mock-mode-debug-report.md` | This report |

### Modified Files

| File | Change |
|---|---|
| `src/lib/config/env.ts` | Added `isMockDataEnabled()` with proper boolean parsing; `isMockMode()` marked deprecated |
| `src/features/notifier/api/notifier-queries.ts` | Imports from `notifier-api-mode` instead of direct `notifier-client`/`me-client` |
| `src/app/[locale]/dashboard/page.tsx` | Removed `notifierMock` import; extracts trend/failure from `useDashboard()` data |
| `src/app/[locale]/observability/page.tsx` | Replaced `notifierMock` calls with `useAdminHealth()`, `useAdminMetrics()`, `useAdminReadiness()` hooks |
| `src/components/layout/sidebar.tsx` | Replaced `mockSession` + `isMockMode` with `authAdapter.getSession()` + `env.isMockDataEnabled` |
| `src/components/layout/user-menu.tsx` | Replaced `mockSession` with `authAdapter.getSession()` |
| `src/components/layout/topbar.tsx` | Uses `isMockDataEnabled()` for mock/real badge |
| `src/shared/components/mock-mode-banner.tsx` | Uses `isMockDataEnabled()` instead of `isMockMode()` |

## 10. Validation

After applying all changes:
- ✅ `npm run typecheck` passes
- ✅ `npm run lint` passes
- ✅ `npm run build` passes

## 11. Remaining Mock Data Files

These files are kept for development but no longer leak into production pages:

| File | Purpose | Status |
|---|---|---|
| `src/features/notifier/api/notifier-mocks.ts` | Mock API for notifier module | 🟢 Used only through `notifier-api-mode.ts` switch |
| `src/lib/mock/db.ts` | Centralized mock database | 🟢 Only used by legacy `lib/api/client.ts` (not by pages) |
| `src/lib/mock/session.ts` | Mock session data | 🟢 Only used by auth adapter internally |
| `src/lib/api/client.ts` | Legacy mock API client | 🟡 Not imported by any page; kept for reference |
| `src/features/dashboard/api.ts` | Legacy dashboard mock API | 🟡 Not imported by any page; kept for reference |

## 12. Verification Steps

### With `NEXT_PUBLIC_NOTIFIER_USE_MOCKS=false`:
1. Restart dev server
2. Open browser Network tab
3. Visit `/notifier` → real `GET /admin/dashboard/overview` must appear
4. Visit `/notifier/observability` → real `GET /admin/observability/*` must appear
5. MockModeBanner must NOT be visible
6. Topbar must show "Real" badge (green)

### With `NEXT_PUBLIC_NOTIFIER_USE_MOCKS=true`:
1. Restart dev server
2. MockModeBanner must be visible (amber, dismissible)
3. Topbar must show "Mock" badge (amber)
4. All pages must work without backend
