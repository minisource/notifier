'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Building2, ChevronDown } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import { useTenants } from '@/features/tenants/hooks/use-tenants';

export function TenantSwitcher() {
  const t = useTranslations();
  const { data: tenants } = useTenants();
  const [current, setCurrent] = useState<string | null>(null);

  const currentTenant = current
    ? tenants?.find(t => t.id === current)
    : tenants?.[0];

  if (!tenants || tenants.length === 0) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="gap-2 text-sm">
          <Building2 className="h-4 w-4" />
          <span className="hidden md:inline max-w-[120px] truncate">{currentTenant?.name ?? t('navigation.tenants')}</span>
          <ChevronDown className="h-3 w-3" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        {tenants.map((tenant) => (
          <DropdownMenuItem
            key={tenant.id}
            onClick={() => setCurrent(tenant.id)}
            className={tenant.id === currentTenant?.id ? 'bg-accent font-medium' : ''}
          >
            <div className="flex flex-col">
              <span>{tenant.name}</span>
              <span className="text-xs text-muted-foreground">{tenant.slug}</span>
            </div>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem disabled className="text-xs text-muted-foreground">
          {t('navigation.tenants')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
