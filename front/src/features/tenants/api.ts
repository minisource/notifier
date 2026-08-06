import { http } from '@/shared/api/http-client';
import type { Tenant } from './types';

/**
 * Tenants are owned by the Auth service. The Notifier lists the tenants the
 * current user belongs to via Auth's GET /users/me/tenants (routed through
 * the gateway, same base URL as every other API call). Notifier never
 * creates/updates/deletes tenants itself.
 */

interface AuthTenant {
  id: string;
  name: string;
  slug?: string;
  displayName?: string;
  logo?: string;
  status?: string;
  plan?: string;
  isDefault?: boolean;
  role?: string;
}

type AuthTenantListResult = AuthTenant[] | { data?: AuthTenant[] };

function toArray(res: AuthTenantListResult): AuthTenant[] {
  if (Array.isArray(res)) {
    return res;
  }
  const payload = res as { data?: AuthTenant[] };
  return payload?.data ?? [];
}

function toTenant(raw: AuthTenant): Tenant {
  const status = raw.status || 'active';
  return {
    id: raw.id,
    name: raw.name,
    slug: raw.slug ?? raw.id,
    displayName: raw.displayName,
    logo: raw.logo,
    status,
    plan: raw.plan,
    isDefault: raw.isDefault,
    role: raw.role,
    isActive: status !== 'inactive' && status !== 'suspended' && status !== 'disabled',
    enabledChannels: [],
    createdAt: '',
  };
}

export async function listTenants(): Promise<Tenant[]> {
  const res = await http.get<AuthTenantListResult>('/users/me/tenants');
  return toArray(res).map(toTenant);
}
