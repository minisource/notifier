'use client';

import { useParams } from 'next/navigation';
import { Sidebar } from '@/components/layout/sidebar';
import { Topbar } from '@/components/layout/topbar';
import { MobileNav } from '@/components/layout/mobile-nav';
import { MockModeBanner } from '@/shared/components/mock-mode-banner';
import { AdminAuthGuard } from '@/features/auth/admin-auth-guard';
import { useAdminAuth } from '@/features/auth/admin-auth-context';
import { useState } from 'react';

export function AppShell({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';
  const { isAuthenticated } = useAdminAuth();

  if (!isAuthenticated) {
    return <AdminAuthGuard>{children}</AdminAuthGuard>;
  }

  return (
    <div className="flex min-h-screen bg-muted/30 dark:bg-muted/10" dir={isRtl ? 'rtl' : 'ltr'}>
      {/* Desktop sidebar */}
      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />

      {/* Mobile overlay */}
      <MobileNav open={sidebarOpen} onClose={() => setSidebarOpen(false)} />

      <div className="flex flex-1 flex-col">
        <MockModeBanner />
        <Topbar onMenuClick={() => setSidebarOpen(true)} />
        <main className="flex-1 p-4 md:p-6 lg:p-8">
          <div className="mx-auto max-w-7xl">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
