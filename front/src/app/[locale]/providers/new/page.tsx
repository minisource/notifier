'use client';

import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { PageHeader } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { ArrowLeft } from 'lucide-react';
import { ProviderForm, type ProviderFormData } from '@/features/providers/components/provider-form';
import { useCreateProvider } from '@/features/providers/hooks/use-providers';
import { ALL_TENANTS } from '@/stores/tenant.store';
import { toast } from 'sonner';

export default function NewProviderPage() {
  const t = useTranslations();
  const router = useRouter();
  const createMutation = useCreateProvider();

  const handleSave = async (data: ProviderFormData) => {
    try {
      await createMutation.mutateAsync({
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
      });
      toast.success(t('common.saved'));
      router.push(`/providers`);
    } catch (err: any) {
      toast.error(err?.message || t('errors.generic'));
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title={t('providers.new_title') || 'Create Provider'}>
        <Button variant="ghost" onClick={() => router.push(`/providers`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>
      <ProviderForm onSave={handleSave} saving={createMutation.isPending} mode="create" />
    </div>
  );
}
