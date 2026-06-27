'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ErrorState } from '@/components/shared/error-state';
import { CardSkeleton } from '@/components/shared/loading-state';
import { EmptyState } from '@/components/shared/empty-state';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { useState } from 'react';
import { Server, RefreshCw, Plus, CheckCircle, AlertTriangle, XCircle, Ban, TestTube, Wifi, Star, Edit, Trash2, Eye, Loader2 } from 'lucide-react';
import { useProviders, useToggleProviderStatus, useSetDefaultProvider } from '@/features/providers/hooks/use-providers';
import { MetricCard } from '@/components/shared/metric-card';
import { ProviderTestDialog } from '@/features/providers/components/provider-test-dialog';
import { DeleteProviderDialog } from '@/features/providers/components/delete-provider-dialog';
import { toast } from 'sonner';

export default function ProvidersPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const { data: providers, isLoading, isError, error, refetch, isFetching } = useProviders();
  const toggleStatusMutation = useToggleProviderStatus();
  const setDefaultMutation = useSetDefaultProvider();
  const [testProvider, setTestProvider] = useState<{ id: string; name: string; channel: string; status: string } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string; channel: string } | null>(null);

  const activeCount = providers?.filter(p => p.status === 'active').length || 0;
  const inactiveCount = providers?.filter(p => p.status === 'inactive' || p.status === 'disabled').length || 0;
  const errorCount = providers?.filter(p => p.status === 'error').length || 0;
  const defaultCount = providers?.filter(p => p.isDefault).length || 0;

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge variant="default" className="bg-green-600"><CheckCircle className="h-3 w-3 ml-1" /> {t('providers.healthy') || 'Active'}</Badge>;
      case 'inactive':
        return <Badge variant="secondary"><AlertTriangle className="h-3 w-3 ml-1" /> Inactive</Badge>;
      case 'disabled':
        return <Badge variant="outline"><Ban className="h-3 w-3 ml-1" /> Disabled</Badge>;
      case 'error':
        return <Badge variant="destructive"><XCircle className="h-3 w-3 ml-1" /> Error</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
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

  if (isLoading) return <CardSkeleton cards={6} context="providers" />;

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title={t('providers.title')} subtitle={t('providers.subtitle')}>
          <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
            <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
            {t('observability.refresh')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || t('errors.generic')}
          onRetry={() => refetch()}
          autoRetrySeconds={15}
        />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader title={t('providers.title')} subtitle={t('providers.subtitle')}>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
            <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
            {t('observability.refresh')}
          </Button>
          <Button size="sm" onClick={() => router.push(`/${locale}/providers/new`)}>
            <Plus className="ml-1.5 h-4 w-4" />
            {t('providers.create') || 'Create Provider'}
          </Button>
        </div>
      </PageHeader>

      <div className="space-y-5">
        {/* Summary cards */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <MetricCard title={t('providers.active') || 'Active'} value={activeCount} icon={CheckCircle} variant="success" />
          <MetricCard title={t('providers.inactive') || 'Inactive'} value={inactiveCount} icon={Ban} variant="default" />
          <MetricCard title={t('providers.error_count') || 'Errors'} value={errorCount} icon={XCircle} variant={errorCount > 0 ? 'danger' : 'default'} />
          <MetricCard title={t('providers.default') || 'Default'} value={defaultCount} icon={Star} variant="warning" />
        </div>

        {/* Provider Cards */}
        {providers && providers.length > 0 ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {providers.map((provider) => (
              <div
                key={provider.id}
                className="rounded-lg border border-border/70 bg-card p-4 shadow-sm transition-all hover:border-border hover:shadow-md"
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2 min-w-0">
                    <Server className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <span className="font-medium text-sm truncate">{provider.name}</span>
                    {provider.isDefault && (
                      <Star className="h-3.5 w-3.5 shrink-0 text-amber-500 fill-amber-500" />
                    )}
                  </div>
                  {getStatusBadge(provider.status)}
                </div>

                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">{t('common.channel')}</span>
                    <ChannelBadge channel={provider.channel} size="sm" />
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">{t('providers.provider_type') || 'Type'}</span>
                    <span className="capitalize">{provider.type || 'N/A'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">{t('providers.priority') || 'Priority'}</span>
                    <span>{provider.priority}</span>
                  </div>
                </div>

                <div className="mt-3 pt-3 border-t border-border/50 flex flex-wrap gap-1.5">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-xs h-8"
                    onClick={() => router.push(`/${locale}/providers/${provider.id}`)}
                  >
                    <Eye className="h-3.5 w-3.5 ml-1" />
                    {t('common.view')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-xs h-8"
                    onClick={() => router.push(`/${locale}/providers/${provider.id}/edit`)}
                  >
                    <Edit className="h-3.5 w-3.5 ml-1" />
                    {t('common.edit')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-xs h-8"
                    onClick={() => setTestProvider({ id: provider.id, name: provider.name, channel: provider.channel, status: provider.status })}
                  >
                    <TestTube className="h-3.5 w-3.5 ml-1" />
                    {t('providers.test')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-xs h-8"
                    onClick={() => handleToggleStatus(provider.id, provider.status)}
                    disabled={toggleStatusMutation.isPending}
                  >
                    {toggleStatusMutation.isPending ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : provider.status === 'disabled' ? (
                      <CheckCircle className="h-3.5 w-3.5 ml-1 text-green-500" />
                    ) : (
                      <Ban className="h-3.5 w-3.5 ml-1" />
                    )}
                    {provider.status === 'disabled' ? (t('common.enable') || 'Enable') : (t('common.disable') || 'Disable')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-xs h-8"
                    onClick={() => handleSetDefault(provider.id, provider.isDefault)}
                    disabled={setDefaultMutation.isPending}
                  >
                    {setDefaultMutation.isPending ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <Star className={`h-3.5 w-3.5 ml-1 ${provider.isDefault ? 'text-amber-500 fill-amber-500' : ''}`} />
                    )}
                    {provider.isDefault ? (t('providers.unset_default') || 'Unset') : (t('providers.set_default') || 'Default')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-xs h-8 text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/20"
                    onClick={() => setDeleteTarget({ id: provider.id, name: provider.name, channel: provider.channel })}
                  >
                    <Trash2 className="h-3.5 w-3.5 ml-1" />
                    {t('common.delete')}
                  </Button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <EmptyState
            icon={Wifi}
            title={t('providers.no_providers')}
            description="Configure one or more providers to enable notification delivery through different channels."
            actionLabel={t('providers.create') || 'Create Provider'}
            onAction={() => router.push(`/${locale}/providers/new`)}
            tips={[
              'SMS providers: Kavenegar, Twilio, or custom SMPP',
              'Email providers: SMTP server, SendGrid, or custom API',
              'Push providers: Firebase Cloud Messaging (FCM) or APNs',
              'Each provider can be tested before enabling in production',
            ]}
          />
        )}
      </div>

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
    </PageContainer>
  );
}
