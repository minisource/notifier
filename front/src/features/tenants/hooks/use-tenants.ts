import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { tenantsKeys } from '../query-keys';
import { listTenants } from '../api';
import {
  useTenantStore,
  ALL_TENANTS,
} from '@/stores/tenant.store';
import type { Tenant } from '../types';

/**
 * Fetches the current user's tenants from the Auth service and syncs them into
 * the tenant store. Defaults the active tenant to "All Tenants (Global)" when
 * nothing is selected.
 *
 * Tenants are read-only in the Notifier: they are managed in the Auth service.
 */
export function useTenants() {
  const queryClient = useQueryClient();
  const { activeTenant, setActiveTenant, setAvailableTenants } = useTenantStore();

  const query = useQuery({
    queryKey: tenantsKeys.list(),
    queryFn: listTenants,
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
  });

  // Sync available tenants to store & default to ALL_TENANTS
  useEffect(() => {
    if (query.data) {
      setAvailableTenants(query.data);
      if (!activeTenant) {
        setActiveTenant(ALL_TENANTS);
      }
    }
  }, [query.data, activeTenant, setActiveTenant, setAvailableTenants]);

  const switchTenant = (tenant: Tenant) => {
    setActiveTenant(tenant);
    // Reload page to re-fetch all tenant-scoped API queries seamlessly
    if (typeof window !== 'undefined') {
      window.location.reload();
    }
  };

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: tenantsKeys.all });

  return {
    data: query.data,
    tenants: query.data ?? [],
    activeTenant: activeTenant ?? ALL_TENANTS,
    switchTenant,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
    invalidate,
  };
}
