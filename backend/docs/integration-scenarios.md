# Integration Scenarios — Minisource Notifier

> Integration flows documented for production validation.

---

## Scenario A — Internal Service Sends Notification

**Goal**: Verify that an internal service can create a notification via the service API.

### Flow

1. Internal service authenticates with service token (JWT with `notifications:send` scope)
2. `POST /api/v1/service/notifications` with:
   - `userId`, `type`, `body`
   - `Idempotency-Key` header
3. Notification is created with status `pending`
4. Notification is enqueued for worker processing
5. Response returns `201 Created` with notification ID and status

### Expected Results

- ✅ Service auth required — normal user JWT is rejected
- ✅ Tenant/project persisted from service context
- ✅ Idempotency key prevents duplicates (second request returns 409)
- ✅ For in-app notifications, WebSocket broadcast occurs
- ✅ Preference filter checked before creation

### Test Commands

```bash
# Create notification
curl -X POST "{{baseUrl}}/api/v1/service/notifications" \
  -H "Authorization: Bearer {{serviceToken}}" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-123" \
  -d '{
    "userId": "123e4567-e89b-12d3-a456-426614174000",
    "type": "email",
    "body": "Hello from service"
  }'

# Duplicate (should return 409)
curl -X POST "{{baseUrl}}/api/v1/service/notifications" \
  -H "Authorization: Bearer {{serviceToken}}" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-123" \
  -d '{
    "userId": "123e4567-e89b-12d3-a456-426614174000",
    "type": "email",
    "body": "Hello from service"
  }'
```

---

## Scenario B — Logged-in User Reads Notifications

**Goal**: Verify that a user can read their own notifications via `/me` endpoints.

### Flow

1. User has JWT with `userId` claim
2. `GET /api/v1/me/notifications` — lists user's notifications
3. `GET /api/v1/me/notifications/{id}` — reads specific notification
4. `PUT /api/v1/me/notifications/{id}/read` — marks as read
5. `POST /api/v1/me/notifications/read-all` — marks all as read

### Expected Results

- ✅ userId derived from JWT (not from path/query/body)
- ✅ User cannot read other user's notifications (403)
- ✅ Read/read-all affect only current user's notifications
- ✅ Unread count decrements after marking read

### Test Commands

```bash
# List notifications
curl "{{baseUrl}}/api/v1/me/notifications?page=1&pageSize=10" \
  -H "Authorization: Bearer {{userToken}}"

# Mark all as read
curl -X POST "{{baseUrl}}/api/v1/me/notifications/read-all" \
  -H "Authorization: Bearer {{userToken}}"
```

---

## Scenario C — Admin Manages Notifications

**Goal**: Verify admin can list/retry/cancel any notification.

### Flow

1. Admin has JWT with `admin` or `super_admin` role
2. `GET /api/v1/admin/notifications?userId=...` — filter by user
3. `POST /api/v1/admin/notifications/{id}/retry` — retry failed/dead
4. `POST /api/v1/admin/notifications/{id}/cancel` — cancel pending
5. Normal user attempts same actions

### Expected Results

- ✅ Admin role required — normal user gets 403
- ✅ Retry enforces state transition (fails for `sent`/`delivered`)
- ✅ Cancel enforces state transition (fails for `sent`/`failed`)
- ✅ 409 for invalid state transitions

### Test Commands

```bash
# Admin lists all notifications
curl "{{baseUrl}}/api/v1/admin/notifications?status=failed&page=1" \
  -H "Authorization: Bearer {{adminToken}}"

# Admin retries
curl -X POST "{{baseUrl}}/api/v1/admin/notifications/{id}/retry" \
  -H "Authorization: Bearer {{adminToken}}"

# Normal user forbidden
curl "{{baseUrl}}/api/v1/admin/notifications" \
  -H "Authorization: Bearer {{userToken}}"
# Expected: 403 Forbidden
```

---

## Scenario D — Template Render + Notification Send

**Goal**: Verify template creation, rendering, and usage in notifications.

### Flow

1. Admin creates template with variables
2. Admin render-preview to validate
3. Service API sends notification with `templateId` + variables in metadata
4. Notification body rendered from template

### Expected Results

- ✅ Template key/locale/channel respected
- ✅ Variables substituted correctly
- ✅ Missing variables handled safely (no crash)
- ✅ Preview returns used and missing variables

---

## Scenario E — User Preferences Affect Notification

**Goal**: Verify that user notification preferences block or redirect notifications.

### Flow

1. User disables SMS preference via `PUT /api/v1/me/preferences`
2. Service creates an SMS notification for the user
3. Preference filter checks and blocks the notification

### Expected Results

- ✅ Preference is checked before notification creation
- ✅ Blocked notification returns error with reason
- ✅ Critical bypass (`priority=urgent`) works if implemented

### Notes

- Preference filter is applied at service level in `NotificationService.CreateNotification`
- If preference check blocks the notification, the API returns an error
- For digest-only preferences, notification is created with `digested` status

---

## Scenario F — Reminder Lifecycle

**Goal**: Verify reminder CRUD and processing.

### Flow

1. User creates a scheduled reminder
2. User lists reminders
3. User updates scheduled reminder
4. User cancels reminder before due time
5. Due reminder is processed (integrated with worker)

### Expected Results

- ✅ Reminder correctly scheduled with `scheduledAt`
- ✅ Scheduled reminders can be updated/cancelled
- ✅ Past/completed reminders cannot be invalidly mutated (409)
- ✅ Tenant/user ownership enforced

---

## Scenario G — Delivery Failure and Retry

**Goal**: Verify that delivery failures are recorded and retries work correctly.

### Flow

1. Provider fails to send (configured for simulation)
2. Delivery attempt is recorded in notification log
3. Notification status becomes `failed`
4. Admin reviews delivery via `GET /admin/deliveries/{id}`
5. Admin retries via `POST /admin/deliveries/{id}/retry`
6. New attempt recorded
7. After max retries, status becomes `dead`

### Expected Results

- ✅ No infinite retry loop (configurable `maxRetries`, default 3)
- ✅ Dead-letter items counted in queue overview
- ✅ Dead items can be manually retried by admin
- ✅ Provider response sanitized (no secrets)
- ✅ 409 for invalid state transitions

---

## Scenario H — Dashboard and Observability

**Goal**: Verify admin dashboard, metrics, and queue overview return real data.

### Flow

1. Admin opens dashboard
2. `GET /api/v1/admin/dashboard/overview` — aggregates
3. `GET /api/v1/admin/observability/metrics` — operational metrics
4. `GET /api/v1/admin/observability/queue` — queue status
5. `GET /api/v1/admin/providers/health` — provider status

### Expected Results

- ✅ Admin only — normal user gets 403
- ✅ Real aggregated data (not faked)
- ✅ No secrets in responses
- ✅ Useful health/metrics for operations
- ✅ All responses typed

---

## Scenario I — Rate Limiting

**Goal**: Verify rate limiter blocks excessive requests.

### Flow

1. Send 100+ requests per minute to any protected endpoint
2. Rate limiter returns 429 after threshold

### Expected Results

- ✅ 429 Too Many Requests
- ✅ ErrorResponse format with `RATE_LIMITED` code
- ✅ `requestId` included in error response

---

## Scenario J — Request ID Propagation

**Goal**: Verify X-Request-Id generation and propagation.

### Flow

1. Send request without X-Request-Id header
2. Send request with X-Request-Id header

### Expected Results

- ✅ Request without header gets auto-generated UUID
- ✅ Request with header preserves the provided ID
- ✅ Response includes X-Request-Id header
- ✅ Error responses include requestId in body
