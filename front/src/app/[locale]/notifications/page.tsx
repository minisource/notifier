'use client';

import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { useState, useCallback, useEffect } from 'react';
import { PageHeader } from '@minisource/ui';
import { Card } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { ConfirmDialog } from '@minisource/ui';
import { NotificationTable } from '@/features/notifications/components/notification-table';
import { NotificationFilters } from '@/features/notifications/components/notification-filters';
import { Pagination } from '@minisource/ui';
import { EmptyState } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { Skeleton } from '@minisource/ui';
import { useNotifications, useRetryNotifications, useRetryAllFailed } from '@/features/notifications/hooks/use-notifications';
import { Plus, RefreshCw, Inbox, RotateCcw, X } from 'lucide-react';
import type { Notification } from '@/features/notifications/types';

const PAGE_SIZE = 20;

export default function NotificationsPage() {
  const t = useTranslations();
  const router = useRouter();

  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [channelFilter, setChannelFilter] = useState('all');
  const [priorityFilter, setPriorityFilter] = useState('all');
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [confirmAction, setConfirmAction] = useState<'selected' | 'all' | null>(null);

  const bulkRetryMutation = useRetryNotifications();
  const retryAllMutation = useRetryAllFailed();

  const queryParams = {
    page,
    pageSize: PAGE_SIZE,
    ...(search ? { search } : {}),
    ...(statusFilter !== 'all' ? { status: statusFilter as Notification['status'] } : {}),
    ...(channelFilter !== 'all' ? { type: channelFilter as Notification['type'] } : {}),
    ...(priorityFilter !== 'all' ? { priority: priorityFilter as Notification['priority'] } : {}),
  };

  const { data, isLoading, isError, error, refetch, isRefetching } = useNotifications(queryParams);

  // Clear selection whenever the visible page changes so the checkboxes always
  // reflect the currently loaded rows.
  useEffect(() => {
    setSelectedIds(new Set());
  }, [page, search, statusFilter, channelFilter, priorityFilter]);

  const toggleSelect = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const toggleSelectAll = useCallback(() => {
    setSelectedIds((prev) => {
      const rows = data?.data ?? [];
      const allSelected = rows.length > 0 && rows.every((n) => prev.has(n.id));
      const next = new Set(prev);
      if (allSelected) {
        rows.forEach((n) => next.delete(n.id));
      } else {
        rows.forEach((n) => next.add(n.id));
      }
      return next;
    });
  }, [data]);

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
      router.push(`/notifications/${notification.id}`);
    },
    [router]
  );

  const handleConfirmRetry = useCallback(() => {
    if (confirmAction === 'selected') {
      const ids = Array.from(selectedIds);
      bulkRetryMutation.mutate(ids);
    } else if (confirmAction === 'all') {
      retryAllMutation.mutate();
    }
    setConfirmAction(null);
    setSelectedIds(new Set());
  }, [confirmAction, selectedIds, bulkRetryMutation, retryAllMutation]);

  return (
    <div className="space-y-6">
      <PageHeader title={t('notifications.title')} description={t('notifications.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isRefetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isRefetching ? 'animate-spin' : ''}`} />
          {t('dashboard.view_all')}
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setConfirmAction('all')}
          disabled={retryAllMutation.isPending}
        >
          <RotateCcw className={`ml-1.5 h-4 w-4 ${retryAllMutation.isPending ? 'animate-spin' : ''}`} />
          {t('notifications.actions.retry_all_failed')}
        </Button>
        <Button size="sm" onClick={() => router.push(`/notifications/new`)}>
          <Plus className="ml-1.5 h-4 w-4" />
          {t('notifications.send')}
        </Button>
      </PageHeader>

      <Card title={t('notifications.list.all_notifications')}>
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
            <Skeleton className="h-64 w-full" />
          ) : isError ? (
            <ErrorState
              title={t('errors.generic')}
              message={(error as Error)?.message || t('errors.generic')}
              onRetry={() => refetch()}
              autoRetrySeconds={15}
            />
          ) : data && data.data.length > 0 ? (
            <>
              {selectedIds.size > 0 && (
                <div className="flex flex-wrap items-center gap-2 rounded-md border bg-muted/40 px-3 py-2">
                  <span className="text-sm font-medium">
                    {selectedIds.size} {t('notifications.list.selected')}
                  </span>
                  <Button
                    size="sm"
                    onClick={() => setConfirmAction('selected')}
                    disabled={bulkRetryMutation.isPending}
                  >
                    <RotateCcw className={`ml-1.5 h-4 w-4 ${bulkRetryMutation.isPending ? 'animate-spin' : ''}`} />
                    {t('notifications.actions.retry_selected')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedIds(new Set())}
                    disabled={bulkRetryMutation.isPending}
                  >
                    <X className="ml-1.5 h-4 w-4" />
                    {t('common.clear')}
                  </Button>
                </div>
              )}
              <NotificationTable
                notifications={data.data}
                loading={false}
                onView={handleView}
                selectedIds={selectedIds}
                onToggleSelect={toggleSelect}
                onToggleSelectAll={toggleSelectAll}
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
              onAction={() => hasActiveFilters ? clearFilters() : router.push(`/notifications/new`)}
            />
          )}
        </div>
      </Card>

      <ConfirmDialog
        open={confirmAction !== null}
        onOpenChange={(open) => { if (!open) setConfirmAction(null); }}
        onConfirm={handleConfirmRetry}
        title={confirmAction === 'all' ? t('notifications.actions.confirm_retry_all_title') : t('notifications.actions.confirm_retry_selected_title')}
        description={confirmAction === 'all' ? t('notifications.actions.confirm_retry_all_desc') : t('notifications.actions.confirm_retry_selected_desc')}
        confirmLabel={t('notifications.actions.retry')}
        cancelLabel={t('common.cancel')}
        destructive={false}
      />
    </div>
  );
}
