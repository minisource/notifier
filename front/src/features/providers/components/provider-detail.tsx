'use client';

import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { PageHeader } from '@minisource/ui';
import { Card, CardContent, CardHeader, CardTitle } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { Badge } from '@minisource/ui';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { Skeleton } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { DeleteProviderDialog } from '@/features/providers/components/delete-provider-dialog';
import { ProviderTestDialog } from '@/features/providers/components/provider-test-dialog';
import { ProviderBalanceCard } from '@/features/providers/components/provider-balance-card';
import { useProvider, useSetDefaultProvider } from '@/features/providers/hooks/use-providers';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { useState } from 'react';
import { ArrowLeft, Edit, Trash2, TestTube, Star, Loader2, CheckCircle, XCircle, Ban, Building2, Globe, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';

interface ProviderDetailProps {
  providerId: string;
}

export function ProviderDetail({ providerId }: ProviderDetailProps) {
  const t = useTranslations();
  const router = useRouter();
  const { data: provider, isLoading, isError, error, refetch } = useProvider(providerId);
  const { tenants } = useTenants();
  const tenantName = (id?: string) => id ? tenants.find(t => t.id === id)?.name : undefined;
  const setDefaultMutation = useSetDefaultProvider();
  const [showDelete, setShowDelete] = useState(false);
  const [showTest, setShowTest] = useState(false);

  if (isLoading) return <Skeleton className="h-64 w-full" />;

  if (isError) {
    return (
      <div className="space-y-6">
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || 'Failed to load provider'}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  if (!provider) {
    return (
      <div className="space-y-6">
        <ErrorState title={t('providers.not_found') || 'Provider not found'} message="The requested provider does not exist." />
      </div>
    );
  }

  const handleSetDefault = async () => {
    try {
      await setDefaultMutation.mutateAsync({ id: providerId, isDefault: !provider.isDefault });
    } catch (err: any) {
      toast.error(err?.message || 'Failed to update default status');
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active': return <Badge variant="default" className="bg-green-600"><CheckCircle className="h-3 w-3 ml-1" /> Active</Badge>;
      case 'inactive': return <Badge variant="secondary"><XCircle className="h-3 w-3 ml-1" /> Inactive</Badge>;
      case 'disabled': return <Badge variant="outline"><Ban className="h-3 w-3 ml-1" /> Disabled</Badge>
      case 'error': return <Badge variant="destructive"><XCircle className="h-3 w-3 ml-1" /> Error</Badge>;
      default: return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getHealthBadge = (p: NonNullable<typeof provider>) => {
    if (p.status === 'disabled') {
      return <Badge variant="outline"><Ban className="h-3 w-3 ml-1" /> {t('providers.health_disabled') || 'Disabled'}</Badge>;
    }
    if (p.status === 'error') {
      return <Badge variant="destructive"><XCircle className="h-3 w-3 ml-1" /> {t('providers.health_down') || 'Down'}</Badge>;
    }
    if (p.successRate !== undefined && p.successRate < 70) {
      return <Badge variant="destructive" className="bg-orange-600"><AlertTriangle className="h-3 w-3 ml-1" /> {t('providers.health_degraded') || 'Degraded'}</Badge>;
    }
    if (p.status === 'inactive') {
      return <Badge variant="secondary"><XCircle className="h-3 w-3 ml-1" /> {t('providers.health_inactive') || 'Inactive'}</Badge>;
    }
    return <Badge variant="default" className="bg-green-600"><CheckCircle className="h-3 w-3 ml-1" /> {t('providers.health_healthy') || 'Healthy'}</Badge>;
  };

  return (
    <div className="space-y-6">
      <PageHeader title={provider.name}>
        <Button variant="ghost" onClick={() => router.push(`/providers`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <div className="space-y-5">
        {/* Health banner */}
        <div className="flex items-center justify-between rounded-lg border p-4">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium">{t('providers.health_title') || 'Provider Health'}</span>
            {getHealthBadge(provider)}
          </div>
          <div className="flex items-center gap-6 text-sm">
            <div className="text-right">
              <p className="text-xs text-muted-foreground">{t('providers.success_rate') || 'Success Rate'}</p>
              <p className="text-lg font-semibold">{provider.successRate !== undefined ? `${Math.round(provider.successRate)}%` : '—'}</p>
            </div>
            <div className="text-right">
              <p className="text-xs text-muted-foreground">{t('providers.avg_latency') || 'Avg Latency'}</p>
              <p className="text-lg font-semibold">{provider.averageLatencyMs !== undefined ? `${Math.round(provider.averageLatencyMs)}ms` : '—'}</p>
            </div>
          </div>
        </div>
        {/* Overview Cards */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-xs font-medium text-muted-foreground">{t('providers.tenant') || 'Tenant / Project'}</CardTitle>
            </CardHeader>
            <CardContent>
              {provider.tenantId ? (
                <span className="inline-flex items-center gap-1.5 rounded-full border bg-muted/50 px-2.5 py-1 text-sm font-medium">
                  <Building2 className="h-4 w-4 text-muted-foreground" />
                  {tenantName(provider.tenantId) || provider.tenantId.slice(0, 8)}
                </span>
              ) : (
                <span className="inline-flex items-center gap-1.5 rounded-full border bg-primary/5 px-2.5 py-1 text-sm font-medium text-primary">
                  <Globe className="h-4 w-4" />
                  {t('providers.tenant_global') || 'Global'}
                </span>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-xs font-medium text-muted-foreground">{t('common.channel')}</CardTitle>
            </CardHeader>
            <CardContent>
              <ChannelBadge channel={provider.channel} size="md" />
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-xs font-medium text-muted-foreground">{t('common.status')}</CardTitle>
            </CardHeader>
            <CardContent>
              {getStatusBadge(provider.status)}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-xs font-medium text-muted-foreground">{t('providers.priority') || 'Priority'}</CardTitle>
            </CardHeader>
            <CardContent>
              <span className="text-lg font-semibold">{provider.priority}</span>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-xs font-medium text-muted-foreground">{t('providers.is_default') || 'Default'}</CardTitle>
            </CardHeader>
            <CardContent>
              {provider.isDefault ? (
                <span className="inline-flex items-center gap-1 text-sm font-medium text-amber-600 dark:text-amber-400">
                  <Star className="h-4 w-4 fill-current" /> {t('providers.yes') || 'Yes'}
                </span>
              ) : (
                <span className="text-sm text-muted-foreground">{t('common.no')}</span>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Account Balance & Credit monitoring */}
        <ProviderBalanceCard providerId={providerId} />

        {/* Provider Details */}
        <Card>
          <CardHeader>
            <CardTitle>{t('providers.details') || 'Provider Details'}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <p className="text-xs text-muted-foreground">{t('providers.provider_type') || 'Provider Type'}</p>
                <p className="text-sm font-medium capitalize">{provider.type || 'N/A'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('providers.description') || 'Description'}</p>
                <p className="text-sm">{provider.description || '—'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('common.created_at')}</p>
                <p className="text-sm">{provider.createdAt ? new Date(provider.createdAt).toLocaleString() : '—'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('common.updated_at')}</p>
                <p className="text-sm">{provider.updatedAt ? new Date(provider.updatedAt).toLocaleString() : '—'}</p>
              </div>
            </div>

            {/* Health / activity details */}
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <p className="text-xs text-muted-foreground">{t('providers.last_success') || 'Last Success'}</p>
                <p className="text-sm">{provider.lastSuccessAt ? new Date(provider.lastSuccessAt).toLocaleString() : '—'}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{t('providers.last_failure') || 'Last Failure'}</p>
                <p className="text-sm">{provider.lastFailureAt ? new Date(provider.lastFailureAt).toLocaleString() : '—'}</p>
              </div>
            </div>
            {provider.lastError && (
              <div className="space-y-1">
                <p className="text-xs font-medium text-muted-foreground">{t('providers.last_error') || 'Last Error'}</p>
                <p className="text-xs text-red-600 dark:text-red-400">{provider.lastError}</p>
              </div>
            )}

            {/* Config */}
            {provider.config && Object.keys(provider.config).length > 0 && (
              <div className="space-y-2">
                <p className="text-xs font-medium text-muted-foreground">{t('providers.config') || 'Configuration'}</p>
                <pre className="rounded-md bg-muted p-3 text-xs font-mono overflow-x-auto">
                  {JSON.stringify(provider.config, null, 2)}
                </pre>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Actions */}
        <Card>
          <CardHeader>
            <CardTitle>{t('providers.actions') || 'Actions'}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-3">
              <Button variant="default" onClick={() => router.push(`/providers/${providerId}/edit`)}>
                <Edit className="ml-1.5 h-4 w-4" />
                {t('common.edit')}
              </Button>
              <Button variant="outline" onClick={() => setShowTest(true)}>
                <TestTube className="ml-1.5 h-4 w-4" />
                {t('providers.test')}
              </Button>
              <Button variant="outline" onClick={handleSetDefault} disabled={setDefaultMutation.isPending}>
                {setDefaultMutation.isPending ? (
                  <Loader2 className="ml-1.5 h-4 w-4 animate-spin" />
                ) : (
                  <Star className="ml-1.5 h-4 w-4" />
                )}
                {provider.isDefault ? (t('providers.unset_default') || 'Unset Default') : (t('providers.set_default') || 'Set as Default')}
              </Button>
              <Button variant="destructive" onClick={() => setShowDelete(true)}>
                <Trash2 className="ml-1.5 h-4 w-4" />
                {t('common.delete')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Delete Dialog */}
        <DeleteProviderDialog
          open={showDelete}
          onOpenChange={setShowDelete}
          providerId={providerId}
          providerName={provider.name}
          channel={provider.channel}
        />

        {/* Test Dialog */}
        {showTest && (
          <ProviderTestDialog
            open={showTest}
            onOpenChange={(o) => { if (!o) setShowTest(false); }}
            provider={{ id: providerId, name: provider.name, channel: provider.channel, status: provider.status }}
          />
        )}
      </div>
    </div>
  );
}
