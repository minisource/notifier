'use client';

import { useTranslations } from 'next-intl';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { SectionCard } from '@/components/shared/section-card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { EmptyState } from '@/components/shared/empty-state';
import { ErrorState } from '@/components/shared/error-state';
import { TableSkeleton } from '@/components/shared/loading-state';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { Building2, Globe, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ChannelBadge } from '@/components/shared/channel-badge';

export default function TenantsPage() {
  const t = useTranslations();
  const { data: tenants, isLoading, isError, error, refetch, isFetching } = useTenants();

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title={t('tenants.title')} subtitle={t('tenants.subtitle')} />
        <TableSkeleton rows={4} columns={5} context="tenants" />
      </PageContainer>
    );
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title={t('tenants.title')} subtitle={t('tenants.subtitle')} />
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || t('errors.generic')}
          onRetry={() => refetch()}
        />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader title={t('tenants.title')} subtitle={t('tenants.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('common.refresh')}
        </Button>
      </PageHeader>

      <SectionCard title={t('tenants.title')}>
        {tenants && tenants.length > 0 ? (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[220px]">{t('tenants.name') || 'Name'}</TableHead>
                  <TableHead className="w-[120px]">{t('tenants.slug') || 'Slug'}</TableHead>
                  <TableHead className="w-[160px]">{t('tenants.channels') || 'Channels'}</TableHead>
                  <TableHead className="w-[100px]">{t('common.status')}</TableHead>
                  <TableHead className="w-[180px]">{t('tenants.usage') || 'Monthly Usage'}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tenants.map((tenant) => {
                  const usagePercent = tenant.monthlyQuota > 0
                    ? Math.round((tenant.usedThisMonth / tenant.monthlyQuota) * 100)
                    : 0;
                  return (
                    <TableRow key={tenant.id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Building2 className="h-4 w-4 shrink-0 text-muted-foreground" />
                          <span className="text-sm font-medium truncate">{tenant.name}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <code className="text-xs font-mono text-muted-foreground">{tenant.slug}</code>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {tenant.enabledChannels.map(ch => (
                            <ChannelBadge key={ch} channel={ch as any} size="sm" />
                          ))}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={tenant.isActive ? 'default' : 'secondary'} className="text-xs">
                          {tenant.isActive ? t('common.active') || 'Active' : t('common.inactive') || 'Inactive'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                            <div
                              className={`h-full rounded-full transition-all ${
                                usagePercent > 90 ? 'bg-destructive' : usagePercent > 70 ? 'bg-amber-500' : 'bg-primary'
                              }`}
                              style={{ width: `${Math.min(usagePercent, 100)}%` }}
                            />
                          </div>
                          <span className="text-xs text-muted-foreground whitespace-nowrap">
                            {tenant.usedThisMonth.toLocaleString()} / {tenant.monthlyQuota.toLocaleString()}
                          </span>
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
            icon={Globe}
            title={t('tenants.no_tenants')}
            description="Projects (tenants) help you organize notifications across different apps, services, or client environments."
            tips={[
              'Each project has its own monthly quota and usage tracking',
              'Configure allowed channels per project for fine-grained control',
              'Use the default project for quick setup, then create custom ones',
            ]}
          />
        )}
      </SectionCard>
    </PageContainer>
  );
}
