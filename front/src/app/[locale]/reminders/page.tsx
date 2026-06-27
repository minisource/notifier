'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SectionCard } from '@/components/shared/section-card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { StatusBadge } from '@/components/shared/status-badge';
import { EmptyState } from '@/components/shared/empty-state';
import { ErrorState } from '@/components/shared/error-state';
import { TableSkeleton } from '@/components/shared/loading-state';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useReminders } from '@/features/reminders/hooks/use-reminders';
import { Plus, Clock, RefreshCw } from 'lucide-react';
import { formatRelativeTime } from '@/lib/utils/date';
import { maskEmail, maskPhone } from '@/lib/utils/format';

export default function RemindersPage() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';

  const [statusFilter, setStatusFilter] = useState('all');

  const { data: reminders, isLoading, isError, error, refetch, isFetching } = useReminders();

  const filtered = reminders
    ? statusFilter === 'all'
      ? reminders
      : reminders.filter(r => r.status === statusFilter)
    : [];

  return (
    <PageContainer>
      <PageHeader title={t('reminders.title')} subtitle={t('reminders.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('dashboard.view_all') as string}
        </Button>
        <Button size="sm" onClick={() => router.push(`/${locale}/reminders/new`)}>
          <Plus className="ml-1.5 h-4 w-4" />
          {t('reminders.schedule')}
        </Button>
      </PageHeader>

      <SectionCard title={t('reminders.title')}>
        {isLoading ? (
          <TableSkeleton rows={5} columns={5} context="reminders" />
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
            <div className="flex flex-wrap items-center gap-2">
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder={t('common.all') as string} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('common.all')}</SelectItem>
                  <SelectItem value="scheduled">{t('statuses.scheduled')}</SelectItem>
                  <SelectItem value="sent">{t('statuses.sent')}</SelectItem>
                  <SelectItem value="cancelled">{t('statuses.cancelled')}</SelectItem>
                  <SelectItem value="failed">{t('statuses.failed')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Table */}
            {filtered.length > 0 ? (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[120px]">{t('reminders.scheduled_at')}</TableHead>
                      <TableHead className="w-[120px]">{t('common.channel')}</TableHead>
                      <TableHead className="w-[180px]">{t('reminders.recipient')}</TableHead>
                      <TableHead className="w-[100px]">{t('common.status')}</TableHead>
                      <TableHead className="w-[140px]">{t('common.created_at')}</TableHead>
                      <TableHead className="w-[48px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filtered.map((reminder) => (
                      <TableRow
                        key={reminder.id}
                        className="cursor-pointer"
                        onClick={() => router.push(`/${locale}/reminders/${reminder.id}`)}
                      >
                        <TableCell>
                          <span className="text-sm whitespace-nowrap">
                            {formatRelativeTime(reminder.scheduledAt, locale)}
                          </span>
                        </TableCell>
                        <TableCell>
                          <ChannelBadge channel={reminder.type as any} size="sm" />
                        </TableCell>
                        <TableCell>
                          <span className="text-sm text-muted-foreground">
                            {reminder.recipientEmail ? maskEmail(reminder.recipientEmail) :
                             reminder.recipientPhone ? maskPhone(reminder.recipientPhone) :
                             reminder.userId}
                          </span>
                        </TableCell>
                        <TableCell>
                          <StatusBadge status={reminder.status} size="sm" />
                        </TableCell>
                        <TableCell>
                          <span className="text-sm text-muted-foreground whitespace-nowrap">
                            {formatRelativeTime(reminder.createdAt, locale)}
                          </span>
                        </TableCell>
                        <TableCell onClick={(e) => e.stopPropagation()}>
                          <Button variant="ghost" size="sm" onClick={() => router.push(`/${locale}/reminders/${reminder.id}`)}>
                            {t('common.view_details') as string} →
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <EmptyState
                icon={Clock}
                title={t('reminders.no_reminders')}
                description="Schedule reminders for time-sensitive notifications that need to be sent at a specific date and time."
                actionLabel={t('reminders.schedule')}
                onAction={() => router.push(`/${locale}/reminders/new`)}
                tips={[
                  'Reminders can be sent via SMS, Email, or Push',
                  'Use templates for consistent message formatting',
                  'Cancel or reschedule reminders anytime',
                ]}
              />
            )}
          </div>
        )}
      </SectionCard>
    </PageContainer>
  );
}
