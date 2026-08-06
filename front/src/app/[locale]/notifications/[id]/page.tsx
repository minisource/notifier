'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@minisource/ui';
import { Card, CardContent, CardHeader, CardTitle } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { Badge } from '@minisource/ui';
import { NotificationSummaryCard } from '@/features/notifications/components/notification-summary-card';
import { NotificationTimeline } from '@/features/notifications/components/notification-timeline';
import { NotificationAttemptsList } from '@/features/notifications/components/notification-attempts-list';
import { NotificationMetadataViewer } from '@/features/notifications/components/notification-metadata-viewer';
import { useNotification, useNotificationDeliveries, useRetryNotification, useCancelNotification, useMarkNotificationRead } from '@/features/notifications/hooks/use-notifications';
import { Skeleton } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { ArrowLeft } from 'lucide-react';
import { maskEmail, maskPhone, shortId } from '@/lib/utils/format';
import { formatDateTime } from '@/lib/utils/date';

export default function NotificationDetailPage() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'en';
  const id = params?.id as string;

  const { data: notification, isLoading, isError, error, refetch } = useNotification(id);
  const { data: deliveries, isLoading: deliveriesLoading } = useNotificationDeliveries(id);

  const retryMutation = useRetryNotification();
  const cancelMutation = useCancelNotification();
  const markReadMutation = useMarkNotificationRead();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('notifications.detail_title')}>
          <Button variant="ghost" onClick={() => router.push(`/notifications`)} disabled>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (isError || !notification) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('notifications.detail_title')}>
          <Button variant="ghost" onClick={() => router.push(`/notifications`)}>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.not_found')}
          message={(error as Error)?.message || t('notifications.detail.not_found')}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={notification.subject || t('notifications.detail_title')}
        description={`${shortId(notification.id)} · ${formatDateTime(notification.createdAt, locale)}`}
      >
        <Button variant="ghost" size="sm" onClick={() => router.push(`/notifications`)}>
          <ArrowLeft className="ml-1.5 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <div className="space-y-6">
        {/* Summary + Actions */}
        <NotificationSummaryCard
          notification={notification}
          onRetry={() => retryMutation.mutate(notification.id)}
          onCancel={() => cancelMutation.mutate(notification.id)}
          onMarkRead={() => markReadMutation.mutate(notification.id)}
          isRetrying={retryMutation.isPending}
          isCancelling={cancelMutation.isPending}
          isMarkingRead={markReadMutation.isPending}
        />

        {/* Detail Grid */}
        <div className="grid gap-6 lg:grid-cols-2">
          {/* Recipient Info */}
          <Card><CardHeader><CardTitle>{t('notifications.recipient')}</CardTitle></CardHeader><CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('notifications.list.notification')} ID</span>
                <code className="font-mono text-xs">{notification.id}</code>
              </div>
              {notification.recipientEmail && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{t('channels.email')}</span>
                  <span className="font-mono text-xs">{maskEmail(notification.recipientEmail)}</span>
                </div>
              )}
              {notification.recipientPhone && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{t('channels.sms')}</span>
                  <span className="font-mono text-xs">{maskPhone(notification.recipientPhone)}</span>
                </div>
              )}
              {notification.recipientId && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">{t('notifications.recipient')}</span>
                  <code className="font-mono text-xs">{shortId(notification.recipientId)}</code>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('notifications.locale')}</span>
                <Badge variant="outline" className="text-xs">
                  {notification.locale === 'fa' ? 'فارسی' : 'English'}
                </Badge>
              </div>
            </div>
          </CardContent></Card>

          {/* Message Content */}
          <Card><CardHeader><CardTitle>{t('notifications.form.content_section')}</CardTitle></CardHeader><CardContent>
            <div className="space-y-3 text-sm">
              {notification.subject && (
                <div>
                  <span className="text-xs font-medium text-muted-foreground">{t('notifications.subject')}</span>
                  <p className="mt-0.5">{notification.subject}</p>
                </div>
              )}
              <div>
                <span className="text-xs font-medium text-muted-foreground">{t('notifications.body')}</span>
                <p className="mt-0.5 whitespace-pre-wrap text-sm">{notification.body}</p>
              </div>
              {notification.templateKey && (
                <div className="flex items-center gap-2 pt-1 text-xs text-muted-foreground">
                  <span>{t('notifications.template')}:</span>
                  <code className="rounded bg-muted px-1.5 py-0.5 font-mono">{notification.templateKey}</code>
                </div>
              )}
              {notification.provider && (
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>{t('deliveries.provider')}:</span>
                  <Badge variant="secondary" className="text-xs">{notification.provider}</Badge>
                </div>
              )}
            </div>
          </CardContent></Card>
        </div>

        {/* Timeline */}
        <Card><CardHeader><CardTitle>{t('notifications.timeline')}</CardTitle></CardHeader><CardContent>
          <NotificationTimeline notification={notification} />
        </CardContent></Card>

        {/* Delivery Attempts */}
        <Card><CardHeader><CardTitle>{t('notifications.delivery_attempts')}</CardTitle></CardHeader><CardContent>
          <NotificationAttemptsList
            deliveries={deliveries || []}
            loading={deliveriesLoading}
          />
        </CardContent></Card>

        {/* Metadata */}
        {notification.metadata && Object.keys(notification.metadata).length > 0 && (
          <Card><CardHeader><CardTitle>{t('notifications.metadata')}</CardTitle></CardHeader><CardContent>
            <NotificationMetadataViewer metadata={notification.metadata} />
          </CardContent></Card>
        )}
      </div>
    </div>
  );
}
