'use client';

import { useTranslations } from 'next-intl';
import { RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface RefreshButtonProps {
  onRefresh: () => void;
  isRefreshing?: boolean;
  lastUpdated?: string;
  label?: string;
  className?: string;
  variant?: 'outline' | 'ghost' | 'default';
  size?: 'sm' | 'default';
}

export function RefreshButton({
  onRefresh, isRefreshing, lastUpdated, label, className,
  variant = 'outline', size = 'sm',
}: RefreshButtonProps) {
  const t = useTranslations();

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {lastUpdated && (
        <span className="text-xs text-muted-foreground">
          {t('observability.last_updated') || 'Updated'}: {lastUpdated}
        </span>
      )}
      <Button
        variant={variant}
        size={size}
        onClick={onRefresh}
        disabled={isRefreshing}
        className="gap-1.5"
      >
        <RefreshCw className={cn('h-4 w-4', isRefreshing && 'animate-spin')} />
        {label || t('observability.refresh')}
      </Button>
    </div>
  );
}
