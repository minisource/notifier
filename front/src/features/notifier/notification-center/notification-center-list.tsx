'use client';

import { NotificationCenterItem } from './notification-center-item';
import { NotificationCenterEmpty } from './notification-center-empty';
import { NotificationCenterSkeleton } from './notification-center-skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import type { Notification } from '@/features/notifier/api/notifier-types';

interface NotificationCenterListProps {
  notifications: Notification[];
  isLoading: boolean;
  onMarkRead: (id: string) => void;
  onNotificationClick: (notification: Notification) => void;
}

export function NotificationCenterList({
  notifications, isLoading, onMarkRead, onNotificationClick,
}: NotificationCenterListProps) {

  if (isLoading) {
    return <NotificationCenterSkeleton count={5} />;
  }

  if (notifications.length === 0) {
    return <NotificationCenterEmpty />;
  }

  return (
    <ScrollArea className="h-[400px]">
      <div className="divide-y">
        {notifications.map((notification) => (
          <NotificationCenterItem
            key={notification.id}
            notification={notification}
            onMarkRead={onMarkRead}
            onClick={onNotificationClick}
          />
        ))}
      </div>
    </ScrollArea>
  );
}
