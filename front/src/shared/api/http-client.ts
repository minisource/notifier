import { authAdapter } from '@/shared/auth/auth-adapter';
import { ApiError, type ApiErrorResponse } from './api-error';

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

interface RequestConfig {
  method?: HttpMethod;
  headers?: Record<string, string>;
  body?: unknown;
  params?: Record<string, string | number | boolean | undefined | null>;
  timeout?: number;
  signal?: AbortSignal;
}

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
  const url = new URL(path.startsWith('http') ? path : `${baseUrl}${path.startsWith('/') ? '' : '/'}${path}`, baseUrl);

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

  // Only add Content-Type for requests with a body (POST, PUT, PATCH)
  // GET and DELETE requests don't need it — avoiding it eliminates
  // unnecessary CORS preflight (OPTIONS) requests.
  if (config?.body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }

  const session = authAdapter.getSession();

  if (session.accessToken) {
    headers['Authorization'] = `Bearer ${session.accessToken}`;
  }

  if (session.tenantId) {
    headers['X-Tenant-Id'] = session.tenantId;
  }

  if (session.projectId) {
    headers['X-Project-Id'] = session.projectId;
  }

  return { ...headers, ...config?.headers };
}

async function handleResponse(response: Response): Promise<never | unknown> {
  if (!response.ok) {
    let body: ApiErrorResponse | undefined;
    try {
      body = await response.json() as ApiErrorResponse;
    } catch {
      // Body not JSON
    }
    throw body ? ApiError.fromResponse(response, body) : new ApiError(
      response.status,
      'HTTP_ERROR',
      response.statusText || 'Request failed',
    );
  }

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
  // The backend (go-common/response) wraps all successful responses in
  // { "success": true, "data": <payload> }. The http client extracts
  // the data field so consumers get the payload directly.
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

function getBaseUrl(): string {
  if (typeof process !== 'undefined' && process.env && process.env.NEXT_PUBLIC_NOTIFIER_API_URL) {
    return process.env.NEXT_PUBLIC_NOTIFIER_API_URL;
  }
  return 'http://127.0.0.1:9002/v1';
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
    return (await handleResponse(response)) as T;
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
