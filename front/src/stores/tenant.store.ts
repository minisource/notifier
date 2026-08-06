import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { Tenant } from '@/features/tenants/types';

export const ALL_TENANTS: Tenant = {
  id: 'all',
  name: 'All Tenants (Global)',
  slug: 'all',
  status: 'active',
  isActive: true,
  enabledChannels: [],
  createdAt: '',
};

interface TenantState {
  activeTenant: Tenant | null;
  availableTenants: Tenant[];
  _hasHydrated: boolean;

  setActiveTenant: (tenant: Tenant) => void;
  setAvailableTenants: (tenants: Tenant[]) => void;
  clearTenant: () => void;
  setHasHydrated: (v: boolean) => void;
}

/**
 * Tenant store — tracks which tenant/project the user is currently viewing.
 * Persisted to localStorage so tenant selection survives refresh, and mirrored
 * into the `activeTenant` cookie for server-side consistency.
 */
export const useTenantStore = create<TenantState>()(
  persist(
    (set) => ({
      activeTenant: null,
      availableTenants: [],
      _hasHydrated: false,

      setHasHydrated: (v: boolean) => set({ _hasHydrated: v }),

      setActiveTenant: (tenant) => {
        if (typeof window !== 'undefined') {
          document.cookie = `activeTenant=${encodeURIComponent(tenant.id)}; path=/; samesite=lax; max-age=${60 * 60 * 24 * 30}`;
        }
        set({ activeTenant: tenant });
      },

      setAvailableTenants: (tenants) => {
        set({ availableTenants: tenants });
      },

      clearTenant: () => {
        if (typeof window !== 'undefined') {
          document.cookie = 'activeTenant=; path=/; max-age=0; samesite=lax';
        }
        set({ activeTenant: null, availableTenants: [] });
      },
    }),
    {
      name: 'notifier-tenant-storage',
      storage: createJSONStorage(() => {
        if (typeof window === 'undefined') return null as unknown as Storage;
        return window.localStorage;
      }),
      partialize: (state) => ({
        activeTenant: state.activeTenant,
        availableTenants: state.availableTenants,
      }),
      onRehydrateStorage: () => {
        return () => {
          useTenantStore.setState({ _hasHydrated: true });
        };
      },
    },
  ),
);
