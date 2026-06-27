'use client';

import { useEffect, useState } from 'react';
import { authAdapter } from '@/shared/auth/auth-adapter';
import { ForbiddenState } from './forbidden-state';

interface RoleGuardProps {
  children: React.ReactNode;
  requiredRoles?: string[];
  fallback?: React.ReactNode;
}

export function RoleGuard({ children, requiredRoles = ['admin', 'operator', 'super_admin'], fallback }: RoleGuardProps) {
  const [hasAccess, setHasAccess] = useState(true);

  useEffect(() => {
    const session = authAdapter.getSession();
    const allowed = session.roles.some(r => requiredRoles.includes(r));
    setHasAccess(allowed);
  }, [requiredRoles]);

  if (!hasAccess) {
    return fallback || <ForbiddenState />;
  }

  return <>{children}</>;
}
