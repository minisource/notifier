'use client';

import { useTranslations } from 'next-intl';
import { useRouter, useParams } from 'next/navigation';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { CardSkeleton } from '@/components/shared/loading-state';
import { ErrorState } from '@/components/shared/error-state';
import { DeleteProviderDialog } from '@/features/providers/components/delete-provider-dialog';
import { ProviderTestDialog } from '@/features/providers/components/provider-test-dialog';
import { useProvider, useSetDefaultProvider } from '@/features/providers/hooks/use-providers';
import { useState } from 'react';
import { ArrowLeft, Edit, Trash2, TestTube, Star, Loader2, CheckCircle, XCircle, Ban } from 'lucide-react';
import { toast } from 'sonner';

interface ProviderDetailProps {
  providerId: string;
}

export function ProviderDetail({ providerId }: ProviderDetailProps) {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const { data: provider, isLoading, isError, error, refetch } = useProvider(providerId);
  const setDefaultMutation = useSetDefaultProvider();
  const [showDelete, setShowDelete] = useState(false);
  const [showTest, setShowTest] = useState(false);

  if (isLoading) return <CardSkeleton cards={3} context="providers" />;

  if (isError) {
    return (
      <PageContainer>
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || 'Failed to load provider'}
          onRetry={() => refetch()}
        />
      </PageContainer>
    );
  }

  if (!provider) {
    return (
      <PageContainer>
        <ErrorState title={t('providers.not_found') || 'Provider not found'} message="The requested provider does not exist." />
      </PageContainer>
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
      case 'disabled': return <Badge variant="outline"><Ban className="h-3 w-3 ml-1" /> Disabled</Badge>;
      case 'error': return <Badge variant="destructive"><XCircle className="h-3 w-3 ml-1" /> Error</Badge>;
      default: return <Badge variant="outline">{status}</Badge>;
    }
  };

  return (
    <PageContainer>
      <PageHeader title={provider.name}>
        <Button variant="ghost" onClick={() => router.push(`/${locale}/providers`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <div className="space-y-5">
        {/* Overview Cards */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
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
              <Button variant="default" onClick={() => router.push(`/${locale}/providers/${providerId}/edit`)}>
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
    </PageContainer>
  );
}
