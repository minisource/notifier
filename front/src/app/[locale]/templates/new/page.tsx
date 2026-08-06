'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageHeader, Card, CardContent, CardHeader, CardTitle, Button, Input, Label, Textarea, Select, SelectContent, SelectItem, SelectTrigger, SelectValue, toast } from '@minisource/ui';
import { ArrowLeft, Plus, X } from 'lucide-react';
import { createTemplate } from '@/features/templates/api';
import type { CreateTemplateInput, ProviderTemplateMap } from '@/features/templates/types';
import { ProviderTemplatesEditor } from '@/features/templates/components/provider-templates-editor';

export default function NewTemplatePage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'en';
  const isRtl = locale === 'fa';

  const [name, setName] = useState('');
  const [key, setKey] = useState('');
  const [type, setType] = useState('email');
  const [templateLocale, setTemplateLocale] = useState('fa');
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [description, setDescription] = useState('');
  const [variables, setVariables] = useState<string[]>([]);
  const [providerTemplates, setProviderTemplates] = useState<ProviderTemplateMap[]>([]);
  const [newVariable, setNewVariable] = useState('');
  const [saving, setSaving] = useState(false);

  const addVariable = () => {
    const trimmed = newVariable.trim().replace(/\s+/g, '_');
    if (trimmed && !variables.includes(trimmed)) {
      setVariables([...variables, trimmed]);
      setNewVariable('');
    }
  };

  const removeVariable = (v: string) => {
    setVariables(variables.filter(x => x !== v));
  };

  const handleSubmit = async () => {
    const isSms = type === 'sms';
    if (!name.trim() || (!isSms && !body.trim())) {
      toast.error(t('forms.required') || 'Name and Body are required');
      return;
    }

    setSaving(true);
    try {
      const input: CreateTemplateInput = {
        key: key.trim() || undefined,
        name: name.trim(),
        type,
        locale: templateLocale as 'fa' | 'en',
        subject: subject.trim() || undefined,
        body: body.trim() || undefined,
        description: description.trim() || undefined,
        variables: variables.length > 0 ? variables : undefined,
        providerTemplates: providerTemplates.length > 0 ? providerTemplates : undefined,
      };
      await createTemplate(input);
      toast.success(t('templates.title') as string, t('common.save') as string);
      router.push(`/templates`);
    } catch {
      toast.error(t('errors.generic') || 'An error occurred');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title={t('templates.new_title')}>
        <Button variant="ghost" onClick={() => router.push(`/templates`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <Card>
        <CardHeader>
          <CardTitle>{t('templates.new_title')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Name & Key */}
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>{t('templates.name')} *</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g., OTP via SMS" />
            </div>
            <div className="space-y-2">
              <Label>{t('templates.key') || 'Programmatic Key'}</Label>
              <Input value={key} onChange={(e) => setKey(e.target.value)} placeholder="e.g., auth.otp" />
            </div>
          </div>

          {/* Type + Locale */}
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>{t('common.type')} *</Label>
              <Select value={type} onValueChange={setType}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sms">SMS</SelectItem>
                  <SelectItem value="email">Email</SelectItem>
                  <SelectItem value="push">Push</SelectItem>
                  <SelectItem value="in_app">In-App</SelectItem>
                  <SelectItem value="webhook">Webhook</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>{t('templates.locale')}</Label>
              <Select value={templateLocale} onValueChange={setTemplateLocale}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="fa">فارسی</SelectItem>
                  <SelectItem value="en">English</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label>{t('templates.description') || 'Description'}</Label>
            <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Optional description" />
          </div>

          {/* Subject */}
          {type !== 'sms' && (
            <div className="space-y-2">
              <Label>{t('notifications.subject')}</Label>
              <Input value={subject} onChange={(e) => setSubject(e.target.value)} placeholder="Notification subject line..." />
            </div>
          )}

          {/* Body */}
          <div className="space-y-2">
            <Label>{t('notifications.body')}{type !== 'sms' ? ' *' : ''}</Label>
            <Textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder={type === 'sms' ? 'Optional body for SMS. E.g., Your code is {{code}}' : 'Enter template body with {{variables}}...'}
              className="min-h-[150px] font-mono text-sm"
            />
          </div>

          {/* Variables */}
          <div className="space-y-2">
            <Label>{t('templates.variables')}</Label>
            <div className="flex gap-2">
              <Input
                value={newVariable}
                onChange={(e) => setNewVariable(e.target.value)}
                placeholder="Add variable name..."
                onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addVariable(); } }}
              />
              <Button variant="outline" onClick={addVariable} type="button">
                <Plus className="h-4 w-4" />
              </Button>
            </div>
            {variables.length > 0 && (
              <div className="flex flex-wrap gap-2 pt-1">
                {variables.map((v) => (
                  <span key={v} className="inline-flex items-center gap-1 rounded-md bg-muted px-2.5 py-1 text-xs font-medium">
                    <code>{v}</code>
                    <button onClick={() => removeVariable(v)} className="text-muted-foreground hover:text-foreground">
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Provider Templates Editor */}
          <div className="border-t pt-4 mt-2">
            <ProviderTemplatesEditor
              value={providerTemplates}
              onChange={setProviderTemplates}
            />
          </div>

          {/* Actions */}
          <div className="flex items-center gap-3 pt-2" dir={isRtl ? 'rtl' : 'ltr'}>
            <Button onClick={handleSubmit} disabled={saving || !name.trim() || (type !== 'sms' && !body.trim())}>
              {saving ? t('common.loading') : t('common.save')}
            </Button>
            <Button variant="outline" onClick={() => router.push(`/templates`)}>
              {t('common.cancel')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
