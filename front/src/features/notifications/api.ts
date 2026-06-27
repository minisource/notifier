import type { PaginatedResponse } from '@/lib/api/types';
import type {
  Notification,
  SendNotificationInput,
  SendBatchNotificationInput,
  ListNotificationsParams,
  NotificationDelivery,
  DeliveryAttempt,
} from './types';
import { adminNotificationsApi } from '@/features/notifier/api/notifier-api-mode';
import type {
  Notification as NotifierNotification,
  NotificationDelivery as NotifierDelivery,
  DeliveryAttempt as NotifierAttempt,
  CreateNotificationInput,
  RecipientInput,
  Template as NotifierTemplate,
} from '@/features/notifier/api/notifier-types';

function mapNotification(n: NotifierNotification): Notification {
  return {
    id: n.id,
    userId: n.userId,
    type: n.type as Notification['type'],
    status: n.status as Notification['status'],
    priority: n.priority as Notification['priority'],
    recipientEmail: n.recipientEmail,
    recipientPhone: n.recipientPhone,
    recipientId: n.recipientId,
    subject: n.subject,
    body: n.body,
    metadata: n.metadata,
    templateId: n.templateId,
    templateKey: n.templateKey,
    locale: n.locale || 'en',
    variables: n.variables,
    scheduledAt: n.scheduledAt,
    sentAt: n.sentAt,
    deliveredAt: n.deliveredAt,
    seenAt: n.seenAt,
    readAt: n.readAt,
    clickedAt: n.clickedAt,
    retryCount: n.retryCount,
    maxRetries: n.maxRetries,
    errorMessage: n.errorMessage,
    provider: n.provider,
    providerMsgId: n.providerMsgId,
    createdAt: n.createdAt,
    updatedAt: n.updatedAt,
  };
}

function mapDelivery(d: NotifierDelivery): NotificationDelivery {
  return {
    id: d.id,
    notificationId: d.notificationId,
    provider: d.provider,
    channel: d.channel as Notification['type'],
    status: d.status as NotificationDelivery['status'],
    attemptCount: d.attemptCount,
    maxAttempts: d.maxAttempts,
    lastError: d.lastError,
    nextRetryAt: d.nextRetryAt,
    createdAt: d.createdAt,
    updatedAt: d.updatedAt,
    attempts: d.attempts.map((a: NotifierAttempt): DeliveryAttempt => ({
      id: a.id,
      deliveryId: a.deliveryId,
      attemptNumber: a.attemptNumber,
      status: a.status as DeliveryAttempt['status'],
      errorMessage: a.errorMessage,
      errorCode: a.errorCode,
      providerResponse: a.providerResponse,
      processingTimeMs: a.processingTimeMs,
      createdAt: a.createdAt,
      completedAt: a.completedAt,
    })),
  };
}

export async function listNotifications(params?: ListNotificationsParams): Promise<PaginatedResponse<Notification>> {
  const result = await adminNotificationsApi.list(params as Record<string, string | number | boolean | undefined>);
  return {
    data: (result.data || []).map(mapNotification),
    total: result.total || 0,
    page: result.page || 1,
    pageSize: result.pageSize || 20,
    totalPages: result.totalPages || 0,
  };
}

export async function getNotification(id: string): Promise<Notification> {
  const result = await adminNotificationsApi.get(id);
  return mapNotification(result);
}

export async function sendNotification(input: SendNotificationInput): Promise<Notification> {
  const notifierInput: CreateNotificationInput = {
    userId: input.userId,
    channel: input.channel as NotifierNotification['type'],
    type: input.type as NotifierNotification['type'],
    priority: input.priority as NotifierNotification['priority'],
    recipient: input.recipient as RecipientInput,
    recipientEmail: input.recipientEmail,
    recipientPhone: input.recipientPhone,
    recipientId: input.recipientId,
    subject: input.subject,
    body: input.body,
    locale: input.locale || 'en',
    scheduledAt: input.scheduledAt,
    templateId: input.templateId,
    templateKey: input.templateKey,
    metadata: input.metadata as Record<string, unknown>,
    variables: input.variables,
    idempotencyKey: input.idempotencyKey,
    providerId: input.providerId,
  };
  const result = await adminNotificationsApi.create(notifierInput);
  return mapNotification(result);
}

export async function sendBatchNotifications(input: SendBatchNotificationInput): Promise<Notification[]> {
  const results = await Promise.all(
    input.notifications.map(n => sendNotification(n))
  );
  return results;
}

export async function retryNotification(id: string): Promise<Notification> {
  const result = await adminNotificationsApi.retry(id);
  return mapNotification(result);
}

export async function cancelNotification(id: string): Promise<void> {
  await adminNotificationsApi.cancel(id);
}

export async function markNotificationRead(id: string): Promise<void> {
  await adminNotificationsApi.markRead(id);
}

export async function markAllNotificationsRead(userId: string): Promise<void> {
  await adminNotificationsApi.readAll(userId);
}

export async function getNotificationDeliveries(notificationId: string): Promise<NotificationDelivery[]> {
  const deliveries = await adminNotificationsApi.getDeliveries(notificationId);
  return (deliveries || []).map(mapDelivery);
}

export async function getTemplatesForSelect(): Promise<Array<{ id: string; key?: string; name: string; type: string; locale: string }>> {
  const { adminTemplatesApi } = await import('@/features/notifier/api/notifier-api-mode');
  const result = await adminTemplatesApi.list();
  // Backend returns paginated { items: [...], total, ... }; mock returns array
  const templates = Array.isArray(result) ? result : (result as any).items || [];
  return templates.map((t: NotifierTemplate) => ({
    id: t.id,
    key: t.key,
    name: t.name,
    type: t.type,
    locale: t.locale,
  }));
}
