'use client';

import { useTranslations } from 'next-intl';
import { RefreshCw, Clock } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface AutoRefreshControlProps {
  isRefreshing: boolean;
  onRefresh: () => void;
  lastUpdated?: string;
  autoRefreshEnabled: boolean;
  onToggleAutoRefresh: (enabled: boolean) => void;
  intervalSeconds?: number;
  className?: string;
}

export function AutoRefreshControl({
  isRefreshing, onRefresh, lastUpdated,
  autoRefreshEnabled, onToggleAutoRefresh,
  intervalSeconds = 30, className,
}: AutoRefreshControlProps) {
  const t = useTranslations();

  return (
    <div className={cn('flex items-center gap-2 text-xs', className)}>
      {lastUpdated && (
        <span className="flex items-center gap-1 text-muted-foreground">
          <Clock className="h-3 w-3" />
          {t('notifier.autoRefresh.lastUpdated') || 'Updated'}: {lastUpdated}
        </span>
      )}
      <Button
        variant="outline"
        size="sm"
        onClick={onRefresh}
        disabled={isRefreshing}
        className="h-7 gap-1 px-2"
        aria-label={t('notifier.autoRefresh.refresh') || 'Refresh'}
      >
        <RefreshCw className={cn('h-3 w-3', isRefreshing && 'animate-spin')} />
        <span className="hidden sm:inline">{t('observability.refresh')}</span>
      </Button>
      <label className="flex items-center gap-1.5 cursor-pointer select-none">
        <input
          type="checkbox"
          checked={autoRefreshEnabled}
          onChange={(e) => onToggleAutoRefresh(e.target.checked)}
          className="h-3 w-3 rounded border-gray-300"
          aria-label={t('notifier.autoRefresh.toggle') || 'Auto refresh'}
        />
        <span className="text-muted-foreground hidden sm:inline">
          {t('notifier.autoRefresh.auto') || 'Auto'}
        </span>
        <span className="text-muted-foreground/50">
          {autoRefreshEnabled ? `${intervalSeconds}s` : ''}
        </span>
      </label>
    </div>
  );
}
