'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SectionCard } from '@/components/shared/section-card';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ErrorState } from '@/components/shared/error-state';
import { PageSkeleton } from '@/components/shared/loading-state';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { NotificationAttemptsList } from '@/features/notifications/components/notification-attempts-list';
import { getNotificationDeliveries } from '@/features/notifications/api';
import { ArrowLeft, Server, RotateCcw, AlertTriangle, Timer } from 'lucide-react';
import { formatDateTime } from '@/lib/utils/date';
import { shortId } from '@/lib/utils/format';
import { toast } from 'sonner';
import type { NotificationDelivery } from '@/features/notifications/types';

export default function DeliveryDetailPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const id = params?.id as string;
  const [showRetryDialog, setShowRetryDialog] = useState(false);

  const [delivery, setDelivery] = useState<NotificationDelivery | null>(null);
  const [deliveries, setDeliveries] = useState<NotificationDelivery[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [isRetrying, setIsRetrying] = useState(false);

  const loadDelivery = async () => {
    setIsLoading(true);
    setIsError(false);
    setError(null);
    try {
      const result = await getNotificationDeliveries(id);
      if (result.length > 0) {
        setDelivery(result[0]);
        setDeliveries(result);
      } else {
        setIsError(true);
        setError(new Error(t('deliveries.no_deliveries')));
      }
    } catch (err) {
      setIsError(true);
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { loadDelivery(); }, [id]);

  const handleRetry = () => {
    setIsRetrying(true);
    setTimeout(() => {
      setIsRetrying(false);
      setShowRetryDialog(false);
      toast.success(t('notifications.actions.retry'));
      loadDelivery();
    }, 500);
  };

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title={t('deliveries.title')}>
          <Button variant="ghost" onClick={() => router.push(`/${locale}/deliveries`)} disabled>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <PageSkeleton context="deliveries" layout="detail" />
      </PageContainer>
    );
  }

  if (isError || !delivery) {
    return (
      <PageContainer>
        <PageHeader title={t('deliveries.title')}>
          <Button variant="ghost" onClick={() => router.push(`/${locale}/deliveries`)}>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.not_found')}
          message={error?.message || t('deliveries.no_deliveries')}
          onRetry={() => loadDelivery()}
        />
      </PageContainer>
    );
  }

  const canRetry = delivery.status === 'failed' || delivery.status === 'dead';
  const hasError = delivery.status === 'failed' || delivery.status === 'dead';

  return (
    <PageContainer>
      <PageHeader
        title={t('deliveries.title')}
        subtitle={`${shortId(delivery.id)} · ${delivery.provider}`}
      >
        <Button variant="ghost" size="sm" onClick={() => router.push(`/${locale}/deliveries`)}>
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
                  <Badge variant={delivery.status === 'delivered' ? 'default' : hasError ? 'destructive' : delivery.status === 'processing' ? 'secondary' : 'outline'}>
                    {t(`statuses.${delivery.status}`)}
                  </Badge>
                  <Badge variant="outline">{delivery.provider}</Badge>
                  <Badge variant="outline">{delivery.channel}</Badge>
                </div>

                <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <Server className="h-3.5 w-3.5" />
                    <span>{delivery.provider}</span>
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
                </div>

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
                  disabled={isRetrying}
                >
                  <RotateCcw className="ml-1.5 h-4 w-4" />
                  {t('notifications.actions.retry')}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Attempts */}
        <SectionCard title={t('notifications.delivery_attempts')} icon={Server}>
          <NotificationAttemptsList
            deliveries={deliveries}
            loading={false}
          />
        </SectionCard>
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
    </PageContainer>
  );
}
