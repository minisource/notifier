import type { Delivery, ListDeliveriesParams } from './types';
import { adminDeliveriesApi } from '@/features/notifier/api/notifier-api-mode';
import type { NotificationDelivery as NotifierDelivery, DeliveryAttempt as NotifierAttempt } from '@/features/notifier/api/notifier-types';

function mapDelivery(d: NotifierDelivery): Delivery {
  return {
    id: d.id,
    notificationId: d.notificationId,
    provider: d.provider,
    channel: d.channel,
    status: d.status as Delivery['status'],
    attemptCount: d.attemptCount,
    maxAttempts: d.maxAttempts,
    // Backend returns lastErrorMessage.
    lastError: d.lastErrorMessage ?? d.lastError,
    nextRetryAt: d.nextRetryAt,
    recipientEmail: d.recipientEmail,
    recipientPhone: d.recipientPhone,
    recipientId: d.recipientId,
    subject: d.subject,
    body: d.body,
    completedAt: d.completedAt,
    createdAt: d.createdAt,
    updatedAt: d.updatedAt,
    attempts: (d.attempts || []).map((a: NotifierAttempt) => ({
      id: a.id,
      deliveryId: a.deliveryId,
      attemptNumber: a.attemptNumber,
      status: a.status as Delivery['status'],
      errorMessage: a.errorMessage,
      errorCode: a.errorCode,
      // Backend returns latencyMs + providerResponseSanitized.
      providerResponse: a.providerResponseSanitized ?? a.providerResponse,
      // Backend omits latencyMs when 0, so fall back to 0 to keep the number type honest.
      processingTimeMs: a.latencyMs ?? a.processingTimeMs ?? 0,
      createdAt: a.createdAt,
      completedAt: a.completedAt,
    })),
  };
}

export async function listDeliveries(params?: ListDeliveriesParams): Promise<Delivery[]> {
  const result = await adminDeliveriesApi.list(params as Record<string, string | number | boolean | undefined>);
  // Backend returns paginated { items: [...], total, ... }
  const items = (result as any).items || (result as any).data || [];
  return items.map(mapDelivery);
}

export async function getDelivery(id: string): Promise<Delivery> {
  const result = await adminDeliveriesApi.get(id);
  return mapDelivery(result);
}

export async function retryDelivery(id: string): Promise<void> {
  await adminDeliveriesApi.retry(id);
}
