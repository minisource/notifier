'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { PageHeader, Card, Table, TableBody, TableCell, TableHead, TableHeader, TableRow, Button, EmptyState, ErrorState, Skeleton, Pagination, Select, SelectContent, SelectItem, SelectTrigger, SelectValue, Input } from '@minisource/ui';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { RefreshCw, Eye, Activity, Search } from 'lucide-react';
import { useAdminProviderAttempts } from '@/features/notifier/api/notifier-queries';
import { shortId } from '@/lib/utils/format';
import { formatRelativeTime, formatDateTime } from '@/lib/utils/date';
import { formatMilliseconds } from '@/lib/utils/format';

const PAGE_SIZE = 20;

const ATTEMPT_STATUSES = ['all', 'queued', 'preparing', 'sending', 'accepted', 'pending', 'delivered', 'failed', 'rejected', 'timed_out', 'cancelled', 'bounced', 'complained', 'unknown'];
const CHANNELS = ['all', 'sms', 'email', 'push', 'in_app'];
const PROVIDERS = ['all', 'kavenegar', 'twilio', 'smtp', 'sendgrid', 'fcm', 'mock', 'tencent', 'huawei', 'infobip', 'msg91', 'netgsm', 'oson', 'smsbao', 'submail'];

export default function ProviderLogsPage() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const locale = (params?.locale as string) || 'en';

  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('all');
  const [channelFilter, setChannelFilter] = useState('all');
  const [providerFilter, setProviderFilter] = useState('all');
  const [notificationFilter, setNotificationFilter] = useState('');
  const [searchInput, setSearchInput] = useState('');

  // Support deep-linking with ?notificationId=… (e.g. from the error-investigation
  // link on a failed Kavenegar send) — prefill the filter once on mount.
  useEffect(() => {
    const nid = searchParams.get('notificationId');
    if (nid) {
      setNotificationFilter(nid);
      setSearchInput(nid);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const { data, isLoading, isError, error, refetch, isFetching } = useAdminProviderAttempts({
    page,
    pageSize: PAGE_SIZE,
    ...(statusFilter !== 'all' ? { status: statusFilter } : {}),
    ...(channelFilter !== 'all' ? { channel: channelFilter } : {}),
    ...(providerFilter !== 'all' ? { provider: providerFilter } : {}),
    ...(notificationFilter ? { notificationId: notificationFilter } : {}),
  });

  const hasActiveFilters = statusFilter !== 'all' || channelFilter !== 'all' || providerFilter !== 'all' || notificationFilter !== '';

  const clearFilters = () => {
    setStatusFilter('all');
    setChannelFilter('all');
    setProviderFilter('all');
    setNotificationFilter('');
    setSearchInput('');
    setPage(1);
  };

  const submitSearch = () => {
    setNotificationFilter(searchInput.trim());
    setPage(1);
  };

  const totalPages = data?.totalPages ?? 1;

  return (
    <div className="space-y-6">
      <PageHeader title={t('providerLogs.title')} description={t('providerLogs.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('providerLogs.actions.refresh')}
        </Button>
      </PageHeader>

      <Card title={t('providerLogs.title')}>
        {isLoading ? (
          <Skeleton className="h-64 w-full" />
        ) : isError ? (
          <ErrorState
            title={t('errors.generic')}
            message={(error as Error)?.message || t('errors.generic')}
            onRetry={() => refetch()}
          />
        ) : (
          <div className="space-y-4">
            {/* Filters */}
            <div className="flex flex-wrap items-center gap-2">
              <Select value={statusFilter} onValueChange={(v: string) => { setStatusFilter(v); setPage(1); }}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder={t('providerLogs.filters.all_statuses') as string} />
                </SelectTrigger>
                <SelectContent>
                  {ATTEMPT_STATUSES.map(s => (
                    <SelectItem key={s} value={s}>
                      {s === 'all' ? t('providerLogs.filters.all_statuses') : t(`providerLogs.status.${s}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select value={channelFilter} onValueChange={(v: string) => { setChannelFilter(v); setPage(1); }}>
                <SelectTrigger className="w-[130px]">
                  <SelectValue placeholder={t('providerLogs.filters.all_channels') as string} />
                </SelectTrigger>
                <SelectContent>
                  {CHANNELS.map(c => (
                    <SelectItem key={c} value={c}>
                      {c === 'all' ? t('providerLogs.filters.all_channels') : t(`channels.${c}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select value={providerFilter} onValueChange={(v: string) => { setProviderFilter(v); setPage(1); }}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder={t('providerLogs.filters.all_providers') as string} />
                </SelectTrigger>
                <SelectContent>
                  {PROVIDERS.map(p => (
                    <SelectItem key={p} value={p}>
                      {p === 'all' ? t('providerLogs.filters.all_providers') : p.charAt(0).toUpperCase() + p.slice(1)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <div className="flex items-center gap-2">
                <Input
                  className="w-[220px]"
                  placeholder={t('providerLogs.search_placeholder') as string}
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') submitSearch(); }}
                />
                <Button variant="outline" size="icon" onClick={submitSearch} title={t('common.search') as string}>
                  <Search className="h-3.5 w-3.5" />
                </Button>
              </div>

              {hasActiveFilters && (
                <Button variant="ghost" size="sm" onClick={clearFilters}>
                  {t('common.clear')}
                </Button>
              )}
            </div>

            {/* Table */}
            {data && data.items.length > 0 ? (
              <>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[160px]">{t('common.created_at')}</TableHead>
                        <TableHead className="w-[110px]">{t('providerLogs.attempt')}</TableHead>
                        <TableHead className="w-[90px]">{t('common.channel')}</TableHead>
                        <TableHead className="w-[120px]">{t('providerLogs.provider')}</TableHead>
                        <TableHead className="w-[150px]">{t('providerLogs.recipient')}</TableHead>
                        <TableHead className="w-[110px]">{t('providerLogs.status')}</TableHead>
                        <TableHead className="w-[90px]">{t('providerLogs.duration')}</TableHead>
                        <TableHead className="w-[90px]">{t('providerLogs.http_status')}</TableHead>
                        <TableHead className="w-[48px]"></TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {data.items.map((a) => (
                        <TableRow
                          key={a.id}
                          className="cursor-pointer"
                          onClick={() => router.push(`/provider-logs/${a.id}`)}
                        >
                          <TableCell>
                            <span className="block text-xs text-muted-foreground whitespace-nowrap" title={formatDateTime(a.createdAt, locale)}>
                              {formatRelativeTime(a.createdAt, locale)}
                            </span>
                            <code className="text-[10px] font-mono text-muted-foreground/70">{shortId(a.id)}</code>
                          </TableCell>
                          <TableCell>
                            <span className="text-sm text-muted-foreground">
                              #{a.attemptNumber}
                              {a.fallbackSequence > 0 ? ` +${a.fallbackSequence}` : ''}
                            </span>
                          </TableCell>
                          <TableCell>
                            <ChannelBadge channel={a.channel} size="sm" />
                          </TableCell>
                          <TableCell>
                            <span className="inline-flex items-center gap-1.5 text-sm capitalize">
                              <Activity className="h-3.5 w-3.5 text-muted-foreground" />
                              {a.provider || '—'}
                            </span>
                          </TableCell>
                          <TableCell>
                            {a.recipientMasked ? (
                              <span className="block truncate text-xs font-mono text-muted-foreground">{a.recipientMasked}</span>
                            ) : (
                              <span className="text-sm text-muted-foreground">—</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <StatusBadge status={a.status} size="sm" />
                          </TableCell>
                          <TableCell>
                            <span className="text-sm text-muted-foreground">{a.durationMs ? formatMilliseconds(a.durationMs) : '—'}</span>
                          </TableCell>
                          <TableCell>
                            {a.responseStatusCode ? (
                              <span className="text-xs font-mono text-muted-foreground">{a.responseStatusCode}</span>
                            ) : (
                              <span className="text-sm text-muted-foreground">—</span>
                            )}
                          </TableCell>
                          <TableCell onClick={(e: React.MouseEvent) => e.stopPropagation()}>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              title={t('common.view_details') as string}
                              onClick={() => router.push(`/provider-logs/${a.id}`)}
                            >
                              <Eye className="h-3.5 w-3.5" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
                <Pagination page={page} totalPages={totalPages} onPageChange={setPage} className="pt-2" />
              </>
            ) : (
              <EmptyState
                icon={Activity}
                title={hasActiveFilters ? t('providerLogs.no_results') : t('providerLogs.no_attempts')}
                description={hasActiveFilters ? undefined : t('providerLogs.no_attempts_desc')}
                actionLabel={hasActiveFilters ? t('common.clear') : undefined}
                onAction={hasActiveFilters ? clearFilters : undefined}
              />
            )}
          </div>
        )}
      </Card>
    </div>
  );
}
