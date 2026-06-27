# Auth Adapter — Before Real Auth

> This document explains the current mock auth adapter and how to replace it with a real auth service.

## Important Distinction: Mock Auth vs Mock Data

The Notifier frontend separates **mock auth** from **mock data**. These are two independent concerns:

| Concern | Controlled By | Purpose |
|---|---|---|
| **Mock Auth** | `NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED` | Provides fake session/token until real Auth service exists |
| **Mock Data** | `NEXT_PUBLIC_NOTIFIER_USE_MOCKS` | Serves fake API responses instead of real backend calls |

**These are independent.** You can (and should, for testing against real backend) set:
- `NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED=true` — use mock session/token
- `NEXT_PUBLIC_NOTIFIER_USE_MOCKS=false` — call real backend APIs

This config sends mock auth headers (`Authorization: Bearer <mock-token>`, `X-Tenant-Id`, etc.) to the real backend, while making actual network requests.

When replacing mock auth with real Auth service, the real auth adapter should implement the same `AuthAdapter` interface. The mock data switch (`USE_MOCKS`) remains independent.

## Current Architecture

```
src/shared/auth/
  auth-types.ts       — AuthSession, UserRole, AuthAdapter interfaces
  auth-adapter.ts     — Singleton adapter (mock, localStorage-backed)
  mock-session.ts     — Default mock session from env vars
```

### Adapter Interface

```typescript
interface AuthAdapter {
  getSession(): AuthSession;
  getAccessToken(): string | null;
  getUserId(): string | null;
  getTenantId(): string | null;
  getProjectId(): string | null;
  getRoles(): UserRole[];
  isAuthenticated(): boolean;
  hasRole(role: UserRole): boolean;
  hasAnyRole(roles: UserRole[]): boolean;
  isAdminLike(): boolean;
}
```

### Session Shape

```typescript
interface AuthSession {
  accessToken: string | null;
  userId: string | null;
  tenantId: string | null;
  projectId: string | null;
  roles: UserRole[];
  permissions: string[];
  isAuthenticated: boolean;
}
```

### How Mock Auth Works

1. On first `getSession()`, it calls `createMockSession()` which reads env vars:
   - `NEXT_PUBLIC_NOTIFIER_MOCK_ACCESS_TOKEN`
   - `NEXT_PUBLIC_NOTIFIER_MOCK_USER_ID`
   - `NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID`
   - `NEXT_PUBLIC_NOTIFIER_MOCK_PROJECT_ID`
   - `NEXT_PUBLIC_NOTIFIER_MOCK_ROLES` (comma-separated)
2. Session can be overridden via `localStorage` key `notifier-mock-session`
3. User can change mock roles from Settings page
4. Session is cached in memory and refreshed only on explicit `refreshMockSession()`

### Headers Sent by HTTP Client

The HTTP client (`src/shared/api/http-client.ts`) automatically adds:
```http
Authorization: Bearer <accessToken>
X-Tenant-Id: <tenantId>
X-Project-Id: <projectId>
X-Request-Id: <uuid>
Content-Type: application/json
```

## JWT Assumptions

The mock adapter uses a plain string as `accessToken`, not a real JWT. When real auth is implemented:

### Expected JWT Claims
```json
{
  "sub": "user_id",
  "tenant_id": "tenant_123",
  "project_id": "project_123",
  "roles": ["admin", "operator"],
  "permissions": ["notifications:read", "notifications:write"]
}
```

### How Real Auth Should Replace Mock

**Step 1:** Create a real auth service that returns an `AuthSession`:
```typescript
// src/shared/auth/real-auth-adapter.ts
import type { AuthAdapter, AuthSession } from './auth-types';

async function getRealSession(): Promise<AuthSession> {
  const token = await getTokenFromAuthService(); // Your auth service
  const decoded = decodeJWT(token); // or call /userinfo endpoint
  return {
    accessToken: token,
    userId: decoded.sub,
    tenantId: decoded.tenant_id,
    projectId: decoded.project_id,
    roles: decoded.roles,
    permissions: decoded.permissions,
    isAuthenticated: true,
  };
}
```

**Step 2:** Replace the `authAdapter` export in `auth-adapter.ts`:
```typescript
// Change this:
export const authAdapter: AuthAdapter = mockAuthAdapter;

// To this:
export const authAdapter: AuthAdapter = realAuthAdapter;
```

No other frontend code needs to change because all pages use the `authAdapter` import.

### Where Auth Adapter Is Used

| File | Usage |
|------|-------|
| `src/shared/api/http-client.ts` | Gets token/tenant/project for headers |
| `src/shared/components/role-guard.tsx` | Checks user roles for access control |
| `src/shared/components/forbidden-state.tsx` | Displays forbidden message |
| `src/app/[locale]/settings/page.tsx` | Shows current session + allows mock role editing |
| `src/features/notifier/notification-center/notification-center-wrapper.tsx` | Gets session context (indirectly via hooks) |

## Important Rules

### Do Not
- Do NOT send `userId` with `/me` API calls — the backend derives it from the auth token
- Do NOT store real tokens in localStorage without encryption
- Do NOT log tokens or auth headers
- Do NOT expose auth adapter implementation details to page components

### Do
- Use `authAdapter.hasRole('admin')` for admin checks
- Use `authAdapter.isAdminLike()` for combined admin/operator checks
- Use `authAdapter.getTenantId()` only where tenant-scoping is needed
- Keep mock adapter as the default until real auth is tested

## Current Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No real JWT validation | Mock token accepted by proxy but rejected by real backend | Use mock mode for UI, real backend with dev auth bypass |
| No token refresh | Token never expires in mock | Manual refreshMockSession() call |
| No login/logout flow | User must use Settings page to change roles | Acceptable until real auth |
| No auth middleware on routes | Pages accessible without valid session | RoleGuard component provides UI-level protection |
| No permission-based checks | Only role-based (hasRole) is implemented | Extend hasAnyRole with permission check when needed |
