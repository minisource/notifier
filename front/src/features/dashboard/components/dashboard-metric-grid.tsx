'use client';

import { useTranslations } from 'next-intl';
import { Bell, TrendingUp, AlertTriangle, BarChart3 } from 'lucide-react';
import { MetricCard } from '@/components/shared/metric-card';
import type { DashboardMetrics } from '@/features/dashboard/types';

interface MetricGridProps {
  metrics: DashboardMetrics;
}

export function DashboardMetricGrid({ metrics }: MetricGridProps) {
  const t = useTranslations();

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <MetricCard
        title={t('dashboard.total_notifications')}
        value={metrics.totalNotifications.toLocaleString()}
        icon={Bell}
        variant="default"
      />
      <MetricCard
        title={t('dashboard.sent_today')}
        value={metrics.sentToday.toLocaleString()}
        icon={TrendingUp}
        variant="success"
        trend={{ value: '+12%', positive: true }}
        progress={metrics.sentToday > 0 ? Math.min(100, (metrics.sentToday / 5000) * 100) : 0}
      />
      <MetricCard
        title={t('dashboard.failed_today')}
        value={metrics.failedToday}
        icon={AlertTriangle}
        variant={metrics.failedToday > 0 ? 'warning' : 'success'}
        trend={metrics.failedToday > 0 ? { value: '+3', positive: false } : { value: '0', positive: true }}
      />
      <MetricCard
        title={t('dashboard.delivery_success_rate')}
        value={`${metrics.deliverySuccessRate}%`}
        icon={BarChart3}
        variant={metrics.deliverySuccessRate >= 99 ? 'success' : metrics.deliverySuccessRate >= 95 ? 'default' : 'warning'}
        trend={{ value: '+0.5%', positive: true }}
        progress={metrics.deliverySuccessRate}
      />
    </div>
  );
}
