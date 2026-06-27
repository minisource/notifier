'use client';

import { useTranslations } from 'next-intl';
import { FlaskConical, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { isMockDataEnabled } from '@/lib/config/env';
import { useState, useEffect } from 'react';

export function MockModeBanner() {
  const t = useTranslations();
  const [visible, setVisible] = useState(false);
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    setVisible(isMockDataEnabled());
  }, []);

  if (!visible || dismissed) return null;

  return (
    <div className={cn(
      'flex items-center justify-between gap-2 px-4 py-1.5 text-xs font-medium',
      'bg-amber-50 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300',
      'border-b border-amber-200 dark:border-amber-800/50'
    )}>
      <div className="flex items-center gap-2">
        <FlaskConical className="h-3.5 w-3.5" />
        <span>{t('notifier.settings.mockModeActive') || 'Mock Mode Active'}</span>
        <span className="text-amber-600 dark:text-amber-400">
          — {t('notifier.settings.mockModeHint') || 'Data is simulated'}
        </span>
      </div>
      <button
        onClick={() => setDismissed(true)}
        className="flex items-center gap-1 rounded p-0.5 hover:bg-amber-200/50 dark:hover:bg-amber-800/50"
        aria-label={t('common.close')}
      >
        <X className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}
