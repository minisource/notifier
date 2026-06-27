import { useQuery } from '@tanstack/react-query';
import { tenantsKeys } from '../query-keys';
import { listTenants } from '../api';

export function useTenants() {
  return useQuery({
    queryKey: tenantsKeys.list(),
    queryFn: listTenants,
  });
}
