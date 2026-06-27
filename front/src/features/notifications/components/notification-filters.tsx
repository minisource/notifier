'use client';

import { useTranslations } from 'next-intl';
import { SearchInput } from '@/components/shared/search-input';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { X } from 'lucide-react';
import { useParams } from 'next/navigation';
import type { NotificationChannel, NotificationStatus, NotificationPriority } from '../types';

interface NotificationFiltersProps {
  search: string;
  onSearchChange: (value: string) => void;
  statusFilter: string;
  onStatusChange: (value: string) => void;
  channelFilter: string;
  onChannelChange: (value: string) => void;
  priorityFilter: string;
  onPriorityChange: (value: string) => void;
  onClearFilters: () => void;
  hasActiveFilters: boolean;
}

const channels: NotificationChannel[] = ['sms', 'email', 'push', 'in_app', 'webhook'];
const statuses: NotificationStatus[] = ['pending', 'queued', 'processing', 'sent', 'failed', 'dead', 'cancelled'];
const priorities: NotificationPriority[] = ['low', 'normal', 'high', 'urgent'];

export function NotificationFilters({
  search, onSearchChange,
  statusFilter, onStatusChange,
  channelFilter, onChannelChange,
  priorityFilter, onPriorityChange,
  onClearFilters, hasActiveFilters,
}: NotificationFiltersProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2" dir={isRtl ? 'rtl' : 'ltr'}>
        <SearchInput
          value={search}
          onChange={onSearchChange}
          placeholder={t('notifications.list.search_placeholder')}
          className="w-full sm:w-64"
        />

        <Select value={statusFilter} onValueChange={onStatusChange}>
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder={t('notifications.filters.all_statuses')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.all')}</SelectItem>
            {statuses.map(s => (
              <SelectItem key={s} value={s}>{t(`statuses.${s}`)}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={channelFilter} onValueChange={onChannelChange}>
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder={t('notifications.filters.all_channels')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.all')}</SelectItem>
            {channels.map(c => (
              <SelectItem key={c} value={c}>{t(`channels.${c}`)}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={priorityFilter} onValueChange={onPriorityChange}>
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder={t('notifications.filters.all_priorities')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.all')}</SelectItem>
            {priorities.map(p => (
              <SelectItem key={p} value={p}>{t(`notifications.filters.priority_${p}`)}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        {hasActiveFilters && (
          <Button variant="ghost" size="sm" onClick={onClearFilters} className="gap-1">
            <X className="h-3.5 w-3.5" />
            {t('common.clear')}
          </Button>
        )}
      </div>

      {hasActiveFilters && (
        <div className="flex flex-wrap items-center gap-1.5" dir={isRtl ? 'rtl' : 'ltr'}>
          {statusFilter !== 'all' && (
            <span className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-medium">
              {t('common.status')}: {t(`statuses.${statusFilter}`)}
            </span>
          )}
          {channelFilter !== 'all' && (
            <span className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-medium">
              {t('common.type')}: {t(`channels.${channelFilter}`)}
            </span>
          )}
          {priorityFilter !== 'all' && (
            <span className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-medium">
              {t('notifications.priority')}: {t(`notifications.filters.priority_${priorityFilter}`)}
            </span>
          )}
          {search && (
            <span className="inline-flex items-center gap-1 rounded-md bg-muted px-2 py-0.5 text-xs font-medium">
              {t('common.search')}: &quot;{search}&quot;
            </span>
          )}
        </div>
      )}
    </div>
  );
}
