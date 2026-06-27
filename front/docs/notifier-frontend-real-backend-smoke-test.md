# Notifier Frontend — Real Backend Smoke Test Guide

> Use this guide to verify the frontend works against a real Notifier backend.

## Prerequisites

- Notifier backend is running and healthy
- Frontend dev server is running (`npm run dev`)
- You have access to a backend auth token or have enabled dev auth bypass

## Step 1 — Configure Environment

```bash
# Stop frontend dev server
# Edit .env.local (create if not exists):
NEXT_PUBLIC_NOTIFIER_API_BASE_URL=http://localhost:9002/v1
NEXT_PUBLIC_NOTIFIER_USE_MOCKS=false
NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED=true
NEXT_PUBLIC_NOTIFIER_MOCK_ACCESS_TOKEN=<backend-dev-token-or-mock>
NEXT_PUBLIC_NOTIFIER_MOCK_USER_ID=user_001
NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID=tenant_default
NEXT_PUBLIC_NOTIFIER_MOCK_PROJECT_ID=project_default
NEXT_PUBLIC_NOTIFIER_MOCK_ROLES=admin,operator

# Start frontend
npm run dev
```

## Step 2 — Verify Backend Health

```bash
curl http://localhost:9002/v1/admin/observability/health
```

Expected: `200 OK` with JSON body containing `"status": "healthy"`

## Step 3 — Test Overview/Dashboard

1. Open `http://localhost:3000/fa/dashboard`
2. **Expected:** Dashboard loads with metric cards and sections
3. **If 401/403:** Backend rejects mock token. Check:
   - Backend dev auth bypass is enabled
   - Mock token is accepted by backend
   - CORS headers allow frontend origin

## Step 4 — Test Notifications

1. Navigate to `/fa/notifications`
2. **Expected:** Notifications list loads
3. Click a notification to view detail
4. **Expected:** Detail, timeline, attempts render
5. **Known issue:** If backend returns error for `/admin/notifications/{id}/attempts`, the UI shows error state with retry button

## Step 5 — Test Templates

1. Navigate to `/fa/templates`
2. **Expected:** Template list loads (or empty state)
3. Click "Create Template" → fill form → submit
4. **Expected:** Template created and redirects to list

## Step 6 — Test Reminders

1. Navigate to `/fa/reminders`
2. **Expected:** Reminder list loads
3. Create a new reminder with future date
4. **Expected:** Reminder created, appears in list
5. Cancel the reminder → confirm dialog → submit
6. **Expected:** Reminder status updates

## Step 7 — Test Providers

1. Navigate to `/fa/providers`
2. **Expected:** Provider cards with health status
3. Click "Test Provider" on a healthy provider
4. **Expected:** Test dialog opens
5. Fill in test data, keep `dryRun=true`, submit
6. **Expected:** Test result shown (success or error message)

## Step 8 — Test Observability

1. Navigate to `/fa/observability`
2. **Expected:** Health, readiness, metrics, queue, workers sections render

## Step 9 — Test Notification Center

1. Click the bell icon in topbar
2. **Expected:** Popover opens with notifications list (or empty state)
3. Unread count updates as notifications arrive
4. "Mark all read" works (if backend supports)

## Step 10 — Test Preferences

1. Navigate to `/fa/preferences`
2. **Expected:** Channel settings load (may show error if `/me/preferences` requires real auth)

## Expected Errors When Auth Is Not Real

| Error | Likely Cause | Workaround |
|-------|-------------|------------|
| 401 Unauthorized | Backend requires real JWT | Enable backend dev auth bypass |
| 403 Forbidden | Mock token/roles not accepted | Check roles match backend expectations |
| 500 Server Error | Backend provider/config not ready | Check backend logs |
| CORS error | Frontend origin not allowed | Add frontend origin to backend CORS config |
| 404 Not Found | Endpoint path mismatch | Check base URL and path prefix |

## Mock Mode Fallback

If real backend is not available:

```bash
# Set in .env.local:
NEXT_PUBLIC_NOTIFIER_USE_MOCKS=true
NEXT_PUBLIC_API_MODE=mock
```

The topbar shows a "Mock" badge and all data comes from `notifier-mocks.ts`.
