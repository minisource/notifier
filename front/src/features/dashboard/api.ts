import type { DashboardData, DashboardMetrics, RecentFailure } from './types';
import type { Notification } from '@/features/notifications/types';
import { adminDashboardApi } from '@/features/notifier/api/notifier-api-mode';
import type { DashboardOverview } from '@/features/notifier/api/notifier-types';

/**
 * The backend wraps all responses in { success: true, data: { ... } }.
 * This helper unwraps the data field, falling back to the raw response
 * if the envelope is not present (e.g., mock mode returns data directly).
 */
function unwrapResponse<T>(response: unknown): T {
  if (
    response &&
    typeof response === 'object' &&
    'data' in response &&
    response.data !== undefined &&
    response.data !== null
  ) {
    return (response as { data: T }).data;
  }
  return response as T;
}

function overviewToDashboardData(raw: unknown): DashboardData {
  const overview = unwrapResponse<DashboardOverview>(raw);

  const metrics: DashboardMetrics = {
    totalNotifications: overview.totalNotifications ?? 0,
    sentToday: overview.sentToday ?? 0,
    failedToday: overview.failedToday ?? 0,
    queued: overview.queuedCount ?? 0,
    deadLetter: overview.deadLetterCount ?? 0,
    deliverySuccessRate: overview.successRate ?? 0,
    avgDeliveryTimeMs: overview.averageDeliveryMs ?? 0,
    activeReminders: overview.activeReminders ?? 0,
    queueDepth: overview.queue?.pendingCount ?? 0,
    channelBreakdown: Array.isArray(overview.channelBreakdown)
      ? (overview.channelBreakdown as Array<{ channel: string; count: number }>).reduce<Record<string, number>>((acc, item) => {
          acc[item.channel] = item.count;
          return acc;
        }, {})
      : (overview.channelBreakdown as Record<string, number>) ?? {},
  };

  return {
    metrics,
    recentNotifications: (overview.recentNotifications ?? []) as unknown as Notification[],
    recentFailures: (overview.recentFailures ?? []).map((f): RecentFailure => ({
      id: f.id ?? f.notificationId ?? '',
      userId: '',
      type: f.channel ?? '',
      errorMessage: f.errorMessage ?? '',
      createdAt: f.createdAt ?? '',
    })),
  };
}

export async function getDashboardData(): Promise<DashboardData> {
  const overview = await adminDashboardApi.getOverview();
  return overviewToDashboardData(overview);
}
