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

export const config = {
  app: {
    name: process.env.NEXT_PUBLIC_APP_NAME || 'Notifier Admin',
    url: process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3004',
    version: '1.0.0',
  },
  api: {
    get baseUrl() {
      return getApiBaseUrl();
    },
    timeout: Number(process.env.NEXT_PUBLIC_API_TIMEOUT) || 30000,
  },
};
