'use client';

import Link from 'next/link';
import { usePathname, useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { cn } from '@/lib/utils';
import { ScrollArea } from '@/components/ui/scroll-area';
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
  ChevronDown,
  BellRing,
  Send,
  Clock,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { useState } from 'react';
import { authAdapter } from '@/shared/auth/auth-adapter';
import { env } from '@/lib/config/env';

interface NavItem {
  labelKey: string;
  href?: string;
  icon: React.ElementType;
  children?: { labelKey: string; href: string; icon: React.ElementType }[];
}

interface NavGroup {
  labelKey: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    labelKey: 'navigation.group_overview',
    items: [
      { labelKey: 'navigation.dashboard', href: '/dashboard', icon: LayoutDashboard },
      { labelKey: 'navigation.observability', href: '/observability', icon: Activity },
    ],
  },
  {
    labelKey: 'navigation.group_messaging',
    items: [
      {
        labelKey: 'navigation.notifications', icon: Bell,
        children: [
          { labelKey: 'navigation.all_notifications', href: '/notifications', icon: BellRing },
          { labelKey: 'navigation.send_notification', href: '/notifications/new', icon: Send },
        ],
      },
      {
        labelKey: 'navigation.templates', icon: FileText,
        children: [
          { labelKey: 'navigation.all_templates', href: '/templates', icon: FileText },
          { labelKey: 'navigation.create_template', href: '/templates/new', icon: Send },
        ],
      },
      {
        labelKey: 'navigation.reminders', icon: AlarmClock,
        children: [
          { labelKey: 'navigation.all_reminders', href: '/reminders', icon: Clock },
          { labelKey: 'navigation.schedule_reminder', href: '/reminders/new', icon: Send },
        ],
      },
    ],
  },
  {
    labelKey: 'navigation.group_operations',
    items: [
      { labelKey: 'navigation.deliveries', href: '/deliveries', icon: Truck },
      { labelKey: 'navigation.providers', href: '/providers', icon: Server },
    ],
  },
  {
    labelKey: 'navigation.group_management',
    items: [
      { labelKey: 'navigation.preferences', href: '/preferences', icon: Settings2 },
      { labelKey: 'navigation.tenants', href: '/tenants', icon: Building2 },
      { labelKey: 'navigation.settings', href: '/settings', icon: Cog },
    ],
  },
];

interface SidebarProps {
  open: boolean;
  onClose: () => void;
}

function NavItemLink({ item, locale, onClose, isActiveFn }: { item: NavItem; locale: string; onClose: () => void; isActiveFn: (href: string) => boolean }) {
  const t = useTranslations();
  const href = item.href!;
  const isActive = isActiveFn(href);
  const isRtl = locale === 'fa';

  return (
    <Link
      href={`/${locale}${href}`}
      onClick={onClose}
      className={cn(
        'relative flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all',
        isActive
          ? 'bg-primary/10 text-primary dark:bg-primary/20'
          : 'text-muted-foreground hover:bg-accent/60 hover:text-accent-foreground'
      )}
    >
      {/* Active indicator bar */}
      {isActive && (
        <span className={cn(
          'absolute inset-y-1 w-0.5 rounded-full bg-primary',
          isRtl ? 'right-0' : 'left-0'
        )} />
      )}
      <item.icon className={cn('h-4 w-4', isActive ? 'text-primary' : '')} />
      {t(item.labelKey)}
    </Link>
  );
}

export function Sidebar({ open: _open, onClose }: SidebarProps) {
  const pathname = usePathname();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const t = useTranslations();
  const isRtl = locale === 'fa';
  const [collapsed, setCollapsed] = useState(false);
  const [expanded, setExpanded] = useState<string[]>(['navigation.notifications', 'navigation.templates', 'navigation.reminders']);

  const toggleExpanded = (key: string) => {
    setExpanded(prev => prev.includes(key) ? prev.filter(i => i !== key) : [...prev, key]);
  };

  const isActive = (href: string) => {
    if (href === '#') return false;
    const fullPath = `/${locale}${href}`;
    if (pathname === fullPath) return true;
    if (pathname.startsWith(fullPath + '/')) {
      const remaining = pathname.slice(fullPath.length + 1);
      // Don't match form pages (new, create, edit) — they have their own menu item
      const firstSegment = remaining.split('/')[0];
      if (firstSegment === 'new' || firstSegment === 'create' || firstSegment === 'edit') return false;
      return true;
    }
    return false;
  };

  const session = authAdapter.getSession();
  const userName = session.userId || 'Unknown';
  const userRole = session.roles[0] || 'user';
  const mode = env.isMockDataEnabled ? 'mock' : 'real';
  const user = { name: userName, role: userRole as string };
  const role = user.role;

  const sidebarWidth = collapsed ? 'w-16' : 'w-[var(--sidebar-width)]';

  return (
    <aside className={cn(
      `hidden flex-shrink-0 border-r bg-sidebar lg:flex lg:flex-col transition-all duration-200`,
      sidebarWidth,
      isRtl ? 'border-l' : 'border-r'
    )}>
      {/* Product Header */}
      <div className={cn(
        'flex h-16 items-center border-b border-border/50 px-4',
        collapsed ? 'justify-center' : 'justify-between'
      )}>
        <Link href={`/${locale}/dashboard`} className={cn('flex items-center gap-2', collapsed && 'justify-center')}>
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Bell className="h-4 w-4" />
          </div>
          {!collapsed && (
            <div className="flex flex-col">
              <span className="text-sm font-bold leading-tight">Notifier</span>
              <span className="text-[10px] leading-tight text-muted-foreground">Admin Console</span>
            </div>
          )}
        </Link>
        {!collapsed && (
          <button
            onClick={() => setCollapsed(true)}
            className="rounded p-1 text-muted-foreground hover:bg-accent"
            aria-label="Collapse sidebar"
          >
            {isRtl ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </button>
        )}
        {collapsed && (
          <button
            onClick={() => setCollapsed(false)}
            className="rounded p-1 text-muted-foreground hover:bg-accent"
            aria-label="Expand sidebar"
          >
            {isRtl ? <ChevronLeft className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </button>
        )}
      </div>

      {/* Environment Pill */}
      {!collapsed && (
        <div className="px-3 pt-3">
          <div className="flex items-center gap-2 rounded-md border border-border/50 bg-muted/50 px-3 py-2">
            <span className={cn(
              'h-2 w-2 rounded-full',
              mode === 'mock' ? 'bg-amber-500' : 'bg-green-500'
            )} />
            <span className="text-xs font-medium text-muted-foreground">
              {mode === 'mock' ? t('settings.mock') : t('settings.real')} Mode
            </span>
          </div>
        </div>
      )}

      {/* Navigation */}
      <ScrollArea className="flex-1 px-3 py-4">
        <nav className="space-y-5">
          {navGroups.map((group) => (
            <div key={group.labelKey}>
              {!collapsed && (
                <p className="mb-1 px-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/70">
                  {t(group.labelKey)}
                </p>
              )}
              <div className="space-y-0.5">
                {group.items.map(item => {
                  if (item.children) {
                    const isExpanded = expanded.includes(item.labelKey);
                    const hasActiveChild = item.children.some(c => isActive(c.href));
                    return (
                      <div key={item.labelKey}>
                        <button
                          onClick={() => toggleExpanded(item.labelKey)}
                          className={cn(
                            'flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all',
                            hasActiveChild
                              ? 'text-primary'
                              : 'text-muted-foreground hover:bg-accent/60 hover:text-accent-foreground'
                          )}
                        >
                          <item.icon className={cn('h-4 w-4', hasActiveChild ? 'text-primary' : '')} />
                          {!collapsed && (
                            <>
                              <span className="flex-1 text-left">{t(item.labelKey)}</span>
                              <ChevronDown className={cn('h-3.5 w-3.5 transition-transform', isExpanded && 'rotate-180')} />
                            </>
                          )}
                        </button>
                        {isExpanded && !collapsed && (
                          <div className={cn(
                            'mt-0.5 space-y-0.5',
                            isRtl ? 'pr-3' : 'pl-3'
                          )}>
                            {item.children.map(child => (
                              <NavItemLink
                                key={child.href}
                                item={child}
                                locale={locale}
                                onClose={onClose}
                                isActiveFn={isActive}
                              />
                            ))}
                          </div>
                        )}
                      </div>
                    );
                  }
                  return (
                    <NavItemLink
                      key={item.href}
                      item={item}
                      locale={locale}
                      onClose={onClose}
                      isActiveFn={isActive}
                    />
                  );
                })}
              </div>
            </div>
          ))}
        </nav>
      </ScrollArea>

      {/* User footer */}
      {!collapsed && (
        <div className="border-t border-border/50 p-3">
          <div className="flex items-center gap-3 rounded-lg px-3 py-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
              {(user.name.split(' ').length > 1 ? user.name.split(' ').map(n => n[0]).join('') : user.name.charAt(0)).toUpperCase().slice(0, 2)}
            </div>
            <div className="flex flex-col">
              <span className="text-xs font-medium">{user.name}</span>
              <span className="text-[10px] text-muted-foreground capitalize">{role}</span>
            </div>
          </div>
        </div>
      )}
    </aside>
  );
}
