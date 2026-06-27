import { http } from '@/shared/api/http-client';

export interface NotificationSettings {
  defaultEmailProviderId?: string | null;
  defaultSmsProviderId?: string | null;
  defaultPushProviderId?: string | null;
  defaultWebhookProviderId?: string | null;
  enabledChannels: {
    email: boolean;
    sms: boolean;
    push: boolean;
    webhook: boolean;
    inApp: boolean;
  };
  retryPolicy: {
    enabled: boolean;
    maxAttempts: number;
    backoffStrategy: 'fixed' | 'linear' | 'exponential';
    initialDelaySeconds: number;
    maxDelaySeconds: number;
  };
  rateLimit: {
    enabled: boolean;
    perMinute: number;
    perHour: number;
  };
  quietHours?: {
    enabled: boolean;
    timezone: string;
    start: string;
    end: string;
  } | null;
  retentionDays: number;
}

export async function fetchNotificationSettings(): Promise<NotificationSettings> {
  return http.get<NotificationSettings>('/admin/settings/notifications');
}

export async function updateNotificationSettings(
  input: Partial<NotificationSettings>,
): Promise<NotificationSettings> {
  return http.patch<NotificationSettings>('/admin/settings/notifications', input);
}
