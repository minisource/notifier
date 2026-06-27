'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { NotificationBell } from '@/features/notifier/notification-center/notification-bell';
import { NotificationCenterPopover } from '@/features/notifier/notification-center/notification-center-popover';
import { NotificationCenterSheet } from '@/features/notifier/notification-center/notification-center-sheet';
import { useAdminNotifications, useMeUnreadCount, useMeReadAllNotifications } from '@/features/notifier/api/notifier-queries';
import { notifierKeys } from '@/features/notifier/api/notifier-queries';
import { adminNotificationsApi } from '@/features/notifier/api/notifier-api-mode';
import { useNotifierRealtime, useShowLiveToast } from '@/features/notifier/realtime/use-notifier-realtime';
import type { Notification } from '@/features/notifier/api/notifier-types';

export function NotificationCenterWrapper() {
  const queryClient = useQueryClient();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';
  const [open, setOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<'all' | 'unread'>('all');
  const prevUnreadRef = useRef(0);
  const { showNewNotificationToast } = useShowLiveToast();

  // Initialize realtime polling
  useNotifierRealtime();

  const { data: unreadCountData, isLoading: unreadLoading } = useMeUnreadCount();
  const { data: notificationsData, isLoading: notificationsLoading } = useAdminNotifications({
    pageSize: 20,
  });

  const markAllReadMutation = useMeReadAllNotifications();

  const unreadCount = unreadCountData?.count ?? 0;
  const allNotifications = notificationsData?.data ?? [];

  const filteredNotifications = activeTab === 'unread'
    ? allNotifications.filter(n => ['pending', 'queued', 'processing', 'sent'].includes(n.status))
    : allNotifications;

  // Live toast for new unread notifications
  useEffect(() => {
    if (unreadCount > prevUnreadRef.current && prevUnreadRef.current > 0) {
      const diff = unreadCount - prevUnreadRef.current;
      showNewNotificationToast(diff, undefined, locale);
    }
    prevUnreadRef.current = unreadCount;
  }, [unreadCount, showNewNotificationToast]);

  const markReadMutation = useMutation({
    mutationFn: (notificationId: string) => adminNotificationsApi.markRead(notificationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.notifications.all });
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.notifications.unreadCount() });
    },
  });

  const handleMarkRead = useCallback((notificationId: string) => {
    markReadMutation.mutate(notificationId);
  }, [markReadMutation]);

  const handleMarkAllRead = useCallback(() => {
    markAllReadMutation.mutate();
  }, [markAllReadMutation]);

  const handleNotificationClick = useCallback((notification: Notification) => {
    setOpen(false);
    if (notification.metadata?.link && typeof notification.metadata.link === 'string') {
      const link = notification.metadata.link;
      if (link.startsWith('http://') || link.startsWith('https://')) {
        window.open(link, '_blank', 'noopener,noreferrer');
        return;
      }
    }
    router.push(`/${locale}/notifications/${notification.id}`);
  }, [locale, router]);

  return (
    <>
      <NotificationBell
        unreadCount={unreadCount}
        isLoading={unreadLoading}
        onClick={() => setOpen(true)}
      />
      <div className="hidden sm:block">
        <NotificationCenterPopover
          open={open}
          onOpenChange={setOpen}
          notifications={filteredNotifications}
          unreadCount={unreadCount}
          isLoading={notificationsLoading}
          activeTab={activeTab}
          onTabChange={setActiveTab}
          onMarkRead={handleMarkRead}
          onMarkAllRead={handleMarkAllRead}
          onNotificationClick={handleNotificationClick}
        />
      </div>
      <div className="sm:hidden">
        <NotificationCenterSheet
          open={open}
          onOpenChange={setOpen}
          notifications={filteredNotifications}
          unreadCount={unreadCount}
          isLoading={notificationsLoading}
          activeTab={activeTab}
          onTabChange={setActiveTab}
          onMarkRead={handleMarkRead}
          onMarkAllRead={handleMarkAllRead}
          onNotificationClick={handleNotificationClick}
        />
      </div>
    </>
  );
}
