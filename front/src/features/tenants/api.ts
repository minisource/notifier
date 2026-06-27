import type { Tenant } from './types';
import { adminTenantsApi } from '@/features/notifier/api/notifier-api-mode';

export async function listTenants(): Promise<Tenant[]> {
  return adminTenantsApi.list() as Promise<Tenant[]>;
}
