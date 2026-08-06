'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { Button } from '@minisource/ui';
import { Input } from '@minisource/ui';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@minisource/ui';
import { useProviders } from '@/features/providers/hooks/use-providers';
import { useTenants } from '@/features/tenants/hooks/use-tenants';
import { Plus, Trash2 } from 'lucide-react';
import type { ProviderTemplateMap } from '../types';

interface ProviderTemplatesEditorProps {
  value?: ProviderTemplateMap[];
  onChange: (value: ProviderTemplateMap[]) => void;
}

export function ProviderTemplatesEditor({ value = [], onChange }: ProviderTemplatesEditorProps) {
  const t = useTranslations();
  const { data: dbProviders } = useProviders();
  const { data: dbTenants } = useTenants();

  // Standard predefined providers for ease of use
  const standardProviders = ['kavenegar', 'twilio', 'bale', 'sms', 'email', 'push'];
  
  // Combine db providers and standard list
  const providerNames = Array.from(new Set([
    ...standardProviders,
    ...(dbProviders?.map(p => p.name.toLowerCase()) || [])
  ]));

  const addMapping = () => {
    onChange([
      ...value,
      { provider: providerNames[0] || 'kavenegar', templateKey: '' }
    ]);
  };

  const removeMapping = (index: number) => {
    onChange(value.filter((_, i) => i !== index));
  };

  const updateMapping = (index: number, fields: Partial<ProviderTemplateMap>) => {
    onChange(value.map((item, i) => (i === index ? { ...item, ...fields } : item)));
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-foreground">{t('templates.provider_mapping') || 'Provider-Specific Mappings'}</h4>
        <Button type="button" variant="outline" size="sm" onClick={addMapping} className="h-8 gap-1">
          <Plus className="h-3.5 w-3.5" />
          {t('common.add') || 'Add Mapping'}
        </Button>
      </div>

      {value.length === 0 ? (
        <p className="text-xs text-muted-foreground italic border border-dashed rounded-lg p-4 text-center">
          {t('templates.no_mappings_hint') || 'No provider-specific templates defined. The main template name or key will be used as a fallback.'}
        </p>
      ) : (
        <div className="space-y-3">
          {value.map((mapping, idx) => (
            <div key={idx} className="flex flex-col sm:flex-row gap-2.5 items-end sm:items-center bg-muted/20 border rounded-lg p-3 sm:p-2.5">
              {/* Provider */}
              <div className="w-full sm:w-1/3 space-y-1">
                <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block sm:hidden">
                  Provider
                </label>
                <Select
                  value={mapping.provider}
                  onValueChange={(val) => updateMapping(idx, { provider: val })}
                >
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Provider" />
                  </SelectTrigger>
                  <SelectContent>
                    {providerNames.map(name => (
                      <SelectItem key={name} value={name}>
                        <span className="capitalize">{name}</span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Tenant (Optional) */}
              <div className="w-full sm:w-1/3 space-y-1">
                <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block sm:hidden">
                  Tenant Scope (Optional)
                </label>
                <Select
                  value={mapping.tenantId || 'all'}
                  onValueChange={(val) => updateMapping(idx, { tenantId: val === 'all' ? undefined : val })}
                >
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="All Tenants (Global)" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">
                      {t('common.all') || 'All Tenants (Global)'}
                    </SelectItem>
                    {dbTenants?.map(tenant => (
                      <SelectItem key={tenant.id} value={tenant.id}>
                        {tenant.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Template Key */}
              <div className="w-full sm:w-1/3 space-y-1 flex gap-2 items-center">
                <div className="flex-1 space-y-1">
                  <label className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block sm:hidden">
                    Template Key
                  </label>
                  <Input
                    value={mapping.templateKey}
                    onChange={(e) => updateMapping(idx, { templateKey: e.target.value })}
                    placeholder="e.g. otp_pattern_key"
                    className="h-9"
                  />
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => removeMapping(idx)}
                  className="text-destructive hover:bg-destructive/10 shrink-0 h-9 w-9 mt-auto"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
