import type { AuthSession, UserRole } from './auth-types';

function getEnvVar(key: string, defaultValue: string): string {
  if (typeof process !== 'undefined' && process.env && process.env[key]) {
    return process.env[key] as string;
  }
  return defaultValue;
}

function parseRoles(raw: string): UserRole[] {
  return raw
    .split(',')
    .map(r => r.trim().toLowerCase())
    .filter(r => ['user', 'admin', 'operator', 'service', 'super_admin'].includes(r)) as UserRole[];
}

function loadFromLocalStorage(): Partial<AuthSession> | null {
  try {
    const stored = localStorage.getItem('notifier-mock-session');
    if (stored) {
      return JSON.parse(stored) as Partial<AuthSession>;
    }
  } catch {
    // Ignore parse errors
  }
  return null;
}

export function createMockSession(overrides?: Partial<AuthSession>): AuthSession {
  const localStorageOverrides = loadFromLocalStorage();

  const defaults: AuthSession = {
    source: 'mock',
    accessToken: getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_ACCESS_TOKEN', 'mock-token-notifier-admin'),
    userId: getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_USER_ID', 'user-mock-001'),
    tenantId: getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_TENANT_ID', '').trim() || null,
    projectId: getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_PROJECT_ID', '').trim() || null,
    roles: parseRoles(getEnvVar('NEXT_PUBLIC_NOTIFIER_MOCK_ROLES', 'admin,operator')),
    permissions: ['notifications:read', 'notifications:write', 'templates:read', 'templates:write',
      'deliveries:read', 'providers:read', 'reminders:read', 'reminders:write',
      'preferences:read', 'preferences:write', 'observability:read', 'dashboard:read'],
    isAuthenticated: true,
  };

  return {
    ...defaults,
    ...localStorageOverrides,
    ...overrides,
  };
}

export const defaultMockSession: AuthSession = createMockSession();
