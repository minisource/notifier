'use client';

import { useTranslations } from 'next-intl';
import { PageHeader, Card, Table, TableBody, TableCell, TableHead, TableHeader, TableRow, Badge, EmptyState, ErrorState, Skeleton, Button, Alert, AlertTitle, AlertDescription } from '@minisource/ui';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { Building2, Globe, RefreshCw, ShieldCheck, Star } from 'lucide-react';
import type { Tenant } from '@/features/tenants/types';

/**
 * Tenants are managed in the Auth service. This page is a read-only view of
 * the tenants the current user belongs to — used for scoping notifications,
 * templates, providers and reminders. Create/edit/delete happen in Auth.
 */
export default function TenantsPage() {
  const t = useTranslations();
  const {
    tenants,
    isLoading,
    isError,
    error,
    refetch,
    isFetching,
  } = useTenants();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('tenants.title')} description={t('tenants.subtitle')} />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('tenants.title')} description={t('tenants.subtitle')} />
        <ErrorState
          title={t('errors.generic')}
          message={(error as Error)?.message || t('errors.generic')}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader title={t('tenants.title')} description={t('tenants.subtitle')}>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw className={`ml-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          {t('common.refresh')}
        </Button>
        <Button size="sm" asChild>
          <a href="/auth/admin/tenants" target="_blank" rel="noreferrer">
            <ShieldCheck className="ml-1.5 h-4 w-4" />
            Manage in Auth Service
          </a>
        </Button>
      </PageHeader>

      <Alert className="border-primary/20 bg-primary/5">
        <ShieldCheck className="h-4 w-4 text-primary" />
        <AlertTitle className="text-sm font-medium">
          {t('tenants.managed_elsewhere_title') || 'Tenants are managed in the Auth service'}
        </AlertTitle>
        <AlertDescription className="text-sm text-muted-foreground">
          {t('tenants.managed_elsewhere_desc') ||
            'This list comes from your Auth account. Create, edit or remove tenants there — they will appear here automatically.'}
        </AlertDescription>
      </Alert>

      <Card title={t('tenants.title')}>
        {tenants && tenants.length > 0 ? (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[220px]">{t('tenants.name') || 'Name'}</TableHead>
                  <TableHead className="w-[140px]">{t('tenants.slug') || 'Slug'}</TableHead>
                  <TableHead className="w-[120px]">{t('tenants.role_label') || 'Role'}</TableHead>
                  <TableHead className="w-[100px]">{t('common.status')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tenants.map((tenant: Tenant) => (
                  <TableRow key={tenant.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Building2 className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <span className="text-sm font-medium truncate">{tenant.name}</span>
                        {tenant.isDefault && (
                          <span title={t('tenants.default')} className="inline-flex">
                            <Star className="h-3.5 w-3.5 shrink-0 text-amber-500 fill-amber-500" aria-label={t('tenants.default')} />
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <code className="text-xs font-mono text-muted-foreground">{tenant.slug}</code>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground capitalize">{tenant.role || 'member'}</span>
                    </TableCell>
                    <TableCell>
                      <Badge variant={tenant.isActive ? 'default' : 'secondary'} className="text-xs">
                        {tenant.isActive ? t('common.active') || 'Active' : t('common.inactive') || 'Inactive'}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <EmptyState
            icon={Globe}
            title={t('tenants.no_tenants')}
            description="You are not a member of any tenant yet. Create one in the Auth service to start scoping notifications."
          />
        )}
      </Card>
    </div>
  );
}
