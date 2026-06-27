'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Loader2 } from 'lucide-react';

type ProviderFormData = {
  name: string;
  channel: string;
  type: string;
  status: string;
  priority: number;
  isDefault: boolean;
  description: string;
  configJson: string;
  secretConfigJson: string;
};

interface ProviderFormProps {
  initialData?: Partial<ProviderFormData>;
  onSave: (data: ProviderFormData) => Promise<void>;
  saving: boolean;
  mode: 'create' | 'edit';
}

const DEFAULT_FORM: ProviderFormData = {
  name: '',
  channel: 'sms',
  type: '',
  status: 'active',
  priority: 1,
  isDefault: false,
  description: '',
  configJson: '{}',
  secretConfigJson: '{}',
};

export function ProviderForm({ initialData, onSave, saving, mode }: ProviderFormProps) {
  const t = useTranslations();
  const [form, setForm] = useState<ProviderFormData>({ ...DEFAULT_FORM, ...initialData });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const updateField = <K extends keyof ProviderFormData>(key: K, value: ProviderFormData[K]) => {
    setForm(prev => ({ ...prev, [key]: value }));
    setErrors(prev => ({ ...prev, [key]: '' }));
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!form.name.trim()) newErrors.name = 'Name is required';
    if (!form.channel) newErrors.channel = 'Channel is required';
    if (!form.type.trim()) newErrors.type = 'Provider type is required';
    if (form.priority < 0) newErrors.priority = 'Priority must be >= 0';

    try {
      if (form.configJson) JSON.parse(form.configJson);
    } catch {
      newErrors.configJson = 'Invalid JSON format';
    }

    try {
      if (form.secretConfigJson) JSON.parse(form.secretConfigJson);
    } catch {
      newErrors.secretConfigJson = 'Invalid JSON format';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    await onSave(form);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{mode === 'create' ? t('providers.new_title') || 'Create Provider' : t('providers.edit_title') || 'Edit Provider'}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
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
            <Select value={form.channel} onValueChange={(v) => updateField('channel', v)}>
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
            <Select value={form.type} onValueChange={(v) => updateField('type', v)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="kavenegar">Kavenegar</SelectItem>
                <SelectItem value="twilio">Twilio</SelectItem>
                <SelectItem value="smtp">SMTP</SelectItem>
                <SelectItem value="sendgrid">SendGrid</SelectItem>
                <SelectItem value="fcm">Firebase FCM</SelectItem>
                <SelectItem value="apns">APNs</SelectItem>
                <SelectItem value="custom">Custom</SelectItem>
              </SelectContent>
            </Select>
            {errors.type && <p className="text-xs text-red-500">{errors.type}</p>}
          </div>
        </div>

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
                ? (t('providers.is_default_enabled') || 'This provider will be used by default for its channel')
                : (t('providers.is_default_disabled') || 'Set as default provider for this channel')}
            </p>
          </div>
          <Switch checked={form.isDefault} onCheckedChange={(v) => updateField('isDefault', v)} />
        </div>

        {/* Config JSON */}
        <div className="space-y-2">
          <Label>{t('providers.config') || 'Configuration (JSON)'}</Label>
          <Textarea
            value={form.configJson}
            onChange={(e) => updateField('configJson', e.target.value)}
            placeholder='{"sender": "1000", "baseUrl": "https://api.example.com"}'
            className="min-h-[100px] font-mono text-sm"
          />
          {errors.configJson && <p className="text-xs text-red-500">{errors.configJson}</p>}
        </div>

        {/* Secret Config JSON */}
        <div className="space-y-2">
          <Label>{t('providers.secret_config') || 'Secret Configuration (JSON)'}</Label>
          <Textarea
            value={form.secretConfigJson}
            onChange={(e) => updateField('secretConfigJson', e.target.value)}
            placeholder='{"apiKey": "your-api-key"}'
            className="min-h-[100px] font-mono text-sm"
          />
          {errors.secretConfigJson && <p className="text-xs text-red-500">{errors.secretConfigJson}</p>}
          {mode === 'edit' && (
            <p className="text-xs text-amber-500">
              {t('providers.secret_config_edit_hint') || 'Leave empty to keep existing secrets. Only provide values you want to replace.'}
            </p>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <Button onClick={handleSubmit} disabled={saving}>
            {saving ? <><Loader2 className="ml-1.5 h-4 w-4 animate-spin" /> {t('common.saving')}</> : t('common.save')}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
