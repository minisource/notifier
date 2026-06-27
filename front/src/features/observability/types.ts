export interface HealthStatus {
  status: string;
  uptime: string;
  workers: number;
  queueDepth: number;
}

export interface ObservabilityMetrics {
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
