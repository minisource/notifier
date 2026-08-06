import { adminDeliveryControlApi } from '@/features/notifier/api/notifier-api-mode';
import type {
  DeliveryControlStatus,
  DeliveryControlEvent,
  HeldDeliveriesResponse,
  PauseDeliveriesInput,
  ResumeDeliveriesInput,
} from '@/features/notifier/api/notifier-types';

export async function getDeliveryControlStatus(): Promise<DeliveryControlStatus> {
  return adminDeliveryControlApi.getStatus();
}

export async function pauseDeliveries(input: PauseDeliveriesInput, idempotencyKey?: string): Promise<DeliveryControlStatus> {
  return adminDeliveryControlApi.pause(input, idempotencyKey);
}

export async function resumeDeliveries(input?: ResumeDeliveriesInput, idempotencyKey?: string): Promise<DeliveryControlStatus> {
  return adminDeliveryControlApi.resume(input, idempotencyKey);
}

export async function listDeliveryControlHistory(limit = 50): Promise<DeliveryControlEvent[]> {
  return adminDeliveryControlApi.getHistory({ limit });
}

export async function listHeldDeliveries(page = 1, pageSize = 20): Promise<HeldDeliveriesResponse> {
  return adminDeliveryControlApi.getHeld({ page, pageSize });
}
