export const config = {
  app: {
    name: process.env.NEXT_PUBLIC_APP_NAME || 'Notifier Admin',
    url: process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000',
    version: '1.0.0',
  },
  api: {
    baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://127.0.0.1:9002/v1',
    timeout: Number(process.env.NEXT_PUBLIC_API_TIMEOUT) || 30000,
  },
};
