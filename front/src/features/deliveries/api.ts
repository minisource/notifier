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
    lastError: d.lastError,
    nextRetryAt: d.nextRetryAt,
    createdAt: d.createdAt,
    updatedAt: d.updatedAt,
    attempts: (d.attempts || []).map((a: NotifierAttempt) => ({
      id: a.id,
      deliveryId: a.deliveryId,
      attemptNumber: a.attemptNumber,
      status: a.status as Delivery['status'],
      errorMessage: a.errorMessage,
      providerResponse: a.providerResponse,
      processingTimeMs: a.processingTimeMs,
      createdAt: a.createdAt,
    })),
  };
}

export async function listDeliveries(params?: ListDeliveriesParams): Promise<Delivery[]> {
  const result = await adminDeliveriesApi.list(params as Record<string, string | number | boolean | undefined>);
  // Backend returns paginated { items: [...], total, ... }; mock returns { data: [...], total, ... }
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
