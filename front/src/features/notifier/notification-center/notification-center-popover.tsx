'use client';

import { useTranslations } from 'next-intl';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { NotificationCenterList } from './notification-center-list';
import type { Notification } from '@/features/notifier/api/notifier-types';

interface NotificationCenterPopoverProps {
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

export function NotificationCenterPopover({
  open, onOpenChange, notifications, unreadCount, isLoading,
  activeTab, onTabChange, onMarkRead, onMarkAllRead, onNotificationClick,
}: NotificationCenterPopoverProps) {
  const t = useTranslations();

  return (
    <Popover open={open} onOpenChange={onOpenChange}>
      <PopoverTrigger asChild>
        <span />
      </PopoverTrigger>
      <PopoverContent
        align="end"
        sideOffset={8}
        className="w-[380px] p-0"
      >
        <div className="flex items-center justify-between border-b px-4 py-3">
          <h3 className="text-sm font-semibold">
            {t('notifier.notificationCenter.title') || 'Notifications'}
          </h3>
          <button
            onClick={onMarkAllRead}
            disabled={unreadCount === 0}
            className="text-xs text-primary hover:underline disabled:text-muted-foreground disabled:no-underline"
            aria-label={t('notifier.notificationCenter.markAllRead') || 'Mark all as read'}
          >
            {t('notifier.notificationCenter.markAllRead') || 'Mark all read'}
          </button>
        </div>
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
        </div>
        <NotificationCenterList
          notifications={notifications}
          isLoading={isLoading}
          onMarkRead={onMarkRead}
          onNotificationClick={onNotificationClick}
        />
      </PopoverContent>
    </Popover>
  );
}
