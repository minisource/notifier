# Notifier Frontend — Security Notes

> Last updated: June 2024

## Auth Limitations

- **No real auth service** — mock auth adapter only
- **No JWT validation** — mock token is a plain string
- **No token refresh** — mock token never expires
- **No login/logout flow** — users change roles via Settings page
- **No auth middleware** — frontend routes are unprotected; only UI-level RoleGuard exists

### Mock Mode Risks

| Risk | Description | Mitigation |
|------|-------------|------------|
| Mock enabled in production | `NEXT_PUBLIC_NOTIFIER_USE_MOCKS` could be accidentally set to `true` | Default is `false` in `.env.example`; visible "Mock" badge in topbar |
| Mock token treated as real | Mock access token is a hardcoded string | Clear "Mock Mode" label in settings; no secret value |
| Roles not enforced server-side | RoleGuard is client-side only | Backend must enforce access control independently |

## Safe Link Handling

The notification center handles external links via `notification.metadata.link`:

```typescript
// Locale-aware safe link check
if (link.startsWith('http://') || link.startsWith('https://')) {
  window.open(link, '_blank', 'noopener,noreferrer');
  return;
}
```

- Only `http://` and `https://` protocols are allowed
- `noopener,noreferrer` prevents tab-napping
- Non-HTTP links (e.g., `javascript:`) are blocked
- If no safe link, navigates to notification detail page internally

## PII Redaction

### Metadata Viewer (`JsonViewer`)
The `JsonViewer` component redacts values for these object keys:
- `password`, `token`, `secret`, `key`, `api_key`, `apiKey`
- `access_key`, `accessKey`, `private_key`, `privateKey`
- `authorization`, `credential`, `auth`

Sensitive values are replaced with the text from `t('notifications.metadata_sensitive')`.

### Recipient Masking
- Email: `maskEmail()` function masks the local part (e.g., `a***@example.com`)
- Phone: `maskPhone()` function masks middle digits (e.g., `+98912*****67`)

## Admin Action Confirmations

All destructive/risky actions require confirmation:

| Action | Dialog | Notes |
|--------|--------|-------|
| Retry notification | ✅ ConfirmDialog | "Notification will be requeued" |
| Cancel notification | ✅ ConfirmDialog | "Cannot be undone" |
| Delete template | ✅ ConfirmDialog | Destructive |
| Cancel reminder | ✅ ConfirmDialog | Permanently cancels |
| Retry delivery | ✅ ConfirmDialog | "Delivery will be requeued" |
| Provider test (dryRun=false) | ⚠️ Warning shown | Warning before real send |

## Provider Test Warnings

The provider test dialog:
- Defaults `dryRun=true` (safe mode)
- Shows warning when `dryRun=false`: "This will send a real notification"
- Sanitizes provider response before displaying
- Does not display raw secrets from provider config

## What Needs Backend Security

The following cannot be fully secured on the frontend:

1. **API authorization** — Backend must validate JWT and check roles
2. **Rate limiting** — Backend must enforce per-tenant rate limits
3. **Data isolation** — Backend must ensure tenants can't access each other's data
4. **Audit logging** — Backend should log all admin actions
5. **Sensitive config** — Provider credentials should never reach frontend

## Remaining Risks Before Real Auth

1. **No real token validation** — Mock token accepted everywhere
2. **No session timeout** — Mock session lasts indefinitely
3. **No CSRF protection** — Not implemented (Next.js has basic CSRF via same-origin)
4. **No content security policy** — Not configured in Next.js headers
5. **No rate limiting** — Frontend polling could be abused (use debounce if needed)
