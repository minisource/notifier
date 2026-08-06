'use client';

import { useTranslations } from 'next-intl';
import Link from 'next/link';
import { AlertTriangle, PauseCircle } from 'lucide-react';
import { useDeliveryControlStatus } from '../hooks/use-delivery-control';

/**
 * Persistent banner shown on admin pages while the global outbound delivery
 * pause is active. It surfaces who paused it, why, and the held count, and
 * links to the full Delivery Control page. It never claims in-flight requests
 * were recalled — that limitation is stated explicitly.
 */
export function DeliveryControlBanner() {
  const t = useTranslations();
  const { data } = useDeliveryControlStatus();

  const isPaused =
    data?.state === 'paused' ||
    data?.state === 'pause_requested' ||
    data?.state === 'resume_requested';

  if (!isPaused) return null;

  const pausedBy = data?.pausedBy || '—';
  const reason = data?.reason || '—';
  const heldLabel = t('deliveryControl.held_count_label', { count: data?.heldCount ?? 0 });

  return (
    <div
      role="alert"
      className="flex flex-col gap-3 rounded-lg border border-red-300 bg-red-50 p-4 dark:border-red-900/50 dark:bg-red-950/30 sm:flex-row sm:items-center"
    >
      <div className="flex items-start gap-3">
        <PauseCircle className="mt-0.5 h-5 w-5 shrink-0 text-red-600 dark:text-red-400" />
        <div className="space-y-1">
          <p className="text-sm font-semibold text-red-800 dark:text-red-200">
            {t('deliveryControl.banner_title')} · {heldLabel}
          </p>
          <p className="text-sm text-red-700 dark:text-red-300">
            {t('deliveryControl.banner_description')}
          </p>
          <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-red-700/90 dark:text-red-300/90">
            <span>
              {t('deliveryControl.actor')}: <strong>{pausedBy}</strong>
            </span>
            <span>
              {t('deliveryControl.reason')}: <strong>{reason}</strong>
            </span>
          </div>
          <div className="flex items-start gap-1.5 pt-0.5 text-xs text-red-700/80 dark:text-red-300/80">
            <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            <span>{t('deliveryControl.in_flight_warning')}</span>
          </div>
        </div>
      </div>
      <Link
        href="/delivery-control"
        className="ml-auto shrink-0 rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-red-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2"
      >
        {t('deliveryControl.go_to_control')}
      </Link>
    </div>
  );
}
