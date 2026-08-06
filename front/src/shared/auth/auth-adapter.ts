import type { AuthAdapter, AuthSession, UserRole } from './auth-types';
import { notifierRuntimeConfig } from '@/features/notifier/config/notifier-config';

let cachedSession: AuthSession | null = null;

/**
 * Create a local session used when authentication is disabled
 * (NEXT_PUBLIC_NOTIFIER_AUTH_ENABLED=false). The Notifier runs standalone:
 * a local admin identity is used for the UI, and the backend serves routes
 * publicly (AUTH_ENABLED=false on the backend). No Auth service is needed.
 */
export function createLocalSession(): AuthSession {
  return {
    source: 'local',
    accessToken: null,
    refreshToken: null,
    userId: process.env.NEXT_PUBLIC_NOTIFIER_LOCAL_USER_ID?.trim() || 'local-admin',
    email: process.env.NEXT_PUBLIC_NOTIFIER_LOCAL_EMAIL?.trim() || 'admin@local',
    name: process.env.NEXT_PUBLIC_NOTIFIER_LOCAL_NAME?.trim() || 'Local Admin',
    tenantId: null,
    projectId: null,
    roles: ['admin', 'super_admin'] as UserRole[],
    permissions: [
      'notifier:admin',
      'notifier:notifications:read',
      'notifier:notifications:manage',
      'notifier:providers:manage',
      'notifier:templates:manage',
      'notifier:reminders:manage',
      'notifier:deliveries:read',
      'notifier:deliveries:manage',
      'notifier:settings:manage',
      'notifications:read', 'notifications:write',
      'templates:read', 'templates:write',
      'deliveries:read', 'providers:read',
      'reminders:read', 'reminders:write',
      'preferences:read', 'preferences:write',
      'observability:read', 'dashboard:read',
    ],
    isAuthenticated: true,
    isSuperAdmin: true,
  };
}

function emptySession(): AuthSession {
  return {
    source: 'none',
    accessToken: null,
    refreshToken: null,
    userId: null,
    email: null,
    name: null,
    tenantId: null,
    projectId: null,
    roles: [],
    permissions: [],
    isAuthenticated: false,
    isSuperAdmin: false,
  };
}

function getSession(): AuthSession {
  if (cachedSession) return cachedSession;
  return emptySession();
}

export function setLocalSession(): AuthSession {
  cachedSession = createLocalSession();
  return cachedSession;
}

export function setRealSession(session: AuthSession): void {
  cachedSession = { ...session, source: 'real' as const };
  try {
    if (session.accessToken) {
      sessionStorage.setItem('notifier-admin-token', session.accessToken);
    }
    if (session.refreshToken) {
      sessionStorage.setItem('notifier-refresh-token', session.refreshToken);
    }
  } catch {
    // Ignore
  }
}

export function clearSession(): void {
  cachedSession = null;
  try {
    sessionStorage.removeItem('notifier-admin-token');
    sessionStorage.removeItem('notifier-refresh-token');
  } catch {
    // Ignore
  }
}

export function isAuthEnabled(): boolean {
  return notifierRuntimeConfig.authEnabled;
}

const adminRoles: UserRole[] = ['admin', 'operator', 'super_admin'];

export const authAdapter: AuthAdapter = {
  getSession: () => getSession(),
  getAccessToken: () => getSession().accessToken,
  getUserId: () => getSession().userId,
  getEmail: () => getSession().email,
  getName: () => getSession().name,
  getTenantId: () => getSession().tenantId,
  getProjectId: () => getSession().projectId,
  getRoles: () => getSession().roles,
  getPermissions: () => getSession().permissions,
  isAuthenticated: () => getSession().isAuthenticated,
  isSuperAdmin: () => getSession().isSuperAdmin ?? false,
  hasRole: (role: UserRole) => getSession().roles.includes(role),
  hasAnyRole: (roles: UserRole[]) => roles.some(r => getSession().roles.includes(r)),
  isAdminLike: () => getSession().roles.some(r => adminRoles.includes(r)) || (getSession().isSuperAdmin ?? false),
  hasPermission: (permission: string) => getSession().permissions.includes(permission),
};
