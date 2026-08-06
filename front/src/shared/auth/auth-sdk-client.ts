/**
 * Auth SDK wiring for the Notifier frontend.
 *
 * Bridges the @minisource/api-core transport to the Notifier runtime:
 * - base URL: resolved at runtime (browser: relative `/v1` through the gateway;
 *   server: the auth backend container URL) — never hardcoded here.
 * - access token: supplied by the existing `authAdapter` (local/real session).
 * - context headers: tenant + project from the active session / tenant store.
 *
 * The public SDK surface (`createAuthClient().users.search(...)`) is used by
 * domain components (e.g. SecureUserPicker) instead of ad-hoc fetch wrappers.
 */

import { createApiClient, type ApiClient, type ContextHeaders } from '@minisource/api-core';
import { createAuthClient, type AuthClient } from '@minisource/auth-sdk';
import { authAdapter } from './auth-adapter';
import { useTenantStore } from '@/stores/tenant.store';
import { notifierRuntimeConfig } from '@/features/notifier/config/notifier-config';

/**
 * Resolve the Auth service base URL at runtime.
 * Browser: relative `/v1` so requests flow through the current origin (gateway).
 * Server: the container hostname used by the backend.
 */
function resolveAuthBaseUrl(): string {
  if (typeof window !== 'undefined') return notifierRuntimeConfig.authApiUrl;
  return notifierRuntimeConfig.authApiUrl;
}

/** Build context headers from the active session + tenant store. */
function buildContextHeaders(): ContextHeaders {
  const headers: ContextHeaders = {};
  const session = authAdapter.getSession();

  const activeTenantId = useTenantStore.getState().activeTenant?.id;
  const effectiveTenantId =
    activeTenantId && activeTenantId !== 'all' ? activeTenantId : session.tenantId;
  if (effectiveTenantId) headers['X-Tenant-Id'] = effectiveTenantId;
  if (session.projectId) headers['X-Project-Id'] = session.projectId;

  return headers;
}

/** The shared api-core transport for Auth calls. */
export const authTransport: ApiClient = createApiClient({
  baseUrl: resolveAuthBaseUrl(),
  unwrapEnvelope: true,
  getAccessToken: () => authAdapter.getAccessToken(),
  getContextHeaders: buildContextHeaders,
});

/** The stable Auth SDK facade used across the Notifier frontend. */
export const authClient: AuthClient = createAuthClient({ transport: authTransport });

export type { AuthClient } from '@minisource/auth-sdk';
export type { ApiClient } from '@minisource/api-core';
