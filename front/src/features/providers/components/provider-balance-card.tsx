'use client';

import { useTranslations } from 'next-intl';
import { Card, CardContent, CardHeader, CardTitle } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { Badge } from '@minisource/ui';
import { Input } from '@minisource/ui';
import { Skeleton } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { useState } from 'react';
import {
  RefreshCw, Wallet, Loader2, CheckCircle2, AlertTriangle, XCircle, Ban,
  HelpCircle, Gauge, Clock, Activity, CheckCheck, Layers,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  useProviderBalanceDetail,
  useRefreshProviderBalance,
  useUpdateProviderBalanceSettings,
  useAcknowledgeCreditAlert,
} from '@/features/providers/hooks/use-providers';
import { formatBalanceValue, balanceUnitLabel } from '@/features/providers/balance-format';

interface ProviderBalanceCardProps {
  providerId: string;
}

// Simple inline sparkline (no heavy chart dependency) for balance history.
function BalanceSparkline({ points, unit }: { points: (number | null | undefined)[]; unit?: string }) {
  if (!points || points.length < 2) {
    return <p className="text-xs text-muted-foreground">—</p>;
  }
  const nums = points.filter((p): p is number => typeof p === 'number');
  if (nums.length < 2) {
    return <p className="text-xs text-muted-foreground">—</p>;
  }
  const min = Math.min(...nums);
  const max = Math.max(...nums);
  const range = max - min || 1;
  const w = 100;
  const h = 28;
  const step = w / (nums.length - 1);
  const coords = nums.map((v, i) => `${(i * step).toFixed(1)},${(h - ((v - min) / range) * (h - 4) - 2).toFixed(1)}`);
  return (
    <div className="flex items-end gap-2">
      <svg viewBox={`0 0 ${w} ${h}`} className="h-8 w-full max-w-[220px]" preserveAspectRatio="none">
        <polyline points={coords.join(' ')} fill="none" stroke="currentColor" strokeWidth="1.5" className="text-primary" />
        {nums.map((v, i) => (
          <circle key={i} cx={(i * step).toFixed(1)} cy={(h - ((v - min) / range) * (h - 4) - 2).toFixed(1)} r="1.8" className="fill-primary" />
        ))}
      </svg>
      <span className="shrink-0 text-xs text-muted-foreground">{unit ? `last ${nums.length}` : ''}</span>
    </div>
  );
}

function healthBadge(level: string | undefined) {
  const Icon = level === 'healthy' ? CheckCircle2 : level === 'warning' ? AlertTriangle : level === 'critical' || level === 'exhausted' ? XCircle : level === 'stale' ? Clock : level === 'unsupported' ? HelpCircle : level === 'disabled' ? Ban : Activity;
  const cls = level === 'healthy' ? 'bg-green-50 text-green-700 dark:bg-green-950/40 dark:text-green-400'
    : level === 'warning' ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400'
    : level === 'critical' || level === 'exhausted' ? 'bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-400'
    : level === 'stale' ? 'bg-yellow-50 text-yellow-700 dark:bg-yellow-950/40 dark:text-yellow-400'
    : level === 'unsupported' ? 'bg-gray-100 text-gray-600 dark:bg-gray-800/50 dark:text-gray-400'
    : 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-400';
  return (
    <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${cls}`}>
      <Icon className="h-3 w-3" />
      {level ?? 'unknown'}
    </span>
  );
}

export function ProviderBalanceCard({ providerId }: ProviderBalanceCardProps) {
  const t = useTranslations();
  const { data, isLoading, isError, error, refetch } = useProviderBalanceDetail(providerId);
  const refreshMutation = useRefreshProviderBalance();
  const settingsMutation = useUpdateProviderBalanceSettings();
  const acknowledgeMutation = useAcknowledgeCreditAlert();

  const [warning, setWarning] = useState('');
  const [critical, setCritical] = useState('');
  const [editing, setEditing] = useState(false);

  const health = data?.health;

  const handleRefresh = async () => {
    try {
      await refreshMutation.mutateAsync(providerId);
      refetch();
    } catch {
      /* toast handled in hook */
    }
  };

  const handleSaveSettings = async () => {
    const w = warning === '' ? null : Number(warning);
    const c = critical === '' ? null : Number(critical);
    if (w !== null && c !== null && c > w) {
      toast.error(t('providers.balance.critical_gt_warning') || 'Critical threshold must be lower than warning');
      return;
    }
    try {
      await settingsMutation.mutateAsync({
        id: providerId,
        settings: {
          enabled: data?.settings?.enabled ?? true,
          warningThreshold: w,
          criticalThreshold: c,
          refreshIntervalSec: data?.settings?.refreshIntervalSec ?? null,
        },
      });
      setEditing(false);
      refetch();
    } catch {
      /* toast handled in hook */
    }
  };

  const handleAcknowledge = async (alertId: string) => {
    try {
      await acknowledgeMutation.mutateAsync(alertId);
      refetch();
    } catch {
      /* toast handled in hook */
    }
  };

  if (isLoading) return <Skeleton className="h-56 w-full" />;

  if (isError) {
    return (
      <ErrorState
        title={t('errors.generic')}
        message={(error as Error)?.message || 'Failed to load balance'}
        onRetry={() => refetch()}
      />
    );
  }

  const balance = health?.balanceValue ?? null;
  const unitLabel = balanceUnitLabel(t, health?.balanceUnit);
  const historyPoints = data?.history?.map((h) => h.balanceValue).filter((v): v is number => v !== null && v !== undefined) ?? [];

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="flex items-center gap-2 text-sm font-medium">
          <Wallet className="h-4 w-4 text-muted-foreground" />
          {t('providers.balance.title') || 'Account Balance & Credit'}
        </CardTitle>
        <Button variant="outline" size="sm" onClick={handleRefresh} disabled={refreshMutation.isPending}>
          {refreshMutation.isPending ? <Loader2 className="ml-1.5 h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="ml-1.5 h-3.5 w-3.5" />}
          {t('providers.balance.refresh_now') || 'Refresh Now'}
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap items-center gap-4">
          <div>
            <p className="text-xs text-muted-foreground">{t('providers.balance.current_balance') || 'Current Balance'}</p>
            <p className="text-2xl font-semibold">
              {balance !== null ? (
                <>
                  {formatBalanceValue(balance)}
                  {unitLabel ? <span className="ml-1 text-sm font-normal text-muted-foreground">{unitLabel}</span> : null}
                  {health?.currency ? <span className="ml-1 text-sm font-normal text-muted-foreground">{health.currency}</span> : null}
                </>
              ) : (
                <span className="text-muted-foreground">—</span>
              )}
            </p>
          </div>
          <div className="flex items-center gap-2">
            {healthBadge(health?.healthLevel)}
            {health?.isEstimated && <Badge variant="outline">{t('providers.balance.estimated') || 'Estimated'}</Badge>}
            {health?.isManual && <Badge variant="outline">{t('providers.balance.manual') || 'Manual'}</Badge>}
          </div>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <div>
            <p className="text-xs text-muted-foreground">{t('providers.balance.last_successful_refresh') || 'Last Successful Refresh'}</p>
            <p className="text-sm">
              {health?.lastSuccessfulRefreshAt ? new Date(health.lastSuccessfulRefreshAt).toLocaleString() : '—'}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">{t('providers.balance.consecutive_failures') || 'Consecutive Failures'}</p>
            <p className="text-sm">{health?.consecutiveFailures ?? 0}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">{t('providers.balance.capability') || 'Capability'}</p>
            <p className="text-sm">{health?.capabilityMode ?? 'unsupported'}</p>
          </div>
        </div>

        {health?.lastErrorMessage ? (
          <div className="rounded-md border border-amber-200 bg-amber-50 p-2 text-xs text-amber-700 dark:border-amber-900/40 dark:bg-amber-950/30 dark:text-amber-400">
            {health.lastErrorMessage}
          </div>
        ) : null}

        {/* History sparkline */}
        <div className="space-y-1">
          <p className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
            <Gauge className="h-3.5 w-3.5" />
            {t('providers.balance.history') || 'Balance History'}
          </p>
          <BalanceSparkline points={historyPoints} unit={health?.balanceUnit} />
        </div>

        {/* Active alerts */}
        {(data?.alerts ?? []).length > 0 && (
          <div className="space-y-2">
            <p className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
              <Activity className="h-3.5 w-3.5" />
              {t('providers.balance.alerts') || 'Alerts'}
            </p>
            <div className="space-y-1.5">
              {(data?.alerts ?? []).slice(0, 5).map((a) => (
                <div key={a.id} className="flex items-center justify-between gap-2 rounded-md border px-2.5 py-1.5 text-xs">
                  <div className="flex min-w-0 items-center gap-2">
                    {healthBadge(a.alertType)}
                    <span className="truncate">{a.message}</span>
                    {a.balanceValue !== null && a.balanceValue !== undefined ? (
                      <span className="shrink-0 text-muted-foreground">({formatBalanceValue(a.balanceValue)}{unitLabel ? ' ' + unitLabel : ''})</span>
                    ) : null}
                  </div>
                  {a.status === 'active' && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 shrink-0 px-1.5 text-xs"
                      onClick={() => handleAcknowledge(a.id)}
                      disabled={acknowledgeMutation.isPending}
                    >
                      <CheckCheck className="ml-1 h-3 w-3" />
                      {t('providers.balance.acknowledge') || 'Acknowledge'}
                    </Button>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Threshold settings */}
        <div className="space-y-2 border-t pt-3">
          <div className="flex items-center justify-between">
            <p className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
              <Layers className="h-3.5 w-3.5" />
              {t('providers.balance.thresholds') || 'Alert Thresholds'}
            </p>
            {!editing && (
              <Button variant="ghost" size="sm" className="h-6 px-1.5 text-xs" onClick={() => {
                setWarning(data?.settings?.warningThreshold != null ? String(data.settings.warningThreshold) : '');
                setCritical(data?.settings?.criticalThreshold != null ? String(data.settings.criticalThreshold) : '');
                setEditing(true);
              }}>
                {t('common.edit')}
              </Button>
            )}
          </div>
          {editing ? (
            <div className="grid gap-3 sm:grid-cols-2">
              <div className="space-y-1">
                <label className="text-xs text-muted-foreground">{t('providers.balance.warning_threshold') || 'Warning below'}</label>
                <Input type="number" value={warning} onChange={(e) => setWarning(e.target.value)} placeholder="100" />
              </div>
              <div className="space-y-1">
                <label className="text-xs text-muted-foreground">{t('providers.balance.critical_threshold') || 'Critical below'}</label>
                <Input type="number" value={critical} onChange={(e) => setCritical(e.target.value)} placeholder="20" />
              </div>
              <div className="flex gap-2 sm:col-span-2">
                <Button size="sm" onClick={handleSaveSettings} disabled={settingsMutation.isPending}>
                  {settingsMutation.isPending ? <Loader2 className="ml-1.5 h-3.5 w-3.5 animate-spin" /> : null}
                  {t('common.save')}
                </Button>
                <Button size="sm" variant="outline" onClick={() => setEditing(false)}>{t('common.cancel')}</Button>
              </div>
            </div>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              <div>
                <p className="text-xs text-muted-foreground">{t('providers.balance.warning_threshold') || 'Warning below'}</p>
                <p className="text-sm">{data?.settings?.warningThreshold ?? '—'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('providers.balance.critical_threshold') || 'Critical below'}</p>
                <p className="text-sm">{data?.settings?.criticalThreshold ?? '—'}</p>
              </div>
            </div>
          )}
          <p className="text-[11px] text-muted-foreground">
            {t('providers.balance.unit_hint') || 'Thresholds use the provider’s reported unit — a monetary credit amount (e.g. Rial) or a message count.'}
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
