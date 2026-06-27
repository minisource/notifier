'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { Button } from '@/components/ui/button';

import { CardSkeleton } from '@/components/shared/loading-state';
import { ErrorState } from '@/components/shared/error-state';
import { ProviderForm } from '@/features/providers/components/provider-form';
import { useProvider, useUpdateProvider } from '@/features/providers/hooks/use-providers';
import { ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';
import { useEffect, useState } from 'react';

export default function EditProviderPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const providerId = params?.providerId as string;
  const { data: provider, isLoading, isError, error } = useProvider(providerId);
  const updateMutation = useUpdateProvider();
  const [initialData, setInitialData] = useState<any>(null);

  useEffect(() => {
    if (provider) {
      setInitialData({
        name: provider.name,
        channel: provider.channel,
        type: provider.type || '',
        status: provider.status,
        priority: provider.priority,
        isDefault: provider.isDefault,
        description: provider.description || '',
        configJson: provider.config ? JSON.stringify(provider.config, null, 2) : '{}',
        secretConfigJson: '{}',
      });
    }
  }, [provider]);

  const handleSave = async (data: { name: string; channel: string; type: string; status: string; priority: number; isDefault: boolean; description: string; configJson: string; secretConfigJson: string }) => {
    try {
      await updateMutation.mutateAsync({
        id: providerId,
        input: {
          name: data.name,
          channel: data.channel,
          type: data.type,
          status: data.status as any,
          priority: data.priority,
          isDefault: data.isDefault,
          description: data.description || undefined,
          config: JSON.parse(data.configJson || '{}'),
          secretConfig: data.secretConfigJson && data.secretConfigJson !== '{}'
            ? JSON.parse(data.secretConfigJson)
            : undefined,
        },
      });
      toast.success(t('common.saved'));
      router.push(`/${locale}/providers/${providerId}`);
    } catch (err: any) {
      toast.error(err?.message || t('errors.generic'));
    }
  };

  if (isLoading) return <PageContainer><CardSkeleton cards={3} context="providers" /></PageContainer>;

  if (isError || !provider) {
    return (
      <PageContainer>
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || 'Failed to load provider'}
        />
      </PageContainer>
    );
  }

  if (!initialData) return null;

  return (
    <PageContainer>
      <PageHeader title={t('providers.edit_title') || 'Edit Provider'}>
        <Button variant="ghost" onClick={() => router.push(`/${locale}/providers/${providerId}`)}>
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
    </PageContainer>
  );
}
