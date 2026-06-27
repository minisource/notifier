import type { Provider } from './types';
import { adminProvidersApi } from '@/features/notifier/api/notifier-api-mode';
import type { ProviderTestInput } from '@/features/notifier/api/notifier-types';

/**
 * Provider API — wired through the centralized notifier API mode switch.
 * When NEXT_PUBLIC_NOTIFIER_USE_MOCKS=false, calls the real backend.
 * When NEXT_PUBLIC_NOTIFIER_USE_MOCKS=true, uses mock data.
 */

export async function listProviders(): Promise<Provider[]> {
  return adminProvidersApi.list() as Promise<Provider[]>;
}

export async function getProvider(id: string): Promise<Provider> {
  return adminProvidersApi.get(id) as Promise<Provider>;
}

export async function createProvider(input: { name: string; channel: string; type?: string; status?: string; priority?: number; isDefault?: boolean; description?: string; config?: Record<string, unknown>; secretConfig?: Record<string, unknown> }): Promise<Provider> {
  return adminProvidersApi.create(input) as Promise<Provider>;
}

export async function updateProvider(id: string, input: { name?: string; channel?: string; type?: string; status?: string; priority?: number; isEnabled?: boolean; isDefault?: boolean; description?: string; config?: Record<string, unknown>; secretConfig?: Record<string, unknown> }): Promise<Provider> {
  return adminProvidersApi.update(id, input) as Promise<Provider>;
}

export async function deleteProvider(id: string): Promise<void> {
  return adminProvidersApi.delete(id);
}

export async function toggleProviderStatus(id: string, isEnabled: boolean): Promise<Provider> {
  return adminProvidersApi.toggleStatus(id, isEnabled) as Promise<Provider>;
}

export async function setDefaultProvider(id: string, isDefault: boolean): Promise<Provider> {
  return adminProvidersApi.setDefault(id, isDefault) as Promise<Provider>;
}

export async function testProvider(id: string, input?: ProviderTestInput): Promise<{ success: boolean; message?: string }> {
  return adminProvidersApi.test(id, input);
}

export async function getProviderHealth() {
  return adminProvidersApi.getHealth();
}
