'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { toast } from 'sonner';
import { ArrowLeft } from 'lucide-react';
import { createReminder } from '@/features/reminders/api';
import type { CreateReminderInput } from '@/features/reminders/types';

const CHANNELS = ['email', 'sms', 'push', 'in_app'] as const;

export default function NewReminderPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';

  const [userId, setUserId] = useState('user-mock-001');
  const [type, setType] = useState('email');
  const [recipientEmail, setRecipientEmail] = useState('');
  const [recipientPhone, setRecipientPhone] = useState('');
  const [templateKey, setTemplateKey] = useState('');
  const [scheduledAt, setScheduledAt] = useState('');
  const [saving, setSaving] = useState(false);

  const handleSubmit = async () => {
    if (!scheduledAt) {
      toast.error(t('forms.required'));
      return;
    }

    setSaving(true);
    try {
      const input: CreateReminderInput = {
        userId: userId.trim(),
        type,
        recipientEmail: recipientEmail.trim() || undefined,
        recipientPhone: recipientPhone.trim() || undefined,
        templateKey: templateKey.trim() || undefined,
        scheduledAt: new Date(scheduledAt).toISOString(),
      };
      await createReminder(input);
      toast.success(t('reminders.title') as string, { description: t('reminders.schedule') as string });
      router.push(`/${locale}/reminders`);
    } catch {
      toast.error(t('errors.generic'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageContainer>
      <PageHeader title={t('reminders.new_title')}>
        <Button variant="ghost" onClick={() => router.push(`/${locale}/reminders`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <Card>
        <CardHeader>
          <CardTitle>{t('reminders.new_title')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* User */}
          <div className="space-y-2">
            <Label>User ID</Label>
            <Input value={userId} onChange={(e) => setUserId(e.target.value)} placeholder="user-mock-001" />
          </div>

          {/* Type */}
          <div className="space-y-2">
            <Label>{t('common.type')}</Label>
            <Select value={type} onValueChange={setType}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {CHANNELS.map(ch => (
                  <SelectItem key={ch} value={ch}>{t(`channels.${ch}`)}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Recipient */}
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>{t('notifications.recipient')} Email</Label>
              <Input value={recipientEmail} onChange={(e) => setRecipientEmail(e.target.value)} placeholder="user@example.com" type="email" />
            </div>
            <div className="space-y-2">
              <Label>{t('notifications.recipient')} Phone</Label>
              <Input value={recipientPhone} onChange={(e) => setRecipientPhone(e.target.value)} placeholder="+989121234567" />
            </div>
          </div>

          {/* Template Key */}
          <div className="space-y-2">
            <Label>{t('templates.key')}</Label>
            <Input value={templateKey} onChange={(e) => setTemplateKey(e.target.value)} placeholder="e.g., auth.otp.sms" />
          </div>

          {/* Scheduled At */}
          <div className="space-y-2">
            <Label>{t('reminders.scheduled_at')} *</Label>
            <Input
              type="datetime-local"
              value={scheduledAt}
              onChange={(e) => setScheduledAt(e.target.value)}
              className="w-full sm:w-72"
            />
          </div>

          {/* Actions */}
          <div className="flex items-center gap-3 pt-2" dir={isRtl ? 'rtl' : 'ltr'}>
            <Button onClick={handleSubmit} disabled={saving || !scheduledAt}>
              {saving ? t('common.loading') : t('reminders.schedule')}
            </Button>
            <Button variant="outline" onClick={() => router.push(`/${locale}/reminders`)}>
              {t('common.cancel')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </PageContainer>
  );
}
