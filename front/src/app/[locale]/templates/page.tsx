'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SectionCard } from '@/components/shared/section-card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { EmptyState } from '@/components/shared/empty-state';
import { ErrorState } from '@/components/shared/error-state';
import { TableSkeleton } from '@/components/shared/loading-state';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { useTemplates, useDeleteTemplate } from '@/features/templates/hooks/use-templates';
import { Plus, FileText, RefreshCw, Trash2, Copy, Play } from 'lucide-react';
import { formatRelativeTime } from '@/lib/utils/date';
import { toast } from 'sonner';
import { TemplateRenderPreview } from '@/features/templates/components/template-render-preview';

const CHANNELS = ['all', 'sms', 'email', 'push', 'in_app', 'webhook'];

export default function TemplatesPage() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';
  const isRtl = locale === 'fa';

  const [typeFilter, setTypeFilter] = useState('all');
  const [localeFilter, setLocaleFilter] = useState('all');
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [previewTemplate, setPreviewTemplate] = useState<{ body: string; subject?: string; variables: string[] } | null>(null);

  const queryParams = {
    ...(typeFilter !== 'all' ? { type: typeFilter } : {}),
    ...(localeFilter !== 'all' ? { locale: localeFilter } : {}),
  };

  const { data: templates, isLoading, isError, error, refetch, isFetching } = useTemplates(queryParams);
  const deleteMutation = useDeleteTemplate();

  const handleDelete = async () => {
    if (!deleteId) return;
    try {
      await deleteMutation.mutateAsync(deleteId);
      toast.success(t('templates.title') as string, { description: t('common.delete') as string });
      setDeleteId(null);
    } catch {
      toast.error(t('errors.generic'));
    }
  };

  const handleCopyKey = (key?: string) => {
    if (!key) return;
    navigator.clipboard.writeText(key);
    toast.success(t('common.copied') as string);
  };

  const openPreview = (body: string, subject?: string) => {
    const variables = [...new Set(body.match(/\{\{(\w+)\}\}/g)?.map(v => v.slice(2, -2)) || [])];
    setPreviewTemplate({ body, subject, variables });
  };

  return (
    <PageContainer>
      <PageHeader title={t('templates.title')} subtitle={t('templates.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('dashboard.view_all') as string}
        </Button>
        <Button size="sm" onClick={() => router.push(`/${locale}/templates/new`)}>
          <Plus className="ml-1.5 h-4 w-4" />
          {t('common.create')}
        </Button>
      </PageHeader>

      <SectionCard title={t('templates.title')}>
        {isLoading ? (
          <TableSkeleton rows={5} columns={7} context="templates" />
        ) : isError ? (
          <ErrorState
            title={t('errors.generic')}
            message={(error as Error)?.message || t('errors.generic')}
            onRetry={() => refetch()}
            autoRetrySeconds={15}
          />
        ) : (
          <div className="space-y-4">
            {/* Filters */}
            <div className="flex flex-wrap items-center gap-2" dir={isRtl ? 'rtl' : 'ltr'}>
              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder={t('common.all') as string} />
                </SelectTrigger>
                <SelectContent>
                  {CHANNELS.map(ch => (
                    <SelectItem key={ch} value={ch}>
                      {ch === 'all' ? t('common.all') : t(`channels.${ch}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select value={localeFilter} onValueChange={setLocaleFilter}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder={t('templates.locale') as string} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t('common.all')}</SelectItem>
                  <SelectItem value="en">English</SelectItem>
                  <SelectItem value="fa">فارسی</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Table */}
            {templates && templates.length > 0 ? (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[220px]">{t('templates.name')}</TableHead>
                      <TableHead className="w-[140px]">{t('templates.key')}</TableHead>
                      <TableHead className="w-[80px]">{t('common.channel')}</TableHead>
                      <TableHead className="w-[60px]">{t('templates.locale')}</TableHead>
                      <TableHead className="w-[80px]">{t('templates.variables')}</TableHead>
                      <TableHead className="w-[80px]">{t('common.status')}</TableHead>
                      <TableHead className="w-[120px]">{t('common.updated_at')}</TableHead>
                      <TableHead className="w-[120px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {templates.map((template) => {
                      const detectedVars = template.body.match(/\{\{(\w+)\}\}/g)?.map(v => v.slice(2, -2)) || [];
                      return (
                        <TableRow key={template.id} className="cursor-pointer" onClick={() => router.push(`/${locale}/templates/${template.id}`)}>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                              <span className="text-sm font-medium truncate">{template.name}</span>
                            </div>
                          </TableCell>
                          <TableCell>
                            {template.key ? (
                              <code className="text-[10px] font-mono text-muted-foreground">{template.key}</code>
                            ) : (
                              <span className="text-xs text-muted-foreground">—</span>
                            )}
                          </TableCell>
                          <TableCell><ChannelBadge channel={template.type as any} size="sm" /></TableCell>
                          <TableCell><span className="text-xs">{template.locale === 'fa' ? 'فا' : 'EN'}</span></TableCell>
                          <TableCell><span className="text-xs text-muted-foreground">{detectedVars.length}</span></TableCell>
                          <TableCell>
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                              template.isActive
                                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-gray-100 text-gray-600 dark:bg-gray-800/50 dark:text-gray-400'
                            }`}>
                              {template.isActive ? t('templates.is_active') : t('templates.not_active')}
                            </span>
                          </TableCell>
                          <TableCell>
                            <span className="text-xs text-muted-foreground whitespace-nowrap">
                              {formatRelativeTime(template.updatedAt, locale)}
                            </span>
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openPreview(template.body, template.subject)} title={t('templates.render_preview')}>
                                <Play className="h-3.5 w-3.5" />
                              </Button>
                              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleCopyKey(template.key)} title={t('common.copy_id')}>
                                <Copy className="h-3.5 w-3.5" />
                              </Button>
                              <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => setDeleteId(template.id)} title={t('common.delete')}>
                                <Trash2 className="h-3.5 w-3.5" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <EmptyState
                icon={FileText}
                title={t('templates.no_templates')}
                description="Start by creating your first notification template. Templates help you reuse message formats across different channels."
                actionLabel={t('common.create')}
                onAction={() => router.push(`/${locale}/templates/new`)}
                tips={[
                  'Use template variables like {{name}} for dynamic content',
                  'Create separate templates for each locale (FA / EN)',
                  'Link templates to providers for automated rendering',
                ]}
              />
            )}
          </div>
        )}
      </SectionCard>

      {/* Delete Confirm */}
      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(o) => { if (!o) setDeleteId(null); }}
        onConfirm={handleDelete}
        title={t('common.confirm_action')}
        description={t('common.delete')}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        destructive
      />

      {/* Render Preview */}
      {previewTemplate && (
        <TemplateRenderPreview
          open={!!previewTemplate}
          onOpenChange={(o) => { if (!o) setPreviewTemplate(null); }}
          templateBody={previewTemplate.body}
          templateSubject={previewTemplate.subject}
          detectedVariables={previewTemplate.variables}
        />
      )}
    </PageContainer>
  );
}
