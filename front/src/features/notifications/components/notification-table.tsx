'use client';

import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { NotificationActionMenu } from './notification-action-menu';
import { formatRelativeTime } from '@/lib/utils/date';
import { maskEmail, maskPhone, shortId, truncate } from '@/lib/utils/format';
import { AlertTriangle, Hash } from 'lucide-react';
import type { Notification } from '../types';

interface NotificationTableProps {
  notifications: Notification[];
  loading?: boolean;
  onView?: (notification: Notification) => void;
}

function getRecipientDisplay(n: Notification): { value: string; type: string } | null {
  if (n.recipientEmail) return { value: maskEmail(n.recipientEmail), type: 'email' };
  if (n.recipientPhone) return { value: maskPhone(n.recipientPhone), type: 'phone' };
  if (n.recipientId) return { value: shortId(n.recipientId), type: 'id' };
  return null;
}

export function NotificationTable({ notifications, loading, onView }: NotificationTableProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';

  if (loading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="h-12 animate-pulse rounded-md bg-muted" />
        ))}
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[280px]">{t('notifications.list.notification')}</TableHead>
            <TableHead className="w-[120px]">{t('common.channel')}</TableHead>
            <TableHead className="w-[160px]">{t('notifications.recipient')}</TableHead>
            <TableHead className="w-[110px]">{t('common.status')}</TableHead>
            <TableHead className="w-[110px]">{t('notifications.list.last_activity')}</TableHead>
            <TableHead className="w-[48px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {notifications.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center py-16 text-muted-foreground">
                {t('notifications.list.empty_state')}
              </TableCell>
            </TableRow>
          ) : (
            notifications.map((n) => {
              const recipient = getRecipientDisplay(n);
              return (
                <TableRow
                  key={n.id}
                  className="cursor-pointer group"
                  onClick={() => onView?.(n)}
                >
                  <TableCell>
                    <div className="flex items-start gap-2">
                      {n.priority === 'urgent' && (
                        <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-red-500" />
                      )}
                      <div className="min-w-0 space-y-0.5">
                        <div className="flex items-center gap-1.5">
                          {n.priority === 'high' && (
                            <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-amber-500" />
                          )}
                          <span className="text-sm font-medium truncate block">
                            {n.subject || truncate(n.body, 40)}
                          </span>
                        </div>
                        <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                          <Hash className="h-3 w-3" />
                          <code className="font-mono text-[10px]">{shortId(n.id)}</code>
                        </div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <ChannelBadge channel={n.type} size="sm" showIcon />
                  </TableCell>
                  <TableCell>
                    {recipient ? (
                      <span className="text-sm text-muted-foreground font-mono text-[11px]">
                        {recipient.value}
                      </span>
                    ) : (
                      <span className="text-sm text-muted-foreground">—</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={n.status} size="sm" />
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-muted-foreground whitespace-nowrap">
                      {formatRelativeTime(n.createdAt, locale)}
                    </span>
                  </TableCell>
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    <NotificationActionMenu
                      notification={n}
                      showView={false}
                    />
                  </TableCell>
                </TableRow>
              );
            })
          )}
        </TableBody>
      </Table>
    </div>
  );
}
