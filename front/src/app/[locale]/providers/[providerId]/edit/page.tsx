'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@minisource/ui';
import { Button } from '@minisource/ui';

import { Skeleton } from '@minisource/ui';
import { ErrorState } from '@minisource/ui';
import { ProviderForm, type ProviderFormData } from '@/features/providers/components/provider-form';
import { useProvider, useUpdateProvider } from '@/features/providers/hooks/use-providers';
import { ALL_TENANTS } from '@/stores/tenant.store';
import { ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';
import { useEffect, useState } from 'react';

export default function EditProviderPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const providerId = params?.providerId as string;
  const { data: provider, isLoading, isError, error } = useProvider(providerId);
  const updateMutation = useUpdateProvider();
  const [initialData, setInitialData] = useState<any>(null);

  useEffect(() => {
    if (provider) {
      setInitialData({
        tenantId: provider.tenantId || ALL_TENANTS.id,
        name: provider.name,
        channel: provider.channel,
        type: provider.type || '',
        status: provider.status,
        priority: provider.priority,
        isDefault: provider.isDefault,
        description: provider.description || '',
        config: provider.config || {},
      });
    }
  }, [provider]);

  const handleSave = async (data: ProviderFormData) => {
    try {
      await updateMutation.mutateAsync({
        id: providerId,
        input: {
          // Pass the explicit 'all' sentinel through so the backend treats a
          // deliberate "Global" choice as authoritative (and does NOT fall back
          // to the active tenant's X-Tenant-Id header).
          tenantId: data.tenantId || ALL_TENANTS.id,
          name: data.name,
          channel: data.channel,
          type: data.type,
          status: data.status as any,
          priority: data.priority,
          isDefault: data.isDefault,
          description: data.description || undefined,
          config: data.config,
          secretConfig: Object.keys(data.secretConfig || {}).length > 0 ? data.secretConfig : undefined,
        },
      });
      toast.success(t('common.saved'));
      router.push(`/providers/${providerId}`);
    } catch (err: any) {
      toast.error(err?.message || t('errors.generic'));
    }
  };

  if (isLoading) return <div className="space-y-6"><Skeleton className="h-64 w-full" /></div>;

  if (isError || !provider) {
    return (
      <div className="space-y-6">
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || 'Failed to load provider'}
        />
      </div>
    );
  }

  if (!initialData) return null;

  return (
    <div className="space-y-6">
      <PageHeader title={t('providers.edit_title') || 'Edit Provider'}>
        <Button variant="ghost" onClick={() => router.push(`/providers/${providerId}`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>
      <ProviderForm
        initialData={initialData}
        onSave={handleSave}
        saving={updateMutation.isPending}
        mode="edit"
      />
    </div>
  );
}
