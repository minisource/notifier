'use client';

import { useTranslations } from 'next-intl';
import { Clock, AlertTriangle, Activity, Bell } from 'lucide-react';
import { MetricCard } from '@/components/shared/metric-card';

interface QueueStatusPanelProps {
  queued: number;
  deadLetter: number;
  activeReminders: number;
  queueDepth: number;
  avgDeliveryTimeMs: number;
}

export function QueueStatusPanel({
  queued,
  deadLetter,
  activeReminders,
  avgDeliveryTimeMs,
}: QueueStatusPanelProps) {
  const t = useTranslations();

  const avgTimeDisplay = avgDeliveryTimeMs < 1000
    ? `${avgDeliveryTimeMs}ms`
    : `${(avgDeliveryTimeMs / 1000).toFixed(1)}s`;

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <MetricCard
        title={t('dashboard.queued')}
        value={queued}
        icon={Clock}
        variant="info"
        description={t('dashboard.queue_depth')}
      />
      <MetricCard
        title={t('dashboard.dead_letter')}
        value={deadLetter}
        icon={AlertTriangle}
        variant={deadLetter > 0 ? 'danger' : 'default'}
        trend={deadLetter > 0 ? { value: t('common.needs_attention'), positive: false } : undefined}
      />
      <MetricCard
        title={t('dashboard.active_reminders')}
        value={activeReminders}
        icon={Bell}
        variant="default"
      />
      <MetricCard
        title={t('dashboard.avg_delivery_time')}
        value={avgTimeDisplay}
        icon={Activity}
        variant="success"
      />
    </div>
  );
}
