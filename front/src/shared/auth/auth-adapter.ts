import type { AuthAdapter, AuthSession, UserRole } from './auth-types';
import { createMockSession } from './mock-session';

function createEmptySession(): AuthSession {
  return {
    source: 'none',
    accessToken: null,
    userId: null,
    tenantId: null,
    projectId: null,
    roles: [],
    permissions: [],
    isAuthenticated: false,
  };
}

let cachedSession: AuthSession | null = null;

function getSession(): AuthSession {
  if (cachedSession) return cachedSession;

  if (!isAuthEnabled()) {
    cachedSession = createEmptySession();
  } else {
    cachedSession = createMockSession();
  }
  return cachedSession;
}

export function refreshMockSession(overrides?: Partial<AuthSession>): AuthSession {
  if (!isAuthEnabled()) {
    cachedSession = createEmptySession();
    return cachedSession;
  }
  cachedSession = createMockSession(overrides);
  return cachedSession;
}

export function clearMockSession(): void {
  cachedSession = null;
  try {
    localStorage.removeItem('notifier-mock-session');
  } catch {
    // Ignore
  }
}

export function updateMockInLocalStorage(updates: Partial<AuthSession>): void {
  if (!isAuthEnabled()) return;
  try {
    const current = getSession();
    const merged = { ...current, ...updates };
    localStorage.setItem('notifier-mock-session', JSON.stringify(merged));
    cachedSession = merged;
  } catch {
    // Ignore
  }
}

export function isAuthEnabled(): boolean {
  if (typeof process !== 'undefined' && process.env) {
    return process.env.NEXT_PUBLIC_NOTIFIER_MOCK_AUTH_ENABLED !== 'false';
  }
  return true;
}

const adminRoles: UserRole[] = ['admin', 'operator', 'super_admin'];

export const authAdapter: AuthAdapter = {
  getSession: () => getSession(),
  getAccessToken: () => getSession().accessToken,
  getUserId: () => getSession().userId,
  getTenantId: () => getSession().tenantId,
  getProjectId: () => getSession().projectId,
  getRoles: () => getSession().roles,
  isAuthenticated: () => getSession().isAuthenticated,
  hasRole: (role: UserRole) => getSession().roles.includes(role),
  hasAnyRole: (roles: UserRole[]) => roles.some(r => getSession().roles.includes(r)),
  isAdminLike: () => getSession().roles.some(r => adminRoles.includes(r)),
  hasPermission: (permission: string) => getSession().permissions.includes(permission),
};
