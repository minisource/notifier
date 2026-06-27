'use client';

import { useTranslations } from 'next-intl';
import { Card, CardContent } from '@/components/ui/card';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import {
  RotateCcw, XCircle, CheckCheck, Copy,
  AlertTriangle, Timer, Hash, User,
  Calendar,
} from 'lucide-react';
import { toast } from 'sonner';
import { formatDateTime } from '@/lib/utils/date';
import { useParams } from 'next/navigation';
import type { Notification } from '../types';

interface NotificationSummaryCardProps {
  notification: Notification;
  onRetry?: () => void;
  onCancel?: () => void;
  onMarkRead?: () => void;
  isRetrying?: boolean;
  isCancelling?: boolean;
  isMarkingRead?: boolean;
}

export function NotificationSummaryCard({
  notification,
  onRetry, onCancel, onMarkRead,
  isRetrying, isCancelling, isMarkingRead,
}: NotificationSummaryCardProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';

  const handleCopyId = () => {
    navigator.clipboard.writeText(notification.id);
    toast.success(t('common.copied'));
  };

  const canRetry = notification.status === 'failed' || notification.status === 'dead';
  const canCancel = notification.status === 'pending' || notification.status === 'queued' || notification.status === 'processing';
  const canMarkRead = notification.type === 'in_app' && !notification.readAt;

  return (
    <Card className="overflow-hidden">
      <CardContent className="p-5">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-3">
            {/* Status + Channel + Priority row */}
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status={notification.status} />
              <ChannelBadge channel={notification.type} />
              {notification.priority === 'urgent' && (
                <span className="inline-flex items-center gap-1 rounded-md bg-red-50 px-2 py-0.5 text-xs font-medium text-red-700 dark:bg-red-950/40 dark:text-red-400">
                  <AlertTriangle className="h-3 w-3" />
                  {t('notifications.filters.priority_urgent')}
                </span>
              )}
              {notification.priority === 'high' && (
                <span className="inline-flex items-center gap-1 rounded-md bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-950/40 dark:text-amber-400">
                  {t('notifications.filters.priority_high')}
                </span>
              )}
            </div>

            {/* Title */}
            {notification.subject && (
              <h2 className="text-lg font-semibold leading-tight">{notification.subject}</h2>
            )}

            {/* ID */}
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Hash className="h-3.5 w-3.5" />
              <code className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">{notification.id}</code>
              <Button variant="ghost" size="icon" className="h-6 w-6" onClick={handleCopyId}>
                <Copy className="h-3 w-3" />
              </Button>
            </div>

            {/* Metadata grid */}
            <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm sm:grid-cols-3">
              {notification.userId && (
                <div className="flex items-center gap-1.5 text-muted-foreground">
                  <User className="h-3.5 w-3.5" />
                  <span>{notification.userId}</span>
                </div>
              )}
              <div className="flex items-center gap-1.5 text-muted-foreground">
                <Calendar className="h-3.5 w-3.5" />
                <span>{formatDateTime(notification.createdAt, locale)}</span>
              </div>
              {(notification.sentAt || notification.deliveredAt) && (
                <div className="flex items-center gap-1.5 text-muted-foreground">
                  <Timer className="h-3.5 w-3.5" />
                  <span>
                    {notification.deliveredAt
                      ? t('notifications.summary.delivered')
                      : notification.sentAt
                        ? t('notifications.summary.sent')
                        : ''}
                  </span>
                </div>
              )}
              {notification.retryCount > 0 && (
                <div className="flex items-center gap-1.5 text-muted-foreground">
                  <RotateCcw className="h-3.5 w-3.5" />
                  <span>{notification.retryCount}/{notification.maxRetries} {t('notifications.retry_count')}</span>
                </div>
              )}
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex shrink-0 flex-wrap items-center gap-2">
            {canRetry && onRetry && (
              <Button size="sm" onClick={onRetry} disabled={isRetrying} variant={notification.status === 'dead' ? 'destructive' : 'default'}>
                <RotateCcw className="ml-1.5 h-4 w-4" />
                {t('notifications.actions.retry')}
              </Button>
            )}
            {canCancel && onCancel && (
              <Button size="sm" variant="outline" onClick={onCancel} disabled={isCancelling}>
                <XCircle className="ml-1.5 h-4 w-4" />
                {t('notifications.actions.cancel')}
              </Button>
            )}
            {canMarkRead && onMarkRead && (
              <Button size="sm" variant="outline" onClick={onMarkRead} disabled={isMarkingRead}>
                <CheckCheck className="ml-1.5 h-4 w-4" />
                {t('notifications.actions.mark_read')}
              </Button>
            )}
          </div>
        </div>

        {notification.errorMessage && (
          <>
            <Separator className="my-4" />
            <div className="rounded-lg bg-red-50 p-3 text-sm dark:bg-red-950/20">
              <div className="flex items-center gap-1.5 font-medium text-red-700 dark:text-red-400">
                <AlertTriangle className="h-4 w-4" />
                {t('notifications.error_message')}
              </div>
              <p className="mt-1 text-red-600 dark:text-red-300">{notification.errorMessage}</p>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
