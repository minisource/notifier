import { useQuery } from '@tanstack/react-query';
import { dashboardKeys } from '../query-keys';
import { getDashboardData } from '../api';

export function useDashboard() {
  return useQuery({
    queryKey: dashboardKeys.metrics(),
    queryFn: getDashboardData,
    refetchInterval: 30000,
  });
}
