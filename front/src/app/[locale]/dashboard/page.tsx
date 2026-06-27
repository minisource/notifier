'use client';

import { useTranslations } from 'next-intl';
import { useEffect, useState } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { AutoRefreshControl } from '@/shared/components/auto-refresh-control';
import { ErrorState } from '@/components/shared/error-state';
import { PageSkeleton } from '@/components/shared/loading-state';
import { RoleGuard } from '@/shared/components/role-guard';
import { useDashboard } from '@/features/dashboard/hooks/use-dashboard';
import { DashboardStatusStrip } from '@/features/dashboard/components/dashboard-status-strip';
import { DashboardMetricGrid } from '@/features/dashboard/components/dashboard-metric-grid';
import { QueueStatusPanel } from '@/features/dashboard/components/queue-status-panel';
import { ChannelBreakdownPanel } from '@/features/dashboard/components/channel-breakdown-panel';
import { ProviderHealthPanel } from '@/features/dashboard/components/provider-health-panel';
import { RecentNotificationsPanel } from '@/features/dashboard/components/recent-notifications-panel';
import { DashboardTrendChart } from '@/features/dashboard/components/dashboard-trend-chart';
import { DashboardRecentFailures } from '@/features/dashboard/components/dashboard-recent-failures';
import type { DailyTrendItem, RecentFailure } from '@/features/notifier/api/notifier-types';

export default function DashboardPage() {
  const t = useTranslations();
  const { data, isLoading, isError, error, refetch, isFetching, dataUpdatedAt } = useDashboard();
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Extract trend and failure data from the dashboard response
  // (useDashboard returns DashboardData which includes these fields)
  const trendData: DailyTrendItem[] = (data as any)?.metrics?.dailyTrend ?? [];
  const failureData: RecentFailure[] = (data as any)?.recentFailures ?? [];

  useEffect(() => {
    if (!autoRefresh) return;
    const interval = setInterval(() => { refetch(); }, 30000);
    return () => clearInterval(interval);
  }, [refetch, autoRefresh]);

  const lastUpdated = dataUpdatedAt
    ? new Date(dataUpdatedAt).toLocaleTimeString()
    : undefined;

  if (isLoading) return <PageSkeleton context="dashboard" layout="cards" cards={8} />;

  if (isError || !data) {
    return (
      <RoleGuard>
        <PageHeader title={t('dashboard.title')} subtitle={t('dashboard.subtitle')} />
        <ErrorState
          message={error instanceof Error ? error.message : t('common.error_occurred')}
          onRetry={() => refetch()}
        />
      </RoleGuard>
    );
  }

  const { metrics, recentNotifications } = data;

  return (
    <RoleGuard>
      <div className="space-y-5">
        <PageHeader title={t('dashboard.title')} subtitle={t('dashboard.subtitle')}>
          <AutoRefreshControl
            isRefreshing={isFetching}
            onRefresh={() => refetch()}
            lastUpdated={lastUpdated}
            autoRefreshEnabled={autoRefresh}
            onToggleAutoRefresh={setAutoRefresh}
            intervalSeconds={30}
          />
        </PageHeader>

        <DashboardStatusStrip data={data} />
        <DashboardMetricGrid metrics={metrics} />

        <QueueStatusPanel
          queued={metrics.queued}
          deadLetter={metrics.deadLetter}
          activeReminders={metrics.activeReminders}
          queueDepth={metrics.queueDepth}
          avgDeliveryTimeMs={metrics.avgDeliveryTimeMs}
        />

        <div className="grid gap-5 lg:grid-cols-2">
          <DashboardTrendChart data={trendData} />
          <ChannelBreakdownPanel breakdown={metrics.channelBreakdown} />
        </div>

        <div className="grid gap-5 lg:grid-cols-2">
          <ProviderHealthPanel />
          <DashboardRecentFailures failures={failureData} deadLetters={[]} />
        </div>

        <RecentNotificationsPanel notifications={recentNotifications} />
      </div>
    </RoleGuard>
  );
}
