'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { SectionCard } from '@/components/shared/section-card';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { Button } from '@/components/ui/button';
import { AlertTriangle, ExternalLink } from 'lucide-react';
import { formatRelativeTime } from '@/lib/utils/date';
import { shortId } from '@/lib/utils/format';
import type { RecentFailure } from '@/features/notifier/api/notifier-types';

interface DashboardRecentFailuresProps {
  failures: RecentFailure[];
  deadLetters: RecentFailure[];
}

export function DashboardRecentFailures({ failures, deadLetters }: DashboardRecentFailuresProps) {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';

  const all = [...failures, ...deadLetters.filter(dl => !failures.find(f => f.notificationId === dl.notificationId))]
    .slice(0, 5);

  if (all.length === 0) {
    return (
      <SectionCard title={t('dashboard.recent_failures')} icon={AlertTriangle}>
        <div className="flex flex-col items-center justify-center py-8">
          <AlertTriangle className="h-8 w-8 text-muted-foreground/40" />
          <p className="mt-2 text-sm text-muted-foreground">{t('dashboard.no_recent_failures')}</p>
        </div>
      </SectionCard>
    );
  }

  return (
    <SectionCard
      title={t('dashboard.recent_failures')}
      icon={AlertTriangle}
      action={
        all.length > 0 && (
          <Button
            variant="ghost"
            size="sm"
            className="h-auto px-2 text-xs text-muted-foreground"
            onClick={() => router.push(`/${locale}/notifications?status=failed`)}
          >
            {t('dashboard.view_all')} →
          </Button>
        )
      }
    >
      <div className="space-y-1">
        {all.map((failure) => (
          <div
            key={`${failure.notificationId}-${failure.errorCode}`}
            className="flex items-center justify-between rounded-lg px-3 py-2.5 transition-colors hover:bg-muted/40 cursor-pointer"
            onClick={() => router.push(`/${locale}/notifications/${failure.notificationId}`)}
          >
            <div className="flex items-center gap-3 min-w-0 flex-1">
              <ChannelBadge channel={failure.channel} size="sm" />
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium truncate">
                  {failure.errorMessage ? (
                    <span className="text-red-600 dark:text-red-400">{failure.errorMessage}</span>
                  ) : (
                    <span className="text-muted-foreground">{t('common.error_occurred')}</span>
                  )}
                </p>
                <p className="text-xs text-muted-foreground">
                  <code className="font-mono text-[10px]">{shortId(failure.notificationId)}</code>
                  {failure.provider && <span className="ml-2">via {failure.provider}</span>}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              {failure.status && <StatusBadge status={failure.status} size="sm" />}
              <span className="text-xs text-muted-foreground hidden sm:inline">
                {formatRelativeTime(failure.createdAt, locale)}
              </span>
              <ExternalLink className="h-3.5 w-3.5 text-muted-foreground" />
            </div>
          </div>
        ))}
      </div>
    </SectionCard>
  );
}
