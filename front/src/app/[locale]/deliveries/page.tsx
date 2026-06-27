'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SectionCard } from '@/components/shared/section-card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { EmptyState } from '@/components/shared/empty-state';
import { ErrorState } from '@/components/shared/error-state';
import { TableSkeleton } from '@/components/shared/loading-state';
import { Pagination } from '@/components/shared/pagination';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Truck, RefreshCw } from 'lucide-react';
import { useDeliveries } from '@/features/deliveries/hooks/use-deliveries';
import { shortId } from '@/lib/utils/format';

const PAGE_SIZE = 20;
const STATUS_OPTIONS = ['all', 'delivered', 'failed', 'processing', 'dead', 'retrying'];
const PROVIDER_OPTIONS = ['all', 'smtp', 'kavenegar', 'fcm', 'sendgrid', 'apns'];

export default function DeliveriesPage() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';

  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('all');
  const [providerFilter, setProviderFilter] = useState('all');

  const { data, isLoading, isError, error, refetch, isFetching } = useDeliveries();

  // Client-side filter since mock data doesn't support server-side
  const allDeliveries = data || [];
  let filtered = [...allDeliveries];
  if (statusFilter !== 'all') filtered = filtered.filter(d => d.status === statusFilter);
  if (providerFilter !== 'all') filtered = filtered.filter(d => d.provider === providerFilter);
  const totalPages = Math.ceil(filtered.length / PAGE_SIZE);
  const paged = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  return (
    <PageContainer>
      <PageHeader title={t('deliveries.title')} subtitle={t('deliveries.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('dashboard.view_all') as string}
        </Button>
      </PageHeader>

      <SectionCard title={t('deliveries.title')}>
        {isLoading ? (
          <TableSkeleton rows={8} columns={6} context="deliveries" />
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
              <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v); setPage(1); }}>
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

              <Select value={providerFilter} onValueChange={(v) => { setProviderFilter(v); setPage(1); }}>
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
                        <TableHead className="w-[180px]">{t('notifications.list.notification')}</TableHead>
                        <TableHead className="w-[100px]">{t('deliveries.provider')}</TableHead>
                        <TableHead className="w-[80px]">{t('common.channel')}</TableHead>
                        <TableHead className="w-[100px]">{t('common.status')}</TableHead>
                        <TableHead className="w-[80px]">{t('deliveries.attempts')}</TableHead>
                        <TableHead className="w-[100px]">{t('deliveries.last_error')}</TableHead>
                        <TableHead className="w-[48px]"></TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {paged.map((delivery) => (
                        <TableRow
                          key={delivery.id}
                          className="cursor-pointer"
                          onClick={() => router.push(`/${locale}/deliveries/${delivery.id}`)}
                        >
                          <TableCell>
                            <code className="text-xs font-mono">{shortId(delivery.notificationId)}</code>
                          </TableCell>
                          <TableCell>
                            <span className="text-sm capitalize">{delivery.provider}</span>
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
                            <span className="text-sm text-muted-foreground truncate block max-w-[200px]">
                              {delivery.lastError || '—'}
                            </span>
                          </TableCell>
                          <TableCell onClick={(e) => e.stopPropagation()}>
                            <Button variant="ghost" size="sm" onClick={() => router.push(`/${locale}/deliveries/${delivery.id}`)}>
                              {t('common.view_details') as string} →
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
                description="Delivery logs show the status and history of each notification attempt. They appear automatically when notifications are sent."
                tips={[
                  'Send a notification to see its delivery logs',
                  'Filter by status to find failed or retrying deliveries',
                  'Click on a delivery to view detailed attempt history',
                ]}
              />
            )}
          </div>
        )}
      </SectionCard>
    </PageContainer>
  );
}
