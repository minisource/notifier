import { authAdapter, clearSession, setRealSession, isAuthEnabled } from '@/shared/auth/auth-adapter';
import { authApi } from '@/shared/auth/auth-api';
import { ApiError, type ApiErrorResponse } from './api-error';
import type { AuthSession } from '@/shared/auth/auth-types';
import { useTenantStore } from '@/stores/tenant.store';

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

interface RequestConfig {
  method?: HttpMethod;
  headers?: Record<string, string>;
  body?: unknown;
  params?: Record<string, string | number | boolean | undefined | null>;
  timeout?: number;
  signal?: AbortSignal;
  skipAuth?: boolean; // Skip Authorization header
}

// Token refresh state to avoid concurrent refresh calls
let refreshPromise: Promise<AuthSession | null> | null = null;

function generateRequestId(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16);
  });
}

function buildUrl(baseUrl: string, path: string, params?: Record<string, string | number | boolean | undefined | null>): string {
  // If the path itself is already an absolute URL, use it directly
  if (path.startsWith('http://') || path.startsWith('https://')) {
    const url = new URL(path);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          url.searchParams.set(key, String(value));
        }
      });
    }
    return url.toString();
  }

  let resolvedBase = baseUrl;
  if (baseUrl.startsWith('/')) {
    if (typeof window !== 'undefined') {
      resolvedBase = `${window.location.origin}${baseUrl}`;
    } else {
      // Server-side default
      resolvedBase = `http://minisource-notifier-backend:9002${baseUrl}`;
    }
  }

  // Ensure resolvedBase is absolute
  if (!resolvedBase.startsWith('http://') && !resolvedBase.startsWith('https://')) {
    resolvedBase = typeof window !== 'undefined'
      ? `${window.location.origin}/v1`
      : 'http://minisource-notifier-backend:9002/v1';
  }

  const baseWithSlash = resolvedBase.endsWith('/') ? resolvedBase : `${resolvedBase}/`;
  const cleanPath = path.startsWith('/') ? path.slice(1) : path;
  
  const url = new URL(`${baseWithSlash}${cleanPath}`);

  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        url.searchParams.set(key, String(value));
      }
    });
  }

  return url.toString();
}

function buildHeaders(config?: RequestConfig): Record<string, string> {
  const headers: Record<string, string> = {
    'X-Request-Id': generateRequestId(),
  };

  // Only add Content-Type for requests with a body
  if (config?.body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }

  if (!config?.skipAuth) {
    const session = authAdapter.getSession();
    if (session.accessToken) {
      headers['Authorization'] = `Bearer ${session.accessToken}`;
    }

    // Tenant scoping: prefer the active tenant selected in the UI, fall back
    // to the session tenant. The 'all' pseudo-tenant means no tenant filter,
    // so the header is omitted.
    const activeTenantId = useTenantStore.getState().activeTenant?.id;
    const effectiveTenantId =
      activeTenantId && activeTenantId !== 'all'
        ? activeTenantId
        : session.tenantId;
    if (effectiveTenantId) {
      headers['X-Tenant-Id'] = effectiveTenantId;
    }

    if (session.projectId) {
      headers['X-Project-Id'] = session.projectId;
    }
  }

  return { ...headers, ...config?.headers };
}

async function handleResponse(response: Response, path: string, config?: RequestConfig): Promise<never | unknown> {
  // Handle 401 Unauthorized — attempt token refresh once
  if (response.status === 401 && !config?.skipAuth) {
    const refreshed = await attemptTokenRefresh();
    if (refreshed) {
      // Retry original request with new token using the original path
      const retryHeaders = buildHeaders(config);
      const baseUrl = getBaseUrl();
      const retryUrl = buildUrl(baseUrl, path, config?.params);

      const retryResponse = await fetch(
        retryUrl,
        {
          method: config?.method || 'GET',
          headers: retryHeaders,
          body: config?.body ? JSON.stringify(config.body) : undefined,
          signal: config?.signal,
        },
      );

      if (retryResponse.ok) {
        return processSuccessResponse(retryResponse);
      }

      // If refresh failed again with 401, clear session
      if (retryResponse.status === 401) {
        clearSession();
        redirectToLoginIfAuthEnabled();
        throw ApiError.unauthorized('Session expired. Please sign in again.');
      }

      // Handle the retry response
      return processErrorResponse(retryResponse);
    }

    // Refresh failed or no refresh token — clear session
    clearSession();
    redirectToLoginIfAuthEnabled();
    throw ApiError.unauthorized('Session expired. Please sign in again.');
  }

  // Handle 403 Forbidden
  if (response.status === 403) {
    let body: ApiErrorResponse | undefined;
    try {
      body = await response.json() as ApiErrorResponse;
    } catch {
      // Body not JSON
    }
    const message = body?.error?.message || 'Forbidden. Admin access required.';
    throw ApiError.forbidden(message);
  }

  if (!response.ok) {
    return processErrorResponse(response);
  }

  return processSuccessResponse(response);
}

async function attemptTokenRefresh(): Promise<boolean> {
  const session = authAdapter.getSession();
  if (!session.refreshToken) {
    return false;
  }

  // Deduplicate concurrent refresh attempts
  if (!refreshPromise) {
    refreshPromise = (async () => {
      try {
        const resp = await authApi.refresh(session.refreshToken!);
        const newSession: AuthSession = {
          ...session,
          accessToken: resp.accessToken,
          refreshToken: resp.refreshToken,
        };
        setRealSession(newSession);
        return newSession;
      } catch {
        clearSession();
        return null;
      }
    })();
  }

  const result = await refreshPromise;
  refreshPromise = null;
  return result !== null;
}

async function processSuccessResponse(response: Response): Promise<unknown> {
  // 204 No Content
  if (response.status === 204) {
    return undefined;
  }

  let body: unknown;
  try {
    body = await response.json() as unknown;
  } catch {
    return undefined;
  }

  // Unwrap standard response envelope: { success: true, data: ... }
  if (
    body !== null &&
    typeof body === 'object' &&
    !Array.isArray(body) &&
    'success' in (body as Record<string, unknown>) &&
    (body as Record<string, unknown>).success === true &&
    'data' in (body as Record<string, unknown>)
  ) {
    return (body as Record<string, unknown>).data;
  }

  return body;
}

async function processErrorResponse(response: Response): Promise<never> {
  let body: ApiErrorResponse | undefined;
  try {
    body = await response.json() as ApiErrorResponse;
  } catch {
    // Body not JSON
  }

  if (response.status === 403) {
    const message = body?.error?.message || 'Forbidden';
    throw ApiError.forbidden(message);
  }

  throw body
    ? ApiError.fromResponse(response, body)
    : new ApiError(
        response.status,
        'HTTP_ERROR',
        response.statusText || 'Request failed',
      );
}

/**
 * Only redirect to /login if auth is explicitly enabled.
 * When auth is disabled, don't redirect — the app runs without login.
 */
function redirectToLoginIfAuthEnabled(): void {
  if (isAuthEnabled() && typeof window !== 'undefined' && !window.location.pathname.startsWith('/auth/login')) {
    window.location.href = '/auth/login';
  }
}

function getBaseUrl(): string {
  if (typeof process !== 'undefined' && process.env && process.env.NEXT_PUBLIC_NOTIFIER_API_URL) {
    return process.env.NEXT_PUBLIC_NOTIFIER_API_URL;
  }
  if (typeof window !== 'undefined') {
    return '/v1';
  }
  return 'http://minisource-notifier-backend:9002/v1';
}

export async function request<T = unknown>(path: string, config?: RequestConfig): Promise<T> {
  const baseUrl = getBaseUrl();
  const url = buildUrl(baseUrl, path, config?.params);
  const headers = buildHeaders(config);
  const timeout = config?.timeout ?? 30000;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  // Combine external signal with timeout signal
  const signal = config?.signal
    ? combineAbortSignals(config.signal, controller.signal)
    : controller.signal;

  try {
    const response = await fetch(url, {
      method: config?.method || 'GET',
      headers,
      body: config?.body ? JSON.stringify(config.body) : undefined,
      signal,
    });

    clearTimeout(timeoutId);
    return (await handleResponse(response, path, config)) as T;
  } catch (error) {
    clearTimeout(timeoutId);
    if (error instanceof ApiError) {
      throw error;
    }
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw ApiError.timeout();
    }
    throw ApiError.networkError(error as Error);
  }
}

function combineAbortSignals(...signals: AbortSignal[]): AbortSignal {
  const controller = new AbortController();
  for (const signal of signals) {
    if (signal.aborted) {
      controller.abort(signal.reason);
      return controller.signal;
    }
    signal.addEventListener('abort', () => controller.abort(signal.reason), { once: true });
  }
  return controller.signal;
}

// Convenience methods
export const http = {
  get: <T = unknown>(path: string, config?: RequestConfig) => request<T>(path, { ...config, method: 'GET' }),
  post: <T = unknown>(path: string, body?: unknown, config?: RequestConfig) => request<T>(path, { ...config, method: 'POST', body }),
  put: <T = unknown>(path: string, body?: unknown, config?: RequestConfig) => request<T>(path, { ...config, method: 'PUT', body }),
  patch: <T = unknown>(path: string, body?: unknown, config?: RequestConfig) => request<T>(path, { ...config, method: 'PATCH', body }),
  delete: <T = unknown>(path: string, config?: RequestConfig) => request<T>(path, { ...config, method: 'DELETE' }),
};
