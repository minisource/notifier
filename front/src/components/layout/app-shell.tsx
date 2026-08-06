'use client';

import * as React from 'react';
import { useParams, usePathname } from 'next/navigation';
import Link from 'next/link';
import { SidebarProvider, SidebarInset, Topbar as DSTopbar, SidebarTrigger, Footer } from '@minisource/app-shell';
import {
  Breadcrumb,
  BreadcrumbList,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbPage,
  BreadcrumbSeparator,
  ModeToggle,
} from '@minisource/ui';
import { Sidebar } from '@/components/layout/sidebar';
import { LanguageSwitcher } from '@/components/layout/language-switcher';
import { UserMenu } from '@/components/layout/user-menu';
import { NotificationCenterWrapper } from '@/features/notifier/notification-center/notification-center-wrapper';
import { useAdminAuth } from '@/features/auth/admin-auth-context';
import { useTranslations } from 'next-intl';

export function AppShell({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const pathname = usePathname();
  const locale = (params?.locale as string) || 'en';
  const isRtl = locale === 'fa';
  const dir = isRtl ? 'rtl' : 'ltr';
  const { isAuthenticated, isLoading } = useAdminAuth();
  const t = useTranslations();

  // Simple breadcrumb generator (URLs are locale-free — localePrefix: 'never')
  const breadcrumbs = React.useMemo(() => {
    const segments = pathname.split('/').filter(Boolean);
    if (segments.length === 0 || segments[0] === 'dashboard') {
      return [{ label: t('navigation.dashboard'), href: '/dashboard' }];
    }
    const crumbs = [{ label: t('navigation.dashboard'), href: '/dashboard' }];
    let accPath = '';
    for (const seg of segments) {
      accPath += `/${seg}`;
      const labelKey = `navigation.${seg.replace('-', '_')}`;
      const label = t.has(labelKey) ? t(labelKey) : seg.charAt(0).toUpperCase() + seg.slice(1);
      crumbs.push({ label, href: accPath });
    }
    return crumbs;
  }, [pathname, t]);

  // Redirect to auth/front login if not authenticated
  React.useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      const returnUrl = encodeURIComponent(
        window.location.pathname + window.location.search
      );
      window.location.href = `/auth/login?returnUrl=${returnUrl}`;
    }
  }, [isLoading, isAuthenticated]);

  // Show loading skeleton while checking auth state
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  // Don't render shell until authenticated (redirect happens via useEffect)
  if (!isAuthenticated) {
    return null;
  }

  return (
    <SidebarProvider defaultOpen dir={dir}>
      <Sidebar />
      <SidebarInset
        dir={dir}
        topbar={
          <div className="flex flex-col w-full">
            <DSTopbar
              left={
                <div className="flex items-center gap-2">
                  <SidebarTrigger />
                  <div className="mx-1.5 h-4 w-px bg-border/60" />
                  <Breadcrumb className="hidden sm:block">
                    <BreadcrumbList>
                      {breadcrumbs.map((crumb, i) => (
                        <React.Fragment key={crumb.href || crumb.label}>
                          <BreadcrumbItem>
                            {i < breadcrumbs.length - 1 ? (
                              <BreadcrumbLink asChild>
                                <Link href={crumb.href}>{crumb.label}</Link>
                              </BreadcrumbLink>
                            ) : (
                              <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                            )}
                          </BreadcrumbItem>
                          {i < breadcrumbs.length - 1 && <BreadcrumbSeparator />}
                        </React.Fragment>
                      ))}
                    </BreadcrumbList>
                  </Breadcrumb>
                </div>
              }
              right={
                <div className="flex items-center gap-1.5">
                  <NotificationCenterWrapper />
                  <div className="mx-1 h-6 w-px bg-border/50" />
                  <LanguageSwitcher />
                  <ModeToggle />
                  <div className="mx-1 h-6 w-px bg-border/50" />
                  <UserMenu />
                </div>
              }
              dir={dir}
            />
          </div>
        }
        footer={
          <Footer dir={dir}>
            © {new Date().getFullYear()} MiniSource Notifier. All rights reserved.
          </Footer>
        }
      >
        <main className="flex-1 p-4 md:p-6 lg:p-8">
          <div className="mx-auto max-w-7xl">{children}</div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}

