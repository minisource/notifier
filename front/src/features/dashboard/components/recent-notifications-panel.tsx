'use client';

import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { Send } from 'lucide-react';
import { SectionCard } from '@/components/shared/section-card';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { Button } from '@/components/ui/button';
import { useRouter } from 'next/navigation';
import { formatRelativeTime } from '@/lib/utils/date';
import { maskEmail, maskPhone, truncate } from '@/lib/utils/format';
import type { Notification } from '@/features/notifications/types';

interface RecentNotificationsPanelProps {
  notifications: Notification[];
}

export function RecentNotificationsPanel({ notifications }: RecentNotificationsPanelProps) {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';

  return (
    <SectionCard
      title={t('dashboard.recent_notifications')}
      icon={Send}
      action={
        <Button
          variant="ghost"
          size="sm"
          className="h-auto px-2 text-xs text-muted-foreground"
          onClick={() => router.push(`/${locale}/notifications`)}
        >
          {t('dashboard.view_all')}
          <span className={locale === 'fa' ? 'mr-1' : 'ml-1'}>→</span>
        </Button>
      }
    >
      {notifications.length > 0 ? (
        <div className="space-y-1">
          {notifications.map(n => (
            <div
              key={n.id}
              className="flex items-center justify-between rounded-lg px-3 py-2.5 transition-colors hover:bg-muted/40 cursor-pointer"
              onClick={() => router.push(`/${locale}/notifications/${n.id}`)}
            >
              <div className="flex items-center gap-3 min-w-0 flex-1">
                <ChannelBadge channel={n.type} size="sm" />
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium truncate">
                    {n.subject || truncate(n.body, 40)}
                  </p>
                  <p className="text-xs text-muted-foreground truncate">
                    {n.recipientEmail ? maskEmail(n.recipientEmail) :
                     n.recipientPhone ? maskPhone(n.recipientPhone) :
                     n.userId}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2 flex-shrink-0">
                <span className="text-xs text-muted-foreground hidden sm:inline">
                  {formatRelativeTime(n.createdAt, locale)}
                </span>
                <StatusBadge status={n.status} size="sm" />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-8">
          <Send className="h-8 w-8 text-muted-foreground/40" />
          <p className="mt-2 text-sm text-muted-foreground">{t('dashboard.no_recent_notifications')}</p>
        </div>
      )}
    </SectionCard>
  );
}
