'use client';

import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface TrendData {
  value: string;
  positive?: boolean;
  neutral?: boolean;
}

interface MetricCardProps {
  title: string;
  value: string | number;
  description?: string;
  icon?: React.ElementType;
  trend?: TrendData;
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info';
  progress?: number; // 0-100
  className?: string;
  accentBar?: boolean;
}

const variantStyles = {
  default: {
    iconBg: 'bg-primary/10 text-primary',
    accent: 'bg-primary',
    trendUp: 'text-green-600 dark:text-green-400',
    trendDown: 'text-red-600 dark:text-red-400',
  },
  success: {
    iconBg: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    accent: 'bg-green-500',
    trendUp: 'text-green-600 dark:text-green-400',
    trendDown: 'text-red-600 dark:text-red-400',
  },
  warning: {
    iconBg: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    accent: 'bg-amber-500',
    trendUp: 'text-green-600 dark:text-green-400',
    trendDown: 'text-amber-600 dark:text-amber-400',
  },
  danger: {
    iconBg: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    accent: 'bg-red-500',
    trendUp: 'text-green-600 dark:text-green-400',
    trendDown: 'text-red-600 dark:text-red-400',
  },
  info: {
    iconBg: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    accent: 'bg-blue-500',
    trendUp: 'text-green-600 dark:text-green-400',
    trendDown: 'text-blue-600 dark:text-blue-400',
  },
};

export function MetricCard({
  title,
  value,
  description,
  icon: Icon,
  trend,
  variant = 'default',
  progress,
  className,
  accentBar = true,
}: MetricCardProps) {
  const styles = variantStyles[variant];

  return (
    <Card className={cn('relative overflow-hidden transition-shadow hover:shadow-md', className)}>
      {/* Accent bar */}
      {accentBar && (
        <div className={cn('h-0.5 w-full', styles.accent)} />
      )}

      <CardContent className="p-4 md:p-5">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <p className="text-xs font-medium text-muted-foreground">{title}</p>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold tracking-tight">{value}</span>
              {trend && !trend.neutral && (
                <span className={cn(
                  'flex items-center gap-0.5 text-xs font-medium',
                  trend.positive ? styles.trendUp : styles.trendDown
                )}>
                  {trend.positive ? (
                    <TrendingUp className="h-3 w-3" />
                  ) : (
                    <TrendingDown className="h-3 w-3" />
                  )}
                  {trend.value}
                </span>
              )}
              {trend?.neutral && (
                <span className="flex items-center gap-0.5 text-xs text-muted-foreground">
                  <Minus className="h-3 w-3" />
                  {trend.value}
                </span>
              )}
            </div>
            {description && (
              <p className="text-xs text-muted-foreground">{description}</p>
            )}
          </div>
          {Icon && (
            <div className={cn('flex h-9 w-9 items-center justify-center rounded-lg', styles.iconBg)}>
              <Icon className="h-[18px] w-[18px]" />
            </div>
          )}
        </div>

        {/* Progress bar */}
        {progress !== undefined && (
          <div className="mt-3 h-1.5 w-full rounded-full bg-muted">
            <div
              className={cn('h-full rounded-full transition-all', styles.accent)}
              style={{ width: `${Math.min(100, Math.max(0, progress))}%` }}
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}
