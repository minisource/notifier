import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';

export interface ApiError {
  message: string;
  code?: string;
  status?: number;
}

function createApiClient(): AxiosInstance {
  const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://127.0.0.1:9002/v1';
  const timeout = Number(process.env.NEXT_PUBLIC_API_TIMEOUT) || 30000;

  const client = axios.create({ baseURL, timeout, headers: { 'Content-Type': 'application/json' } });

  client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
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
        localStorage.removeItem('accessToken');
        if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login')) {
          window.location.href = '/login';
        }
      }
      const apiError: ApiError = {
        message: error.response?.data?.message || error.message || 'An error occurred',
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
