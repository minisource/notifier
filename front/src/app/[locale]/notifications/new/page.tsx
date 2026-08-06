'use client';

import { useTranslations } from 'next-intl';
import { PageHeader } from '@minisource/ui';
import { SendNotificationForm } from '@/features/notifications/components/send-notification-form';

export default function NewNotificationPage() {
  const t = useTranslations();

  return (
    <div className="space-y-6">
      <PageHeader title={t('notifications.new_title')} description={t('notifications.form.subtitle')} />
      <SendNotificationForm />
    </div>
  );
}
