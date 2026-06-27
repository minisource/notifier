'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState, useCallback } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SectionCard } from '@/components/shared/section-card';
import { Button } from '@/components/ui/button';
import { NotificationTable } from '@/features/notifications/components/notification-table';
import { NotificationFilters } from '@/features/notifications/components/notification-filters';
import { Pagination } from '@/components/shared/pagination';
import { EmptyState } from '@/components/shared/empty-state';
import { ErrorState } from '@/components/shared/error-state';
import { TableSkeleton } from '@/components/shared/loading-state';
import { useNotifications } from '@/features/notifications/hooks/use-notifications';
import { Plus, RefreshCw, Inbox } from 'lucide-react';
import type { Notification } from '@/features/notifications/types';

const PAGE_SIZE = 20;

export default function NotificationsPage() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';

  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [channelFilter, setChannelFilter] = useState('all');
  const [priorityFilter, setPriorityFilter] = useState('all');

  const queryParams = {
    page,
    pageSize: PAGE_SIZE,
    ...(search ? { search } : {}),
    ...(statusFilter !== 'all' ? { status: statusFilter as Notification['status'] } : {}),
    ...(channelFilter !== 'all' ? { type: channelFilter as Notification['type'] } : {}),
    ...(priorityFilter !== 'all' ? { priority: priorityFilter as Notification['priority'] } : {}),
  };

  const { data, isLoading, isError, error, refetch, isRefetching } = useNotifications(queryParams);

  const hasActiveFilters = search !== '' || statusFilter !== 'all' || channelFilter !== 'all' || priorityFilter !== 'all';

  const clearFilters = useCallback(() => {
    setSearch('');
    setStatusFilter('all');
    setChannelFilter('all');
    setPriorityFilter('all');
    setPage(1);
  }, []);

  const handleView = useCallback(
    (notification: Notification) => {
      router.push(`/${locale}/notifications/${notification.id}`);
    },
    [locale, router]
  );

  return (
    <PageContainer>
      <PageHeader title={t('notifications.title')} subtitle={t('notifications.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isRefetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isRefetching ? 'animate-spin' : ''}`} />
          {t('dashboard.view_all')}
        </Button>
        <Button size="sm" onClick={() => router.push(`/${locale}/notifications/new`)}>
          <Plus className="ml-1.5 h-4 w-4" />
          {t('notifications.send')}
        </Button>
      </PageHeader>

      <SectionCard title={t('notifications.list.all_notifications')}>
        <div className="space-y-4">
          <NotificationFilters
            search={search}
            onSearchChange={(v) => { setSearch(v); setPage(1); }}
            statusFilter={statusFilter}
            onStatusChange={(v) => { setStatusFilter(v); setPage(1); }}
            channelFilter={channelFilter}
            onChannelChange={(v) => { setChannelFilter(v); setPage(1); }}
            priorityFilter={priorityFilter}
            onPriorityChange={(v) => { setPriorityFilter(v); setPage(1); }}
            onClearFilters={clearFilters}
            hasActiveFilters={hasActiveFilters}
          />

          {isLoading ? (
            <TableSkeleton rows={8} columns={6} context="notifications" />
          ) : isError ? (
            <ErrorState
              title={t('errors.generic')}
              message={(error as Error)?.message || t('errors.generic')}
              onRetry={() => refetch()}
              autoRetrySeconds={15}
            />
          ) : data && data.data.length > 0 ? (
            <>
              <NotificationTable
                notifications={data.data}
                loading={false}
                onView={handleView}
              />
              <Pagination
                page={data.page}
                totalPages={data.totalPages}
                onPageChange={setPage}
                className="pt-2"
              />
              <div className="text-center text-xs text-muted-foreground">
                {data.total} {t('common.total')} · {t('common.page')} {data.page} {t('common.of')} {data.totalPages}
              </div>
            </>
          ) : (
            <EmptyState
              icon={Inbox}
              title={t('notifications.list.empty_state')}
              description={hasActiveFilters ? t('notifications.list.no_results') : t('notifications.list.no_notifications_yet')}
              actionLabel={hasActiveFilters ? t('common.clear') : t('notifications.send')}
              onAction={() => hasActiveFilters ? clearFilters() : router.push(`/${locale}/notifications/new`)}
            />
          )}
        </div>
      </SectionCard>
    </PageContainer>
  );
}
