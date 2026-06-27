'use client';

import { useTranslations } from 'next-intl';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { NotificationCenterList } from './notification-center-list';
import type { Notification } from '@/features/notifier/api/notifier-types';

interface NotificationCenterSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;
  activeTab: 'all' | 'unread';
  onTabChange: (tab: 'all' | 'unread') => void;
  onMarkRead: (id: string) => void;
  onMarkAllRead: () => void;
  onNotificationClick: (notification: Notification) => void;
}

export function NotificationCenterSheet({
  open, onOpenChange, notifications, unreadCount, isLoading,
  activeTab, onTabChange, onMarkRead, onMarkAllRead, onNotificationClick,
}: NotificationCenterSheetProps) {
  const t = useTranslations();

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-md p-0">
        <SheetHeader className="px-4 py-3 border-b">
          <SheetTitle className="text-sm font-semibold">
            {t('notifier.notificationCenter.title') || 'Notifications'}
          </SheetTitle>
        </SheetHeader>
        <div className="flex border-b">
          <button
            onClick={() => onTabChange('all')}
            className={`flex-1 px-4 py-2 text-xs font-medium ${
              activeTab === 'all'
                ? 'border-b-2 border-primary text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {t('common.all')}
          </button>
          <button
            onClick={() => onTabChange('unread')}
            className={`flex-1 px-4 py-2 text-xs font-medium ${
              activeTab === 'unread'
                ? 'border-b-2 border-primary text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {t('notifier.notificationCenter.unread') || 'Unread'}
            {unreadCount > 0 && (
              <span className="ml-1.5 rounded-full bg-destructive px-1.5 py-0.5 text-[10px] text-destructive-foreground">
                {unreadCount}
              </span>
            )}
          </button>
          <button
            onClick={onMarkAllRead}
            disabled={unreadCount === 0}
            className="px-3 py-2 text-xs text-primary hover:underline disabled:text-muted-foreground disabled:no-underline"
          >
            {t('notifier.notificationCenter.markAllRead') || 'Mark all read'}
          </button>
        </div>
        <NotificationCenterList
          notifications={notifications}
          isLoading={isLoading}
          onMarkRead={onMarkRead}
          onNotificationClick={onNotificationClick}
        />
      </SheetContent>
    </Sheet>
  );
}
