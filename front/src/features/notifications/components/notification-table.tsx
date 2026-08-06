'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@minisource/ui';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { NotificationActionMenu } from './notification-action-menu';
import { NotificationDetailRow } from './notification-detail-row';
import { formatRelativeTime, formatDateTime } from '@/lib/utils/date';
import { maskEmail, maskPhone, shortId, truncate } from '@/lib/utils/format';
import { AlertTriangle, Hash, ChevronDown, ChevronRight } from 'lucide-react';
import type { Notification } from '../types';

interface NotificationTableProps {
  notifications: Notification[];
  loading?: boolean;
  onView?: (notification: Notification) => void;
  /** IDs currently selected (checkbox column). When provided, selection UI is shown. */
  selectedIds?: Set<string>;
  onToggleSelect?: (id: string) => void;
  onToggleSelectAll?: () => void;
}

function getRecipientDisplay(n: Notification): { value: string; type: string } | null {
  if (n.recipientEmail) return { value: maskEmail(n.recipientEmail), type: 'email' };
  if (n.recipientPhone) return { value: maskPhone(n.recipientPhone), type: 'phone' };
  if (n.recipientId) return { value: shortId(n.recipientId), type: 'id' };
  return null;
}

export function NotificationTable({
  notifications, loading,
  selectedIds, onToggleSelect, onToggleSelectAll,
}: NotificationTableProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'en';
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  const selectable = !!selectedIds && !!onToggleSelect;
  const allSelected = selectable && notifications.length > 0 && notifications.every((n) => selectedIds!.has(n.id));
  const colSpan = selectable ? 8 : 7;

  const toggleExpand = (id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

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
            {selectable && (
              <TableHead className="w-[36px]">
                <input
                  type="checkbox"
                  checked={allSelected}
                  onChange={onToggleSelectAll}
                  aria-label={t('notifications.list.select_all')}
                  className="h-4 w-4 cursor-pointer rounded border-gray-300 accent-primary"
                />
              </TableHead>
            )}
            <TableHead className="w-[40px]"></TableHead>
            <TableHead className="w-[280px]">{t('notifications.list.notification')}</TableHead>
            <TableHead className="w-[120px]">{t('common.channel')}</TableHead>
            <TableHead className="w-[160px]">{t('notifications.recipient')}</TableHead>
            <TableHead className="w-[110px]">{t('common.status')}</TableHead>
            <TableHead className="w-[140px]">{t('notifications.list.last_activity')}</TableHead>
            <TableHead className="w-[48px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {notifications.length === 0 ? (
            <TableRow>
              <TableCell colSpan={colSpan} className="text-center py-16 text-muted-foreground">
                {t('notifications.list.empty_state')}
              </TableCell>
            </TableRow>
          ) : (
            notifications.map((n) => {
              const recipient = getRecipientDisplay(n);
              const isExpanded = expandedIds.has(n.id);
              return (
                <>
                  <TableRow
                    key={n.id}
                    className="cursor-pointer group"
                    onClick={() => toggleExpand(n.id)}
                    data-state={isExpanded ? 'expanded' : undefined}
                  >
                    {selectable && (
                      <TableCell className="w-[36px] pr-0" onClick={(e) => e.stopPropagation()}>
                        <input
                          type="checkbox"
                          checked={selectedIds!.has(n.id)}
                          onChange={() => onToggleSelect!(n.id)}
                          aria-label={t('notifications.list.select_row')}
                          className="h-4 w-4 cursor-pointer rounded border-gray-300 accent-primary"
                        />
                      </TableCell>
                    )}
                    <TableCell className="w-[40px] pr-0">
                      <div className="flex items-center justify-center">
                        {isExpanded ? (
                          <ChevronDown className="h-4 w-4 text-muted-foreground transition-transform" />
                        ) : (
                          <ChevronRight className="h-4 w-4 text-muted-foreground transition-transform group-hover:text-foreground" />
                        )}
                      </div>
                    </TableCell>
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
                      <div className="text-sm text-muted-foreground whitespace-nowrap">
                        <span>{formatRelativeTime(n.createdAt, locale)}</span>
                        {n.sentAt && (
                          <span className="block text-[10px] text-muted-foreground/70">
                            {t('notifications.sent_at')}: {formatDateTime(n.sentAt, locale)}
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell onClick={(e) => e.stopPropagation()}>
                      <NotificationActionMenu
                        notification={n}
                        showView={false}
                      />
                    </TableCell>
                  </TableRow>
                  {isExpanded && (
                    <TableRow key={`${n.id}-detail`} className="hover:bg-transparent">
                      <TableCell colSpan={colSpan} className="p-0">
                        <NotificationDetailRow notification={n} />
                      </TableCell>
                    </TableRow>
                  )}
                </>
              );
            })
          )}
        </TableBody>
      </Table>
    </div>
  );
}
