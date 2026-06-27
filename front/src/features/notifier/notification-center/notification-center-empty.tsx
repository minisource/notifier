'use client';

import { useTranslations } from 'next-intl';
import { BellOff } from 'lucide-react';

export function NotificationCenterEmpty() {
  const t = useTranslations();

  return (
    <div className="flex flex-col items-center justify-center py-12 px-4">
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
        <BellOff className="h-6 w-6 text-muted-foreground" />
      </div>
      <p className="mt-3 text-sm font-medium text-foreground">
        {t('notifications.list.no_notifications_yet') || 'No notifications'}
      </p>
      <p className="mt-1 text-xs text-muted-foreground text-center">
        {t('notifier.notificationCenter.emptyDescription') || 'You have no notifications yet. They will appear here when you receive them.'}
      </p>
    </div>
  );
}
