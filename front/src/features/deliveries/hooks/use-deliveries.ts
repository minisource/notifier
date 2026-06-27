import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { deliveriesKeys } from '../query-keys';
import { listDeliveries, getDelivery, retryDelivery } from '../api';
import type { ListDeliveriesParams } from '../types';

export function useDeliveries(params?: ListDeliveriesParams) {
  return useQuery({
    queryKey: deliveriesKeys.list(params as Record<string, unknown>),
    queryFn: () => listDeliveries(params),
  });
}

export function useDelivery(id: string) {
  return useQuery({
    queryKey: deliveriesKeys.detail(id),
    queryFn: () => getDelivery(id),
    enabled: !!id,
  });
}

export function useRetryDelivery() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => retryDelivery(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: deliveriesKeys.lists() });
    },
  });
}
