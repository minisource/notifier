# MiniSource Notifier Test Center — Implementation & Audit Plan

This document defines the comprehensive architecture, endpoint inventory, multi-step flow suite, safety policies, component structure, and verification roadmap for building the **Notifier API Test Client & Flow Test Center** inside `C:\ActiveProjects\MiniSource\notifier\front`.

---

## 1. Executive Summary & Runtime Discovery

### Discovered Active Services & Ports
- **Notifier Frontend**: Running at `http://localhost:3003` / `http://localhost:3004` (Next.js with `next-intl` localization, Dokploy-inspired UI refactor active).
- **Notifier Backend**: Running at `http://localhost:9002` (Go Fiber service with GORM, Prometheus metrics, JWKS/Mock Auth, structured logging, tenant validation).
- **Auth Backend**: Running at `http://localhost:9001` (Auth reference service).
- **Gateway (Traefik / Proxy)**: Running at `http://localhost:8080` (routes `/v1/*` to backend services).
- **Runtime Rules**: All existing processes, containers, and hot-reload dev servers will be preserved. No processes will be restarted or killed unless explicitly justified.

---

## 2. Auth Test Client Architecture & Reusable Patterns

The Auth frontend (`C:\ActiveProjects\MiniSource\auth\front`) provides the reference implementation:
- **Client Page**: `src/app/(main)/admin/api-lab/page.tsx`
- **Execution Proxy Route**: `src/app/api/admin/api-lab/execute/route.ts`

### Adapted & Reused Concepts
1. **Server-side Execution Proxy**: Next.js Route Handler executing requests to backend to prevent CORS issues, handle correlation IDs, propagate trace headers (`traceparent`), measure millisecond timing, and enforce structured audit logging.
2. **Recursive Redaction Utility**: Centralized secret filtering case-insensitively masking authorization headers, JWT tokens, cookies, provider credentials, and secret parameters.
3. **Execution Stepper UI**: Visual step status tracker (`idle` | `running` | `success` | `failed`) with per-step timing, HTTP status badges, correlation IDs, and request/response inspectors.
4. **Local Execution History**: In-memory execution log tracking previous operations, durations, status codes, and request/response payloads.
5. **Permitted Headers Builder**: Custom header support for `X-Correlation-ID`, `X-Tenant-ID`, `Idempotency-Key`, `Accept-Language`.

### Key Differences for Notifier Test Center
- Adapt domain-specific catalog from Auth endpoints (OTP, login) to Notifier endpoints (Providers, Templates, Notifications, Deliveries, Preferences, Reminders).
- Add explicit **External Delivery Warnings** and **Side-Effect Safeguards** before executing notification dispatches.
- Include **Flow Engine Assertion & Value Extraction** for passing output IDs (e.g. `notificationId`, `templateKey`) between steps.

---

## 3. Discovered Notifier API Surface Inventory

| Domain | Method | Endpoint / Resolved Path | Auth Required | Safe Classification | Parameters / Body Highlights |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Health** | `GET` | `/v1/health` | Public | SAFE READ | Health status, service name |
| **Readiness** | `GET` | `/v1/ready` | Public | SAFE READ | System readiness |
| **Observability** | `GET` | `/v1/observability/health` | JWT / Admin | SAFE READ | Health breakdown |
| **Observability** | `GET` | `/v1/observability/metrics` | JWT / Admin | SAFE READ | System metrics |
| **Observability** | `GET` | `/v1/observability/queue` | JWT / Admin | SAFE READ | Queue depth, active workers |
| **Dashboard** | `GET` | `/v1/dashboard/overview` | JWT / Admin | SAFE READ | System analytics overview |
| **Providers** | `GET` | `/v1/providers` | JWT / Admin | SAFE READ | Provider list |
| **Providers** | `POST` | `/v1/providers` | JWT / Admin | LOCAL MUTATION | Create provider |
| **Providers** | `GET` | `/v1/providers/:id` | JWT / Admin | SAFE READ | Provider detail |
| **Providers** | `PUT` | `/v1/providers/:id` | JWT / Admin | LOCAL MUTATION | Update provider details |
| **Providers** | `POST` | `/v1/providers/:id/test` | JWT / Admin | SAFE VALIDATION | Test provider connectivity |
| **Providers** | `PATCH` | `/v1/providers/:id/status` | JWT / Admin | LOCAL MUTATION | Toggle provider active status |
| **Templates** | `GET` | `/v1/templates` | JWT / Admin | SAFE READ | Template list |
| **Templates** | `POST` | `/v1/templates` | JWT / Admin | LOCAL MUTATION | Create test template |
| **Templates** | `GET` | `/v1/templates/key/:key` | JWT / Admin | SAFE READ | Get template by key |
| **Templates** | `POST` | `/v1/templates/render-preview` | JWT / Admin | SAFE VALIDATION | Render template with sample variables |
| **Templates** | `PUT` | `/v1/templates/:id` | JWT / Admin | LOCAL MUTATION | Update template content |
| **Notifications**| `GET` | `/v1/notifications` | JWT / Admin | SAFE READ | List all notifications |
| **Notifications**| `POST` | `/v1/notifications/send` | JWT / Admin | EXTERNAL DELIVERY | Dispatch notification |
| **Notifications**| `GET` | `/v1/notifications/:id` | JWT / Admin | SAFE READ | Get notification status |
| **Notifications**| `GET` | `/v1/notifications/:id/deliveries` | JWT / Admin | SAFE READ | Get delivery logs |
| **Notifications**| `GET` | `/v1/notifications/:id/attempts` | JWT / Admin | SAFE READ | Get delivery attempts |
| **Notifications**| `POST` | `/v1/notifications/:id/retry` | JWT / Admin | EXTERNAL DELIVERY | Manual notification retry |
| **Notifications**| `POST` | `/v1/notifications/:id/cancel` | JWT / Admin | LOCAL MUTATION | Cancel pending notification |
| **Deliveries** | `GET` | `/v1/deliveries` | JWT / Admin | SAFE READ | List delivery attempts |
| **Deliveries** | `GET` | `/v1/deliveries/:id` | JWT / Admin | SAFE READ | Delivery detail |
| **Preferences** | `GET` | `/v1/me/preferences` | JWT | SAFE READ | Get user preferences |
| **Preferences** | `PUT` | `/v1/me/preferences` | JWT | LOCAL MUTATION | Update user preferences |
| **Reminders** | `GET` | `/v1/me/reminders` | JWT | SAFE READ | List user reminders |
| **Reminders** | `POST` | `/v1/me/reminders` | JWT | LOCAL MUTATION | Create scheduled reminder |
| **Settings** | `GET` | `/v1/admin/settings/notifications` | JWT / Admin | SAFE READ | Global settings |

---

## 4. Business Flow Inventory & Multi-Step Scenarios

### Flow A: Connectivity & Gateway Verification
1. Resolve frontend env configuration.
2. Call Gateway health (`/health`).
3. Call Notifier backend health (`/v1/health`).
4. Validate header propagation (`X-Correlation-ID`, `X-Tenant-ID`).
5. Compare Gateway route vs Direct Backend response.

### Flow B: Provider Lifecycle & Connectivity Check
1. Fetch provider list (`GET /v1/providers`).
2. Create test provider with unique key (`MINISOURCE_TEST_PROV_<timestamp>`).
3. Run provider connectivity test (`POST /v1/providers/:id/test`).
4. Update metadata & status.
5. Perform safe cleanup of test provider.

### Flow C: Template Lifecycle & Variable Rendering
1. Create unique test template (`MINISOURCE_TEST_TMPL_<timestamp>`).
2. Fetch created template by key (`GET /v1/templates/key/:key`).
3. Render preview with test payload (`POST /v1/templates/render-preview`).
4. Test negative case: render with missing variable.
5. Update template content.
6. Clean up test template.

### Flow D: Notification Preview & Validation (No Side Effect)
1. Select/Construct provider and template parameters.
2. Validate payload structure.
3. Perform dry-run preview (verifying variable replacement & target resolution).
4. Assert zero external delivery calls were dispatched.

### Flow E: End-to-End Notification Dispatch & Status Polling
1. *Requires user confirmation modal*.
2. Construct safe notification request (`type`, `recipient`, `templateKey`, `variables`).
3. Dispatch notification (`POST /v1/notifications/send`).
4. Capture returned `notificationId`.
5. Poll notification status with bounded attempts (max 5 attempts, 2s interval).
6. Inspect delivery attempts (`GET /v1/notifications/:id/attempts`).
7. Verify terminal status (`delivered`, `sent`, `failed`).

### Flow F: Authorization & Tenant Isolation Test
1. Request endpoint without token $\rightarrow$ Assert `401 Unauthorized`.
2. Request with invalid token $\rightarrow$ Assert `401 Unauthorized`.
3. Request with valid token but invalid tenant ID $\rightarrow$ Assert tenant boundary policy.

### Flow G: Idempotency & Duplicate Prevention
1. Send notification request with `Idempotency-Key: test-key-<uuid>`.
2. Repeat exact same request with identical key.
3. Verify backend handles duplicate submission without spawning duplicate dispatch jobs.

---

## 5. Information Architecture & Proposed File Structure

We will place the Test Center in the Notifier frontend app structure following the project refactor pattern:

```text
front/src/
├── app/
│   ├── [locale]/
│   │   └── test-center/
│   │       └── page.tsx                     # Main Test Center Page
│   └── api/
│       └── admin/
│           └── test-center/
│               └── execute/
│                   └── route.ts              # API Execution Proxy Handler
├── features/
│   └── test-center/
│       ├── components/
│       │   ├── operation-tester.tsx          # Individual Endpoint Form & Controls
│       │   ├── flow-tester.tsx               # Multi-Step Flow Runner
│       │   ├── request-response-inspector.tsx# Header, Body, Status, & Redaction Viewer
│       │   ├── execution-history-list.tsx    # Execution History Panel
│       │   ├── header-config-builder.tsx     # Custom Header & Token Input
│       │   └── external-delivery-modal.tsx   # Safeguard Confirmation Modal
│       ├── lib/
│       │   ├── catalog.ts                    # Approved Endpoints Catalog & Types
│       │   ├── flow-definitions.ts           # Flow Definitions (Flow A-G)
│       │   ├── flow-runner.ts                # Flow Engine & Step State Machine
│       │   ├── redaction.ts                 # Recursive Redaction Engine
│       │   └── error-normalizer.ts           # Error Classification Helper
│       └── types/
│           └── index.ts                      # Test Center TypeScript Interfaces
```

---

## 6. Security, Redaction & Safe Side-Effect Policy

1. **Centralized Redaction**: Redact sensitive headers (`Authorization`, `Cookie`, `Set-Cookie`) and body keys (`token`, `access_token`, `refresh_token`, `secret`, `password`, `api_key`, `client_secret`) recursively before display or copy.
2. **External Delivery Safeguard**: Any operation classified as `EXTERNAL DELIVERY` will be gated behind a non-bypassable user confirmation modal warning that real SMS/Email/Push may be sent if external credentials are enabled.
3. **Clean Test-Data Ownership**: All test resources created during flow execution will be prefixed with `MINISOURCE_TEST_` and cleaned up at the end of the flow. Cleanup failures will be reported separately without masking test flow success.

---

## 7. Verification Plan & Definition of Done

### Automated Tests
1. **Unit Tests**:
   - `redaction.test.ts`: Verify case-insensitive recursive secret masking.
   - `error-normalizer.test.ts`: Verify error classification (network, 401, 403, CORS, validation, server error).
   - `flow-runner.test.ts`: Verify value extraction, step dependency execution, and error handling.
2. **Typecheck & Lint**:
   - `npm run type-check` (or `tsc --noEmit`) in `notifier/front`.
   - `npm run lint` in `notifier/front`.

### Manual & Runtime Verification
- Execute runtime health check via Test Center against running Notifier backend (`:9002`) and Gateway (`:8080`).
- Execute Flow A (Connectivity), Flow C (Template Lifecycle), and Flow D (Preview & Validation).
- Verify request/response inspection, trace IDs, and duration timing.

---

## User Review Required

> [!IMPORTANT]
> **No Services Will Be Restarted**: Implementation will proceed using the already-running processes (`:9002` backend, `:8080` gateway, `:3003`/`:3004` frontend).
> **External Delivery Safeguards**: Live notification dispatches (SMS/Email) require explicit user interaction via confirmation modals in the Test Center.

---

## Plan Approval

Please review and confirm to proceed with execution.
