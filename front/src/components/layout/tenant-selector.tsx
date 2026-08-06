'use client';

import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { ChevronsUpDown, Star, Building2, Globe, Check, Settings2 } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Button,
} from '@minisource/ui';
import { cn } from '@/lib/utils';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { ALL_TENANTS } from '@/stores/tenant.store';
import type { Tenant } from '@/features/tenants/types';

interface TenantSelectorProps {
  /** Custom class name */
  className?: string;
}

export function TenantSelector({ className }: TenantSelectorProps) {
  const t = useTranslations();
  const { tenants, activeTenant, switchTenant } = useTenants();

  const isGlobalActive = activeTenant.id === ALL_TENANTS.id;

  return (
    <div
      className={cn(
        'border-b pb-3 px-1',
        'group-data-[collapsible=icon]:border-b-0 group-data-[collapsible=icon]:px-0 group-data-[collapsible=icon]:pb-0',
        className
      )}
    >
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            aria-label={activeTenant?.name ?? t('tenants.all_tenants')}
            className={cn(
              'w-full justify-start gap-2.5 border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground transition-all',
              'group-data-[collapsible=icon]:h-10 group-data-[collapsible=icon]:w-10 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:gap-0 group-data-[collapsible=icon]:p-0 group-data-[collapsible=icon]:mx-auto'
            )}
          >
            <div
              className={cn(
                'flex items-center justify-center rounded-md shrink-0 size-7',
                isGlobalActive ? 'bg-primary/10 text-primary' : 'bg-secondary text-secondary-foreground'
              )}
            >
              {isGlobalActive ? <Globe className="size-4" /> : <Building2 className="size-4" />}
            </div>

            <div
              className={cn(
                'flex flex-col items-start min-w-0 flex-1 text-start',
                'group-data-[collapsible=icon]:hidden'
              )}
            >
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider leading-none mb-1">
                {t('tenants.organization_context')}
              </span>
              <span className="text-sm font-semibold truncate leading-none">
                {activeTenant?.name ?? t('tenants.all_tenants')}
              </span>
            </div>
            <ChevronsUpDown className="ms-auto size-4 shrink-0 text-muted-foreground/70 group-data-[collapsible=icon]:hidden" />
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent
          className="w-72 rounded-lg max-h-[min(70vh,28rem)] flex flex-col p-1.5 shadow-lg"
          align="start"
          side="bottom"
          sideOffset={6}
        >
          <DropdownMenuLabel className="text-xs font-semibold uppercase tracking-wider text-muted-foreground px-2 py-1.5">
            {t('tenants.switch_context')}
          </DropdownMenuLabel>

          <div className="overflow-y-auto overflow-x-hidden min-h-0 space-y-0.5">
            {/* Global / All Tenants Option */}
            <DropdownMenuItem
              onClick={() => switchTenant(ALL_TENANTS)}
              className={cn(
                'flex items-center justify-between p-2 rounded-md cursor-pointer text-sm font-medium',
                isGlobalActive && 'bg-accent text-accent-foreground font-semibold'
              )}
            >
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="flex size-7 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary border border-primary/20">
                  <Globe className="size-4" />
                </div>
                <div className="flex flex-col min-w-0">
                  <span className="truncate">{t('tenants.all_tenants')}</span>
                  <span className="text-xs text-muted-foreground font-normal">
                    {t('tenants.all_tenants_desc')}
                  </span>
                </div>
              </div>
              {isGlobalActive && <Check className="size-4 text-primary shrink-0 ms-2" />}
            </DropdownMenuItem>

            <DropdownMenuSeparator className="my-1" />

            {/* List of Specific Tenants */}
            {tenants.length === 0 ? (
              <DropdownMenuItem disabled className="text-xs text-muted-foreground cursor-default">
                {t('tenants.no_tenants')}
              </DropdownMenuItem>
            ) : (
              tenants.map((tenant) => (
                <TenantItem
                  key={tenant.id}
                  tenant={tenant}
                  isActive={activeTenant?.id === tenant.id}
                  onSelect={() => switchTenant(tenant)}
                />
              ))
            )}
          </div>

          <DropdownMenuSeparator className="my-1" />
          <DropdownMenuItem asChild className="cursor-pointer gap-2 p-2 rounded-md text-xs font-medium text-primary hover:text-primary focus:text-primary">
            <Link href={`/tenants`}>
              <Settings2 className="size-4" />
              {t('tenants.manage')}
            </Link>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* TenantItem                                                                 */
/* -------------------------------------------------------------------------- */

function TenantItem({
  tenant,
  isActive,
  onSelect,
}: {
  tenant: Tenant;
  isActive: boolean;
  onSelect: () => void;
}) {
  const t = useTranslations();

  return (
    <DropdownMenuItem
      onClick={onSelect}
      className={cn(
        'flex items-center justify-between p-2 rounded-md cursor-pointer text-sm font-medium',
        isActive && 'bg-accent text-accent-foreground font-semibold'
      )}
    >
      <div className="flex items-center gap-2.5 min-w-0">
        <div className="flex size-7 shrink-0 items-center justify-center rounded-md border bg-muted/50">
          <Building2 className="size-4 text-muted-foreground" />
        </div>
        <div className="flex flex-col min-w-0">
          <span className="truncate">{tenant.name}</span>
          <span className="text-xs text-muted-foreground font-normal">
            {t('tenants.slug_label')}: {tenant.slug}
          </span>
        </div>
      </div>

      <div className="flex items-center gap-1.5 shrink-0 ms-2">
        {tenant.isDefault && (
          <span title={t('tenants.default')}>
            <Star className="size-3.5 text-amber-500 fill-amber-500" />
          </span>
        )}
        {isActive && <Check className="size-4 text-primary" />}
      </div>
    </DropdownMenuItem>
  );
}
