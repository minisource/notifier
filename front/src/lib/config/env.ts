import { parseBooleanEnv } from "@/features/notifier/config/notifier-config";

export const env = {
  apiMode: (process.env.NEXT_PUBLIC_API_MODE ?? "mock") as "mock" | "real",
  apiUrl:
    process.env.NEXT_PUBLIC_NOTIFIER_API_URL ??
    "http://localhost:9002/v1",
  defaultLocale: (process.env.NEXT_PUBLIC_DEFAULT_LOCALE ?? "fa") as string,
  appName: process.env.NEXT_PUBLIC_APP_NAME ?? "Notifier Admin",
  appUrl: process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000",
  appVersion: process.env.NEXT_PUBLIC_APP_VERSION ?? "1.0.0",
  /** Whether the app-level API mode is 'mock' */
  isMockMode: (process.env.NEXT_PUBLIC_API_MODE ?? "mock") === "mock",
  /** Whether NOTIFIER-specific mock data is explicitly enabled */
  isMockDataEnabled: parseBooleanEnv(
    process.env.NEXT_PUBLIC_NOTIFIER_USE_MOCKS,
    false,
  ),
} as const;

/**
 * @deprecated Use `env.isMockMode` directly or check `NEXT_PUBLIC_API_MODE`.
 * This checks the general app-level API mode, NOT the notifier-specific mock data flag.
 * For notifier mock data, use `env.isMockDataEnabled`.
 */
export function isMockMode(): boolean {
  return env.apiMode === "mock";
}

/**
 * Returns true when the Notifier module should use mock data.
 * Controlled by NEXT_PUBLIC_NOTIFIER_USE_MOCKS env var (default: false).
 * Independent of mock auth/session.
 */
export function isMockDataEnabled(): boolean {
  return env.isMockDataEnabled;
}
