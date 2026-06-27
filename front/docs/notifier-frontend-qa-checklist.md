# Notifier Frontend QA Checklist

## API Integration
- [ ] Admin endpoints return expected data shapes
- [ ] Me endpoints return expected data shapes
- [ ] Error responses display properly (message, code, requestId)
- [ ] Loading states show during API calls
- [ ] Empty states display for empty responses

## Mock Mode
- [ ] Dashboard shows realistic mock data
- [ ] Notifications list/detail work with mock data
- [ ] Templates list/create/edit work with mock data
- [ ] Reminders list/create/edit/cancel work with mock data
- [ ] Deliveries list/detail/retry work with mock data
- [ ] Providers list/test dialog work with mock data
- [ ] Preferences display channel toggles
- [ ] Observability shows health/queue/workers

## Admin Pages
- [ ] Dashboard metrics, trend, failures, provider health render
- [ ] Notifications table with filters, pagination, sorting
- [ ] Notification detail with timeline, attempts, metadata
- [ ] Templates list with filters, create form, detail
- [ ] Reminders list, create form, detail with cancel
- [ ] Deliveries list, detail with retry
- [ ] Providers list with health status, test dialog
- [ ] Preferences channel toggles
- [ ] Observability health, readiness, metrics, queue, workers

## User Notification Center
- [ ] Bell icon shows in topbar
- [ ] Unread count badge displays correctly
- [ ] Popover opens on desktop
- [ ] Sheet opens on mobile
- [ ] All/unread tabs work
- [ ] Notification items show channel/status badges
- [ ] Mark as read action works
- [ ] Mark all as read works
- [ ] Loading state shows skeleton
- [ ] Empty state shows when no notifications

## Realtime / Polling
- [ ] Polling interval configuration respected
- [ ] Query invalidations trigger on poll
- [ ] No overly aggressive refetching
- [ ] Tab visibility pause works

## RTL/LTR
- [ ] Dashboard layout works in RTL
- [ ] Notifications table works in LTR
- [ ] Detail pages stack correctly in RTL
- [ ] Forms render correctly in LTR
- [ ] Dialog/sheet positions work in RTL

## fa/en
- [ ] All pages display in Persian
- [ ] All pages display in English
- [ ] No hardcoded UI strings
- [ ] Date/time formats respect locale
- [ ] Direction (rtl/ltr) switches correctly

## Responsive
- [ ] Desktop >= 1280px - full layout
- [ ] Laptop 1024px - readable
- [ ] Tablet 768px - stacked layout
- [ ] Mobile < 640px - single column, horizontal scroll tables
- [ ] Topbar does not overflow
- [ ] Sidebar works on mobile
- [ ] Dialogs fit mobile screens

## Dark/Light
- [ ] Light mode - all pages readable
- [ ] Dark mode - all pages readable
- [ ] Badges/status colors visible in both modes
- [ ] Chart/bar colors visible in both modes

## Accessibility
- [ ] Icon buttons have aria-label
- [ ] Dialog has accessible title/description
- [ ] Form errors linked to inputs
- [ ] Focus states visible on interactive elements
- [ ] Color is not the only signal for status
- [ ] Tables have proper headers
- [ ] Loading states do not trap focus
- [ ] Keyboard navigation works for main actions

## PWA
- [ ] Manifest.json exists with correct metadata
- [ ] Icons referenced in manifest
- [ ] Layout has theme-color meta
- [ ] apple-mobile-web-app-capable set

## Performance
- [ ] Dashboard staleTime reasonable (15-30s)
- [ ] Unread count staleTime reasonable (10-30s)
- [ ] Static data staleTime reasonable (30-60s)
- [ ] No unnecessary refetches on re-render
- [ ] Polling respects tab visibility

## Build Validation
- [ ] TypeScript passes (0 errors)
- [ ] Lint passes (0 errors)
- [ ] Build passes
- [ ] Tests pass

## Known Limitations
- No real API integration — all data is mock-only
- No WebSocket/SSE realtime — using polling fallback
- No charts library — trend shown as CSS bars
- Auth still mock-only — no real login/register
- Drive.js not installed — tours documented, not implemented
- No E2E tests
- New API hooks (notifier-queries.ts) not wired to pages
