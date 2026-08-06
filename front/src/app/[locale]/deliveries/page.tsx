'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageHeader } from '@minisource/ui';
import { Card } from '@minisource/ui';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { EmptyState } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { Skeleton } from '@minisource/ui';
import { Pagination } from '@minisource/ui';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@minisource/ui';
import { Truck, RefreshCw, Eye, Mail, MessageSquare, Smartphone, Webhook, Server } from 'lucide-react';
import { useDeliveries } from '@/features/deliveries/hooks/use-deliveries';
import { shortId, maskEmail, maskPhone } from '@/lib/utils/format';
import { formatRelativeTime, formatDateTime } from '@/lib/utils/date';

const PAGE_SIZE = 20;
const STATUS_OPTIONS = ['all', 'delivered', 'failed', 'processing', 'dead', 'retrying'];
const PROVIDER_OPTIONS = ['all', 'smtp', 'kavenegar', 'fcm', 'sendgrid', 'apns'];

const providerIcons: Record<string, React.ElementType> = {
  smtp: Mail,
  sendgrid: Mail,
  kavenegar: MessageSquare,
  twilio: MessageSquare,
  fcm: Smartphone,
  apns: Smartphone,
  push: Smartphone,
  webhook: Webhook,
};

export default function DeliveriesPage() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'en';
  const isRtl = locale === 'fa';

  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('all');
  const [providerFilter, setProviderFilter] = useState('all');

  const { data, isLoading, isError, error, refetch, isFetching } = useDeliveries();

  // Client-side filter
  const allDeliveries = data || [];
  let filtered = [...allDeliveries];
  if (statusFilter !== 'all') filtered = filtered.filter(d => d.status === statusFilter);
  if (providerFilter !== 'all') filtered = filtered.filter(d => d.provider === providerFilter);
  const hasActiveFilters = statusFilter !== 'all' || providerFilter !== 'all';
  const totalPages = Math.ceil(filtered.length / PAGE_SIZE);
  const paged = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  const clearFilters = () => {
    setStatusFilter('all');
    setProviderFilter('all');
    setPage(1);
  };

  const getProviderIcon = (provider: string) => {
    const Icon = providerIcons[provider.toLowerCase()] || Server;
    return <Icon className="h-3.5 w-3.5 text-muted-foreground" />;
  };

  return (
    <div className="space-y-6">
      <PageHeader title={t('deliveries.title')} description={t('deliveries.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('dashboard.view_all') as string}
        </Button>
      </PageHeader>

      <Card title={t('deliveries.title')}>
        {isLoading ? (
          <Skeleton className="h-64 w-full" />
        ) : isError ? (
          <ErrorState
            title={t('errors.generic')}
            message={(error as Error)?.message || t('errors.generic')}
            onRetry={() => refetch()}
            autoRetrySeconds={15}
          />
        ) : (
          <div className="space-y-4">
            {/* Filters */}
            <div className="flex flex-wrap items-center gap-2" dir={isRtl ? 'rtl' : 'ltr'}>
              <Select value={statusFilter} onValueChange={(v: string) => { setStatusFilter(v); setPage(1); }}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder={t('common.all') as string} />
                </SelectTrigger>
                <SelectContent>
                  {STATUS_OPTIONS.map(s => (
                    <SelectItem key={s} value={s}>
                      {s === 'all' ? t('common.all') : t(`statuses.${s}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select value={providerFilter} onValueChange={(v: string) => { setProviderFilter(v); setPage(1); }}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder={t('deliveries.provider') as string} />
                </SelectTrigger>
                <SelectContent>
                  {PROVIDER_OPTIONS.map(p => (
                    <SelectItem key={p} value={p}>
                      {p === 'all' ? t('common.all') : p.charAt(0).toUpperCase() + p.slice(1)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Table */}
            {paged.length > 0 ? (
              <>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[200px]">{t('notifications.list.notification')}</TableHead>
                        <TableHead className="w-[140px]">{t('common.recipient')}</TableHead>
                        <TableHead className="w-[110px]">{t('deliveries.provider')}</TableHead>
                        <TableHead className="w-[80px]">{t('common.channel')}</TableHead>
                        <TableHead className="w-[100px]">{t('common.status')}</TableHead>
                        <TableHead className="w-[80px]">{t('deliveries.attempts')}</TableHead>
                        <TableHead className="w-[120px]">{t('deliveries.next_retry')}</TableHead>
                        <TableHead className="w-[160px]">{t('deliveries.last_error')}</TableHead>
                        <TableHead className="w-[110px]">{t('common.updated_at')}</TableHead>
                        <TableHead className="w-[48px]"></TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {paged.map((delivery) => (
                        <TableRow
                          key={delivery.id}
                          className="cursor-pointer"
                          onClick={() => router.push(`/deliveries/${delivery.id}`)}
                        >
                          <TableCell>
                            <div className="min-w-0">
                              <p className="truncate text-sm font-medium">{delivery.subject || t('notifications.list.notification')}</p>
                              <code className="text-xs font-mono text-muted-foreground">{shortId(delivery.notificationId)}</code>
                            </div>
                          </TableCell>
                          <TableCell>
                            {delivery.recipientEmail || delivery.recipientPhone || delivery.recipientId ? (
                              <span className="block truncate text-xs font-mono text-muted-foreground">
                                {delivery.recipientEmail
                                  ? maskEmail(delivery.recipientEmail)
                                  : delivery.recipientPhone
                                    ? maskPhone(delivery.recipientPhone)
                                    : shortId(delivery.recipientId ?? '')}
                              </span>
                            ) : (
                              <span className="text-sm text-muted-foreground">—</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <span className="inline-flex items-center gap-1.5 text-sm capitalize">
                              {getProviderIcon(delivery.provider)}
                              {delivery.provider || '—'}
                            </span>
                          </TableCell>
                          <TableCell>
                            <ChannelBadge channel={delivery.channel} size="sm" />
                          </TableCell>
                          <TableCell>
                            <StatusBadge status={delivery.status} size="sm" />
                          </TableCell>
                          <TableCell>
                            <span className="text-sm text-muted-foreground">
                              {delivery.attemptCount}/{delivery.maxAttempts}
                            </span>
                          </TableCell>
                          <TableCell>
                            {delivery.nextRetryAt ? (
                              <span className="text-xs text-muted-foreground whitespace-nowrap" title={formatDateTime(delivery.nextRetryAt, locale)}>
                                {formatRelativeTime(delivery.nextRetryAt, locale)}
                              </span>
                            ) : (
                              <span className="text-sm text-muted-foreground">—</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {delivery.lastError ? (
                              <span className="block truncate text-xs text-muted-foreground" title={delivery.lastError}>
                                {delivery.lastError}
                              </span>
                            ) : (
                              <span className="text-sm text-muted-foreground">—</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <span className="text-xs text-muted-foreground whitespace-nowrap" title={formatDateTime(delivery.updatedAt, locale)}>
                              {formatRelativeTime(delivery.updatedAt, locale)}
                            </span>
                          </TableCell>
                          <TableCell onClick={(e: React.MouseEvent) => e.stopPropagation()}>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              title={t('common.view_details') as string}
                              onClick={() => router.push(`/deliveries/${delivery.id}`)}
                            >
                              <Eye className="h-3.5 w-3.5" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
                <Pagination
                  page={page}
                  totalPages={totalPages}
                  onPageChange={setPage}
                  className="pt-2"
                />
              </>
            ) : (
              <EmptyState
                icon={Truck}
                title={t('deliveries.no_deliveries')}
                description={hasActiveFilters
                  ? 'No deliveries match the selected filters.'
                  : 'Delivery logs show the status and history of each notification attempt. They appear automatically when notifications are sent.'}
                actionLabel={hasActiveFilters ? t('common.clear') : undefined}
                onAction={hasActiveFilters ? clearFilters : undefined}
                tips={hasActiveFilters
                  ? ['Clear the filters to see all deliveries']
                  : [
                      'Send a notification to see its delivery logs',
                      'Filter by status to find failed or retrying deliveries',
                      'Click on a delivery to view detailed attempt history',
                    ]}
              />
            )}
          </div>
        )}
      </Card>
    </div>
  );
}
