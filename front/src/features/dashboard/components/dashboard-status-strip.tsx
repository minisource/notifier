'use client';

import { useTranslations } from 'next-intl';
import { cn } from '@/lib/utils';
import { Activity, AlertTriangle, Server, Clock, Users } from 'lucide-react';
import type { DashboardData } from '@/features/dashboard/types';

interface StatusStripProps {
  data: DashboardData;
  className?: string;
}

export function DashboardStatusStrip({ data, className }: StatusStripProps) {
  const t = useTranslations();
  const { metrics } = data;
  const degradedProviders = 0; // would come from real provider health data

  // Safe access to numeric fields (defensive against undefined/missing data)
  const deadLetter = metrics.deadLetter ?? 0;
  const queueDepth = metrics.queueDepth ?? 0;
  const sentToday = metrics.sentToday ?? 0;
  const failedToday = metrics.failedToday ?? 0;

  const hasDeadLetter = deadLetter > 0;
  const systemDegraded = failedToday > 50 || degradedProviders > 0 || hasDeadLetter;

  const items = [
    {
      icon: Activity,
      label: t('dashboard.system_status'),
      value: systemDegraded ? t('statuses.degraded') : t('statuses.healthy'),
      variant: systemDegraded ? 'warning' : 'success',
    },
    {
      icon: Clock,
      label: t('dashboard.queue_depth'),
      value: queueDepth.toString(),
      variant: queueDepth > 100 ? 'warning' : 'default',
    },
    {
      icon: Server,
      label: t('dashboard.provider_health'),
      value: '7 configured',
      variant: 'default',
    },
    {
      icon: AlertTriangle,
      label: t('dashboard.dead_letter'),
      value: deadLetter.toString(),
      variant: hasDeadLetter ? 'danger' : 'default',
    },
    {
      icon: Users,
      label: t('dashboard.recent_notifications'),
      value: t('dashboard.sent_today'),
      detail: sentToday.toString(),
      variant: 'default',
    },
  ];

  return (
    <div className={cn(
      'flex flex-wrap items-center gap-3 rounded-lg border border-border/60 bg-card p-3 shadow-sm',
      className
    )}>
      {items.map((item, i) => (
        <div key={i} className="flex items-center gap-2">
          <div className={cn(
            'flex h-7 w-7 items-center justify-center rounded-md',
            item.variant === 'warning' && 'bg-amber-50 text-amber-600 dark:bg-amber-950/40 dark:text-amber-400',
            item.variant === 'danger' && 'bg-red-50 text-red-600 dark:bg-red-950/40 dark:text-red-400',
            item.variant === 'success' && 'bg-green-50 text-green-600 dark:bg-green-950/40 dark:text-green-400',
            item.variant === 'default' && 'bg-muted text-muted-foreground'
          )}>
            <item.icon className="h-3.5 w-3.5" />
          </div>
          <div className="flex flex-col leading-tight">
            <span className="text-[11px] text-muted-foreground">{item.label}</span>
            <span className={cn(
              'text-xs font-semibold',
              item.variant === 'danger' && 'text-red-600 dark:text-red-400',
              item.variant === 'warning' && 'text-amber-600 dark:text-amber-400',
              item.variant === 'success' && 'text-green-600 dark:text-green-400',
            )}>
              {item.detail || item.value}
            </span>
          </div>
          {i < items.length - 1 && <div className="mx-1 h-8 w-px bg-border/50 hidden sm:block" />}
        </div>
      ))}
    </div>
  );
}
