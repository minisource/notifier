'use client';

import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { StatusBadge } from '@/components/shared/status-badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { formatRelativeTime } from '@/lib/utils/date';
import { shortId } from '@/lib/utils/format';
import { MailOpen } from 'lucide-react';
import type { Notification } from '@/features/notifier/api/notifier-types';

interface NotificationCenterItemProps {
  notification: Notification;
  onMarkRead: (id: string) => void;
  onClick: (notification: Notification) => void;
}

export function NotificationCenterItem({ notification, onMarkRead, onClick }: NotificationCenterItemProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';

  const isUnread = notification.status === 'pending' || notification.status === 'queued' || notification.status === 'processing' || notification.status === 'sent';

  return (
    <div
      className={cn(
        'flex items-start gap-3 px-4 py-3 transition-colors hover:bg-muted/40 cursor-pointer',
        isUnread && 'bg-muted/20'
      )}
      onClick={() => onClick(notification)}
      role="button"
      tabIndex={0}
      aria-label={`${notification.subject || 'Notification'} - ${notification.type}`}
    >
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <ChannelBadge channel={notification.type} size="sm" />
          <StatusBadge status={notification.status} size="sm" />
        </div>
        <p className={cn(
          'text-sm truncate',
          isUnread ? 'font-semibold' : 'font-medium text-muted-foreground'
        )}>
          {notification.subject || t('notifications.list.notification')}
        </p>
        <p className="text-xs text-muted-foreground truncate mt-0.5">
          {notification.body?.replace(/<[^>]*>/g, '').substring(0, 100)}
        </p>
        <div className="flex items-center gap-2 mt-1">
          <code className="text-[10px] text-muted-foreground font-mono">
            {shortId(notification.userId)}
          </code>
          <span className="text-[10px] text-muted-foreground">
            {formatRelativeTime(notification.createdAt, locale)}
          </span>
        </div>
      </div>
      {isUnread && (
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7 shrink-0"
          onClick={(e) => { e.stopPropagation(); onMarkRead(notification.id); }}
          aria-label={t('notifications.actions.mark_read')}
        >
          <MailOpen className="h-3.5 w-3.5" />
        </Button>
      )}
    </div>
  );
}
