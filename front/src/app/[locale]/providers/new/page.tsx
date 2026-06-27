'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { ProviderForm } from '@/features/providers/components/provider-form';
import { useCreateProvider } from '@/features/providers/hooks/use-providers';
import { toast } from 'sonner';

export default function NewProviderPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const createMutation = useCreateProvider();

  const handleSave = async (data: { name: string; channel: string; type: string; status: string; priority: number; isDefault: boolean; description: string; configJson: string; secretConfigJson: string }) => {
    try {
      await createMutation.mutateAsync({
        name: data.name,
        channel: data.channel,
        type: data.type,
        status: data.status as any,
        priority: data.priority,
        isDefault: data.isDefault,
        description: data.description || undefined,
        config: JSON.parse(data.configJson || '{}'),
        secretConfig: JSON.parse(data.secretConfigJson || '{}'),
      });
      toast.success(t('common.saved'));
      router.push(`/${locale}/providers`);
    } catch (err: any) {
      toast.error(err?.message || t('errors.generic'));
    }
  };

  return (
    <PageContainer>
      <PageHeader title={t('providers.new_title') || 'Create Provider'}>
        <Button variant="ghost" onClick={() => router.push(`/${locale}/providers`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>
      <ProviderForm onSave={handleSave} saving={createMutation.isPending} mode="create" />
    </PageContainer>
  );
}
