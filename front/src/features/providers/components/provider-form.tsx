'use client';

import { useMemo, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Card, CardContent, CardHeader, CardTitle } from '@minisource/ui';
import { Button } from '@minisource/ui';
import { Input } from '@minisource/ui';
import { Label } from '@minisource/ui';
import { Textarea } from '@minisource/ui';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@minisource/ui';
import { Switch } from '@minisource/ui';
import { Loader2, KeyRound, Eye, EyeOff } from 'lucide-react';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { ALL_TENANTS } from '@/stores/tenant.store';
import {
  PROVIDER_TYPES,
  typesForChannel,
  buildProviderConfig,
  configToFieldValues,
  fieldLabelKey,
  typeLabelKey,
  type ProviderFieldDef,
} from '../provider-schemas';

export type ProviderFormData = {
  tenantId: string;
  name: string;
  channel: string;
  type: string;
  status: string;
  priority: number;
  isDefault: boolean;
  description: string;
  config: Record<string, unknown>;
  secretConfig: Record<string, unknown>;
};

interface ProviderFormProps {
  initialData?: Partial<ProviderFormData>;
  onSave: (data: ProviderFormData) => Promise<void>;
  saving: boolean;
  mode: 'create' | 'edit';
}

const BASE_FORM = {
  tenantId: ALL_TENANTS.id,
  name: '',
  channel: 'sms',
  type: '',
  status: 'active',
  priority: 1,
  isDefault: false,
  description: '',
};

export function ProviderForm({ initialData, onSave, saving, mode }: ProviderFormProps) {
  const t = useTranslations();
  const { tenants, activeTenant } = useTenants();

  const [form, setForm] = useState<Omit<ProviderFormData, 'config' | 'secretConfig'>>({
    ...BASE_FORM,
    tenantId:
      initialData?.tenantId ??
      (activeTenant && activeTenant.id !== ALL_TENANTS.id ? activeTenant.id : ALL_TENANTS.id),
    name: initialData?.name ?? '',
    channel: initialData?.channel ?? 'sms',
    type: initialData?.type ?? '',
    status: initialData?.status ?? 'active',
    priority: initialData?.priority ?? 1,
    isDefault: initialData?.isDefault ?? false,
    description: initialData?.description ?? '',
  });

  const [fieldValues, setFieldValues] = useState<Record<string, unknown>>(() =>
    configToFieldValues(initialData?.type ?? '', initialData?.config),
  );
  // Advanced fallback: raw JSON editor for the 'custom' type
  const [customConfigJson, setCustomConfigJson] = useState(() =>
    initialData?.config && Object.keys(initialData.config).length
      ? JSON.stringify(initialData.config, null, 2)
      : '{}',
  );
  const [customSecretConfigJson, setCustomSecretConfigJson] = useState('{}');
  const [showSecrets, setShowSecrets] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const typeOptions = useMemo(() => typesForChannel(form.channel), [form.channel]);
  const typeDef = PROVIDER_TYPES[form.type];
  const typeAvailable = !form.type || typeOptions.some((o) => o.type === form.type);

  const updateField = <K extends keyof typeof form>(key: K, value: (typeof form)[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    setErrors((prev) => ({ ...prev, [key]: '' }));
  };

  const handleChannelChange = (channel: string) => {
    setForm((prev) => ({ ...prev, channel, type: '' }));
    setFieldValues({});
    setCustomConfigJson('{}');
    setCustomSecretConfigJson('{}');
    setErrors((prev) => ({ ...prev, channel: '', type: '' }));
  };

  const handleTypeChange = (type: string) => {
    setForm((prev) => ({ ...prev, type }));
    setFieldValues({});
    setCustomConfigJson('{}');
    setCustomSecretConfigJson('{}');
    setErrors((prev) => ({ ...prev, type: '' }));
  };

  const setValue = (key: string, value: unknown) => {
    setFieldValues((prev) => ({ ...prev, [key]: value }));
    setErrors((prev) => ({ ...prev, [key]: '' }));
  };

  const tField = (field: ProviderFieldDef) => {
    const key = fieldLabelKey(form.type, field.key);
    const label = t(key);
    return label === key ? field.key : label;
  };

  const tType = (type: string) => {
    const key = typeLabelKey(type);
    const label = t(key);
    return label === key ? type : label;
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!form.name.trim()) newErrors.name = 'Name is required';
    if (!form.channel) newErrors.channel = 'Channel is required';
    if (!form.type.trim()) newErrors.type = 'Provider type is required';
    if (form.priority < 0) newErrors.priority = 'Priority must be >= 0';

    if (form.type === 'custom' || !typeDef) {
      try {
        JSON.parse(customConfigJson);
      } catch {
        newErrors.customConfig = 'Invalid JSON format';
      }
      try {
        if (customSecretConfigJson.trim()) JSON.parse(customSecretConfigJson);
      } catch {
        newErrors.customSecretConfig = 'Invalid JSON format';
      }
    } else {
      for (const field of typeDef.fields) {
        if (!field.required) continue;
        // In edit mode, secret fields may stay empty to keep existing values.
        if (field.secret && mode === 'edit') continue;
        // Number fields are already converted to numbers in onChange, so an
        // empty-string check is sufficient (NaN text values are impossible).
        const v = fieldValues[field.key];
        if (v === undefined || v === null || v === '') {
          newErrors[field.key] = 'Required';
        }
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;

    let config: Record<string, unknown> = {};
    let secretConfig: Record<string, unknown> = {};
    if (form.type === 'custom' || !typeDef) {
      config = JSON.parse(customConfigJson || '{}');
      secretConfig = customSecretConfigJson.trim() ? JSON.parse(customSecretConfigJson) : {};
    } else {
      ({ config, secretConfig } = buildProviderConfig(form.type, fieldValues));
    }

    await onSave({
      ...form,
      config,
      secretConfig,
    });
  };

  const renderField = (field: ProviderFieldDef) => {
    const err = errors[field.key];
    const isSecret = !!field.secret;

    if (field.input === 'boolean') {
      return (
        <div key={field.key} className="space-y-2 sm:col-span-2">
          <div className="flex items-center justify-between rounded-lg border p-3">
            <Label>{tField(field)}</Label>
            <Switch
              checked={!!fieldValues[field.key]}
              onCheckedChange={(v) => setValue(field.key, v)}
            />
          </div>
          {err && <p className="text-xs text-red-500">{err}</p>}
        </div>
      );
    }

    const inputProps = {
      id: field.key,
      className: 'w-full',
      value: String(fieldValues[field.key] ?? ''),
      onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const raw = e.target.value;
        setValue(field.key, field.input === 'number' ? (raw === '' ? '' : Number(raw)) : raw);
      },
      placeholder: field.key,
      ...(field.input === 'number' ? { type: 'number' as const, min: 0 } : {}),
      ...(isSecret && field.input === 'password'
        ? { type: showSecrets ? ('text' as const) : ('password' as const) }
        : {}),
    };

    return (
      <div key={field.key} className="space-y-2">
        <div className="flex items-center gap-1.5">
          <Label htmlFor={field.key} className="flex items-center gap-1.5">
            {isSecret && <KeyRound className="h-3 w-3 text-muted-foreground" />}
            {tField(field)}
            {field.required && <span className="text-red-500">*</span>}
          </Label>
        </div>
        {field.input === 'textarea' ? (
          <Textarea {...inputProps} className="min-h-[80px] w-full font-mono text-xs" />
        ) : (
          <div className="relative">
            <Input {...inputProps} />
            {isSecret && field.input === 'password' && (
              <button
                type="button"
                onClick={() => setShowSecrets((s) => !s)}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                tabIndex={-1}
              >
                {showSecrets ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            )}
          </div>
        )}
        {err && <p className="text-xs text-red-500">{err}</p>}
        {isSecret && mode === 'edit' && (
          <p className="text-xs text-amber-500">
            {t('providers.secret_config_edit_hint') ||
              'Leave empty to keep existing secrets. Only provide values you want to replace.'}
          </p>
        )}
      </div>
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {mode === 'create'
            ? t('providers.new_title') || 'Create Provider'
            : t('providers.edit_title') || 'Edit Provider'}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Tenant / Project */}
        <div className="space-y-2">
          <Label>{t('providers.tenant') || 'Tenant / Project'}</Label>
          <Select value={form.tenantId} onValueChange={(v) => updateField('tenantId', v)}>
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
            {t('providers.tenant_hint') ||
              'Providers created for a specific tenant are only used by that tenant. Global providers are available to every tenant.'}
          </p>
        </div>

        {/* Name */}
        <div className="space-y-2">
          <Label>{t('providers.name') || 'Name'} *</Label>
          <Input
            value={form.name}
            onChange={(e) => updateField('name', e.target.value)}
            placeholder="e.g., Kavenegar SMS"
          />
          {errors.name && <p className="text-xs text-red-500">{errors.name}</p>}
        </div>

        {/* Channel + Type */}
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>{t('common.channel')} *</Label>
            <Select value={form.channel} onValueChange={handleChannelChange}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="sms">SMS</SelectItem>
                <SelectItem value="email">Email</SelectItem>
                <SelectItem value="push">Push</SelectItem>
                <SelectItem value="webhook">Webhook</SelectItem>
                <SelectItem value="in_app">In-App</SelectItem>
              </SelectContent>
            </Select>
            {errors.channel && <p className="text-xs text-red-500">{errors.channel}</p>}
          </div>
          <div className="space-y-2">
            <Label>{t('providers.provider_type') || 'Provider Type'} *</Label>
            <Select value={form.type} onValueChange={handleTypeChange}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {typeOptions.map((option) => (
                  <SelectItem key={option.type} value={option.type}>
                    {tType(option.type)}
                  </SelectItem>
                ))}
                {form.type && !typeAvailable && (
                  <SelectItem value={form.type}>{form.type}</SelectItem>
                )}
              </SelectContent>
            </Select>
            {errors.type && <p className="text-xs text-red-500">{errors.type}</p>}
            <p className="text-xs text-muted-foreground">
              {t('providers.type_hint') ||
                'Available provider types depend on the selected channel.'}
            </p>
          </div>
        </div>

        {/* Provider-specific configuration */}
        {form.type && (
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">
                {t('providers.config_section') || 'Provider Configuration'}
              </span>
              <span className="inline-flex items-center rounded-full border bg-muted/50 px-2 py-0.5 text-xs font-medium">
                {tType(form.type)}
              </span>
            </div>

            {form.type !== 'custom' && typeDef ? (
              typeDef.fields.length > 0 ? (
                <div className="grid gap-4 sm:grid-cols-2">
                  {typeDef.fields.map(renderField)}
                </div>
              ) : (
                <p className="text-xs text-muted-foreground">
                  {t('providers.no_config_needed') ||
                    'This provider type does not require additional configuration.'}
                </p>
              )
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>{t('providers.config') || 'Configuration (JSON)'}</Label>
                  <Textarea
                    value={customConfigJson}
                    onChange={(e) => setCustomConfigJson(e.target.value)}
                    placeholder='{"sender": "1000", "baseUrl": "https://api.example.com"}'
                    className="min-h-[100px] font-mono text-sm"
                  />
                  {errors.customConfig && (
                    <p className="text-xs text-red-500">{errors.customConfig}</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label>{t('providers.secret_config') || 'Secret Configuration (JSON)'}</Label>
                  <Textarea
                    value={customSecretConfigJson}
                    onChange={(e) => setCustomSecretConfigJson(e.target.value)}
                    placeholder='{"apiKey": "your-api-key"}'
                    className="min-h-[100px] font-mono text-sm"
                  />
                  {errors.customSecretConfig && (
                    <p className="text-xs text-red-500">{errors.customSecretConfig}</p>
                  )}
                  {mode === 'edit' && (
                    <p className="text-xs text-amber-500">
                      {t('providers.secret_config_edit_hint') ||
                        'Leave empty to keep existing secrets. Only provide values you want to replace.'}
                    </p>
                  )}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Status + Priority */}
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>{t('common.status')}</Label>
            <Select value={form.status} onValueChange={(v) => updateField('status', v)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="active">{t('settings.active') || 'Active'}</SelectItem>
                <SelectItem value="inactive">{t('settings.inactive') || 'Inactive'}</SelectItem>
                <SelectItem value="disabled">{t('settings.disabled') || 'Disabled'}</SelectItem>
                <SelectItem value="error">{t('settings.error') || 'Error'}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>{t('providers.priority') || 'Priority'}</Label>
            <Input
              type="number"
              value={form.priority}
              onChange={(e) => updateField('priority', parseInt(e.target.value) || 0)}
              min={0}
              max={100}
            />
            {errors.priority && <p className="text-xs text-red-500">{errors.priority}</p>}
          </div>
        </div>

        {/* Description */}
        <div className="space-y-2">
          <Label>{t('providers.description') || 'Description'}</Label>
          <Textarea
            value={form.description}
            onChange={(e) => updateField('description', e.target.value)}
            placeholder="Optional description"
            className="min-h-[60px]"
          />
        </div>

        {/* Is Default */}
        <div className="flex items-center justify-between rounded-lg border p-3">
          <div>
            <p className="text-sm font-medium">{t('providers.is_default') || 'Default Provider'}</p>
            <p className="text-xs text-muted-foreground">
              {form.isDefault
                ? t('providers.is_default_enabled') ||
                  'This provider will be used by default for its channel'
                : t('providers.is_default_disabled') ||
                  'Set as default provider for this channel'}
            </p>
          </div>
          <Switch checked={form.isDefault} onCheckedChange={(v) => updateField('isDefault', v)} />
        </div>

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <Button onClick={handleSubmit} disabled={saving}>
            {saving ? (
              <>
                <Loader2 className="ml-1.5 h-4 w-4 animate-spin" /> {t('common.saving')}
              </>
            ) : (
              t('common.save')
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
