'use client';

import { useEffect, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslations } from 'next-intl';
import { realtimeClient } from './notifier-realtime-client';
import { getRealtimeConfig, shouldShowLiveToasts } from './notifier-realtime-types';
import { notifierKeys } from '@/features/notifier/api/notifier-queries';

export function useNotifierRealtime() {
  const queryClient = useQueryClient();
  const config = getRealtimeConfig();
  shouldShowLiveToasts(); // Evaluate once

  useEffect(() => {
    if (config.mode === 'disabled') return;

    realtimeClient.connect();

    const unsubNotification = realtimeClient.on('notification', () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.notifications.all });
      queryClient.invalidateQueries({ queryKey: notifierKeys.notifications.lists() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.dashboard.all });
    });

    const unsubUnread = realtimeClient.on('unread_change', () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.notifications.unreadCount() });
    });

    const unsubHealth = realtimeClient.on('health_change', () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.dashboard.health() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.providers.health() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.dashboard.all });
    });

    const unsubQueue = realtimeClient.on('queue_change', () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.dashboard.queue() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.dashboard.overview() });
    });

    return () => {
      unsubNotification();
      unsubUnread();
      unsubHealth();
      unsubQueue();
      realtimeClient.disconnect();
    };
  }, [queryClient, config.mode]);
}

export function usePollingWithVisibility(_pollIntervalMs?: number) {
  const queryClient = useQueryClient();

  useEffect(() => {
    const handleVisibility = () => {
      if (document.visibilityState === 'visible') {
        queryClient.invalidateQueries({ queryKey: notifierKeys.dashboard.all });
        queryClient.invalidateQueries({ queryKey: notifierKeys.me.notifications.unreadCount() });
      }
    };

    document.addEventListener('visibilitychange', handleVisibility);
    return () => document.removeEventListener('visibilitychange', handleVisibility);
  }, [queryClient]);
}

export function useShowLiveToast() {
  const t = useTranslations();
  const showToasts = shouldShowLiveToasts();

  const showNewNotificationToast = useCallback((count: number, title?: string, locale?: string) => {
    if (!showToasts) return;
    if (count > 1) {
      const desc = t('notifier.notificationCenter.newNotifications').replace('{count}', String(count));
      toast(desc, {
        description: t('common.click_to_view') || 'Click to view',
        action: {
          label: t('notifier.notificationCenter.view'),
          onClick: () => { window.location.href = `/${locale || 'en'}/notifications`; },
        },
        duration: 5000,
      });
    } else if (title) {
      toast(title, {
        description: t('notifier.notificationCenter.newNotification'),
        duration: 5000,
      });
    }
  }, [showToasts, t]);

  return { showNewNotificationToast };
}
