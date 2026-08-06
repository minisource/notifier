# Notifier Frontend

> Admin panel and user notification center for the Minisource Notifier service.

## Overview

The Notifier Frontend provides a complete admin console for managing notifications, templates, reminders, deliveries, providers, and observability. It also includes a user notification center with realtime polling.

## Tech Stack

- **Framework:** Next.js 15 (App Router)
- **Language:** TypeScript
- **UI:** shadcn/ui, Tailwind CSS, Lucide icons
- **State:** TanStack Query, Zustand
- **Forms:** React Hook Form, Zod
- **i18n:** next-intl (fa + en, RTL/LTR)
- **Themes:** next-themes (dark/light/system)

## Routes

All routes are under `/{locale}/` where `locale` is `fa` or `en`.

### Admin Pages

| Route | Page | Description |
|-------|------|-------------|
| `/dashboard` | Dashboard | Overview metrics, trend, provider health, queue |
| `/notifications` | Notifications | List with filters, pagination, actions |
| `/notifications/[id]` | Notification Detail | Full detail, timeline, attempts |
| `/notifications/new` | Send Notification | Compose and send form |
| `/templates` | Templates | List with filters |
| `/templates/new` | Create Template | Form with body, variables, locale |
| `/templates/[id]` | Template Detail | View/edit, render preview |
| `/reminders` | Reminders | List with status filter |
| `/reminders/new` | Create Reminder | Form with scheduling |
| `/reminders/[id]` | Reminder Detail | View/cancel |
| `/deliveries` | Deliveries | List with filters |
| `/deliveries/[id]` | Delivery Detail | View/retry |
| `/providers` | Providers | Health cards, test dialog |
| `/preferences` | Preferences | Channel settings |
| `/observability` | Observability | Health, metrics, queue, workers |
| `/settings` | Settings | Theme, language, debug |

## Environment Variables

See `.env.example` for full list.

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_NOTIFIER_API_URL` | `http://localhost:9002/v1` | Backend API URL |
| `NEXT_PUBLIC_NOTIFIER_AUTH_ENABLED` | `true` | Enforce Auth-service login (`false` = standalone local session) |
| `NEXT_PUBLIC_NOTIFIER_LOCAL_USER_ID` | `local-admin` | Local session user ID when auth is disabled |
| `NEXT_PUBLIC_NOTIFIER_LOCAL_EMAIL` | `admin@local` | Local session email when auth is disabled |
| `NEXT_PUBLIC_NOTIFIER_LOCAL_NAME` | `Local Admin` | Local session display name when auth is disabled |
| `NEXT_PUBLIC_NOTIFIER_REALTIME_MODE` | `polling` | Realtime mode |
| `NEXT_PUBLIC_NOTIFIER_REALTIME_POLL_INTERVAL` | `30000` | Poll interval ms |
| `NEXT_PUBLIC_NOTIFIER_SHOW_LIVE_TOASTS` | `true` | Show toast on new notifications |

## Auth-Enabled vs Standalone Mode

Mock data has been removed — the frontend always calls the real backend. The only toggle is authentication:

| Variable | Controls | Default |
|---|---|---|
| `NEXT_PUBLIC_NOTIFIER_AUTH_ENABLED` | Require login via the Auth service | `true` |

- **Auth enabled** (`true`): the app shows the login page and requires a real session/token from the Auth service. The backend must run with `AUTH_ENABLED=true`.
- **Standalone** (`false`): the app boots directly with a local admin session (identity from `NEXT_PUBLIC_NOTIFIER_LOCAL_*`, defaults to `local-admin`). The backend must run with `AUTH_ENABLED=false` so routes are public; `/me` endpoints fall back to `AUTH_LOCAL_USER_ID` on the backend. No Auth service required.

## Admin vs Me API Usage

| API Group | Prefix | Used By | Role Required |
|-----------|--------|---------|---------------|
| Admin | `/admin/*` | Dashboard, CRUD pages | admin/operator |
| Me | `/me/*` | Notification center, preferences | any authenticated |

Pages use `/admin` endpoints for the admin console. The notification center uses `/me` endpoints for the current user's data.

## Realtime Modes

| Mode | Description | Config |
|------|-------------|--------|
| `polling` (default) | TanStack Query refetchInterval + query invalidation | 30s interval |
| `websocket` | WebSocket connection (backend must support) | Set URL |
| `sse` | Server-Sent Events (backend must support) | Set URL |
| `disabled` | No realtime updates | — |

## i18n / RTL

- Languages: Persian (fa) and English (en)
- Direction: RTL for fa, LTR for en
- Messages: `src/messages/{locale}.json`

## PWA

- Manifest: `public/manifest.json`
- Icons: 192x192 and 512x512 (add to `public/icons/`)
- No service worker currently

## Testing

```bash
npm run test          # Unit tests (Vitest)
npm run test:watch    # Watch mode
```

Test files are co-located with components in `__tests__/` directories.

## Build

```bash
npm run build         # Production build
npm run lint          # ESLint
npm run typecheck     # TypeScript check
```

## Known Limitations

1. **No real Auth service** — Mock auth adapter only. Real auth service not integrated.
2. **No real auth** — Mock auth adapter only. Real auth service not integrated.
3. **No E2E tests** — Manual test scenarios documented in `docs/notifier-frontend-manual-test-scenarios.md`
4. **No WebSocket/SSE** — Realtime uses polling only
5. **No charts library** — Trend shown as CSS bars
6. **No `/notifier` route prefix** — Pages under `/[locale]` directly
7. **Drive.js tours not installed** — Onboarding structure documented only
