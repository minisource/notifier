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
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { toast } from 'sonner';
import { ArrowLeft, Plus, X } from 'lucide-react';
import { createTemplate } from '@/features/templates/api';
import type { CreateTemplateInput } from '@/features/templates/types';

export default function NewTemplatePage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';

  const [name, setName] = useState('');
  const [type, setType] = useState('email');
  const [templateLocale, setTemplateLocale] = useState('fa');
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [description, setDescription] = useState('');
  const [variables, setVariables] = useState<string[]>([]);
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
    if (!name.trim() || !body.trim()) {
      toast.error(t('forms.required'));
      return;
    }

    setSaving(true);
    try {
      const input: CreateTemplateInput = {
        name: name.trim(),
        type,
        locale: templateLocale as 'fa' | 'en',
        subject: subject.trim() || undefined,
        body: body.trim(),
        description: description.trim() || undefined,
        variables: variables.length > 0 ? variables : undefined,
      };
      await createTemplate(input);
      toast.success(t('templates.title') as string, { description: t('common.save') as string });
      router.push(`/${locale}/templates`);
    } catch {
      toast.error(t('errors.generic'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageContainer>
      <PageHeader title={t('templates.new_title')}>
        <Button variant="ghost" onClick={() => router.push(`/${locale}/templates`)}>
          <ArrowLeft className="ml-2 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <Card>
        <CardHeader>
          <CardTitle>{t('templates.new_title')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Name */}
          <div className="space-y-2">
            <Label>{t('templates.name')} *</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g., OTP via SMS" />
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
          <div className="space-y-2">
            <Label>{t('notifications.subject')}</Label>
            <Input value={subject} onChange={(e) => setSubject(e.target.value)} placeholder="Notification subject line..." />
          </div>

          {/* Body */}
          <div className="space-y-2">
            <Label>{t('notifications.body')} *</Label>
            <Textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Enter template body with {{variables}}..."
              className="min-h-[200px] font-mono text-sm"
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

          {/* Actions */}
          <div className="flex items-center gap-3 pt-2" dir={isRtl ? 'rtl' : 'ltr'}>
            <Button onClick={handleSubmit} disabled={saving || !name.trim() || !body.trim()}>
              {saving ? t('common.loading') : t('common.save')}
            </Button>
            <Button variant="outline" onClick={() => router.push(`/${locale}/templates`)}>
              {t('common.cancel')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </PageContainer>
  );
}
