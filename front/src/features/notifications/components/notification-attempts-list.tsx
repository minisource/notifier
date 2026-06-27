'use client';

import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { Card, CardContent } from '@/components/ui/card';
import { StatusBadge } from '@/components/shared/status-badge';
import { Badge } from '@/components/ui/badge';
import { formatRelativeTime } from '@/lib/utils/date';
import { formatMilliseconds } from '@/lib/utils/format';
import { cn } from '@/lib/utils';
import { AlertTriangle, Timer, RotateCcw, Server } from 'lucide-react';
import type { NotificationDelivery } from '../types';

interface NotificationAttemptsListProps {
  deliveries: NotificationDelivery[];
  loading?: boolean;
}

export function NotificationAttemptsList({ deliveries, loading }: NotificationAttemptsListProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';

  if (loading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-20 animate-pulse rounded-lg bg-muted" />
        ))}
      </div>
    );
  }

  if (deliveries.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Timer className="h-10 w-10 text-muted-foreground/50" />
        <p className="mt-3 text-sm text-muted-foreground">{t('notifications.attempts.no_attempts')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {deliveries.map((delivery) => (
        <Card key={delivery.id} className="overflow-hidden">
          <CardContent className="p-4">
            {/* Delivery header */}
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <Server className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-medium">{delivery.provider}</span>
                <StatusBadge status={delivery.status} size="sm" />
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>{delivery.attemptCount}/{delivery.maxAttempts} {t('deliveries.attempts')}</span>
                {delivery.nextRetryAt && (
                  <Badge variant="outline" className="gap-1 text-xs">
                    <RotateCcw className="h-3 w-3" />
                    {formatRelativeTime(delivery.nextRetryAt, locale)}
                  </Badge>
                )}
              </div>
            </div>

            {/* Attempts list */}
            <div className="space-y-1.5" dir={isRtl ? 'rtl' : 'ltr'}>
              {delivery.attempts.map((attempt) => (
                <div
                  key={attempt.id}
                  className={cn(
                    'flex flex-wrap items-center justify-between gap-2 rounded-md border p-2.5 text-sm',
                    attempt.status === 'failed' || attempt.status === 'dead'
                      ? 'border-red-200 bg-red-50/50 dark:border-red-900/50 dark:bg-red-950/10'
                      : attempt.status === 'delivered'
                        ? 'border-green-200 bg-green-50/50 dark:border-green-900/50 dark:bg-green-950/10'
                        : 'border-border bg-muted/20'
                  )}
                >
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium text-muted-foreground">
                      #{attempt.attemptNumber}
                    </span>
                    <StatusBadge status={attempt.status} size="sm" />
                    {attempt.errorCode && (
                      <code className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-mono text-red-600 dark:text-red-400">
                        {attempt.errorCode}
                      </code>
                    )}
                  </div>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <Timer className="h-3 w-3" />
                      {formatMilliseconds(attempt.processingTimeMs)}
                    </span>
                    <span>{formatRelativeTime(attempt.createdAt, locale)}</span>
                  </div>

                  {/* Error details */}
                  {attempt.errorMessage && (
                    <div className="w-full mt-1 flex items-start gap-1.5 text-xs">
                      <AlertTriangle className="mt-0.5 h-3 w-3 shrink-0 text-red-500" />
                      <span className="text-red-600 dark:text-red-400">{attempt.errorMessage}</span>
                    </div>
                  )}

                  {/* Provider response */}
                  {attempt.providerResponse && (
                    <div className="w-full mt-1 text-xs text-green-600 dark:text-green-400">
                      <span className="font-medium">{t('notifications.attempts.provider_response')}:</span>{' '}
                      {attempt.providerResponse}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
