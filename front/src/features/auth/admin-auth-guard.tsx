'use client';

import { useAdminAuth } from './admin-auth-context';
import { AdminLoginPage } from './admin-login-page';
import { LoadingState } from '@/components/shared/loading-state';

interface AdminAuthGuardProps {
  children: React.ReactNode;
}

export function AdminAuthGuard({ children }: AdminAuthGuardProps) {
  const { state } = useAdminAuth();

  if (state.status === 'loading') {
    return <LoadingState rows={4} columns={2} />;
  }

  if (state.status === 'unauthenticated') {
    return <AdminLoginPage />;
  }

  // Authenticated — render children
  return <>{children}</>;
}
