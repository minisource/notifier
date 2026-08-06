'use client';

import { useAdminAuth } from './admin-auth-context';
import { AdminLoginPage } from './admin-login-page';
import { LoadingState } from '@minisource/ui';

interface AdminAuthGuardProps {
  children: React.ReactNode;
}

export function AdminAuthGuard({ children }: AdminAuthGuardProps) {
  const { state } = useAdminAuth();

  if (state.status === 'loading') {
    return <LoadingState   />;
  }

  if (state.status === 'unauthenticated') {
    return <AdminLoginPage />;
  }

  // Authenticated — render children
  return <>{children}</>;
}
