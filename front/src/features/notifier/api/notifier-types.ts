// Backend-aligned TypeScript types for the Notifier API
// Based on the backend Swagger/OpenAPI specification

// ==================== Common ====================

export type NotificationChannel = 'sms' | 'email' | 'push' | 'in_app' | 'webhook' | 'security';
export type NotificationStatus = 'pending' | 'queued' | 'processing' | 'sent' | 'delivered' | 'failed' | 'retrying' | 'dead' | 'cancelled' | 'canceled';
export type NotificationPriority = 'low' | 'normal' | 'high' | 'urgent';
export type RecipientType = 'email' | 'phone' | 'user_id' | 'device_token' | 'webhook_url';
export type ProviderStatus = 'healthy' | 'degraded' | 'down' | 'disabled' | 'unknown';
export type DeliveryStatus = 'pending' | 'processing' | 'sent' | 'delivered' | 'failed' | 'retrying' | 'dead' | 'read' | 'seen' | 'clicked';
export type ReminderStatus = 'scheduled' | 'processing' | 'sent' | 'cancelled' | 'failed';
export type TemplateLocale = 'fa' | 'en';
export type TemplateStatus = 'active' | 'inactive' | 'archived';
export type UserRole = 'user' | 'admin' | 'operator' | 'service' | 'super_admin';

// ==================== Pagination ====================

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface PaginationParams {
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

// ==================== Error ====================

export interface ErrorDetail {
  code: string;
  message: string;
  details?: unknown;
}

export interface ErrorResponse {
  error: ErrorDetail;
  requestId?: string;
}

// ==================== Notification ====================

export interface Notification {
  id: string;
  tenantId?: string;
  projectId?: string;
  userId: string;
  type: NotificationChannel;
  status: NotificationStatus;
  priority: NotificationPriority;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  recipientType?: RecipientType;
  subject?: string;
  body: string;
  metadata?: Record<string, unknown>;
  templateId?: string;
  templateKey?: string;
  locale: string;
  variables?: Record<string, string>;
  scheduledAt?: string;
  sentAt?: string;
  deliveredAt?: string;
  seenAt?: string;
  readAt?: string;
  clickedAt?: string;
  failedAt?: string;
  retryCount: number;
  maxRetries: number;
  errorMessage?: string;
  errorCode?: string;
  provider?: string;
  providerMsgId?: string;
  idempotencyKey?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ListNotificationsParams extends PaginationParams {
  type?: NotificationChannel;
  status?: NotificationStatus;
  priority?: NotificationPriority;
  userId?: string;
  search?: string;
  startDate?: string;
  endDate?: string;
  tenantId?: string;
  projectId?: string;
}

export interface RecipientInput {
  phone?: string;
  email?: string;
  userId?: string;
  deviceToken?: string;
  webhookUrl?: string;
}

export interface CreateNotificationInput {
  userId?: string;
  channel?: NotificationChannel;
  type?: NotificationChannel;
  priority?: NotificationPriority;
  recipient?: RecipientInput;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  recipientType?: RecipientType;
  subject?: string;
  body: string;
  metadata?: Record<string, unknown>;
  templateId?: string;
  templateKey?: string;
  locale?: string;
  variables?: Record<string, string>;
  scheduledAt?: string;
  idempotencyKey?: string;
  tenantId?: string;
  projectId?: string;
  providerId?: string;
}

export interface BatchNotificationInput {
  notifications: CreateNotificationInput[];
}

// ==================== Notification Delivery / Attempt ====================

export interface DeliveryAttempt {
  id: string;
  deliveryId: string;
  attemptNumber: number;
  status: DeliveryStatus;
  errorMessage?: string;
  errorCode?: string;
  providerResponse?: string;
  processingTimeMs: number;
  createdAt: string;
  completedAt?: string;
}

export interface NotificationDelivery {
  id: string;
  notificationId: string;
  provider: string;
  channel: NotificationChannel;
  status: DeliveryStatus;
  attemptCount: number;
  maxAttempts: number;
  lastError?: string;
  nextRetryAt?: string;
  createdAt: string;
  updatedAt: string;
  attempts: DeliveryAttempt[];
}

export interface ListDeliveriesParams extends PaginationParams {
  status?: DeliveryStatus;
  provider?: string;
  channel?: NotificationChannel;
  failedOnly?: boolean;
  tenantId?: string;
  projectId?: string;
}

// ==================== Provider ====================

export interface Provider {
  id: string;
  name: string;
  channel: NotificationChannel;
  type?: string;
  status: ProviderStatus;
  description?: string;
  successRate: number;
  latencyMs?: number;
  lastFailure?: string;
  isEnabled: boolean;
  isPrimary?: boolean;
  priority: number;
  config?: Record<string, unknown>;
}

export interface ProviderHealth {
  provider: string;
  channel: NotificationChannel;
  status: ProviderStatus;
  successRate: number;
  latencyMs?: number;
  lastChecked?: string;
  error?: string;
}

export interface ProviderTestInput {
  recipient?: string;
  body?: string;
}

export interface ProviderTestResult {
  success: boolean;
  message?: string;
  error?: string;
  processingTimeMs?: number;
}

// ==================== Template ====================

export interface Template {
  id: string;
  key?: string;
  name: string;
  type: NotificationChannel;
  locale: TemplateLocale;
  subject?: string;
  body: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
  status?: TemplateStatus;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTemplateInput {
  name: string;
  type: NotificationChannel;
  locale: TemplateLocale;
  subject?: string;
  body: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
}

export interface UpdateTemplateInput extends Partial<CreateTemplateInput> {
  isActive?: boolean;
  status?: TemplateStatus;
}

export interface RenderPreviewInput {
  templateId: string;
  variables: Record<string, string>;
}

export interface RenderPreviewResult {
  subject?: string;
  body: string;
  missingVariables?: string[];
}

// ==================== Reminder ====================

export interface Reminder {
  id: string;
  tenantId?: string;
  projectId?: string;
  userId: string;
  type: NotificationChannel;
  recipientEmail?: string;
  recipientPhone?: string;
  templateKey?: string;
  variables?: Record<string, string>;
  scheduledAt: string;
  status: ReminderStatus;
  notificationId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateReminderInput {
  userId: string;
  type: NotificationChannel;
  recipientEmail?: string;
  recipientPhone?: string;
  templateKey?: string;
  variables?: Record<string, string>;
  scheduledAt: string;
  tenantId?: string;
  projectId?: string;
}

export interface UpdateReminderInput {
  type?: NotificationChannel;
  recipientEmail?: string;
  recipientPhone?: string;
  templateKey?: string;
  variables?: Record<string, string>;
  scheduledAt?: string;
}

// ==================== Preference ====================

export interface ChannelPreference {
  channel: NotificationChannel;
  isEnabled: boolean;
  allowInstant: boolean;
  allowDigest: boolean;
  digestFrequency: 'daily' | 'weekly' | 'monthly';
  quietHours?: {
    start: string;
    end: string;
    timezone: string;
  };
  categorySettings?: Record<string, boolean>;
}

export interface UserPreference {
  id: string;
  userId: string;
  tenantId?: string;
  projectId?: string;
  channels: ChannelPreference[];
  updatedAt: string;
}

export interface UpdatePreferenceInput {
  channels?: Partial<{
    channel: NotificationChannel;
    isEnabled: boolean;
    allowInstant: boolean;
    allowDigest: boolean;
    digestFrequency: 'daily' | 'weekly' | 'monthly';
    quietHours?: { start: string; end: string; timezone: string };
    categorySettings?: Record<string, boolean>;
  }>[];
  channelOverrides?: Record<string, boolean>;
}

// ==================== Preference ====================

export interface PreferenceResponse {
  id: string;
  userId: string;
  type: string;
  isEnabled: boolean;
  allowInstant: boolean;
  allowDigest: boolean;
  digestFrequency: string;
  quietHours?: { start: string; end: string; timezone: string };
  categorySettings?: Record<string, boolean>;
  updatedAt?: string;
}

// ==================== Tenant ====================

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  isActive: boolean;
  enabledChannels: string[];
  monthlyQuota: number;
  usedThisMonth: number;
  createdAt: string;
}

// ==================== Dashboard ====================

export interface ChannelBreakdownItem {
  channel: NotificationChannel;
  count: number;
  sent: number;
  failed: number;
  successRate: number;
}

export interface StatusBreakdownItem {
  status: NotificationStatus;
  count: number;
}

export interface DailyTrendItem {
  date: string;
  total: number;
  sent: number;
  failed: number;
  dead: number;
}

export interface ProviderHealthSummary {
  healthyCount: number;
  degradedCount: number;
  downCount: number;
  disabledCount: number;
  unknownCount: number;
}

export interface RecentFailure {
  id: string;
  notificationId: string;
  channel: NotificationChannel;
  provider?: string;
  status: string;
  errorCode?: string;
  errorMessage?: string;
  createdAt: string;
  lastAttemptAt?: string;
}

export interface DashboardOverview {
  totalNotifications: number;
  notificationsToday: number;
  sentToday: number;
  failedToday: number;
  queuedCount: number;
  processingCount: number;
  retryingCount: number;
  deadLetterCount: number;
  cancelledCount: number;
  successRate: number;
  failureRate: number;
  averageDeliveryMs: number;
  activeReminders: number;
  scheduledReminders: number;
  failedReminders: number;
  providers: ProviderHealthSummary;
  channelBreakdown: ChannelBreakdownItem[];
  statusBreakdown: StatusBreakdownItem[];
  dailyTrend: DailyTrendItem[];
  recentNotifications: Notification[];
  recentFailures: RecentFailure[];
  recentDeadLetters: RecentFailure[];
  queue: {
    pendingCount: number;
    queuedCount: number;
    processingCount: number;
    retryingCount: number;
    deadCount: number;
  };
  generatedAt: string;
}

// ==================== Observability ====================

export interface DependencyHealth {
  name: string;
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  message?: string;
  latencyMs?: number;
}

export interface ObservabilityHealth {
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  service: string;
  version: string;
  environment: string;
  uptimeSeconds: number;
  dependencies: DependencyHealth[];
  generatedAt: string;
}

export interface ReadinessCheck {
  name: string;
  status: 'ready' | 'not_ready' | 'degraded';
  message?: string;
}

export interface ReadinessResult {
  ready: boolean;
  overall: 'ready' | 'not_ready' | 'degraded';
  checks: ReadinessCheck[];
  generatedAt: string;
}

export interface ObservabilityMetrics {
  notifications: {
    total: number;
    createdToday: number;
    sentToday: number;
    failedToday: number;
    deadToday: number;
    successRate: number;
    failureRate: number;
    averageDeliveryMs: number;
  };
  deliveries: {
    totalAttempts: number;
    failedAttempts: number;
    retrying: number;
    dead: number;
    averageLatencyMs: number;
    p95LatencyMs?: number;
  };
  providers: Record<string, {
    sent: number;
    failed: number;
    successRate: number;
    averageLatencyMs: number;
    health: ProviderStatus;
  }>;
  http?: {
    requestsTotal: number;
    errorsTotal: number;
    averageLatencyMs: number;
    statusCodeBreakdown: Record<string, number>;
  };
  queue: QueueOverview;
  workers: WorkerOverview;
  generatedAt: string;
}

export interface QueueOverview {
  pendingCount: number;
  queuedCount: number;
  processingCount: number;
  retryingCount: number;
  deadCount: number;
  scheduledCount: number;
  oldestPendingAt?: string;
  nextRetryAt?: string;
  throughputPerMinute: number;
  averageLatencyMs: number;
  generatedAt: string;
}

export interface WorkerInfo {
  workerName: string;
  enabled: boolean;
  status: string;
  lastRunAt?: string;
  lastError?: string;
  pollInterval: string;
  batchSize: number;
}

export interface WorkerOverview {
  workers: WorkerInfo[];
  activeCount: number;
  idleCount: number;
  failedCount: number;
  lastHeartbeatAt?: string;
  generatedAt: string;
}
