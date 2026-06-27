# Notifier Frontend — Manual Test Scenarios

> These scenarios can be run in mock mode (no backend required).

---

## Scenario A — Admin Dashboard Loads

**Setup:** Mock session with `admin` role

**Steps:**
1. Visit `/{locale}/dashboard`
2. ✅ Dashboard cards render (Total, Sent Today, Failed Today)
3. ✅ Status strip shows operational status
4. ✅ Provider health section renders with health summaries
5. ✅ Queue panel renders
6. ✅ Channel breakdown renders
7. ✅ Daily trend chart renders
8. ✅ Recent failures list renders
9. ✅ Auto-refresh toggle works (enable/disable)
10. ✅ Refresh button works, last-updated timestamp updates
11. ✅ No forbidden state shown

---

## Scenario B — Normal User Blocked From Admin Pages

**Setup:** Mock session with `user` role only

**Steps:**
1. Visit `/{locale}/dashboard`
2. ✅ ForbiddenState renders
3. ✅ Shows "Access forbidden" message
4. ✅ Shows "Back" button that navigates to dashboard

---

## Scenario C — Notifications List and Detail

**Setup:** Mock session with `admin` role

**Steps:**
1. Visit `/{locale}/notifications`
2. ✅ Table renders with notification data
3. ✅ Channel badges display correctly
4. ✅ Status badges display correctly
5. ✅ Filter by status works (select "failed" → only failed shown)
6. ✅ Filter by channel works
7. ✅ Search by recipient works
8. ✅ Pagination works (next/prev page)
9. ✅ "Clear filters" button resets all filters
10. ✅ Click a notification → navigates to detail page
11. ✅ Detail page shows summary card
12. ✅ Timeline renders with status steps
13. ✅ Delivery attempts section renders
14. ✅ Metadata viewer renders (if metadata exists)
15. ✅ "Back to list" button works
16. ✅ Retry button opens confirm dialog
17. ✅ Cancel button opens confirm dialog

---

## Scenario D — Notification Center

**Setup:** Mock session with `admin` or `user` role

**Steps:**
1. ✅ Bell icon visible in topbar
2. ✅ Unread count badge displays
3. ✅ Click bell → popover opens (desktop)
4. ✅ Click bell → sheet opens (mobile, <640px)
5. ✅ "All" tab shows all notifications
6. ✅ "Unread" tab shows only unread
7. ✅ Each notification shows channel badge + status badge
8. ✅ Mark as read button appears on unread items
9. ✅ "Mark all as read" button calls backend
10. ✅ Click notification → navigates to detail page
11. ✅ Empty state when no notifications
12. ✅ Skeleton loading state while data loads

---

## Scenario E — Template Render Preview

**Setup:** Mock session with `admin` role

**Steps:**
1. Visit `/{locale}/templates`
2. ✅ Template list renders with name, key, channel, locale
3. ✅ Click a template → detail page
4. ✅ Template details show name, key, body, variables
5. ✅ Click "Render Preview" → dialog opens
6. ✅ Enter variables JSON (e.g., `{"code": "12345"}`)
7. ✅ Click "Render" → preview output shown
8. ✅ Missing variables display (if variables input is incomplete)
9. ✅ Close dialog

---

## Scenario F — Reminder Create/Cancel

**Setup:** Mock session with `admin` role

**Steps:**
1. Visit `/{locale}/reminders`
2. ✅ Reminder list renders with status badges
3. ✅ Click "New Reminder" → form page
4. ✅ Fill form with: userId, channel, templateKey, scheduledAt (future date)
5. ✅ Submit form → reminder created
6. ✅ Back to list → new reminder appears
7. ✅ Click new reminder → detail page
8. ✅ Click "Cancel" → confirm dialog
9. ✅ Confirm → status updates to cancelled
10. ✅ Back to list → status reflects cancellation

---

## Scenario G — Provider Test Dry Run

**Setup:** Mock session with `admin` role

**Steps:**
1. Visit `/{locale}/providers`
2. ✅ Provider cards render with health badges
3. ✅ Click "Test Provider" on a provider
4. ✅ Dialog opens with form fields
5. ✅ `dryRun` toggle defaults to true
6. ✅ Fill in recipient + body
7. ✅ Submit → result panel shows success/error
8. ✅ Close dialog
9. ✅ Test with `dryRun=false` (warning shown before send)
10. ✅ Provider response does not display raw secrets

---

## Scenario H — API Error Shows RequestId

**Setup:** This scenario requires a real backend or modifying mock to simulate errors

**Steps:**
1. Simulate backend returning `ErrorResponse` with `requestId`
2. ✅ Error message displays user-friendly text
3. ✅ Error code displays
4. ✅ RequestId displays with copy button
5. ✅ Retry button works

---

## Scenario I — RTL/LTR and i18n

**Steps:**
1. Visit `/{locale}/dashboard` with `locale=fa`
2. ✅ All text is in Persian
3. ✅ Layout is RTL (sidebar on right, text right-aligned)
4. ✅ Visit `/{locale}/dashboard` with `locale=en`
5. ✅ All text is in English
6. ✅ Layout is LTR (sidebar on left)
7. ✅ Toggle language from settings page
8. ✅ Language changes without breaking layout

---

## Scenario J — Responsive Breakpoints

**Steps:**
1. Desktop (≥1280px): ✅ Full layout, sidebar visible
2. Laptop (1024px): ✅ Readable, sidebar visible
3. Tablet (768px): ✅ Stacked layout, sidebar collapsed
4. Mobile (<640px): ✅ Single column, hamburger menu, notification sheet
5. ✅ Tables horizontally scroll on mobile
6. ✅ Detail cards stack vertically on mobile
7. ✅ Topbar doesn't overflow on any breakpoint

---

## Scenario K — Dark/Light Mode

**Steps:**
1. ✅ Toggle to dark mode → all pages readable
2. ✅ Badges and status colors visible in dark mode
3. ✅ Toggle to light mode → all pages readable
4. ✅ Chart/bar colors visible in both modes

---

## Scenario L — Accessibility

**Steps:**
1. ✅ Tab through interactive elements → focus ring visible
2. ✅ Icon buttons have tooltip or aria-label
3. ✅ Dialogs close with Escape key
4. ✅ Form fields show validation errors
5. ✅ Tables have proper headers
6. ✅ Loading states do not trap focus
