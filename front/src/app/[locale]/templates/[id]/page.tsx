'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';
import {
  PageHeader,
  Button,
  Badge,
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  Separator,
  ErrorState,
  Skeleton,
  Input,
  Label,
  Textarea,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  toast,
} from '@minisource/ui';
import { useTemplate, useUpdateTemplate } from '@/features/templates/hooks/use-templates';
import { ArrowLeft, Send, Edit, Save, X, Plus } from 'lucide-react';
import { formatDateTime } from '@/lib/utils/date';
import { ProviderTemplatesEditor } from '@/features/templates/components/provider-templates-editor';
import type { ProviderTemplateMap, UpdateTemplateInput } from '@/features/templates/types';
import { useTenants } from '@/features/tenants/hooks/use-tenants';

export default function TemplateDetailPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'en';
  const id = params?.id as string;

  const { data: template, isLoading, isError, error, refetch } = useTemplate(id);
  const updateMutation = useUpdateTemplate();
  const { data: dbTenants } = useTenants();

  const [isEditing, setIsEditing] = useState(false);
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

  // Sync state with template data when not editing
  useEffect(() => {
    if (template && !isEditing) {
      setName(template.name || '');
      setKey(template.key || '');
      setType(template.type || 'email');
      setTemplateLocale(template.locale || 'fa');
      setSubject(template.subject || '');
      setBody(template.body || '');
      setDescription(template.description || '');
      setVariables(template.variables || []);
      setProviderTemplates(template.providerTemplates || []);
    }
  }, [template, isEditing]);

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

  const handleSave = async () => {
    const isSms = type === 'sms';
    if (!name.trim() || (!isSms && !body.trim())) {
      toast.error(t('forms.required') || 'Name and Body are required');
      return;
    }

    setSaving(true);
    try {
      const input: UpdateTemplateInput = {
        key: key.trim() || undefined,
        name: name.trim(),
        type,
        locale: templateLocale as 'fa' | 'en',
        subject: subject.trim() || undefined,
        body: body.trim() || undefined,
        description: description.trim() || undefined,
        variables: variables.length > 0 ? variables : [],
        providerTemplates: providerTemplates,
      };
      await updateMutation.mutateAsync({ id, input });
      toast.success(t('templates.title') as string, t('common.save') as string);
      setIsEditing(false);
    } catch {
      toast.error(t('errors.generic') || 'An error occurred');
    } finally {
      setSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('templates.detail_title')}>
          <Button variant="ghost" onClick={() => router.push(`/templates`)} disabled>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (isError || !template) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('templates.detail_title')}>
          <Button variant="ghost" onClick={() => router.push(`/templates`)}>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.not_found')}
          message={(error as Error)?.message || t('templates.no_templates')}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={isEditing ? `${t('common.edit')}: ${template.name}` : template.name}
        description={template.key ? `Key: ${template.key}` : t('templates.detail_title')}
      >
        <div className="flex items-center gap-2">
          {isEditing ? (
            <>
              <Button size="sm" onClick={handleSave} disabled={saving || !name.trim() || (type !== 'sms' && !body.trim())}>
                <Save className="ml-1.5 h-4 w-4" />
                {saving ? t('common.loading') : t('common.save')}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setIsEditing(false)}>
                <X className="ml-1.5 h-4 w-4" />
                {t('common.cancel')}
              </Button>
            </>
          ) : (
            <>
              <Button variant="ghost" size="sm" onClick={() => router.push(`/templates`)}>
                <ArrowLeft className="ml-1.5 h-4 w-4" />
                {t('common.back')}
              </Button>
              <Button size="sm" onClick={() => setIsEditing(true)}>
                <Edit className="ml-1.5 h-4 w-4" />
                {t('common.edit')}
              </Button>
            </>
          )}
        </div>
      </PageHeader>

      {isEditing ? (
        <Card>
          <CardContent className="space-y-6 pt-6">
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
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-6">
          {/* Summary Card */}
          <Card className="overflow-hidden">
            <CardContent className="p-5">
              <div className="flex items-start justify-between gap-4">
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Badge variant={template.isActive ? 'default' : 'secondary'}>
                      {template.isActive ? t('templates.is_active') : t('templates.not_active')}
                    </Badge>
                    <Badge variant="outline">{template.type}</Badge>
                    <Badge variant="outline">{template.locale === 'fa' ? 'فارسی' : 'English'}</Badge>
                  </div>

                  {template.description && (
                    <p className="text-sm text-muted-foreground">{template.description}</p>
                  )}
                </div>
              </div>

              <Separator className="my-4" />

              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                <div>
                  <p className="text-xs font-medium text-muted-foreground">{t('templates.name')}</p>
                  <p className="text-sm mt-0.5">{template.name}</p>
                </div>
                <div>
                  <p className="text-xs font-medium text-muted-foreground">{t('templates.key')}</p>
                  <code className="text-sm mt-0.5 font-mono">{template.key || '—'}</code>
                </div>
                <div>
                  <p className="text-xs font-medium text-muted-foreground">{t('common.type')}</p>
                  <p className="text-sm mt-0.5 capitalize">{template.type}</p>
                </div>
                <div>
                  <p className="text-xs font-medium text-muted-foreground">{t('templates.locale')}</p>
                  <p className="text-sm mt-0.5">{template.locale === 'fa' ? 'فارسی' : 'English'}</p>
                </div>
                <div>
                  <p className="text-xs font-medium text-muted-foreground">{t('common.created_at')}</p>
                  <p className="text-sm mt-0.5">{formatDateTime(template.createdAt, locale)}</p>
                </div>
                <div>
                  <p className="text-xs font-medium text-muted-foreground">{t('common.updated_at')}</p>
                  <p className="text-sm mt-0.5">{formatDateTime(template.updatedAt, locale)}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Provider Templates Display (if mapped) */}
          {template.providerTemplates && template.providerTemplates.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">
                  {t('templates.provider_mapping') || 'Provider-Specific Mappings'}
                </CardTitle>
              </CardHeader>
              <CardContent className="p-5 pt-0">
                <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                  {template.providerTemplates.map((pt, idx) => {
                    const matchedTenant = dbTenants?.find(t => t.id === pt.tenantId);
                    return (
                      <div key={idx} className="rounded-md border bg-muted/30 p-3 text-sm">
                        <div className="flex justify-between items-center gap-2 mb-1">
                          <span className="font-semibold text-foreground capitalize">{pt.provider}</span>
                          <Badge variant="outline" className="text-[10px]">
                            {matchedTenant ? `${t('common.tenant') || 'Tenant'}: ${matchedTenant.name}` : t('common.all') || 'Global'}
                          </Badge>
                        </div>
                        <div className="text-xs text-muted-foreground">
                          Key: <code className="rounded bg-muted px-1 py-0.5 text-xs font-mono">{pt.templateKey}</code>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Body Content */}
          <div className="grid gap-6 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>{t('notifications.form.content_section')}</CardTitle>
              </CardHeader>
              <CardContent>
                {template.subject && (
                  <div className="mb-3">
                    <p className="text-xs font-medium text-muted-foreground mb-1">{t('notifications.subject')}</p>
                    <p className="text-sm p-3 rounded-md bg-muted/30">{template.subject}</p>
                  </div>
                )}
                <div>
                  <p className="text-xs font-medium text-muted-foreground mb-1">{t('notifications.body')}</p>
                  {template.body ? (
                    <pre className="text-sm p-3 rounded-md bg-muted/30 whitespace-pre-wrap font-mono text-xs overflow-x-auto">
                      {template.body}
                    </pre>
                  ) : (
                    <p className="text-sm text-muted-foreground italic">{t('templates.empty_body') || 'No body content defined'}</p>
                  )}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>{t('templates.variables')}</CardTitle>
              </CardHeader>
              <CardContent>
                {template.variables && template.variables.length > 0 ? (
                  <div className="space-y-2">
                    {template.variables.map((v) => (
                      <div key={v} className="flex items-center gap-2 rounded-md border p-2.5 text-sm">
                        <code className="rounded bg-muted px-2 py-0.5 font-mono text-xs">{v}</code>
                        <span className="text-xs text-muted-foreground">{'{{'}{v}{'}}'}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">{t('templates.no_variables') || 'No variables defined'}</p>
                )}

                {template.provider && (
                  <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
                    <Send className="h-4 w-4" />
                    <span>{t('templates.provider')}: {template.provider}</span>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      )}
    </div>
  );
}
