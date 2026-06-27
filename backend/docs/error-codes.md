# Error Codes — Minisource Notifier

All API errors return the standard `ErrorResponse` format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description",
    "details": {}
  },
  "requestId": "uuid"
}
```

---

## Standard Error Codes

| Code | HTTP Status | Meaning | Example |
|------|-------------|---------|---------|
| `VALIDATION_ERROR` | 400 | Request body validation failed | Required field missing, invalid UUID format |
| `UNAUTHORIZED` | 401 | Authentication required | Missing/wrong JWT or service token |
| `FORBIDDEN` | 403 | Insufficient role/permission | Normal user accessing admin endpoint |
| `NOT_FOUND` | 404 | Resource not found | Notification ID doesn't exist |
| `CONFLICT` | 409 | State transition not allowed | Retrying a `sent` notification |
| `RATE_LIMITED` | 429 | Too many requests | Rate limit threshold exceeded |
| `PROVIDER_ERROR` | 502 | Provider returned an error | SMS provider API failure |
| `NOT_IMPLEMENTED` | 501 | Endpoint not yet implemented | Dashboard future route |
| `INTERNAL_ERROR` | 500 | Unexpected server error | Database connection failure |

---

## Specific Error Context

| Endpoint | Code | HTTP | When |
|----------|------|------|------|
| Any authenticated | `UNAUTHORIZED` | 401 | Missing/invalid Authorization header |
| Any admin | `FORBIDDEN` | 403 | User without admin role |
| `/me/*` | `FORBIDDEN` | 403 | User accessing another user's resource |
| `POST /notifications` | `CONFLICT` | 409 | Duplicate idempotency key |
| `POST /notifications/{id}/retry` | `CONFLICT` | 409 | Notification not in retryable state |
| `POST /notifications/{id}/cancel` | `CONFLICT` | 409 | Notification not in cancellable state |
| `POST /deliveries/{id}/retry` | `CONFLICT` | 409 | Delivery not in retryable state |
| `POST /providers/{id}/test` | `VALIDATION_ERROR` | 400 | Missing provider ID |
| `POST /templates` | `VALIDATION_ERROR` | 400 | Missing required template fields |
| `GET /notifications/{id}` | `NOT_FOUND` | 404 | Notification ID not found |
| Any rate-limited | `RATE_LIMITED` | 429 | Global or per-route rate limit exceeded |

---

## Error Response Examples

### Validation Error
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "userId is required"
  },
  "requestId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

### Authorization Error
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing authentication token"
  },
  "requestId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
}
```

### Forbidden
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Admin access required"
  },
  "requestId": "c3d4e5f6-a7b8-9012-cdef-123456789012"
}
```

### Conflict
```json
{
  "error": {
    "code": "CONFLICT",
    "message": "cannot retry notification with status 'sent': only failed/dead/retrying statuses can be retried"
  },
  "requestId": "d4e5f6a7-b8c9-0123-defa-234567890123"
}
```

### Rate Limited
```json
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Too many requests. Please try again later."
  },
  "requestId": "e5f6a7b8-c9d0-1234-efab-345678901234"
}
```

### Not Found
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Notification not found"
  },
  "requestId": "f6a7b8c9-d0e1-2345-fabc-456789012345"
}
```

### Internal Error
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to list notifications: connection refused"
  },
  "requestId": "a7b8c9d0-e1f2-3456-abcd-567890123456"
}
```

---

## Error Code Constants

Defined in `api/v1/dto/notification_dto.go`:

```go
const (
    ErrorCodeValidation     = "VALIDATION_ERROR"
    ErrorCodeUnauthorized   = "UNAUTHORIZED"
    ErrorCodeForbidden      = "FORBIDDEN"
    ErrorCodeNotFound       = "NOT_FOUND"
    ErrorCodeConflict       = "CONFLICT"
    ErrorCodeRateLimited    = "RATE_LIMITED"
    ErrorCodeProvider       = "PROVIDER_ERROR"
    ErrorCodeNotImplemented = "NOT_IMPLEMENTED"
    ErrorCodeInternal       = "INTERNAL_ERROR"
)
```
