# Plan: Reusable Secure User Picker for Notifier Frontend

## 1. Auth User Search Capabilities & Routes
Auth backend exposes a paginated user list with filters under the path:
*   Route: `GET /api/v1/admin/users` (proxied via Gateway at `http://localhost:8080/api/v1/admin/users`)
*   Query Parameters:
    *   `search` (string): Filters users by email, username, first name, last name, phone.
    *   `page` (int, default: 1): Page number.
    *   `pageSize` (int, default: 20): Bounded limit.
    *   `isActive` (bool, optional): Filter active status.
*   Enforced permissions: RoleAdmin or RoleSuperAdmin.
*   Tenant isolation: Enforced server-side in user repository `ListWithFilters` using the header `X-Tenant-ID`. It queries members belonging to the current tenant or active in it.

## 2. Privacy & Enumeration Mitigations
*   **Minimum Query Length**: Query requests are only triggered when the user types 2 or more characters. Blank input returns an empty list.
*   **PII Masking**: Emails and phones are masked on the frontend (e.g., `a***@example.com` and `+98******1234`) when displayed in the autocomplete dropdown.
*   **Abort & Debounce**: Typing is debounced by 300ms. In-flight requests are automatically aborted via `AbortController` to prevent race conditions and stale response substitution.
*   **Isolation**: Selection values only preserve the canonical stable `userId` (UUID). Device/Push tokens are never exposed or managed in the frontend.

## 3. Frontend Architecture
We will create:
1.  **Component**: `SecureUserPicker` in `notifier/front/src/features/test-center/components/secure-user-picker.tsx`.
2.  **Hook**: `useUserSearch` inside `secure-user-picker.tsx` or a separate hook.
3.  **UI Primitives**: Popover, Command, Combobox from `@minisource/ui`.

## 4. Integration Points in Notifier
*   **Send Notification Form**: For `in_app` and `push` channels, the manual recipient input will be replaced by `<SecureUserPicker />`.
*   **Admin Feed Queries**: Replace manual `filterUserId` input in `AdminNotificationsTester` with `<SecureUserPicker />`.
*   **Notification Actions**: Replace manual user ID fields in other actions.
*   **Test Center Flows**: Ensure selected user ID is passed safely across test flow steps.

## 5. Definition of Done
*   Auth is the source of truth for user search.
*   Query parameter binding is server-side and tenant-isolated.
*   All interactive user-ID inputs are updated to use the Secure User Picker.
*   Tests and typechecks pass successfully.
