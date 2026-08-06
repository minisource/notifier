'use client';

import { Button } from '@minisource/ui';
import { ThemeToggle } from '@/components/layout/theme-toggle';
import { LanguageSwitcher } from '@/components/layout/language-switcher';
import { UserMenu } from '@/components/layout/user-menu';
import { NotificationCenterWrapper } from '@/features/notifier/notification-center/notification-center-wrapper';
import { Menu } from 'lucide-react';
import { useTranslations } from 'next-intl';

interface TopbarProps {
  onMenuClick: () => void;
}

export function Topbar({ onMenuClick }: TopbarProps) {
  const t = useTranslations();

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-border/70 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4 md:px-6">
      {/* Mobile menu trigger */}
      <Button variant="ghost" size="icon" className="lg:hidden" onClick={onMenuClick} aria-label={t('common.menu')}>
        <Menu className="h-5 w-5" />
      </Button>

      {/* Spacer */}
      <div className="flex-1" />

      {/* Right side controls */}
      <div className="flex items-center gap-1.5">
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
