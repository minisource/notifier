'use client';

import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { cn } from '@/lib/utils';
import { formatDateTime } from '@/lib/utils/date';
import {
  Clock, Send, CheckCheck, Eye, BookOpen,
  MousePointerClick, AlertCircle, XCircle, RotateCcw, Ban,
} from 'lucide-react';
import type { Notification, NotificationStatus } from '../types';

type TimelineEvent = {
  status: NotificationStatus | 'created' | 'seen' | 'read' | 'clicked' | 'delivered';
  label: string;
  time?: string;
  icon: React.ElementType;
  color: string;
};

interface NotificationTimelineProps {
  notification: Notification;
}

const statusColors: Record<string, string> = {
  created: 'border-blue-500 bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400',
  pending: 'border-blue-500 bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400',
  queued: 'border-blue-500 bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400',
  processing: 'border-amber-500 bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400',
  sent: 'border-green-500 bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400',
  delivered: 'border-green-500 bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400',
  seen: 'border-teal-500 bg-teal-100 text-teal-700 dark:bg-teal-900/40 dark:text-teal-400',
  read: 'border-emerald-500 bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400',
  clicked: 'border-cyan-500 bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-400',
  failed: 'border-red-500 bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400',
  dead: 'border-red-600 bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-400',
  cancelled: 'border-gray-400 bg-gray-100 text-gray-600 dark:bg-gray-800/40 dark:text-gray-400',
};

export function NotificationTimeline({ notification }: NotificationTimelineProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';

  const events: TimelineEvent[] = [
    { status: 'created', label: t('notifications.timeline.created'), time: notification.createdAt, icon: Clock, color: statusColors.created },
  ];

  if (notification.scheduledAt && notification.scheduledAt !== notification.createdAt) {
    events.push({
      status: 'queued', label: t('notifications.timeline.scheduled'),
      time: notification.scheduledAt, icon: Send, color: statusColors.queued,
    });
  }

  if (notification.status === 'queued' || notification.status === 'pending') {
    events.push({ status: 'queued', label: t('notifications.timeline.queued'), time: undefined, icon: Send, color: statusColors.queued });
  }

  if (notification.status === 'processing') {
    if (!events.find(e => e.status === 'queued')) {
      events.push({ status: 'queued', label: t('notifications.timeline.queued'), time: undefined, icon: Send, color: statusColors.queued });
    }
    events.push({ status: 'processing', label: t('notifications.timeline.processing'), time: undefined, icon: RotateCcw, color: statusColors.processing });
  }

  if (notification.sentAt && notification.status !== 'queued' && notification.status !== 'pending') {
    events.push({ status: 'sent', label: t('notifications.timeline.sent'), time: notification.sentAt, icon: Send, color: statusColors.sent });
  }

  if (notification.deliveredAt) {
    events.push({ status: 'delivered', label: t('notifications.timeline.delivered'), time: notification.deliveredAt, icon: CheckCheck, color: statusColors.delivered });
  }

  if (notification.seenAt) {
    events.push({ status: 'seen', label: t('notifications.timeline.seen'), time: notification.seenAt, icon: Eye, color: statusColors.seen });
  }

  if (notification.readAt) {
    events.push({ status: 'read', label: t('notifications.timeline.read'), time: notification.readAt, icon: BookOpen, color: statusColors.read });
  }

  if (notification.clickedAt) {
    events.push({ status: 'clicked', label: t('notifications.timeline.clicked'), time: notification.clickedAt, icon: MousePointerClick, color: statusColors.clicked });
  }

  if (notification.status === 'failed') {
    events.push({ status: 'failed', label: t('notifications.timeline.failed'), time: notification.updatedAt, icon: AlertCircle, color: statusColors.failed });
  }

  if (notification.status === 'dead') {
    events.push({ status: 'dead', label: t('notifications.timeline.dead'), time: notification.updatedAt, icon: XCircle, color: statusColors.dead });
  }

  if (notification.status === 'cancelled') {
    events.push({ status: 'cancelled', label: t('notifications.timeline.cancelled'), time: notification.updatedAt, icon: Ban, color: statusColors.cancelled });
  }

  return (
    <div className="space-y-0">
      {events.map((event, i) => {
        const Icon = event.icon;
        const isLast = i === events.length - 1;

        return (
          <div key={`${event.status}-${i}`} className="relative flex gap-4 pb-6 last:pb-0">
            {/* Timeline line */}
            {!isLast && (
              <div className={cn('absolute top-8 h-full w-px bg-border', isRtl ? 'left-[15px]' : 'right-[15px]')} />
            )}

            {/* Icon circle */}
            <div className={cn(
              'relative z-10 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border-2',
              event.color,
              event.time ? '' : 'opacity-60'
            )}>
              <Icon className="h-3.5 w-3.5" />
            </div>

            {/* Content */}
            <div className="min-w-0 flex-1 pt-1">
              <div className="flex items-center justify-between gap-2">
                <p className={cn(
                  'text-sm font-medium',
                  event.time ? '' : 'text-muted-foreground'
                )}>
                  {event.label}
                </p>
                {event.time && (
                  <span className="shrink-0 text-xs text-muted-foreground">
                    {formatDateTime(event.time, locale)}
                  </span>
                )}
              </div>
              {!event.time && (
                <p className="text-xs text-muted-foreground">{t('notifications.timeline.pending')}</p>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
