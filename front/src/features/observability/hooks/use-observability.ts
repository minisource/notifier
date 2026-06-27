import { useQuery } from '@tanstack/react-query';
import { observabilityKeys } from '../query-keys';
import { getHealth, getMetrics } from '../api';

export function useHealth() {
  return useQuery({
    queryKey: observabilityKeys.health(),
    queryFn: getHealth,
    refetchInterval: 30000,
  });
}

export function useMetrics() {
  return useQuery({
    queryKey: observabilityKeys.metrics(),
    queryFn: getMetrics,
    refetchInterval: 60000,
  });
}
