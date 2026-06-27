import type { HealthStatus, ObservabilityMetrics } from './types';
import { adminDashboardApi } from '@/features/notifier/api/notifier-api-mode';

export async function getHealth(): Promise<HealthStatus> {
  const result = await adminDashboardApi.getHealth();
  return {
    status: result.status || 'unknown',
    uptime: `${Math.floor((result.uptimeSeconds || 0) / 86400)}d ${Math.floor(((result.uptimeSeconds || 0) % 86400) / 3600)}h`,
    workers: 0,
    queueDepth: 0,
  };
}

export async function getMetrics(): Promise<ObservabilityMetrics> {
  const result = await adminDashboardApi.getMetrics();
  return {
    totalNotifications: result.notifications?.total || 0,
    sentToday: result.notifications?.sentToday || 0,
    failedToday: result.notifications?.failedToday || 0,
    queued: result.queue?.queuedCount || 0,
    deadLetter: result.notifications?.deadToday || 0,
    deliverySuccessRate: result.notifications?.successRate || 0,
    avgDeliveryTimeMs: result.notifications?.averageDeliveryMs || 0,
    activeReminders: result.notifications?.total || 0,
    queueDepth: result.queue?.queuedCount || 0,
    channelBreakdown: {},
  };
}
