'use client';

import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/layout/theme-toggle';
import { LanguageSwitcher } from '@/components/layout/language-switcher';
import { TenantSwitcher } from '@/components/layout/tenant-switcher';
import { UserMenu } from '@/components/layout/user-menu';
import { NotificationCenterWrapper } from '@/features/notifier/notification-center/notification-center-wrapper';
import { Menu } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { cn } from '@/lib/utils';
import { isMockDataEnabled } from '@/lib/config/env';

interface TopbarProps {
  onMenuClick: () => void;
}

export function Topbar({ onMenuClick }: TopbarProps) {
  const t = useTranslations();
  const mockMode = isMockDataEnabled();

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-border/70 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4 md:px-6">
      {/* Mobile menu trigger */}
      <Button variant="ghost" size="icon" className="lg:hidden" onClick={onMenuClick} aria-label={t('common.menu')}>
        <Menu className="h-5 w-5" />
      </Button>

      {/* API Mode Badge */}
      <div className={cn(
        'hidden sm:flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium',
        mockMode
          ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400'
          : 'bg-green-50 text-green-700 dark:bg-green-950/40 dark:text-green-400'
      )}>
        <span className={cn(
          'h-1.5 w-1.5 rounded-full',
          mockMode ? 'bg-amber-500' : 'bg-green-500'
        )} />
        {mockMode ? t('settings.mock') : t('settings.real')}
      </div>

      {/* Spacer */}
      <div className="flex-1" />

      {/* Right side controls */}
      <div className="flex items-center gap-1.5">
        <TenantSwitcher />
        <NotificationCenterWrapper />
        <div className="mx-1 h-6 w-px bg-border/50" />
        <LanguageSwitcher />
        <ThemeToggle />
        <div className="mx-1 h-6 w-px bg-border/50" />
        <UserMenu />
      </div>
    </header>
  );
}
