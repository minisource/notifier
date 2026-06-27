## Phase 2.5 — Visual Design System & Dashboard Redesign

### Current UI Problems (Before)

1. **Metric cards are too plain** — default shadcn style, no semantic variants, no visual hierarchy
2. **Sidebar active state too generic** — flat shading, no product identity
3. **Dashboard has no operational status strip** — no quick-glance system health
4. **Channel breakdown too basic** — just text labels with counts
5. **Provider health panel lacks hierarchy** — no actionable visual cues
6. **Recent notifications too simple** — no scannable layout
7. **Page background too flat** — all pure white, no depth layering
8. **Cards look like default shadcn examples** — no product personality
9. **Topbar feels empty** — just language switcher + theme toggle
10. **No strong product identity** — Notifier brand invisible

### Design Direction

**Style**: Refined operations console — professional, dense, technical, trustworthy.

**Background layering**:
```
app surface: bg-muted/30 (subtle)
content cards: bg-card with border-border/70 + shadow-sm
elevated panels: bg-card with shadow-md
```

**Sidebar**:
- Grouped nav sections (Overview, Messaging, Operations, Management)
- Stronger active indicator (left/right accent bar based on RTL/LTR)
- Section labels
- Compact product header with status pill

**Dashboard hierarchy** (top to bottom):
1. Operational Status Strip — system health, queue, providers, dead letters
2. Critical Metrics Row — 4 key KPIs with semantic variants
3. Queue & Reliability Row — secondary metrics
4. Two-column panels — provider health + recent notifications

**Status color system**:
```
success (sent/delivered/healthy):    green-500/green-100
warning (retrying/degraded):         amber-500/amber-100
danger (failed/dead/down):           red-500/red-100
info (queued/pending/processing):    blue-500/blue-100
muted (cancelled/disabled):          gray-500/gray-100
```

### Files Changed

| File | Change |
|------|--------|
| `components/shared/metric-card.tsx` | Added variants, trend arrows, accent bars, progress |
| `components/shared/status-badge.tsx` | Added colored dot indicator |
| `components/shared/channel-badge.tsx` | Enhanced with channel icons |
| `components/shared/section-card.tsx` | NEW: card with section styling |
| `components/shared/mini-progress.tsx` | NEW: compact progress bar |
| `components/shared/kpi-delta.tsx` | NEW: trend indicator component |
| `components/layout/sidebar.tsx` | Grouped nav, stronger active, section labels, product header |
| `components/layout/topbar.tsx` | Added API mode badge, tenant selector, page context |
| `components/layout/app-shell.tsx` | Layered background, refined spacing |
| `features/dashboard/components/*` | NEW: 6 dashboard panel components |
| `app/[locale]/dashboard/page.tsx` | Refactored with hierarchy |
| `messages/*.json` | Added dashboard design keys |

---

## Phase 4 — Notifications Management Implementation

### Scope

Full notification management experience:
- Notifications List (`/[locale]/notifications`)
- Send Notification (`/[locale]/notifications/new`)
- Notification Detail (`/[locale]/notifications/[id]`)

### Design Notes

- Uses existing design language from Phase 2.5 (StatusBadge, ChannelBadge, SectionCard, MetricCard)
- Rows are dense but readable — operations console density
- Status-driven: semantic colors for quick scan
- Actions are contextual — disabled when invalid for current status
- All new strings translated in fa/en
- Mock mode only, no real backend

### Status Alignment

Mock data statuses aligned with feature types:
```
Feature types: pending | queued | processing | sent | failed | dead | cancelled
Mock types:   pending | queued | processing | sent | failed | dead | cancelled
```

### Files Created

```
src/features/notifications/components/
  notification-table.tsx
  notification-filters.tsx
  notification-action-menu.tsx
  notification-summary-card.tsx
  notification-timeline.tsx
  notification-attempts-list.tsx
  notification-metadata-viewer.tsx
  send-notification-form.tsx
  template-key-combobox.tsx
  variables-editor.tsx
```

### Files Changed

| File | Change |
|------|--------|
| `lib/mock/db.ts` | Aligned status enums, added rich mock data (dead letter, webhook, read/seen) |
| `features/notifications/types.ts` | Added `DeliveryAttempt`, `NotificationDelivery`, enhanced types |
| `features/notifications/api.ts` | Added `markNotificationRead`, `markAllNotificationsRead`, `sendBatchNotifications` |
| `features/notifications/hooks/use-notifications.ts` | Added all new mutation hooks |
| `features/notifications/schemas.ts` | Enhanced form schema |
| `app/[locale]/notifications/page.tsx` | Full implementation with filters, table, pagination |
| `app/[locale]/notifications/new/page.tsx` | Full send form implementation |
| `app/[locale]/notifications/[id]/page.tsx` | Full detail page implementation |
| `messages/en.json` | Added notification section keys |
| `messages/fa.json` | Added notification section keys |

### i18n Namespaces Added

```
notifications.list — columns, empty states
notifications.filters — filter labels, active filters
notifications.form — recipient, channel, template, variables
notifications.detail — summary, timeline, attempts
notifications.actions — contextual action labels
notifications.validation — form validation messages
notifications.timeline — timeline event labels
notifications.attempts — attempt details
```

### Acceptance Checklist

- [ ] Notifications list renders with table/search/filters/pagination
- [ ] Row actions work contextually (retry/cancel/mark-read/copy-id)
- [ ] Send notification form validates by channel type
- [ ] Template key combobox with mock templates
- [ ] Variables editor (add/remove key-value)
- [ ] Detail page with summary/timeline/attempts/metadata
- [ ] PII masked in UI (phone, email, tokens)
- [ ] All new strings translated in fa + en
- [ ] RTL layout correct for all 3 routes
- [ ] Dark mode polished
- [ ] Build/typecheck pass
- [ ] No any
- [ ] No direct mock imports in pages/components
