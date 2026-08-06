/**
 * Notifier Runtime Configuration
 *
 * Centralized env parsing for the Notifier frontend module.
 * All env vars are parsed ONCE and exposed through this config object.
 *
 * IMPORTANT: Never use Boolean(process.env.X) — Boolean("false") === true!
 */

export function parseBooleanEnv(
  value: string | undefined,
  defaultValue = false,
): boolean {
  if (value == null || value.trim() === "") return defaultValue;

  const normalized = value.trim().toLowerCase();

  if (["true", "1", "yes", "on"].includes(normalized)) return true;
  if (["false", "0", "no", "off"].includes(normalized)) return false;

  return defaultValue;
}

/**
 * Notifier-specific runtime config.
 *
 * - authEnabled: controls whether the app enforces authentication through the
 *   Auth service (NEXT_PUBLIC_NOTIFIER_AUTH_ENABLED). When enabled the app
 *   shows the login page and requires a real session/token. When disabled it
 *   boots directly with a local admin session so the Notifier works
 *   standalone — no Auth service required.
 *
 * The Notifier itself always talks to the real backend regardless of this
 * flag (mock API data has been removed).
 */
export const notifierRuntimeConfig = {
  /** Enforce Auth-service authentication (NEXT_PUBLIC_NOTIFIER_AUTH_ENABLED) */
  authEnabled: parseBooleanEnv(
    process.env.NEXT_PUBLIC_NOTIFIER_AUTH_ENABLED,
    true,
  ),

  /** Backend API base URL */
  apiBaseUrl:
    process.env.NEXT_PUBLIC_NOTIFIER_API_URL ??
    (typeof window !== "undefined" ? "/v1" : "http://minisource-notifier-backend:9002/v1"),

  /** Auth service API base URL */
  authApiUrl:
    process.env.NEXT_PUBLIC_AUTH_API_URL ??
    (typeof window !== "undefined" ? "/v1" : "http://minisource-auth-backend:9001/v1"),

  /** Realtime mode */
  realtimeMode: (process.env.NEXT_PUBLIC_NOTIFIER_REALTIME_MODE ??
    "polling") as "websocket" | "sse" | "polling" | "disabled",

  /** Show live toast popups for new notifications */
  showLiveToasts: parseBooleanEnv(
    process.env.NEXT_PUBLIC_NOTIFIER_SHOW_LIVE_TOASTS,
    true,
  ),
} as const;
