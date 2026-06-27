'use client';

import { useTranslations } from 'next-intl';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useParams, useRouter } from 'next/navigation';
import { useState, useMemo } from 'react';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { SectionCard } from '@/components/shared/section-card';
import { TemplateKeyCombobox } from './template-key-combobox';
import { VariablesEditor } from './variables-editor';
import { useTemplatesForSelect } from '../hooks/use-notifications';
import { useSendNotification } from '../hooks/use-notifications';
import { Send, ArrowLeft, Loader2, Calendar, Globe, Layers, Hash } from 'lucide-react';
import type { NotificationChannel } from '../types';

const channels: { value: NotificationChannel; label: string }[] = [
  { value: 'sms', label: 'SMS' },
  { value: 'email', label: 'Email' },
  { value: 'push', label: 'Push' },
  { value: 'in_app', label: 'In-App' },
  { value: 'webhook', label: 'Webhook' },
];

const priorities = [
  { value: 'low', label: 'Low' },
  { value: 'normal', label: 'Normal' },
  { value: 'high', label: 'High' },
  { value: 'urgent', label: 'Urgent' },
];

export function SendNotificationForm() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';
  const sendMutation = useSendNotification();
  const { data: templates, isLoading: templatesLoading } = useTemplatesForSelect();
  const [selectedChannel, setSelectedChannel] = useState<NotificationChannel>('email');
  const [templateVariables, setTemplateVariables] = useState<Record<string, string>>({});
  const [selectedTemplateId, setSelectedTemplateId] = useState<string | undefined>();
  const [showScheduling, setShowScheduling] = useState(false);

  // Build dynamic schema based on channel
  const buildSchema = (channel: NotificationChannel) => {
    let schema = z.object({
      channel: z.string().min(1),
      recipientType: z.string().min(1),
      recipientValue: z.string().min(1, t('forms.required')),
      subject: z.string().optional(),
      body: z.string().min(1, t('forms.required')),
      priority: z.string().optional(),
      locale: z.string().optional(),
      scheduledAt: z.string().optional(),
      idempotencyKey: z.string().optional(),
    });

    // Add channel-specific validation
    if (channel === 'email') {
      schema = schema.extend({
        recipientValue: z.string().email(t('forms.invalid_email')),
      });
    } else if (channel === 'sms') {
      schema = schema.extend({
        recipientValue: z.string().min(5, t('forms.invalid_phone')),
      });
    }

    return schema;
  };

  const schema = useMemo(() => buildSchema(selectedChannel), [selectedChannel]);

  const form = useForm({
    resolver: zodResolver(schema),
    defaultValues: {
      channel: 'email',
      recipientType: 'email',
      recipientValue: '',
      subject: '',
      body: '',
      priority: 'normal',
      locale: 'fa',
      scheduledAt: '',
      idempotencyKey: '',
    },
  });

  // Update recipient type when channel changes
  const handleChannelChange = (value: string) => {
    const channel = value as NotificationChannel;
    setSelectedChannel(channel);
    form.setValue('channel', channel);

    const typeMap: Record<NotificationChannel, string> = {
      email: 'email',
      sms: 'phone',
      push: 'device_token',
      in_app: 'user_id',
      webhook: 'webhook_url',
    };
    const rType = typeMap[channel] || 'email';
    form.setValue('recipientType', rType);
    form.setValue('recipientValue', '');
    form.clearErrors();
  };

  const onSubmit = form.handleSubmit(async (data) => {
    const channel = data.channel as NotificationChannel;

    const recipient: Record<string, string | undefined> = {};
    if (channel === 'sms') recipient.phone = data.recipientValue;
    else if (channel === 'email') recipient.email = data.recipientValue;
    else if (channel === 'in_app') recipient.userId = data.recipientValue;
    else if (channel === 'push') recipient.deviceToken = data.recipientValue;
    else if (channel === 'webhook') recipient.webhookUrl = data.recipientValue;

    // Only include non-empty recipient fields
    const cleanRecipient = Object.fromEntries(
      Object.entries(recipient).filter(([_, v]) => v !== undefined && v !== '')
    );

    const payload: Record<string, unknown> = {
      channel,
      priority: (data.priority as 'low' | 'normal' | 'high' | 'urgent') || 'normal',
      recipient: Object.keys(cleanRecipient).length > 0 ? cleanRecipient : undefined,
      subject: data.subject || undefined,
      body: data.body,
      locale: data.locale || 'fa',
      templateId: selectedTemplateId || undefined,
      scheduledAt: data.scheduledAt || undefined,
      idempotencyKey: data.idempotencyKey || undefined,
    };

    if (Object.keys(templateVariables).length > 0) {
      payload.variables = templateVariables;
    }

    sendMutation.mutate(payload as any, {
      onSuccess: (result) => {
        router.push(`/${locale}/notifications/${result.id}`);
      },
    });
  });

  const channelPlaceholders: Record<NotificationChannel, string> = {
    email: 'user@example.com',
    sms: '+989121234567',
    push: 'device-token-...',
    in_app: 'user-id-...',
    webhook: 'https://api.example.com/hooks',
  };

  return (
    <form onSubmit={onSubmit} className="space-y-6">
      {/* Channel & Recipient */}
      <SectionCard title={t('notifications.form.recipient_section')} icon={Send}>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>{t('notifications.channel')}</Label>
            <Select value={form.watch('channel')} onValueChange={handleChannelChange}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {channels.map(ch => (
                  <SelectItem key={ch.value} value={ch.value}>{t(`channels.${ch.value}`)}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>{t('notifications.form.recipient_value')}</Label>
            <Input
              placeholder={channelPlaceholders[selectedChannel]}
              {...form.register('recipientValue')}
              className={form.formState.errors.recipientValue ? 'border-destructive' : ''}
            />
            {form.formState.errors.recipientValue && (
              <p className="text-xs text-destructive">{form.formState.errors.recipientValue.message}</p>
            )}
          </div>
        </div>
      </SectionCard>

      {/* Message Content */}
      <SectionCard title={t('notifications.form.content_section')} icon={Layers}>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>{t('notifications.subject')}</Label>
            <Input
              placeholder={t('notifications.form.subject_placeholder')}
              {...form.register('subject')}
            />
          </div>
          <div className="space-y-2">
            <Label>{t('notifications.body')} *</Label>
            <Textarea
              placeholder={t('notifications.form.body_placeholder')}
              rows={4}
              {...form.register('body')}
              className={form.formState.errors.body ? 'border-destructive' : ''}
            />
            {form.formState.errors.body && (
              <p className="text-xs text-destructive">{form.formState.errors.body.message}</p>
            )}
          </div>
        </div>
      </SectionCard>

      {/* Template */}
      <SectionCard title={t('notifications.template')} icon={Hash}>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>{t('notifications.form.select_template')}</Label>
            <TemplateKeyCombobox
              templates={templates || []}
              value={selectedTemplateId}
              onChange={(id, _key) => {
                setSelectedTemplateId(id);
              }}
              loading={templatesLoading}
            />
          </div>
          <div className="space-y-2">
            <Label>{t('templates.variables')}</Label>
            <VariablesEditor
              variables={templateVariables}
              onChange={setTemplateVariables}
            />
          </div>
        </div>
      </SectionCard>

      {/* Options */}
      <SectionCard title={t('notifications.form.options_section')} icon={Globe}>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>{t('notifications.priority')}</Label>
            <Select value={form.watch('priority')} onValueChange={(v) => form.setValue('priority', v)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {priorities.map(p => (
                  <SelectItem key={p.value} value={p.value}>
                    {t(`notifications.filters.priority_${p.value}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>{t('notifications.form.locale')}</Label>
            <Select value={form.watch('locale')} onValueChange={(v) => form.setValue('locale', v)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="fa">{t('settings.language')}: فارسی</SelectItem>
                <SelectItem value="en">Language: English</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="mt-4 flex items-center gap-2">
          <Switch checked={showScheduling} onCheckedChange={setShowScheduling} />
          <Label className="cursor-pointer">{t('notifications.form.schedule_later')}</Label>
        </div>

        {showScheduling && (
          <div className="mt-3 grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>{t('notifications.scheduled_at')}</Label>
              <div className="relative">
                <Calendar className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  type="datetime-local"
                  {...form.register('scheduledAt')}
                  className="pr-10"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>{t('notifications.form.idempotency_key')}</Label>
              <Input
                placeholder="optional-unique-key"
                {...form.register('idempotencyKey')}
              />
            </div>
          </div>
        )}
      </SectionCard>

      {/* Actions */}
      <div className="flex items-center justify-between gap-4 border-t border-border/50 pt-4">
        <Button
          type="button"
          variant="ghost"
          onClick={() => router.back()}
        >
          <ArrowLeft className="ml-1.5 h-4 w-4" />
          {t('common.back')}
        </Button>
        <Button
          type="submit"
          disabled={sendMutation.isPending}
          className="min-w-[140px]"
        >
          {sendMutation.isPending ? (
            <>
              <Loader2 className="ml-1.5 h-4 w-4 animate-spin" />
              {t('common.loading')}
            </>
          ) : (
            <>
              <Send className="ml-1.5 h-4 w-4" />
              {t('notifications.send')}
            </>
          )}
        </Button>
      </div>
    </form>
  );
}
