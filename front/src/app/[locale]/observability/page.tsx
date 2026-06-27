'use client';

import { useTranslations } from 'next-intl';
import { useState } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { RefreshButton } from '@/shared/components/refresh-button';
import { MetricCard } from '@/components/shared/metric-card';
import { ErrorState } from '@/components/shared/error-state';
import { PageSkeleton } from '@/components/shared/loading-state';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Activity, HeartPulse, CheckCircle, XCircle, AlertTriangle,
  Clock, Server, Users, BarChart3, Copy, Check, Loader2, RefreshCw,
} from 'lucide-react';

import { useAdminHealth, useAdminMetrics, useAdminReadiness } from '@/features/notifier/api/notifier-queries';
import { cn } from '@/lib/utils';

export default function ObservabilityPage() {
  const t = useTranslations();
  const [copied, setCopied] = useState(false);

  const {
    data: health,
    isLoading: healthLoading,
    isError: healthError,
    error: healthErr,
    refetch: refetchHealth,
    dataUpdatedAt: healthUpdatedAt,
  } = useAdminHealth();

  const {
    data: metrics,
    isLoading: metricsLoading,
    isError: metricsError,
    error: metricsErr,
    refetch: refetchMetrics,
  } = useAdminMetrics();

  const {
    data: readiness,
    isLoading: readinessLoading,
    isError: readinessError,
    error: readinessErr,
    refetch: refetchReadiness,
  } = useAdminReadiness();

  const loading = healthLoading || metricsLoading || readinessLoading;
  const fetchError = healthError || metricsError || readinessError;
  const errorMsg =
    (healthErr ?? metricsErr ?? readinessErr) instanceof Error
      ? ((healthErr ?? metricsErr ?? readinessErr) as Error).message
      : null;
  const isRefreshing = healthLoading || metricsLoading || readinessLoading;
  const lastUpdated = healthUpdatedAt
    ? new Date(healthUpdatedAt).toLocaleTimeString()
    : undefined;

  const handleRefresh = () => {
    refetchHealth();
    refetchMetrics();
    refetchReadiness();
  };

  const copyDiagnostics = () => {
    const info = {
      health: health?.status,
      ready: readiness?.ready,
      uptimeSeconds: health?.uptimeSeconds,
      workers: metrics?.workers?.activeCount,
      notifications: metrics?.notifications?.total,
      successRate: metrics?.notifications?.successRate,
      queueDepth: metrics?.queue?.queuedCount,
    };
    navigator.clipboard.writeText(JSON.stringify(info, null, 2));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (loading) return <PageSkeleton context="observability" layout="cards" cards={8} />;

  if (fetchError || errorMsg) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('observability.title')} subtitle={t('observability.subtitle')} />
        <ErrorState message={errorMsg || t('common.error_occurred')} onRetry={handleRefresh} />
      </div>
    );
  }

  const formatUptime = (seconds: number) => {
    const d = Math.floor(seconds / 86400);
    const h = Math.floor((seconds % 86400) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${d}d ${h}h ${m}m`;
  };

  const getStatusIcon = (status: string, size = 'h-4 w-4') => {
    switch (status) {
      case 'healthy': case 'ready': return <CheckCircle className={`${size} text-green-500`} />;
      case 'degraded': case 'not_ready': return <AlertTriangle className={`${size} text-amber-500`} />;
      case 'unhealthy': case 'down': return <XCircle className={`${size} text-red-500`} />;
      default: return <Activity className={`${size} text-muted-foreground`} />;
    }
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, string> = {
      healthy: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      ready: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      degraded: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
      not_ready: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
      unhealthy: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
      down: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
      running: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      idle: 'bg-slate-100 text-slate-600 dark:bg-slate-800/50 dark:text-slate-400',
    };
    return (
      <span className={cn('inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium', variants[status] || 'bg-gray-100 text-gray-600')}>
        {getStatusIcon(status, 'h-3 w-3')}
        {status}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <PageHeader title={t('observability.title')} subtitle={t('observability.subtitle')}>
        <Button variant="outline" size="sm" onClick={copyDiagnostics}>
          {copied ? <Check className="h-4 w-4 ltr:mr-2 rtl:ml-2" /> : <Copy className="h-4 w-4 ltr:mr-2 rtl:ml-2" />}
          {copied ? t('common.copied') : t('observability.copy_diagnostics')}
        </Button>
        <RefreshButton onRefresh={handleRefresh} isRefreshing={isRefreshing} lastUpdated={lastUpdated} />
      </PageHeader>

      {/* 1. Service Health */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <HeartPulse className="h-4 w-4" />
            {t('observability.health')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-4">
            <MetricCard title={t('common.status')} value={health?.status || 'unknown'} icon={Activity}
              variant={health?.status === 'healthy' ? 'success' : health?.status === 'degraded' ? 'warning' : 'danger'} />
            <MetricCard title={t('observability.uptime')} value={formatUptime(health?.uptimeSeconds || 0)} icon={Clock} />
            <MetricCard title={t('settings.version') || 'Version'} value={health?.version || '1.0.0'} icon={Server} />
            <MetricCard title={t('settings.api_mode')} value={health?.environment || 'development'} icon={BarChart3} />
          </div>
          {health?.dependencies && (
            <div className="mt-4 space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Dependencies</p>
              {health.dependencies.map((dep) => (
                <div key={dep.name} className="flex items-center justify-between rounded-lg border p-2.5 text-sm">
                  <div className="flex items-center gap-2">
                    {getStatusIcon(dep.status, 'h-3.5 w-3.5')}
                    <span className="font-medium capitalize">{dep.name}</span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    {dep.latencyMs && <span>{dep.latencyMs}ms</span>}
                    {getStatusBadge(dep.status)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 2. Readiness */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <CheckCircle className="h-4 w-4" />
            {t('observability.readiness')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3 mb-4">
            <span className="text-sm font-medium">{t('observability.readiness')}:</span>
            {getStatusBadge(readiness?.overall || 'unknown')}
            {readiness?.ready !== undefined && (
              <span className="text-xs text-muted-foreground">
                {readiness.ready ? 'Accepting traffic' : 'Not ready'}
              </span>
            )}
          </div>
          {readiness?.checks && (
            <div className="space-y-1.5">
              {readiness.checks.map((check) => (
                <div key={check.name} className="flex items-center justify-between rounded-lg border p-2.5 text-sm">
                  <div className="flex items-center gap-2">
                    {getStatusIcon(check.status, 'h-3.5 w-3.5')}
                    <span className="capitalize">{check.name}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {check.message && <span className="text-xs text-muted-foreground">{check.message}</span>}
                    {getStatusBadge(check.status)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 3. Metrics */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard title={t('dashboard.total_notifications')} value={metrics?.notifications?.total?.toLocaleString() || 0} icon={BarChart3} />
        <MetricCard title={t('dashboard.sent_today')} value={metrics?.notifications?.sentToday ?? 0} icon={Activity} variant="success" />
        <MetricCard title={t('dashboard.failed_today')} value={metrics?.notifications?.failedToday ?? 0} icon={AlertTriangle} variant={metrics?.notifications?.failedToday ? 'warning' : 'default'} />
        <MetricCard title={t('dashboard.delivery_success_rate')} value={`${metrics?.notifications?.successRate ?? 0}%`} icon={CheckCircle}
          variant={(metrics?.notifications?.successRate ?? 0) >= 99 ? 'success' : (metrics?.notifications?.successRate ?? 0) >= 95 ? 'default' : 'warning'} />
      </div>

      {/* 4. Queue */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Clock className="h-4 w-4" />
            {t('observability.queue_depth')} — {t('dashboard.queued')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-4">
            <MetricCard title={t('statuses.queued')} value={metrics?.queue?.queuedCount ?? 0} icon={Clock} />
            <MetricCard title={t('notifications.timeline.processing')} value={metrics?.queue?.processingCount ?? 0} icon={Loader2} />
            <MetricCard title={t('statuses.retrying')} value={metrics?.queue?.retryingCount ?? 0} icon={RefreshCw} variant="warning" />
            <MetricCard title={t('statuses.dead')} value={metrics?.queue?.deadCount ?? 0} icon={XCircle} variant="danger" />
          </div>
          <div className="mt-3 grid gap-4 sm:grid-cols-3 text-sm">
            <div className="flex justify-between rounded-lg border p-2.5">
              <span className="text-muted-foreground">{t('notifications.filters.priority_low') || 'Throughput'}</span>
              <span className="font-medium">{metrics?.queue?.throughputPerMinute ?? 0}/min</span>
            </div>
            <div className="flex justify-between rounded-lg border p-2.5">
              <span className="text-muted-foreground">{t('dashboard.avg_delivery_time')}</span>
              <span className="font-medium">{metrics?.deliveries?.averageLatencyMs ?? 0}ms</span>
            </div>
            <div className="flex justify-between rounded-lg border p-2.5">
              <span className="text-muted-foreground">{t('deliveries.attempts')}</span>
              <span className="font-medium">{metrics?.deliveries?.totalAttempts ?? 0}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 5. Workers */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Users className="h-4 w-4" />
            {t('observability.workers')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-3 mb-4">
            <MetricCard title={t('common.status') as string} value={`${metrics?.workers?.activeCount ?? 0} active`} icon={CheckCircle} variant="success" />
            <MetricCard title="Idle" value={metrics?.workers?.idleCount ?? 0} icon={Clock} />
            <MetricCard title="Failed" value={metrics?.workers?.failedCount ?? 0} icon={XCircle} variant={(metrics?.workers?.failedCount ?? 0) > 0 ? 'danger' : 'default'} />
          </div>
          {metrics?.workers?.workers && metrics.workers.workers.length > 0 && (
            <div className="space-y-1.5">
              {metrics.workers.workers.map((w: { workerName: string; enabled: boolean; status: string; lastRunAt?: string; pollInterval: string; batchSize: number }) => (
                <div key={w.workerName} className="flex items-center justify-between rounded-lg border p-2.5 text-sm">
                  <div className="flex items-center gap-2">
                    {getStatusIcon(w.status === 'running' ? 'healthy' : 'idle', 'h-3.5 w-3.5')}
                    <span className="font-medium">{w.workerName}</span>
                  </div>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground">
                    <span>{w.pollInterval}</span>
                    <span>{w.batchSize} batch</span>
                    {getStatusBadge(w.status)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
