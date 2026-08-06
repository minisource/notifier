'use client';

import { useTranslations } from 'next-intl';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import { useState, useMemo, useEffect, useRef } from 'react';
import { z } from 'zod';
import { Button } from '@minisource/ui';
import { Textarea } from '@minisource/ui';
import { Label } from '@minisource/ui';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@minisource/ui';
import { Switch } from '@minisource/ui';
import { Card, CardContent, CardHeader, CardTitle } from '@minisource/ui';
import { TemplateKeyCombobox } from './template-key-combobox';
import { VariablesEditor } from './variables-editor';
import { useTemplatesForSelect } from '../hooks/use-notifications';
import { useSendNotification } from '../hooks/use-notifications';
import { useAdminProviders } from '@/features/notifier/api/notifier-queries';
import { Form, FormField, FormInput, FormMessage } from '@minisource/rhf';
import { Send, ArrowLeft, Loader2, Calendar, Server, CheckCircle2, AlertTriangle, XCircle, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { NotificationChannel } from '../types';
import type { Provider } from '@/features/notifier/api/notifier-types';
import { SecureUserPicker } from '@/features/test-center/components/secure-user-picker';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { ALL_TENANTS } from '@/stores/tenant.store';

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
  const router = useRouter();
  const sendMutation = useSendNotification();
  const { data: templates, isLoading: templatesLoading } = useTemplatesForSelect();
  const { tenants, activeTenant } = useTenants();
  const [selectedTenantId, setSelectedTenantId] = useState<string>(() =>
    activeTenant && activeTenant.id !== ALL_TENANTS.id ? activeTenant.id : ALL_TENANTS.id
  );
  const { data: providers, isLoading: providersLoading } = useAdminProviders(selectedTenantId);
  const [selectedChannel, setSelectedChannel] = useState<NotificationChannel>('email');
  const [templateVariables, setTemplateVariables] = useState<Record<string, string>>({});
  const [selectedTemplateId, setSelectedTemplateId] = useState<string | undefined>();
  const [selectedProviderId, setSelectedProviderId] = useState<string | undefined>();
  const [showScheduling, setShowScheduling] = useState(false);

  // Providers configured for the currently selected channel (backend-driven),
  // ordered by default/primary first, then priority (this is the failover order
  // the backend follows when a send fails).
  const channelProviders = useMemo(
    () =>
      (providers || [])
        .filter((p) => p.channel === selectedChannel)
        .sort((a, b) => {
          const aDefault = a.isDefault || a.isPrimary ? 0 : 1;
          const bDefault = b.isDefault || b.isPrimary ? 0 : 1;
          if (aDefault !== bDefault) return aDefault - bDefault;
          return (a.priority ?? 99) - (b.priority ?? 99);
        }),
    [providers, selectedChannel]
  );
  const enabledChannelProviders = channelProviders.filter((p) => p.isEnabled !== false);

  // Default/primary provider for a given channel (shared by channel change + initial load)
  const pickDefaultProvider = (channel: NotificationChannel) => {
    const forChannel = (providers || []).filter((p) => p.channel === channel);
    const enabled = forChannel.filter((p) => p.isEnabled !== false);
    return enabled.find((p) => p.isDefault || p.isPrimary) || enabled[0];
  };

  // Auto-select the channel's default provider exactly once when providers first
  // arrive. Guarded by a ref so an explicit user choice (including "Auto") is
  // never overridden afterwards.
  const didAutoSelectRef = useRef(false);
  useEffect(() => {
    if (providersLoading || didAutoSelectRef.current) return;
    const primary = pickDefaultProvider(selectedChannel);
    if (primary) {
      setSelectedProviderId(primary.id);
      didAutoSelectRef.current = true;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [providersLoading, channelProviders]);

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
      locale: 'en',
      scheduledAt: '',
      idempotencyKey: '',
    },
  });

  // Update recipient type + provider when channel changes
  const handleTenantChange = (value: string) => {
    setSelectedTenantId(value);
    // Provider availability depends on the tenant scope — reset the selection
    // and let the auto-select effect pick the new tenant's default provider.
    setSelectedProviderId(undefined);
    didAutoSelectRef.current = false;
    form.clearErrors();
  };

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

    // Auto-select the channel's default/primary provider if one is configured
    setSelectedProviderId(pickDefaultProvider(channel)?.id);
  };

  const handleTemplateChange = (templateId: string) => {
    setSelectedTemplateId(templateId);
    const tmpl = templates?.find((t) => t.id === templateId);
    if (!tmpl) return;
    // Prefill subject/body from the selected template when not already set
    if (tmpl.body) {
      form.setValue('body', tmpl.body);
    }
    if (tmpl.subject) {
      form.setValue('subject', tmpl.subject);
    }
    // Pre-populate variables editor with the template's declared variables
    if (tmpl.variables && tmpl.variables.length > 0) {
      const nextVars: Record<string, string> = {};
      for (const v of tmpl.variables) nextVars[v] = templateVariables[v] ?? '';
      setTemplateVariables(nextVars);
    }
  };

  // Map backend provider status ('active'|'inactive'|'disabled'|'error') and
  // statuses ('healthy'|'degraded'|'down') to a consistent visual state.
  const providerStatus = (p: Provider) => {
    const status = (p.status || '').toLowerCase();
    if (p.isEnabled === false || status === 'disabled' || status === 'inactive') {
      return { label: t('providers.disabled'), className: 'bg-muted text-muted-foreground border-border/60', Icon: XCircle };
    }
    if (status === 'error' || status === 'down') {
      return { label: t('providers.down'), className: 'bg-destructive/10 text-destructive border-destructive/30', Icon: XCircle };
    }
    if (status === 'degraded') {
      return { label: t('providers.degraded'), className: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30', Icon: AlertTriangle };
    }
    return { label: t('providers.healthy'), className: 'bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/30', Icon: CheckCircle2 };
  };

  const onSubmit = form.handleSubmit(async (data) => {
    const channel = data.channel as NotificationChannel;

    const recipient: Record<string, string | undefined> = {};
    if (channel === 'sms') recipient.phone = data.recipientValue;
    else if (channel === 'email') recipient.email = data.recipientValue;
    else if (channel === 'in_app') recipient.userId = data.recipientValue;
    else if (channel === 'push') recipient.userId = data.recipientValue;
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
      locale: data.locale || 'en',
      templateId: selectedTemplateId || undefined,
      providerId: selectedProviderId || undefined,
      scheduledAt: data.scheduledAt || undefined,
      idempotencyKey: data.idempotencyKey || undefined,
      // Always send the explicit tenant scope: "all" means global (the backend
      // maps it to nil and ignores the X-Tenant-Id header), a UUID scopes the
      // notification to that tenant.
      tenantId: selectedTenantId,
    };

    if (Object.keys(templateVariables).length > 0) {
      payload.variables = templateVariables;
    }

    sendMutation.mutate(payload as any, {
      onSuccess: (result) => {
        router.push(`/notifications/${result.id}`);
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
    <Form form={form} onSubmit={onSubmit} className="space-y-6">
      {/* Channel & Recipient */}
      <Card><CardHeader><CardTitle>{t('notifications.form.recipient_section')}</CardTitle></CardHeader><CardContent>
        <div className="space-y-4">
        {/* Tenant / Project scope — mirrors the provider form so a notification
            can be sent on behalf of a specific tenant and its tenant-scoped
            providers are used during delivery. */}
        <div className="space-y-2">
          <Label>{t('providers.tenant') || 'Tenant / Project'}</Label>
          <Select value={selectedTenantId} onValueChange={handleTenantChange}>
            <SelectTrigger className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_TENANTS.id}>
                {t('providers.tenant_global') || 'All Tenants (Global)'}
              </SelectItem>
              {tenants.map((tenant) => (
                <SelectItem key={tenant.id} value={tenant.id}>
                  {tenant.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">
            {t('notifications.form.tenant_hint') ||
              'Choose the tenant/project this notification belongs to. Only providers available to that tenant (its own + global) will be used.'}
          </p>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField
            name="channel"
            label={t('notifications.channel')}
            error={form.formState.errors.channel?.message}
          >
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
            <FormMessage />
          </FormField>
          <FormField
            name="recipientValue"
            label={t('notifications.form.recipient_value')}
            error={form.formState.errors.recipientValue?.message}
            required
          >
            {selectedChannel === 'in_app' || selectedChannel === 'push' ? (
              <SecureUserPicker
                value={form.watch('recipientValue')}
                onChange={(userId) => {
                  form.setValue('recipientValue', userId || '');
                  form.trigger('recipientValue');
                }}
                placeholder={selectedChannel === 'in_app' ? 'Select recipient user...' : 'Select push target user...'}
              />
            ) : (
              <FormInput
                placeholder={channelPlaceholders[selectedChannel]}
                {...form.register('recipientValue')}
              />
            )}
            <FormMessage />
          </FormField>
        </div>
        </div>
      </CardContent></Card>

      {/* Message Content */}
      <Card><CardHeader><CardTitle>{t('notifications.form.content_section')}</CardTitle></CardHeader><CardContent>
        <div className="space-y-4">
          <FormField
            name="subject"
            label={t('notifications.subject')}
            error={form.formState.errors.subject?.message}
          >
            <FormInput
              placeholder={t('notifications.form.subject_placeholder')}
              {...form.register('subject')}
            />
            <FormMessage />
          </FormField>
          <FormField
            name="body"
            label={t('notifications.body')}
            required
            error={form.formState.errors.body?.message}
          >
            <Textarea
              placeholder={t('notifications.form.body_placeholder')}
              
              {...form.register('body')}
              className={form.formState.errors.body ? 'border-destructive focus-visible:ring-destructive' : ''}
            />
            <FormMessage />
          </FormField>
        </div>
      </CardContent></Card>

      {/* Template */}
      <Card><CardHeader><CardTitle>{t('notifications.template')}</CardTitle></CardHeader><CardContent>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>{t('notifications.form.select_template')}</Label>
            <TemplateKeyCombobox
              templates={templates || []}
              value={selectedTemplateId}
              onChange={handleTemplateChange}
              loading={templatesLoading}
            />
            {selectedTemplateId && templates?.find((t) => t.id === selectedTemplateId)?.body && (
              <p className="text-xs text-muted-foreground">
                {t('notifications.form.template_prefilled')}
              </p>
            )}
          </div>
          <div className="space-y-2">
            <Label>{t('templates.variables')}</Label>
            <VariablesEditor
              variables={templateVariables}
              onChange={setTemplateVariables}
            />
          </div>
        </div>
      </CardContent></Card>

      {/* Provider */}
      <Card><CardHeader><CardTitle>{t('notifications.form.provider_section')}</CardTitle></CardHeader><CardContent>
        <div className="space-y-3">
          <div className="space-y-2">
            <Label>{t('notifications.form.select_provider')}</Label>
            <Select
              value={selectedProviderId ?? 'auto'}
              onValueChange={(v) => setSelectedProviderId(v === 'auto' ? undefined : v)}
              disabled={providersLoading}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('notifications.form.auto_provider')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="auto">{t('notifications.form.auto_failover')}</SelectItem>
                {channelProviders.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    <span className="flex items-center gap-2">
                      <Server className="h-3.5 w-3.5 text-muted-foreground" />
                      {p.name}
                      {p.isDefault && <span className="text-xs text-muted-foreground">({t('providers.default')})</span>}
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {t('notifications.form.provider_hint', { channel: t(`channels.${selectedChannel}`) })}
            </p>
          </div>

          {/* Available providers for this channel, straight from the backend */}
          <div className="space-y-2">
            <Label>{t('notifications.form.available_providers')}</Label>
            {providersLoading ? (
              <p className="text-xs text-muted-foreground">{t('common.loading')}</p>
            ) : enabledChannelProviders.length === 0 ? (
              <p className="text-xs text-muted-foreground">{t('notifications.form.no_providers_channel')}</p>
            ) : (
              <div className="space-y-2">
                {/* Failover order indicator */}
                {enabledChannelProviders.length > 1 && (
                  <div className="rounded-md border bg-muted/40 px-3 py-2 text-xs">
                    <p className="flex items-center gap-1.5 font-medium text-muted-foreground">
                      <RefreshCw className="h-3 w-3" />
                      {t('notifications.form.failover_order')}:{' '}
                      {enabledChannelProviders.map((p, idx) => (
                        <span key={p.id} className="font-semibold text-foreground">
                          {idx > 0 && <span className="text-muted-foreground/60"> → </span>}
                          {p.name}
                        </span>
                      ))}
                    </p>
                    <p className="mt-0.5 text-muted-foreground">{t('notifications.form.failover_hint')}</p>
                  </div>
                )}

                <div className="flex flex-wrap gap-2">
                  {enabledChannelProviders.map((p, idx) => {
                    const st = providerStatus(p);
                    const Icon = st.Icon;
                    return (
                      <button
                        key={p.id}
                        type="button"
                        onClick={() => setSelectedProviderId(p.id)}
                        className={cn(
                          'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors',
                          st.className,
                          selectedProviderId === p.id && 'ring-2 ring-ring ring-offset-1'
                        )}
                        title={`${p.name} — ${st.label}${p.successRate != null ? ` · ${p.successRate}%` : ''}`}
                      >
                        <Icon className="h-3 w-3" />
                        {p.name}
                        {p.isDefault && <span className="opacity-70">· {t('providers.default')}</span>}
                        {enabledChannelProviders.length > 1 && (
                          <span className="opacity-60">· {idx + 1}</span>
                        )}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      </CardContent></Card>

      {/* Options */}
      <Card><CardHeader><CardTitle>{t('notifications.form.options_section')}</CardTitle></CardHeader><CardContent>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField
            name="priority"
            label={t('notifications.priority')}
            error={form.formState.errors.priority?.message}
          >
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
            <FormMessage />
          </FormField>
          <FormField
            name="locale"
            label={t('notifications.form.locale')}
            error={form.formState.errors.locale?.message}
          >
            <Select value={form.watch('locale')} onValueChange={(v) => form.setValue('locale', v)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="fa">{t('settings.language')}: فارسی</SelectItem>
                <SelectItem value="en">Language: English</SelectItem>
              </SelectContent>
            </Select>
            <FormMessage />
          </FormField>
        </div>

        <div className="mt-4 flex items-center gap-2">
          <Switch checked={showScheduling} onCheckedChange={setShowScheduling} />
          <Label className="cursor-pointer">{t('notifications.form.schedule_later')}</Label>
        </div>

        {showScheduling && (
          <div className="mt-3 grid gap-4 sm:grid-cols-2">
            <FormField
              name="scheduledAt"
              label={t('notifications.scheduled_at')}
              error={form.formState.errors.scheduledAt?.message}
            >
              <div className="relative">
                <Calendar className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground" />
                <FormInput
                  type="datetime-local"
                  {...form.register('scheduledAt')}
                  className="pr-10"
                />
              </div>
              <FormMessage />
            </FormField>
            <FormField
              name="idempotencyKey"
              label={t('notifications.form.idempotency_key')}
              error={form.formState.errors.idempotencyKey?.message}
            >
              <FormInput
                placeholder="optional-unique-key"
                {...form.register('idempotencyKey')}
              />
              <FormMessage />
            </FormField>
          </div>
        )}
      </CardContent></Card>

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
    </Form>
  );
}
