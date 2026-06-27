import type { Notification } from '@/features/notifications/types';

export interface DashboardMetrics {
  totalNotifications: number;
  sentToday: number;
  failedToday: number;
  queued: number;
  deadLetter: number;
  deliverySuccessRate: number;
  avgDeliveryTimeMs: number;
  activeReminders: number;
  queueDepth: number;
  channelBreakdown: Record<string, number>;
}

export interface RecentFailure {
  id: string;
  userId: string;
  type: string;
  errorMessage: string;
  createdAt: string;
}

export interface DashboardData {
  metrics: DashboardMetrics;
  recentNotifications: Notification[];
  recentFailures: RecentFailure[];
}
