export type ProviderStatus = 'active' | 'inactive' | 'disabled' | 'error';

export interface Provider {
  id: string;
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
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateProviderInput {
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
  name: string;
  channel: string;
  status: string;
  successRate?: number;
}

export interface ProviderHealthResponse {
  providers: ProviderHealthItem[];
  healthyCount: number;
  degradedCount: number;
  downCount: number;
  disabledCount: number;
  checkedAt: string;
}
