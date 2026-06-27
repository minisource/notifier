'use client';

import { useTranslations } from 'next-intl';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SendNotificationForm } from '@/features/notifications/components/send-notification-form';

export default function NewNotificationPage() {
  const t = useTranslations();

  return (
    <PageContainer>
      <PageHeader title={t('notifications.new_title')} subtitle={t('notifications.form.subtitle')} />
      <SendNotificationForm />
    </PageContainer>
  );
}
