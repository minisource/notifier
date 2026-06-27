'use client';

import { useTranslations } from 'next-intl';
import { ShieldAlert, ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useRouter, useParams } from 'next/navigation';

interface ForbiddenStateProps {
  message?: string;
}

export function ForbiddenState({ message }: ForbiddenStateProps) {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';

  return (
    <div className="flex flex-col items-center justify-center py-20">
      <div className="flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
        <ShieldAlert className="h-8 w-8 text-destructive" />
      </div>
      <h2 className="mt-4 text-lg font-semibold">{t('errors.forbidden')}</h2>
      <p className="mt-2 text-sm text-muted-foreground text-center max-w-md">
        {message || t('errors.forbidden_description') || 'You do not have permission to access this page.'}
      </p>
      <Button variant="outline" className="mt-6" onClick={() => router.push(`/${locale}/dashboard`)}>
        <ArrowLeft className="ml-2 h-4 w-4" />
        {t('common.back')}
      </Button>
    </div>
  );
}
