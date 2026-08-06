'use client';

import * as React from 'react';
import Link from 'next/link';
import { usePathname, useParams, useSearchParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Sidebar as DSSidebar, type NavItem } from '@minisource/app-shell';
import {
  LayoutDashboard,
  Bell,
  FileText,
  AlarmClock,
  Truck,
  Server,
  Settings2,
  Building2,
  Activity,
  Cog,
  BellRing,
  Send,
  Clock,
  Wrench,
  Play,
  Shield,
  Sliders,
  ScrollText,
} from 'lucide-react';
import { authAdapter } from '@/shared/auth/auth-adapter';
import { TenantSelector } from '@/components/layout/tenant-selector';

interface SidebarProps {
  open?: boolean;
  onClose?: () => void;
  className?: string;
}

export function Sidebar({ className }: SidebarProps) {
  const pathname = usePathname();
  const params = useParams();
  const searchParams = useSearchParams();
  const currentSearch = searchParams.toString();
  const activeHref = currentSearch ? `${pathname}?${currentSearch}` : pathname;
  const locale = (params?.locale as string) || 'en';
  const t = useTranslations();
  const isRtl = locale === 'fa';
  const dir = isRtl ? 'rtl' : 'ltr';

  const session = authAdapter.getSession();
  const userName = session.name || session.email?.split('@')[0] || 'Unknown';
  const userRole = session.roles[0] || 'user';

  const navItems: NavItem[] = React.useMemo(
    () => [
      {
        id: 'overview',
        label: t('navigation.group_overview'),
        children: [
          {
            id: 'dashboard',
            label: t('navigation.dashboard'),
            href: `/dashboard`,
            icon: LayoutDashboard,
          },
          {
            id: 'observability',
            label: t('navigation.observability'),
            href: `/observability`,
            icon: Activity,
          },
        ],
      },
      {
        id: 'messaging',
        label: t('navigation.group_messaging'),
        children: [
          {
            id: 'notifications',
            label: t('navigation.notifications'),
            icon: Bell,
            children: [
              {
                id: 'all_notifications',
                label: t('navigation.all_notifications'),
                href: `/notifications`,
                icon: BellRing,
              },
              {
                id: 'send_notification',
                label: t('navigation.send_notification'),
                href: `/notifications/new`,
                icon: Send,
              },
            ],
          },
          {
            id: 'templates',
            label: t('navigation.templates'),
            icon: FileText,
            children: [
              {
                id: 'all_templates',
                label: t('navigation.all_templates'),
                href: `/templates`,
                icon: FileText,
              },
              {
                id: 'create_template',
                label: t('navigation.create_template'),
                href: `/templates/new`,
                icon: Send,
              },
            ],
          },
          {
            id: 'reminders',
            label: t('navigation.reminders'),
            icon: AlarmClock,
            children: [
              {
                id: 'all_reminders',
                label: t('navigation.all_reminders'),
                href: `/reminders`,
                icon: Clock,
              },
              {
                id: 'schedule_reminder',
                label: t('navigation.schedule_reminder'),
                href: `/reminders/new`,
                icon: Send,
              },
            ],
          },
        ],
      },
      {
        id: 'operations',
        label: t('navigation.group_operations'),
        children: [
          {
            id: 'deliveries',
            label: t('navigation.deliveries'),
            href: `/deliveries`,
            icon: Truck,
          },
          {
            id: 'providers',
            label: t('navigation.providers'),
            href: `/providers`,
            icon: Server,
          },
          {
            id: 'provider_logs',
            label: t('navigation.provider_logs'),
            href: `/provider-logs`,
            icon: ScrollText,
          },
          {
            id: 'delivery_control',
            label: t('navigation.delivery_control'),
            href: `/delivery-control`,
            icon: Shield,
          },
          {
            id: 'test_center',
            label: t('navigation.test_center'),
            icon: Wrench,
            children: [
              {
                id: 'tc_inapp',
                label: 'In-App & Self Testing',
                href: `/test-center?tab=inapp`,
                icon: Bell,
              },
              {
                id: 'tc_admin_notifs',
                label: 'Admin Notifications',
                href: `/test-center?tab=admin-notifs`,
                icon: Shield,
              },
              {
                id: 'tc_templates',
                label: 'Templates & Preview',
                href: `/test-center?tab=templates`,
                icon: FileText,
              },
              {
                id: 'tc_providers',
                label: 'Providers & Health',
                href: `/test-center?tab=providers`,
                icon: Server,
              },
              {
                id: 'tc_preferences',
                label: 'Preferences & Reminders',
                href: `/test-center?tab=preferences`,
                icon: Sliders,
              },
              {
                id: 'tc_system',
                label: 'System & Observability',
                href: `/test-center?tab=system`,
                icon: Activity,
              },
              {
                id: 'tc_operations',
                label: 'Operations Catalog',
                href: `/test-center?tab=operations`,
                icon: Play,
              },
              {
                id: 'tc_flows',
                label: 'Flow Scenarios',
                href: `/test-center?tab=flows`,
                icon: Activity,
              },
              {
                id: 'tc_config',
                label: 'Request Config',
                href: `/test-center?tab=config`,
                icon: Wrench,
              },
            ],
          },
        ],
      },
      {
        id: 'management',
        label: t('navigation.group_management'),
        children: [
          {
            id: 'preferences',
            label: t('navigation.preferences'),
            href: `/preferences`,
            icon: Settings2,
          },
          {
            id: 'tenants',
            label: t('navigation.tenants'),
            href: `/tenants`,
            icon: Building2,
          },
          {
            id: 'settings',
            label: t('navigation.settings'),
            href: `/settings`,
            icon: Cog,
          },
        ],
      },
    ],
    [locale, t]
  );

  const brandNode = (
    <div className="flex w-full flex-col gap-2">
      <Link href={`/dashboard`} className="flex items-center gap-2 px-1">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
          <Bell className="h-4 w-4" />
        </div>
        <div className="flex flex-col group-data-[collapsible=icon]:hidden">
          <span className="text-sm font-bold leading-tight">Notifier</span>
          <span className="text-[10px] leading-tight text-muted-foreground">Admin Console</span>
        </div>
      </Link>

      {/* Tenant selector — same pattern as auth/front */}
      <TenantSelector />
    </div>
  );

  const footerNode = (
    <div className="border-t border-border/50 p-2">
      <div className="flex items-center gap-2.5 rounded-lg px-2 py-1.5">
        <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
          {(userName.split(' ').length > 1
            ? userName.split(' ').map((n) => n[0]).join('')
            : userName.charAt(0)
          ).toUpperCase().slice(0, 2)}
        </div>
        <div className="flex flex-col group-data-[collapsible=icon]:hidden">
          <span className="text-xs font-medium truncate">{userName}</span>
          <span className="text-[10px] capitalize text-muted-foreground">{userRole}</span>
        </div>
      </div>
    </div>
  );

  return (
    <DSSidebar
      items={navItems}
      activeHref={activeHref}
      linkComponent={Link}
      brand={brandNode}
      footer={footerNode}
      dir={dir}
      className={className}
    />
  );
}

