'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { X, LayoutDashboard, Bell, FileText, AlarmClock, Truck, Server, Settings2, Building2, Activity, Cog } from 'lucide-react';
import { useTranslations } from 'next-intl';

interface MobileNavProps {
  open: boolean;
  onClose: () => void;
}

const navItems = [
  { labelKey: 'navigation.dashboard', href: '/dashboard', icon: LayoutDashboard },
  { labelKey: 'navigation.notifications', href: '/notifications', icon: Bell },
  { labelKey: 'navigation.templates', href: '/templates', icon: FileText },
  { labelKey: 'navigation.reminders', href: '/reminders', icon: AlarmClock },
  { labelKey: 'navigation.deliveries', href: '/deliveries', icon: Truck },
  { labelKey: 'navigation.providers', href: '/providers', icon: Server },
  { labelKey: 'navigation.preferences', href: '/preferences', icon: Settings2 },
  { labelKey: 'navigation.tenants', href: '/tenants', icon: Building2 },
  { labelKey: 'navigation.observability', href: '/observability', icon: Activity },
  { labelKey: 'navigation.settings', href: '/settings', icon: Cog },
];

export function MobileNav({ open, onClose }: MobileNavProps) {
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const t = useTranslations();

  return (
    <>
      {open && <div className="fixed inset-0 z-40 bg-black/50" onClick={onClose} />}
      <aside className={cn(
        'fixed inset-y-0 z-50 flex w-64 flex-col border-l bg-background transition-transform duration-300 lg:hidden',
        open ? 'translate-x-0' : '-translate-x-full',
        locale === 'fa' && open ? 'right-0 translate-x-0 border-r' : '',
        locale === 'fa' && !open ? '-translate-x-full' : ''
      )}>
        <div className="flex h-16 items-center justify-between border-b px-6">
          <Link href={`/${locale}/dashboard`} className="text-xl font-bold">
            Notifier
          </Link>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>
        <ScrollArea className="flex-1 px-3 py-4">
          <nav className="space-y-1">
            {navItems.map(item => (
              <Link
                key={item.href}
                href={`/${locale}${item.href}`}
                onClick={onClose}
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
              >
                <item.icon className="h-4 w-4" />
                {t(item.labelKey)}
              </Link>
            ))}
          </nav>
        </ScrollArea>
      </aside>
    </>
  );
}
