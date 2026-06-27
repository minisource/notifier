'use client';

import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import { CheckCheck } from 'lucide-react';

interface NotificationCenterActionsProps {
  unreadCount: number;
  onMarkAllRead: () => void;
  isLoading?: boolean;
}

export function NotificationCenterActions({ unreadCount, onMarkAllRead, isLoading }: NotificationCenterActionsProps) {
  const t = useTranslations();

  return (
    <div className="flex items-center justify-between px-4 py-2 border-t">
      <span className="text-xs text-muted-foreground">
        {unreadCount > 0
          ? (t('notifier.notificationCenter.unreadCount') || '{count} unread').replace('{count}', String(unreadCount))
          : t('notifier.notificationCenter.allRead') || 'All caught up'}
      </span>
      <Button
        variant="ghost"
        size="sm"
        className="h-7 text-xs gap-1"
        onClick={onMarkAllRead}
        disabled={unreadCount === 0 || isLoading}
      >
        <CheckCheck className="h-3.5 w-3.5" />
        {t('notifier.notificationCenter.markAllRead') || 'Mark all read'}
      </Button>
    </div>
  );
}
