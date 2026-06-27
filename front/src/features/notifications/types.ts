import type { PaginatedResponse } from '@/lib/api/types';

export type NotificationChannel = 'sms' | 'email' | 'push' | 'in_app' | 'webhook';

export type NotificationStatus = 'pending' | 'queued' | 'processing' | 'sent' | 'failed' | 'dead' | 'cancelled';

export type NotificationPriority = 'low' | 'normal' | 'high' | 'urgent';

export type DeliveryStatus = 'pending' | 'processing' | 'sent' | 'delivered' | 'failed' | 'retrying' | 'dead' | 'read' | 'seen' | 'clicked';

export type RecipientType = 'email' | 'phone' | 'user_id' | 'device_token' | 'webhook_url';

export interface NotificationRecipientDisplay {
  type: RecipientType;
  value: string;
  masked?: string;
}

export interface RecipientPayload {
  phone?: string;
  email?: string;
  userId?: string;
  deviceToken?: string;
  webhookUrl?: string;
}

export interface DeliveryAttempt {
  id: string;
  deliveryId: string;
  attemptNumber: number;
  status: DeliveryStatus;
  errorMessage?: string;
  errorCode?: string;
  providerResponse?: string;
  processingTimeMs: number;
  createdAt: string;
  completedAt?: string;
}

export interface NotificationDelivery {
  id: string;
  notificationId: string;
  provider: string;
  channel: NotificationChannel;
  status: DeliveryStatus;
  attemptCount: number;
  maxAttempts: number;
  lastError?: string;
  nextRetryAt?: string;
  createdAt: string;
  updatedAt: string;
  attempts: DeliveryAttempt[];
}

export interface Notification {
  id: string;
  tenantId?: string;
  projectId?: string;
  userId: string;
  type: NotificationChannel;
  status: NotificationStatus;
  priority: NotificationPriority;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  recipientType?: RecipientType;
  subject?: string;
  body: string;
  metadata?: Record<string, unknown>;
  templateId?: string;
  templateKey?: string;
  locale: string;
  variables?: Record<string, string>;
  scheduledAt?: string;
  sentAt?: string;
  deliveredAt?: string;
  seenAt?: string;
  readAt?: string;
  clickedAt?: string;
  failedAt?: string;
  retryCount: number;
  maxRetries: number;
  errorMessage?: string;
  errorCode?: string;
  provider?: string;
  providerMsgId?: string;
  idempotencyKey?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SendNotificationInput {
  tenantId?: string;
  projectId?: string;
  userId?: string;
  channel?: NotificationChannel;
  type?: NotificationChannel;
  priority?: NotificationPriority;
  recipient?: RecipientPayload;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  recipientType?: RecipientType;
  subject?: string;
  body: string;
  metadata?: Record<string, unknown>;
  templateId?: string;
  templateKey?: string;
  locale?: string;
  variables?: Record<string, string>;
  scheduledAt?: string;
  idempotencyKey?: string;
  providerId?: string;
}

export interface SendBatchNotificationInput {
  notifications: SendNotificationInput[];
}

export interface ListNotificationsParams {
  page?: number;
  pageSize?: number;
  type?: NotificationChannel;
  status?: NotificationStatus;
  priority?: NotificationPriority;
  userId?: string;
  search?: string;
  startDate?: string;
  endDate?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export type NotificationListResponse = PaginatedResponse<Notification>;
