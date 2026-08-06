import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';

export interface ApiError {
  message: string;
  code?: string;
  status?: number;
}

function getApiBaseUrl(): string {
  if (typeof window !== 'undefined') {
    if (process.env.NEXT_PUBLIC_API_URL) {
      return process.env.NEXT_PUBLIC_API_URL;
    }
    // In browser: relative /v1 path routing through Gateway or current origin
    return '/v1';
  }
  return process.env.NEXT_PUBLIC_API_URL || 'http://minisource-notifier-backend:9002/v1';
}

function createApiClient(): AxiosInstance {
  const timeout = Number(process.env.NEXT_PUBLIC_API_TIMEOUT) || 30000;

  const client = axios.create({
    baseURL: getApiBaseUrl(),
    timeout,
    headers: { 'Content-Type': 'application/json' },
  });

  client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    config.baseURL = getApiBaseUrl();
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('accessToken');
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  });

  client.interceptors.response.use(
    (response) => response,
    async (error: AxiosError<ApiError>) => {
      if (error.response?.status === 401) {
        // Clear auth tokens from both localStorage and sessionStorage
        try {
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          sessionStorage.removeItem('accessToken');
          sessionStorage.removeItem('refreshToken');
          sessionStorage.removeItem('notifier-admin-token');
          sessionStorage.removeItem('notifier-refresh-token');
          localStorage.removeItem('notifier-auth');
        } catch { /* ignore */ }

        // Redirect to auth/front login page (served at /auth via the gateway)
        if (typeof window !== 'undefined') {
          const returnUrl = encodeURIComponent(
            window.location.pathname + window.location.search
          );
          window.location.href = `/auth/login?returnUrl=${returnUrl}`;
          return new Promise(() => {});
        }
      }
      const apiError: ApiError = {
        message:
          error.response?.data?.message || error.message || 'An error occurred',
        code: error.response?.data?.code || error.code,
        status: error.response?.status,
      };
      return Promise.reject(apiError);
    }
  );

  return client;
}

export const apiClient = createApiClient();

export const api = {
  get: <T>(url: string, params?: Record<string, unknown>) =>
    apiClient.get<T>(url, { params }).then((res) => res.data),
  post: <T>(url: string, data?: unknown) =>
    apiClient.post<T>(url, data).then((res) => res.data),
  put: <T>(url: string, data?: unknown) =>
    apiClient.put<T>(url, data).then((res) => res.data),
  patch: <T>(url: string, data?: unknown) =>
    apiClient.patch<T>(url, data).then((res) => res.data),
  delete: <T>(url: string) =>
    apiClient.delete<T>(url).then((res) => res.data),
};
