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
 * - useMocks: controls whether mock API data is served instead of real backend calls.
 * - mockAuthEnabled: controls whether mock auth/session is used (temporary until real Auth).
 *
 * These two settings are INDEPENDENT.
 * You can have mockAuthEnabled=true + useMocks=false (mock session, real API data).
 */
export const notifierRuntimeConfig = {
  /** Use mock API data instead of real backend calls */
  useMocks: parseBooleanEnv(
    process.env.NEXT_PUBLIC_NOTIFIER_USE_MOCKS,
    false,
  ),

  /** Use mock auth/session adapter (temporary until real Auth service) */
  mockAuthEnabled: parseBooleanEnv(
    process.env.NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED,
    false,
  ),

  /** Backend API base URL */
  apiBaseUrl:
    process.env.NEXT_PUBLIC_NOTIFIER_API_URL ??
    "http://127.0.0.1:9002/v1",

  /** Realtime mode */
  realtimeMode: (process.env.NEXT_PUBLIC_NOTIFIER_REALTIME_MODE ??
    "polling") as "websocket" | "sse" | "polling" | "disabled",

  /** Show live toast popups for new notifications */
  showLiveToasts: parseBooleanEnv(
    process.env.NEXT_PUBLIC_NOTIFIER_SHOW_LIVE_TOASTS,
    true,
  ),
} as const;
