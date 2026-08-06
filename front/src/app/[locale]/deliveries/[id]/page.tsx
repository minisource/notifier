'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { Card, CardHeader, CardTitle, CardContent } from '@minisource/ui';
import { Badge } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { Skeleton } from '@minisource/ui';
import { ConfirmDialog } from '@minisource/ui';
import { NotificationAttemptsList } from '@/features/notifications/components/notification-attempts-list';
import { useDelivery, useRetryDelivery } from '@/features/deliveries/hooks/use-deliveries';
import { ArrowLeft, Server, RotateCcw, AlertTriangle, Timer, User, FileText } from 'lucide-react';
import { formatDateTime } from '@/lib/utils/date';
import { shortId, maskEmail, maskPhone } from '@/lib/utils/format';
import { toast } from 'sonner';
import { useState } from 'react';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';

export default function DeliveryDetailPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'en';
  const id = params?.id as string;
  const [showRetryDialog, setShowRetryDialog] = useState(false);

  const { data: delivery, isLoading, isError, error, refetch } = useDelivery(id);
  const retryMutation = useRetryDelivery();

  const canRetry = delivery?.status === 'failed' || delivery?.status === 'dead';
  const recipient = delivery?.recipientEmail || delivery?.recipientPhone || delivery?.recipientId;

  const handleRetry = () => {
    retryMutation.mutate(id, {
      onSuccess: () => {
        toast.success(t('notifications.actions.retry_success'));
        setShowRetryDialog(false);
        refetch();
      },
      onError: (err: Error) => {
        toast.error(err?.message || t('errors.generic'));
        setShowRetryDialog(false);
      },
    });
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('deliveries.title')}>
          <Button variant="ghost" onClick={() => router.push(`/deliveries`)} disabled>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (isError || !delivery) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('deliveries.title')}>
          <Button variant="ghost" onClick={() => router.push(`/deliveries`)}>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.not_found')}
          message={(error as Error)?.message || t('deliveries.no_deliveries')}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={delivery.subject || t('deliveries.title')}
        description={`${shortId(delivery.id)} · ${delivery.channel}`}
      >
        <Button variant="ghost" size="sm" onClick={() => router.push(`/deliveries`)}>
          <ArrowLeft className="ml-1.5 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <div className="space-y-6">
        {/* Summary */}
        <Card className="overflow-hidden">
          <CardContent className="p-5">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div className="space-y-3">
                <div className="flex flex-wrap items-center gap-2">
                  <StatusBadge status={delivery.status} size="sm" />
                  <Badge variant="outline">{delivery.provider || '—'}</Badge>
                  <ChannelBadge channel={delivery.channel} size="sm" />
                </div>

                <div className="grid grid-cols-1 gap-x-6 gap-y-2 text-sm sm:grid-cols-2">
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <Server className="h-3.5 w-3.5" />
                    <span>{delivery.provider || '—'}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <Timer className="h-3.5 w-3.5" />
                    <span>{delivery.attemptCount}/{delivery.maxAttempts} {t('deliveries.attempts')}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <span className="text-xs text-muted-foreground">{t('common.id')}:</span>
                    <code className="text-xs font-mono">{delivery.notificationId}</code>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <span className="text-xs">{formatDateTime(delivery.createdAt, locale)}</span>
                  </div>
                  {recipient && (
                    <div className="flex items-center gap-1.5 text-muted-foreground sm:col-span-2">
                      <User className="h-3.5 w-3.5" />
                      <span className="font-mono text-xs">
                        {delivery.recipientEmail
                          ? maskEmail(delivery.recipientEmail)
                          : delivery.recipientPhone
                            ? maskPhone(delivery.recipientPhone)
                            : delivery.recipientId}
                      </span>
                    </div>
                  )}
                </div>

                {/* Message content */}
                {(delivery.subject || delivery.body) && (
                  <div className="rounded-lg bg-muted/40 p-3">
                    <div className="flex items-center gap-1.5 text-sm font-medium">
                      <FileText className="h-4 w-4" />
                      {delivery.subject || t('notifications.body')}
                    </div>
                    {delivery.body && (
                      <p className="mt-1 whitespace-pre-wrap text-sm text-muted-foreground">{delivery.body}</p>
                    )}
                  </div>
                )}

                {delivery.lastError && (
                  <div className="rounded-lg bg-red-50 p-3 dark:bg-red-950/20">
                    <div className="flex items-center gap-1.5 text-sm font-medium text-red-700 dark:text-red-400">
                      <AlertTriangle className="h-4 w-4" />
                      {t('deliveries.last_error')}
                    </div>
                    <p className="mt-1 text-sm text-red-600 dark:text-red-300">{delivery.lastError}</p>
                  </div>
                )}
              </div>

              {canRetry && (
                <Button
                  size="sm"
                  variant={delivery.status === 'dead' ? 'destructive' : 'default'}
                  onClick={() => setShowRetryDialog(true)}
                  disabled={retryMutation.isPending}
                >
                  <RotateCcw className="ml-1.5 h-4 w-4" />
                  {retryMutation.isPending ? t('deliveries.retrying') : t('notifications.actions.retry')}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Attempts */}
        <Card><CardHeader><CardTitle>{t('notifications.delivery_attempts')}</CardTitle></CardHeader><CardContent>
          <NotificationAttemptsList
            deliveries={delivery.attempts?.length ? [delivery] : []}
            loading={false}
          />
        </CardContent></Card>
      </div>

      <ConfirmDialog
        open={showRetryDialog}
        onOpenChange={setShowRetryDialog}
        onConfirm={handleRetry}
        title={t('notifications.actions.confirm_retry_title')}
        description={t('notifications.actions.confirm_retry_desc')}
        confirmLabel={t('notifications.actions.retry')}
        cancelLabel={t('common.cancel')}
        destructive={delivery.status === 'dead'}
      />
    </div>
  );
}
