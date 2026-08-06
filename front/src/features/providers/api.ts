import type { Provider } from './types';
import { adminProvidersApi, adminProviderBalanceApi } from '@/features/notifier/api/notifier-api-mode';
import type {
  ProviderTestInput,
  ProviderAccountHealthSummary,
  ProviderBalanceDetail,
  ProviderBalanceSettings,
  BalanceRefreshResult,
  CreditAlert,
} from '@/features/notifier/api/notifier-types';

/**
 * Provider API — wired through the centralized notifier API mode switch.
 * Calls the real backend (mock API data has been removed).
 */

export async function listProviders(): Promise<Provider[]> {
  return adminProvidersApi.list() as Promise<Provider[]>;
}

export async function getProvider(id: string): Promise<Provider> {
  return adminProvidersApi.get(id) as Promise<Provider>;
}

export async function createProvider(input: { tenantId?: string; name: string; channel: string; type?: string; status?: string; priority?: number; isDefault?: boolean; description?: string; config?: Record<string, unknown>; secretConfig?: Record<string, unknown> }): Promise<Provider> {
  return adminProvidersApi.create(input) as Promise<Provider>;
}

export async function updateProvider(id: string, input: { tenantId?: string; name?: string; channel?: string; type?: string; status?: string; priority?: number; isEnabled?: boolean; isDefault?: boolean; description?: string; config?: Record<string, unknown>; secretConfig?: Record<string, unknown> }): Promise<Provider> {
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

export async function testProvider(id: string, input?: ProviderTestInput) {
  return adminProvidersApi.test(id, input);
}

export async function getProviderHealth() {
  return adminProvidersApi.getHealth();
}

export async function healthCheckAll() {
  return adminProvidersApi.healthCheckAll();
}

// ---- Provider account balance / quota / credit alerting ----

export async function listProviderBalanceHealth(): Promise<ProviderAccountHealthSummary[]> {
  return adminProviderBalanceApi.listHealth();
}

export async function getProviderBalanceDetail(id: string): Promise<ProviderBalanceDetail> {
  return adminProviderBalanceApi.getDetail(id);
}

export async function refreshProviderBalance(id: string): Promise<BalanceRefreshResult> {
  return adminProviderBalanceApi.refresh(id);
}

export async function updateProviderBalanceSettings(id: string, settings: ProviderBalanceSettings): Promise<ProviderBalanceSettings> {
  return adminProviderBalanceApi.updateSettings(id, settings);
}

export async function listCreditAlerts(params?: { status?: string; providerId?: string }): Promise<CreditAlert[]> {
  return adminProviderBalanceApi.listAlerts(params);
}

export async function acknowledgeCreditAlert(alertId: string): Promise<CreditAlert> {
  return adminProviderBalanceApi.acknowledgeAlert(alertId);
}
