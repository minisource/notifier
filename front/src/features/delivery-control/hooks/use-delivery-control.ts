import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ApiError } from '@/shared/api/api-error';
import { deliveryControlKeys } from '../query-keys';
import {
  getDeliveryControlStatus,
  pauseDeliveries,
  resumeDeliveries,
  listDeliveryControlHistory,
  listHeldDeliveries,
} from '../api';
import type { PauseDeliveriesInput, ResumeDeliveriesInput } from '@/features/notifier/api/notifier-types';

/**
 * Generate one idempotency key per INTENTIONAL user action. It is preserved
 * across transport retries for that single action (the mutation promise), so
 * double-clicks or browser retries replay the original result instead of
 * creating a second pause/resume transition.
 */
function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return `dc-${Date.now()}-${Math.random().toString(36).slice(2, 12)}`;
}

function isConflict(err: unknown): boolean {
  return err instanceof ApiError && err.isConflict();
}

function isRateLimited(err: unknown): boolean {
  return err instanceof ApiError && err.isRateLimited();
}

/**
 * On a 409 CONFLICT the authoritative state changed (another operator acted).
 * The UI must reload the fresh state and require a NEW explicit confirmation —
 * never auto-replay the stale high-risk action against the new version.
 */
function handleControlConflict(queryClient: ReturnType<typeof useQueryClient>): void {
  queryClient.invalidateQueries({ queryKey: deliveryControlKeys.all });
  toast.error('Delivery control state changed by another operator', {
    description: 'Latest state reloaded. Please review and confirm again.',
  });
}

function handleControlRateLimited(err: ApiError): void {
  toast.error('Too many control requests', {
    description: err.retryAfter
      ? `Please wait ${err.retryAfter}s before trying again.`
      : 'Please wait a moment before trying again.',
  });
}

export function useDeliveryControlStatus() {
  return useQuery({
    queryKey: deliveryControlKeys.status(),
    queryFn: getDeliveryControlStatus,
    // Bounded polling with backoff: pause while the tab is hidden and stop
    // retrying on persistent failure (no aggressive polling / no DoS).
    refetchInterval: 10000,
    refetchIntervalInBackground: false,
    retry: (failureCount, error) => {
      const e = error as ApiError;
      // Do not hammer after rate limiting or auth/session loss.
      if (isRateLimited(e) || e?.isUnauthorized?.()) return false;
      return failureCount < 2;
    },
  });
}

export function usePauseDeliveries() {
  const queryClient = useQueryClient();
  return useMutation({
    // One idempotency key per intentional action; expectedVersion comes from
    // the caller's last-known status so a stale submit is rejected with 409.
    mutationFn: (input: PauseDeliveriesInput) =>
      pauseDeliveries(input, newIdempotencyKey()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: deliveryControlKeys.all });
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      toast.success('Outbound delivery paused');
    },
    onError: (err: Error) => {
      if (isConflict(err)) {
        handleControlConflict(queryClient);
        return;
      }
      if (isRateLimited(err)) {
        handleControlRateLimited(err as ApiError);
        return;
      }
      toast.error('Failed to pause deliveries', { description: err.message });
    },
  });
}

export function useResumeDeliveries() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input?: ResumeDeliveriesInput) =>
      resumeDeliveries(input, newIdempotencyKey()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: deliveryControlKeys.all });
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      toast.success('Outbound delivery resumed');
    },
    onError: (err: Error) => {
      if (isConflict(err)) {
        handleControlConflict(queryClient);
        return;
      }
      if (isRateLimited(err)) {
        handleControlRateLimited(err as ApiError);
        return;
      }
      toast.error('Failed to resume deliveries', { description: err.message });
    },
  });
}

export function useDeliveryControlHistory(limit = 50) {
  return useQuery({
    queryKey: deliveryControlKeys.history(),
    queryFn: () => listDeliveryControlHistory(limit),
  });
}

export function useHeldDeliveries(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: deliveryControlKeys.held(page, pageSize),
    queryFn: () => listHeldDeliveries(page, pageSize),
  });
}
