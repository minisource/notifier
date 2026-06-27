'use client';

import { useTranslations } from 'next-intl';
import { Server, AlertTriangle } from 'lucide-react';
import { SectionCard } from '@/components/shared/section-card';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { MiniProgress } from '@/components/shared/mini-progress';
import { Button } from '@/components/ui/button';
import { useProviders } from '@/features/providers/hooks/use-providers';

export function ProviderHealthPanel() {
  const t = useTranslations();
  const { data: providers, isLoading } = useProviders();

  const degradedCount = providers?.filter(p => p.status === 'inactive' || p.status === 'error').length || 0;

  return (
    <SectionCard
      title={t('dashboard.provider_health')}
      icon={Server}
      action={
        degradedCount > 0 && (
          <div className="flex items-center gap-1 text-xs text-amber-600 dark:text-amber-400">
            <AlertTriangle className="h-3 w-3" />
            <span>{degradedCount} {t('providers.degraded')}</span>
          </div>
        )
      }
    >
      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-12 animate-pulse rounded-lg bg-muted" />
          ))}
        </div>
      ) : providers && providers.length > 0 ? (
        <div className="space-y-2">
          {providers.map(provider => (
            <div
              key={provider.id}
              className="flex items-center justify-between rounded-lg border border-border/60 bg-muted/20 px-3 py-2.5 transition-colors hover:bg-muted/40"
            >
              <div className="flex items-center gap-3 min-w-0">
                <ChannelBadge channel={provider.channel} size="sm" />
                <span className="text-sm font-medium truncate">{provider.name}</span>
              </div>
              <div className="flex items-center gap-3 flex-shrink-0">
                <div className="hidden sm:block w-20">
                  <MiniProgress value={provider.successRate ?? 0} variant={(provider.successRate ?? 0) >= 95 ? 'success' : (provider.successRate ?? 0) >= 80 ? 'warning' : 'danger'} label={`${provider.successRate ?? 0}%`} />
                </div>
                <StatusBadge status={provider.status} size="sm" />
              </div>
            </div>
          ))}
          <div className="pt-1">
            <Button variant="link" size="sm" className="h-auto px-0 text-xs text-muted-foreground">
              {t('providers.title')} →
            </Button>
          </div>
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">{t('providers.no_providers')}</p>
      )}
    </SectionCard>
  );
}
