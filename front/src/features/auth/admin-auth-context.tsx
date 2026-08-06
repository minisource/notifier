'use client';

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react';
import type { AuthSession, UserRole } from '@/shared/auth/auth-types';
import { setLocalSession, clearSession, setRealSession } from '@/shared/auth/auth-adapter';
import { authApi } from '@/shared/auth/auth-api';
import { notifierRuntimeConfig } from '@/features/notifier/config/notifier-config';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type AdminAuthState =
  | { status: 'loading' }
  | { status: 'unauthenticated'; session: null }
  | { status: 'authenticated'; session: AuthSession };

export interface AdminAuthContextValue {
  state: AdminAuthState;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  session: AuthSession | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

const AdminAuthContext = createContext<AdminAuthContextValue | null>(null);

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

export function AdminAuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AdminAuthState>({ status: 'loading' });

  // Load session from sessionStorage on mount
  useEffect(() => {
    // If auth feature is disabled, boot with a local admin session (bypass login)
    if (!notifierRuntimeConfig.authEnabled) {
      const localSession = setLocalSession();
      setState({ status: 'authenticated', session: localSession });
      return;
    }

    // Check notifier's own session storage first
    const storedToken = sessionStorage.getItem('notifier-admin-token');
    const storedRefreshToken = sessionStorage.getItem('notifier-refresh-token');

    // Also check auth/front's token storage (cross-app auth)
    const authFrontAccessToken =
      localStorage.getItem('accessToken') || sessionStorage.getItem('accessToken');
    const authFrontRefreshToken =
      localStorage.getItem('refreshToken') || sessionStorage.getItem('refreshToken');

    const effectiveToken = storedToken || authFrontAccessToken;
    const effectiveRefreshToken = storedRefreshToken || authFrontRefreshToken;

    if (effectiveToken) {
      // Parse token for basic session info
      const payload = parseJwtPayload(effectiveToken);
      const roles = payload ? extractRolesFromPayload(payload) : [];
      const userId = (payload?.sub as string) || (payload?.userId as string) || '';

      const session: AuthSession = {
        source: 'real',
        accessToken: effectiveToken,
        refreshToken: effectiveRefreshToken,
        userId,
        email: (payload?.email as string) || null,
        name: (payload?.name as string) || (payload?.username as string) || null,
        tenantId: null,
        projectId: null,
        roles,
        permissions: [],
        isAuthenticated: true,
        isSuperAdmin: roles.includes('super_admin'),
      };

      setRealSession(session);
      setState({ status: 'authenticated', session });
      return;
    }

    setState({ status: 'unauthenticated', session: null });
  }, []);

  // Login with email/password via Auth API
  const login = useCallback(
    async (email: string, password: string) => {
      const resp = await authApi.login(email, password);

      // Fetch userinfo for permissions and the canonical display name
      let permissions: string[] = [];
      let isSuperAdmin = false;
      let userinfoName: string | null = null;
      try {
        const userinfo = await authApi.userinfo(resp.accessToken);
        permissions = userinfo.permissions || [];
        isSuperAdmin = userinfo.isSuperAdmin || false;
        userinfoName =
          userinfo.name ||
          [userinfo.givenName, userinfo.familyName].filter(Boolean).join(' ').trim() ||
          null;
      } catch {
        // Userinfo is optional, continue with basic session
      }

      const session: AuthSession = {
        source: 'real',
        accessToken: resp.accessToken,
        refreshToken: resp.refreshToken,
        userId: resp.user.id,
        email: resp.user.email,
        name:
          userinfoName ||
          [resp.user.firstName, resp.user.lastName].filter(Boolean).join(' ').trim() ||
          resp.user.username ||
          null,
        tenantId: null,
        projectId: null,
        roles: resp.user.roles as UserRole[],
        permissions,
        isAuthenticated: true,
        isSuperAdmin,
      };

      // Store session
      setRealSession(session);
      setState({ status: 'authenticated', session });
    },
    [],
  );

  // Logout
  const logout = useCallback(async () => {
    const token = sessionStorage.getItem('notifier-admin-token');
    if (token) {
      // Try to call auth logout
      try {
        await authApi.logout(token);
      } catch {
        // Logout failure is non-critical
      }
    }
    clearSession();

    // Without auth there is no login page — stay signed in with the local session.
    if (!notifierRuntimeConfig.authEnabled) {
      const localSession = setLocalSession();
      setState({ status: 'authenticated', session: localSession });
      return;
    }

    setState({ status: 'unauthenticated', session: null });
  }, []);

  const isLoading = state.status === 'loading';
  const isAuthenticated = state.status === 'authenticated';
  const isAdmin =
    isAuthenticated &&
    (state.session.isSuperAdmin === true ||
      state.session.roles.some((r) =>
        ['admin', 'super_admin', 'system_admin'].includes(r),
      ));
  const session = state.status === 'authenticated' ? state.session : null;

  return (
    <AdminAuthContext.Provider
      value={{
        state,
        isLoading,
        isAuthenticated,
        isAdmin,
        session,
        login,
        logout,
      }}
    >
      {children}
    </AdminAuthContext.Provider>
  );
}

// ---------------------------------------------------------------------------
// JWT parsing helpers (client-side, no verification)
// ---------------------------------------------------------------------------

function parseJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded) as Record<string, unknown>;
  } catch {
    return null;
  }
}

function extractRolesFromPayload(payload: Record<string, unknown>): UserRole[] {
  const raw = payload.roles ?? payload.role ?? [];
  if (typeof raw === 'string') return [raw as UserRole];
  if (Array.isArray(raw)) return raw.filter((r): r is UserRole => typeof r === 'string');
  return [];
}

// ---------------------------------------------------------------------------
// Hook
// ---------------------------------------------------------------------------

export function useAdminAuth(): AdminAuthContextValue {
  const ctx = useContext(AdminAuthContext);
  if (!ctx) {
    throw new Error('useAdminAuth must be used within <AdminAuthProvider>');
  }
  return ctx;
}
