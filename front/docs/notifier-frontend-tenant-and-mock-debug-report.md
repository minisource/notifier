# Notifier Frontend — Tenant & Mock Data Debug Report

## 1. Current Frontend Env Values

| Variable | Value |
|---|---|
| `NEXT_PUBLIC_NOTIFIER_API_BASE_URL` | `http://127.0.0.1:9002/v1` |
| `NEXT_PUBLIC_NOTIFIER_USE_MOCKS` | `false` |
| `NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED` | `true` |
| `NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID` | (empty — no longer defaults to `tenant-default`) |
| `NEXT_PUBLIC_NOTIFIER_MOCK_PROJECT_ID` | (empty — no longer defaults to `project-default`) |
| `NEXT_PUBLIC_NOTIFIER_MOCK_ACCESS_TOKEN` | `mock-token-notifier-admin` |
| `NEXT_PUBLIC_NOTIFIER_MOCK_USER_ID` | `user-mock-001` |
| `NEXT_PUBLIC_NOTIFIER_MOCK_ROLES` | `admin,operator` |

## 2. How Tenant Is Resolved

- **Source:** `mock-session.ts` → `createMockSession()` → `getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID', '').trim() || null`
- **Fallback:** Previously hardcoded to `'tenant-default'`. **Fixed:** Now defaults to `null`.
- **When null:** HTTP client (`http-client.ts`) checks `if (session.tenantId)` before setting `X-Tenant-Id` header — so no header is sent when tenant is null.

## 3. How Project Is Resolved

- Same pattern as tenant: `getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_PROJECT_ID', '').trim() || null`
- **Fixed:** No longer defaults to `'project-default'`.
- When null: `X-Project-Id` header is not sent.

## 4. How Auth Token Is Resolved

- `mock-session.ts` → `getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_ACCESS_TOKEN', 'mock-token-notifier-admin')`
- Fallback: `'mock-token-notifier-admin'` (kept as-is for backward compatibility)
- Sent as `Authorization: Bearer <token>` header

## 5. Headers Sent by HTTP Client

| Header | Condition |
|---|---|
| `Authorization: Bearer <token>` | Only if `session.accessToken` is non-null |
| `X-Tenant-Id` | Only if `session.tenantId` is non-null |
| `X-Project-Id` | Only if `session.projectId` is non-null |
| `X-Request-Id` | Always (generated per request) |
| `Content-Type: application/json` | Always |

## 6. Auth Adapter Behavior

| Scenario | `getSession()` returns |
|---|---|
| `MOCK_AUTH_ENABLED=true` (default) | Mock session with `source: 'mock'`, populated fields |
| `MOCK_AUTH_ENABLED=false` | Empty session with `source: 'none'`, all nullable fields null, `isAuthenticated: false` |

## 7. API Mode Switch

- Located in `src/features/notifier/api/notifier-api-mode.ts`
- Gated by `notifierRuntimeConfig.useMocks` from `notifier-config.ts`
- When `useMocks=false`: exports real API client implementations
- When `useMocks=true`: exports mock API implementations

## 8. Mock Data Leakage Points

| Location | Status |
|---|---|
| `pages/components importing mock data directly` | ✅ None found |
| `query hooks using mock initialData/placeholderData` | ✅ None found |
| `catch blocks falling back to mock data` | ✅ None found |
| `lib/mock/session.ts` hardcoded `tenant-default` | ⚠️ Legacy system, not imported by notifier features |
| `features/notifications/api.ts` importing from `@/lib/mock/db` | ⚠️ Outside notifier feature, separate module |

## 9. Why Backend Returned `Invalid or missing tenant`

**Root cause:** `mock-session.ts` had:
```ts
tenantId: getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID', 'tenant-default')
```

When `NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID` was not explicitly set, it fell back to `'tenant-default'`. This value was sent as `X-Tenant-Id: tenant-default` header. The backend does not recognize `tenant-default` as a valid tenant, so it returned `400 Invalid or missing tenant`.

**Fix:** Changed to:
```ts
tenantId: getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID', '').trim() || null
```

No X-Tenant-Id header is sent unless explicitly configured.

## 10. Files Changed

| File | Change |
|---|---|
| `src/shared/auth/auth-types.ts` | Added `SessionSource` type and `source` field to `AuthSession` |
| `src/shared/auth/mock-session.ts` | Removed hardcoded `tenant-default`/`project-default` fallbacks, added `source: 'mock'` |
| `src/shared/auth/auth-adapter.ts` | Added `createEmptySession()`, gated session creation on `isAuthEnabled()` |
| `.env.example` | Cleared `MOCK_TENANT_ID`/`MOCK_PROJECT_ID` defaults, added documentation |

## 11. Validation Results

| Check | Result |
|---|---|
| Frontend TypeScript (`tsc --noEmit`) | ✅ 0 errors |
| Frontend lint (`npm run lint`) | ✅ 0 new errors |
| No hardcoded tenant default in mock session | ✅ Fixed |
| Auth adapter respects isAuthEnabled() | ✅ Fixed |
| refreshMockSession respects isAuthEnabled() | ✅ Fixed |
| updateMockInLocalStorage respects isAuthEnabled() | ✅ Fixed |

## 12. Remaining Limitations

1. **Backend response envelope not unwrapped** — The HTTP client returns `{ success: true, data: { ... } }` as-is. Components receive the envelope. A `notifier-response.ts` with `unwarpApiEnvelope<T>()` has not been created yet.
2. **`TenantRequiredState` component not created** — No dedicated UI state for when tenant is missing. Components currently show generic error states.
3. **HTTP client dev logging not added** — No dev-mode logging of API mode, base URL, or header presence.
4. **Legacy `lib/mock/session.ts`** still has hardcoded `tenant-default` — but this is a separate system not used by notifier features.
5. **Settings page** still references old env var names (`NEXT_PUBLIC_NOTIFIER_API_URL`).
