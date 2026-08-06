'use client';

import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { PageHeader } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { Card } from '@minisource/ui';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { Skeleton } from '@minisource/ui';
import { EmptyState } from '@minisource/ui';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { useState } from 'react';
import { Server, RefreshCw, Plus, Star, CheckCircle, AlertTriangle, XCircle, Ban, Wifi, Loader2, Gauge, Building2, Globe } from 'lucide-react';
import { useProviders, useToggleProviderStatus, useSetDefaultProvider, useProviderBalanceHealth } from '@/features/providers/hooks/use-providers';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { MetricCard } from '@minisource/ui';
import { ProviderTestDialog } from '@/features/providers/components/provider-test-dialog';
import { ProviderActionsMenu } from '@/features/providers/components/provider-actions-menu';
import { DeleteProviderDialog } from '@/features/providers/components/delete-provider-dialog';
import { formatBalanceValue, balanceUnitLabel } from '@/features/providers/balance-format';
import { healthCheckAll, testProvider as runProviderTest } from '@/features/providers/api';
import { toast } from 'sonner';

export default function ProvidersPage() {
  const t = useTranslations();
  const router = useRouter();
  const { data: providers, isLoading, isError, error, refetch, isFetching } = useProviders();
  const { data: balanceHealth } = useProviderBalanceHealth();
  const { tenants } = useTenants();
  const tenantName = (id?: string) => id ? tenants.find(t => t.id === id)?.name : undefined;
  const toggleStatusMutation = useToggleProviderStatus();
  const setDefaultMutation = useSetDefaultProvider();
  const [testProvider, setTestProvider] = useState<{ id: string; name: string; channel: string; status: string } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string; channel: string } | null>(null);
  const [checkingAll, setCheckingAll] = useState(false);
  const [checkingIds, setCheckingIds] = useState<Record<string, boolean>>({});

  const activeCount = providers?.filter(p => p.status === 'active').length || 0;
  const inactiveCount = providers?.filter(p => p.status === 'inactive' || p.status === 'disabled').length || 0;
  const errorCount = providers?.filter(p => p.status === 'error').length || 0;
  const defaultCount = providers?.filter(p => p.isDefault).length || 0;

  const getSuccessRateBadge = (rate?: number, status?: string) => {
    if (rate === undefined) {
      return status === 'active'
        ? <span className="inline-flex items-center gap-1 text-xs text-muted-foreground"><Gauge className="h-3.5 w-3.5" /> {t('providers.no_data') || 'No data'}</span>
        : <span className="text-xs text-muted-foreground">—</span>;
    }
    const pct = Math.round(rate);
    const color = pct >= 90 ? 'text-green-600 dark:text-green-400' : pct >= 70 ? 'text-amber-600 dark:text-amber-400' : 'text-red-600 dark:text-red-400';
    const barColor = pct >= 90 ? 'bg-green-500' : pct >= 70 ? 'bg-amber-500' : 'bg-red-500';
    return (
      <div className="flex items-center gap-2">
        <div className="h-1.5 w-14 overflow-hidden rounded-full bg-muted">
          <div className={`h-full rounded-full ${barColor}`} style={{ width: `${Math.min(pct, 100)}%` }} />
        </div>
        <span className={`text-sm font-medium ${color}`}>{pct}%</span>
      </div>
    );
  };

  const getLastFailure = (p: { lastFailureAt?: string; lastError?: string; status?: string }) => {
    if (p.lastFailureAt) {
      return `${new Date(p.lastFailureAt).toLocaleString()}${p.lastError ? ` — ${p.lastError}` : ''}`;
    }
    return undefined;
  };

  const balanceHealthFor = (providerId: string) => balanceHealth?.find((b) => b.providerId === providerId);

  const getBalanceBadge = (providerId: string) => {
    const h = balanceHealthFor(providerId);
    if (!h) return <span className="text-xs text-muted-foreground">—</span>;
    const level = h.healthLevel;
    const unitLabel = balanceUnitLabel(t, h.balanceUnit);
    const value = h.balanceValue !== null && h.balanceValue !== undefined ? `${formatBalanceValue(h.balanceValue)}${unitLabel ? ' ' + unitLabel : ''}` : null;
    const cls =
      level === 'healthy' ? 'text-green-600 dark:text-green-400'
      : level === 'warning' ? 'text-amber-600 dark:text-amber-400'
      : level === 'critical' || level === 'exhausted' ? 'text-red-600 dark:text-red-400'
      : level === 'stale' ? 'text-yellow-600 dark:text-yellow-400'
      : 'text-muted-foreground';
    return (
      <div className="flex items-center gap-1.5" title={h.lastErrorMessage || level}>
        <span className={`inline-flex items-center gap-1 text-xs font-medium ${cls}`}>
          {level === 'healthy' ? <CheckCircle className="h-3 w-3" /> : level === 'warning' ? <AlertTriangle className="h-3 w-3" /> : level === 'critical' || level === 'exhausted' ? <XCircle className="h-3 w-3" /> : <Gauge className="h-3 w-3" />}
          {value ?? (t(`providers.health_${level}`) || level)}
        </span>
        {h.activeAlertCount > 0 && (
          <span className="inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-red-100 px-1 text-[10px] font-semibold text-red-600 dark:bg-red-950/40 dark:text-red-400">{h.activeAlertCount}</span>
        )}
      </div>
    );
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-950/40 dark:text-green-400"><CheckCircle className="h-3 w-3" /> {t('providers.active') || 'Active'}</span>;
      case 'inactive':
        return <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-950/40 dark:text-amber-400"><AlertTriangle className="h-3 w-3" /> {t('providers.inactive') || 'Inactive'}</span>;
      case 'disabled':
        return <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600 dark:bg-gray-800/50 dark:text-gray-400"><Ban className="h-3 w-3" /> {t('providers.disabled') || 'Disabled'}</span>;
      case 'error':
        return <span className="inline-flex items-center gap-1 rounded-full bg-red-50 px-2 py-0.5 text-xs font-medium text-red-700 dark:bg-red-950/40 dark:text-red-400"><XCircle className="h-3 w-3" /> {t('providers.error') || 'Error'}</span>;
      default:
        return <span className="text-xs text-muted-foreground">{status}</span>;
    }
  };

  const handleToggleStatus = async (id: string, currentStatus: string) => {
    const newEnabled = currentStatus === 'disabled';
    try {
      await toggleStatusMutation.mutateAsync({ id, isEnabled: newEnabled });
      toast.success(newEnabled ? 'Provider enabled' : 'Provider disabled');
    } catch (err: any) {
      toast.error(err?.message || 'Failed to toggle status');
    }
  };

  const handleSetDefault = async (id: string, isDefault: boolean) => {
    try {
      await setDefaultMutation.mutateAsync({ id, isDefault: !isDefault });
    } catch (err: any) {
      toast.error(err?.message || 'Failed to update default status');
    }
  };

  // Runs a REAL connectivity check against this provider's own API.
  const handleCheckProvider = async (id: string, name: string) => {
    setCheckingIds((prev) => ({ ...prev, [id]: true }));
    try {
      const res = await runProviderTest(id, { dryRun: true });
      if (res.success) {
        toast.success(`${name} ${t('providers.healthy').toLowerCase()}`, {
          description: `${res.message || 'Connection verified'}${res.latencyMs ? ` · ${res.latencyMs}ms` : ''}`,
        });
      } else {
        toast.error(`${name} ${t('providers.degraded').toLowerCase()}`, {
          description: res.message || res.providerResponseSanitized || 'Health check failed',
        });
      }
      refetch();
    } catch (err: any) {
      toast.error(`${name} ${t('providers.degraded').toLowerCase()}`, {
        description: err?.message || 'Health check failed',
      });
    } finally {
      setCheckingIds((prev) => ({ ...prev, [id]: false }));
    }
  };

  // Runs REAL connectivity checks against ALL providers' own APIs.
  const handleCheckAll = async () => {
    setCheckingAll(true);
    try {
      const res = await healthCheckAll();
      toast.success(t('providers.health_check_done') || 'Health check completed', {
        description: `${res.healthyCount} ${t('providers.healthy').toLowerCase()} · ${res.downCount} ${t('providers.down').toLowerCase() || 'down'} · ${res.disabledCount} ${t('providers.disabled').toLowerCase() || 'disabled'}`,
      });
      refetch();
    } catch (err: any) {
      toast.error(err?.message || 'Health check failed');
    } finally {
      setCheckingAll(false);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title={t('providers.title')} description={t('providers.subtitle')}>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleCheckAll} disabled={checkingAll || isFetching}>
            {checkingAll ? <Loader2 className="ml-1.5 h-4 w-4 animate-spin" /> : <Wifi className="ml-1.5 h-4 w-4" />}
            {t('providers.check_all_health') || 'Check All Health'}
          </Button>
          <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
            <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
            {t('observability.refresh')}
          </Button>
          <Button size="sm" onClick={() => router.push(`/providers/new`)}>
            <Plus className="ml-1.5 h-4 w-4" />
            {t('providers.create') || 'Create Provider'}
          </Button>
        </div>
      </PageHeader>

      {/* Summary metrics */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard title={t('providers.active') || 'Active'} value={activeCount} icon={CheckCircle} variant="success" />
        <MetricCard title={t('providers.inactive') || 'Inactive'} value={inactiveCount} icon={Ban} variant="default" />
        <MetricCard title={t('providers.error_count') || 'Errors'} value={errorCount} icon={XCircle} variant={errorCount > 0 ? 'danger' : 'default'} />
        <MetricCard title={t('providers.default') || 'Default'} value={defaultCount} icon={Star} variant="warning" />
      </div>

      <Card title={t('providers.title')}>
        {isLoading ? (
          <Skeleton className="h-64 w-full" />
        ) : isError ? (
          <ErrorState
            title={t('errors.generic')}
            message={(error as Error)?.message || t('errors.generic')}
            onRetry={() => refetch()}
            autoRetrySeconds={15}
          />
        ) : providers && providers.length > 0 ? (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[220px]">{t('providers.name') || 'Name'}</TableHead>
                  <TableHead className="w-[150px]">{t('providers.tenant') || 'Tenant / Project'}</TableHead>
                  <TableHead className="w-[90px]">{t('common.channel')}</TableHead>
                  <TableHead className="w-[110px]">{t('providers.provider_type') || 'Type'}</TableHead>
                  <TableHead className="w-[100px]">{t('common.status')}</TableHead>
                  <TableHead className="w-[70px]">{t('providers.priority') || 'Priority'}</TableHead>
                  <TableHead className="w-[110px]">{t('providers.success_rate') || 'Success Rate'}</TableHead>
                  <TableHead className="w-[120px]">{t('providers.balance.title') || 'Balance'}</TableHead>
                  <TableHead className="w-[70px]">{t('providers.default') || 'Default'}</TableHead>
                  <TableHead className="w-[60px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {providers.map((provider) => (
                  <TableRow
                    key={provider.id}
                    className="cursor-pointer"
                    onClick={() => router.push(`/providers/${provider.id}`)}
                  >
                    <TableCell>
                      <div className="flex items-center gap-2 min-w-0">
                        <Server className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <span className="text-sm font-medium truncate">{provider.name}</span>
                        {provider.isDefault && (
                          <Star className="h-3.5 w-3.5 shrink-0 text-amber-500 fill-amber-500" />
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      {provider.tenantId ? (
                        <span className="inline-flex items-center gap-1.5 rounded-full border bg-muted/50 px-2 py-0.5 text-xs font-medium">
                          <Building2 className="h-3 w-3 text-muted-foreground" />
                          {tenantName(provider.tenantId) || provider.tenantId.slice(0, 8)}
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 rounded-full border bg-primary/5 px-2 py-0.5 text-xs font-medium text-primary">
                          <Globe className="h-3 w-3" />
                          {t('providers.tenant_global') || 'Global'}
                        </span>
                      )}
                    </TableCell>
                    <TableCell>
                      <ChannelBadge channel={provider.channel} size="sm" />
                    </TableCell>
                    <TableCell>
                      <span className="text-sm capitalize">{provider.type || '—'}</span>
                    </TableCell>
                    <TableCell>{getStatusBadge(provider.status)}</TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">{provider.priority}</span>
                    </TableCell>
                    <TableCell title={getLastFailure(provider)}>
                      {getSuccessRateBadge(provider.successRate, provider.status)}
                    </TableCell>
                    <TableCell>{getBalanceBadge(provider.id)}</TableCell>
                    <TableCell>
                      {provider.isDefault ? (
                        <span className="inline-flex items-center gap-1 text-xs font-medium text-amber-600 dark:text-amber-400">
                          <Star className="h-3.5 w-3.5 fill-current" /> {t('providers.yes') || 'Yes'}
                        </span>
                      ) : (
                        <span className="text-xs text-muted-foreground">{t('common.no')}</span>
                      )}
                    </TableCell>
                    <TableCell onClick={(e) => e.stopPropagation()}>
                      <ProviderActionsMenu
                        provider={provider}
                        checking={!!checkingIds[provider.id]}
                        togglePending={toggleStatusMutation.isPending}
                        defaultPending={setDefaultMutation.isPending}
                        onView={() => router.push(`/providers/${provider.id}`)}
                        onEdit={() => router.push(`/providers/${provider.id}/edit`)}
                        onHealthCheck={() => handleCheckProvider(provider.id, provider.name)}
                        onTest={() => setTestProvider({ id: provider.id, name: provider.name, channel: provider.channel, status: provider.status })}
                        onToggleStatus={() => handleToggleStatus(provider.id, provider.status)}
                        onSetDefault={() => handleSetDefault(provider.id, provider.isDefault)}
                        onDelete={() => setDeleteTarget({ id: provider.id, name: provider.name, channel: provider.channel })}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <EmptyState
            icon={Wifi}
            title={t('providers.no_providers')}
            description="Configure one or more providers to enable notification delivery through different channels."
            actionLabel={t('providers.create') || 'Create Provider'}
            onAction={() => router.push(`/providers/new`)}
            tips={[
              'SMS providers: Kavenegar, Twilio, or custom SMPP',
              'Email providers: SMTP server, SendGrid, or custom API',
              'Push providers: Firebase Cloud Messaging (FCM) or APNs',
              'Each provider can be tested before enabling in production',
            ]}
          />
        )}
      </Card>

      {/* Test Dialog */}
      {testProvider && (
        <ProviderTestDialog
          open={!!testProvider}
          onOpenChange={(open) => { if (!open) setTestProvider(null); }}
          provider={testProvider}
        />
      )}

      {/* Delete Dialog */}
      {deleteTarget && (
        <DeleteProviderDialog
          open={!!deleteTarget}
          onOpenChange={(open) => { if (!open) setDeleteTarget(null); }}
          providerId={deleteTarget.id}
          providerName={deleteTarget.name}
          channel={deleteTarget.channel}
        />
      )}
    </div>
  );
}
