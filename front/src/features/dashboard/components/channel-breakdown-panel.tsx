'use client';

import { useTranslations } from 'next-intl';
import { BarChart3 } from 'lucide-react';
import { SectionCard } from '@/components/shared/section-card';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { MiniProgress } from '@/components/shared/mini-progress';

interface ChannelBreakdownPanelProps {
  breakdown: Record<string, number>;
}

export function ChannelBreakdownPanel({ breakdown }: ChannelBreakdownPanelProps) {
  const t = useTranslations();
  const total = Object.values(breakdown).reduce((sum, v) => sum + v, 0);

  const channels = [
    { key: 'sms', variant: 'info' as const },
    { key: 'email', variant: 'success' as const },
    { key: 'push', variant: 'warning' as const },
    { key: 'in_app', variant: 'default' as const },
  ];

  return (
    <SectionCard title={t('dashboard.channel_breakdown')} icon={BarChart3}>
      <div className="space-y-3">
        {channels.map(ch => {
          const count = breakdown[ch.key] || 0;
          const percentage = total > 0 ? (count / total) * 100 : 0;
          return (
            <div key={ch.key} className="space-y-1">
              <div className="flex items-center justify-between">
                <ChannelBadge channel={ch.key} size="sm" />
                <div className="flex items-center gap-3">
                  <span className="text-sm font-semibold">{count.toLocaleString()}</span>
                  <span className="text-xs text-muted-foreground w-10 text-right">
                    {percentage.toFixed(0)}%
                  </span>
                </div>
              </div>
              <MiniProgress value={percentage} variant={ch.variant} size="sm" />
            </div>
          );
        })}
      </div>
    </SectionCard>
  );
}
