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
import { refreshMockSession, clearMockSession } from '@/shared/auth/auth-adapter';
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
  login: (token: string) => Promise<void>;
  logout: () => void;
  devLogin: () => void;
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

const AdminAuthContext = createContext<AdminAuthContextValue | null>(null);

// ---------------------------------------------------------------------------
// Helper: parse JWT and extract payload (client-side only, no verification)
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
// Provider
// ---------------------------------------------------------------------------

export function AdminAuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AdminAuthState>({ status: 'loading' });

  // Build an AuthSession from a real JWT token
  const buildSessionFromToken = useCallback((token: string): AuthSession => {
    const payload = parseJwtPayload(token);
    const roles = payload ? extractRolesFromPayload(payload) : [];
    const userId = (payload?.sub as string) || (payload?.userId as string) || 'unknown';

    return {
      source: 'real',
      accessToken: token,
      userId,
      tenantId: (payload?.tenantId as string) || null,
      projectId: null,
      roles,
      permissions: [],
      isAuthenticated: true,
    };
  }, []);

  // Load session from localStorage on mount
  useEffect(() => {
    const storedToken = sessionStorage.getItem('notifier-admin-token');

    if (storedToken) {
      const session = buildSessionFromToken(storedToken);
      setState({ status: 'authenticated', session });
      return;
    }

    // In dev mode with mock auth enabled, allow mock session
    if (notifierRuntimeConfig.mockAuthEnabled) {
      const mockSession = refreshMockSession();
      setState({ status: 'authenticated', session: mockSession });
      return;
    }

    setState({ status: 'unauthenticated', session: null });
  }, [buildSessionFromToken]);

  // Login: store token and build session
  const login = useCallback(
    async (token: string) => {
      const payload = parseJwtPayload(token);
      if (!payload) {
        throw new Error('Invalid token format. Expected a valid JWT with 3 dot-separated parts.');
      }
      const session = buildSessionFromToken(token);
      sessionStorage.setItem('notifier-admin-token', token);
      // Also update the auth adapter for the HTTP client
      refreshMockSession({ ...session, source: 'real' as const });
      setState({ status: 'authenticated', session });
    },
    [buildSessionFromToken],
  );

  // Logout: clear everything
  const logout = useCallback(() => {
    sessionStorage.removeItem('notifier-admin-token');
    clearMockSession();
    setState({ status: 'unauthenticated', session: null });
  }, []);

  // Dev login: create mock admin session
  const devLogin = useCallback(() => {
    const mockSession = refreshMockSession();
    sessionStorage.removeItem('notifier-admin-token');
    setState({ status: 'authenticated', session: mockSession });
  }, []);

  const isLoading = state.status === 'loading';
  const isAuthenticated = state.status === 'authenticated';
  const isAdmin =
    isAuthenticated &&
    state.session.roles.some((r) =>
      ['admin', 'super_admin'].includes(r),
    );
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
        devLogin,
      }}
    >
      {children}
    </AdminAuthContext.Provider>
  );
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
