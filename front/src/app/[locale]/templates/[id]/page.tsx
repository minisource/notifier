'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SectionCard } from '@/components/shared/section-card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { ErrorState } from '@/components/shared/error-state';
import { PageSkeleton } from '@/components/shared/loading-state';
import { useTemplate } from '@/features/templates/hooks/use-templates';
import { ArrowLeft, Layers, Tag, Send } from 'lucide-react';
import { formatDateTime } from '@/lib/utils/date';

export default function TemplateDetailPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const id = params?.id as string;

  const { data: template, isLoading, isError, error, refetch } = useTemplate(id);

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title={t('templates.detail_title')}>
          <Button variant="ghost" onClick={() => router.push(`/${locale}/templates`)} disabled>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <PageSkeleton context="templates" layout="detail" />
      </PageContainer>
    );
  }

  if (isError || !template) {
    return (
      <PageContainer>
        <PageHeader title={t('templates.detail_title')}>
          <Button variant="ghost" onClick={() => router.push(`/${locale}/templates`)}>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.not_found')}
          message={(error as Error)?.message || t('templates.no_templates')}
          onRetry={() => refetch()}
        />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader title={template.name} subtitle={template.key ? `Key: ${template.key}` : t('templates.detail_title')}>
        <Button variant="ghost" size="sm" onClick={() => router.push(`/${locale}/templates`)}>
          <ArrowLeft className="ml-1.5 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

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

        {/* Body Content */}
        <div className="grid gap-6 lg:grid-cols-2">
          <SectionCard title={t('notifications.form.content_section')} icon={Layers}>
            {template.subject && (
              <div className="mb-3">
                <p className="text-xs font-medium text-muted-foreground mb-1">{t('notifications.subject')}</p>
                <p className="text-sm p-3 rounded-md bg-muted/30">{template.subject}</p>
              </div>
            )}
            <div>
              <p className="text-xs font-medium text-muted-foreground mb-1">{t('notifications.body')}</p>
              <pre className="text-sm p-3 rounded-md bg-muted/30 whitespace-pre-wrap font-mono text-xs overflow-x-auto">
                {template.body}
              </pre>
            </div>
          </SectionCard>

          <SectionCard title={t('templates.variables')} icon={Tag}>
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
          </SectionCard>
        </div>
      </div>
    </PageContainer>
  );
}
