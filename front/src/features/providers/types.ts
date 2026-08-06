export type ProviderStatus = 'active' | 'inactive' | 'disabled' | 'error';

export interface Provider {
  id: string;
  tenantId?: string;
  name: string;
  channel: string;
  type?: string;
  status: ProviderStatus;
  description?: string;
  isEnabled: boolean;
  isPrimary: boolean;
  isDefault: boolean;
  priority: number;
  config?: Record<string, unknown>;
  successRate?: number;
  averageLatencyMs?: number;
  lastSuccessAt?: string;
  lastFailureAt?: string;
  lastError?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateProviderInput {
  tenantId?: string;
  name: string;
  channel: string;
  type?: string;
  status?: ProviderStatus;
  priority?: number;
  isDefault?: boolean;
  description?: string;
  config?: Record<string, unknown>;
  secretConfig?: Record<string, unknown>;
}

export interface UpdateProviderInput {
  tenantId?: string;
  name?: string;
  channel?: string;
  type?: string;
  status?: ProviderStatus;
  priority?: number;
  isEnabled?: boolean;
  isDefault?: boolean;
  description?: string;
  config?: Record<string, unknown>;
  secretConfig?: Record<string, unknown>;
}

export interface ProviderHealthItem {
  providerId?: string;
  name: string;
  channel: string;
  type?: string;
  status: string;
  successRate?: number;
  latencyMs?: number;
  message?: string;
  error?: string;
  checkedAt?: string;
}

export interface ProviderHealthResponse {
  providers: ProviderHealthItem[];
  healthyCount: number;
  degradedCount: number;
  downCount: number;
  disabledCount: number;
  checkedAt: string;
}
