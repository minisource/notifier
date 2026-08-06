/**
 * Tenant as surfaced inside the Notifier app.
 *
 * Tenants are owned by the Auth service — the Notifier only *reads* the list
 * of tenants the current user belongs to (GET /users/me/tenants via the
 * gateway) and uses it for scoping notifications, templates, providers,
 * reminders, etc. Tenants can never be created or edited from the Notifier.
 */
export interface Tenant {
  id: string;
  name: string;
  slug: string;
  displayName?: string;
  logo?: string;
  description?: string;
  /** Auth tenant status, e.g. "active" | "inactive" | "suspended". */
  status: string;
  plan?: string;
  isDefault?: boolean;
  /** Current user's role inside this tenant (owner / member / role name). */
  role?: string;
  /** Derived from `status` for UI badges. */
  isActive: boolean;
  /** Kept for backward-compat with channel-tag UIs; always [] (Auth has no channels). */
  enabledChannels: string[];
  createdAt: string;
}
