'use client';

import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { SectionCard } from '@/components/shared/section-card';
import { BarChart3 } from 'lucide-react';
interface TrendItem {
  date: string;
  total: number;
  sent: number;
  failed: number;
  dead: number;
}

interface DashboardTrendChartProps {
  data: TrendItem[];
}

export function DashboardTrendChart({ data }: DashboardTrendChartProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';

  if (!data || data.length === 0) {
    return (
      <SectionCard title={t('dashboard.daily_trend') || 'Daily Trend'} icon={BarChart3}>
        <p className="text-sm text-muted-foreground py-8 text-center">{t('common.no_data')}</p>
      </SectionCard>
    );
  }

  const maxValue = Math.max(...data.map(d => d.total), 1);

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString(locale === 'fa' ? 'fa-IR' : 'en-US', { month: 'short', day: 'numeric' });
  };

  // Sort by date ascending
  const sorted = [...data].sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());

  return (
    <SectionCard title={t('dashboard.daily_trend') || 'Daily Trend'} icon={BarChart3}>
      <div className="space-y-3">
        {/* Legend */}
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rounded-sm bg-emerald-500" />
            <span>{t('statuses.sent')}</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rounded-sm bg-red-500" />
            <span>{t('statuses.failed')}</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rounded-sm bg-rose-600" />
            <span>{t('statuses.dead')}</span>
          </div>
        </div>

        {/* Bar Chart */}
        <div className="flex items-end gap-2 pt-2" style={{ height: '160px' }} dir="ltr">
          {sorted.map((item, i) => {
            const sentHeight = (item.sent / item.total) * 100;
            const totalHeight = (item.total / maxValue) * 100;

            return (
              <div key={i} className="flex-1 flex flex-col items-center gap-1 h-full justify-end">
                <div className="relative w-full max-w-[32px] flex flex-col-reverse rounded-t-sm overflow-hidden transition-all hover:opacity-80"
                     style={{ height: `${Math.max(totalHeight, 4)}%` }}>
                  {item.sent > 0 && (
                    <div
                      className="w-full bg-emerald-500/80 dark:bg-emerald-600/80"
                      style={{ height: `${sentHeight > 0 ? (item.sent / item.total) * 100 : 0}%` }}
                      title={`${t('statuses.sent')}: ${item.sent}`}
                    />
                  )}
                  {item.failed > 0 && (
                    <div
                      className="w-full bg-red-500/80 dark:bg-red-600/80"
                      style={{ height: `${(item.failed / item.total) * 100}%` }}
                      title={`${t('statuses.failed')}: ${item.failed}`}
                    />
                  )}
                  {item.dead > 0 && (
                    <div
                      className="w-full bg-rose-600/80 dark:bg-rose-700/80"
                      style={{ height: `${(item.dead / item.total) * 100}%` }}
                      title={`${t('statuses.dead')}: ${item.dead}`}
                    />
                  )}
                </div>
                <span className="text-[10px] text-muted-foreground whitespace-nowrap">
                  {formatDate(item.date)}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </SectionCard>
  );
}
