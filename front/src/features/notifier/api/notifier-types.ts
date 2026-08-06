// Backend-aligned TypeScript types for the Notifier API
// Based on the backend Swagger/OpenAPI specification

// ==================== Common ====================

export type NotificationChannel = 'sms' | 'email' | 'push' | 'in_app' | 'webhook' | 'security';
export type NotificationStatus = 'pending' | 'queued' | 'processing' | 'sent' | 'delivered' | 'failed' | 'retrying' | 'dead' | 'cancelled' | 'canceled';
export type NotificationPriority = 'low' | 'normal' | 'high' | 'urgent';
export type RecipientType = 'email' | 'phone' | 'user_id' | 'device_token' | 'webhook_url';
export type ProviderStatus = 'healthy' | 'degraded' | 'down' | 'disabled' | 'unsupported' | 'unknown';
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
  providerResponseSanitized?: string;
  processingTimeMs: number;
  latencyMs?: number;
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
  lastErrorMessage?: string;
  nextRetryAt?: string;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  subject?: string;
  body?: string;
  completedAt?: string;
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
  tenantId?: string;
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
  isDefault?: boolean;
  priority: number;
  config?: Record<string, unknown>;
}

export interface ProviderHealthItem {
  providerId?: string;
  name: string;
  channel: NotificationChannel;
  type?: string;
  status: ProviderStatus;
  successRate?: number;
  latencyMs?: number;
  message?: string;
  error?: string;
  checkedAt?: string;
}

export interface ProviderHealthResponse {
  providers: ProviderHealthItem[];
  healthyCount: number;
  degradedCount: number;
  downCount: number;
  disabledCount: number;
  checkedAt: string;
}

export interface ProviderTestInput {
  recipient?: string;
  subject?: string;
  body?: string;
  dryRun?: boolean;
}

export interface ProviderTestResult {
  providerId: string;
  channel?: string;
  dryRun: boolean;
  success: boolean;
  status: string;
  message?: string;
  providerMessageId?: string;
  providerResponseSanitized?: string;
  latencyMs?: number;
  checkedAt: string;
}

// ==================== Provider Attempts (request lifecycle logging) ====================

export type ProviderAttemptStatus =
  | 'queued'
  | 'preparing'
  | 'sending'
  | 'accepted'
  | 'pending'
  | 'delivered'
  | 'failed'
  | 'rejected'
  | 'timed_out'
  | 'cancelled'
  | 'bounced'
  | 'complained'
  | 'unknown';

export type ProviderErrorKind =
  | 'configuration'
  | 'invalid_recipient'
  | 'invalid_message'
  | 'provider'
  | 'rate_limited'
  | 'timeout'
  | 'network'
  | 'authentication'
  | 'cancelled'
  | 'unknown';

export interface ProviderAttemptSummary {
  id: string;
  notificationId: string;
  providerAccountId?: string;
  tenantId?: string;
  channel: NotificationChannel;
  provider: string;
  attemptNumber: number;
  fallbackSequence: number;
  status: ProviderAttemptStatus | string;
  providerStatus?: string;
  providerMessageId?: string;
  recipientMasked?: string;
  responseStatusCode?: number;
  durationMs?: number;
  retryable: boolean;
  normalizedErrorKind?: string;
  normalizedErrorCode?: string;
  requestId?: string;
  correlationId?: string;
  traceId?: string;
  createdAt: string;
  completedAt?: string;
}

export interface ProviderAttemptEvent {
  id: string;
  attemptId: string;
  eventType: string;
  previousStatus?: string;
  newStatus?: string;
  eventPayloadSanitized?: Record<string, unknown>;
  source?: string;
  requestId?: string;
  correlationId?: string;
  traceId?: string;
  occurredAt: string;
}

export interface ProviderAttemptDetails extends ProviderAttemptSummary {
  parentAttemptId?: string;
  requestMethod?: string;
  requestUrlSanitized?: string;
  requestHeadersSanitized?: Record<string, string>;
  requestBodySanitized?: string;
  requestSizeBytes?: number;
  responseHeadersSanitized?: Record<string, string>;
  responseBodySanitized?: string;
  responseSizeBytes?: number;
  bodyTruncated: boolean;
  originalSizeBytes?: number;
  capturedSizeBytes?: number;
  contentHash?: string;
  bodyPreview?: string;
  providerErrorCode?: string;
  normalizedErrorMessage?: string;
  queuedAt: string;
  startedAt?: string;
  timeoutMs?: number;
  spanId?: string;
  events?: ProviderAttemptEvent[];
}

export interface ListProviderAttemptsParams extends PaginationParams {
  notificationId?: string;
  channel?: NotificationChannel;
  provider?: string;
  status?: ProviderAttemptStatus | string;
  providerMessageId?: string;
  requestId?: string;
  correlationId?: string;
  from?: string;
  to?: string;
}

export interface ProviderAttemptListResponse {
  items: ProviderAttemptSummary[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// ==================== Global Outbound Delivery Pause / Emergency Freeze ====================

export type DeliveryControlState =
  | 'active'
  | 'pause_requested'
  | 'paused'
  | 'resume_requested'
  | 'active_with_uncertain_attempts';

export type DeliveryControlMode = 'immediate' | 'drain';

export interface DeliveryControlStatus {
  state: DeliveryControlState;
  mode: DeliveryControlMode;
  reason?: string;
  pausedBy?: string;
  pausedAt?: string;
  effectiveAt?: string;
  expiresAt?: string;
  resumedBy?: string;
  resumedAt?: string;
  version: number;
  heldCount: number;
  retryingHeld: number;
  uncertainCount: number;
  activeAttemptCount: number;
  canPause: boolean;
  canResume: boolean;
  lastUpdatedAt: string;
}

export interface DeliveryControlEvent {
  id: string;
  action: string;
  actor?: string;
  reason?: string;
  mode?: string;
  fromState?: string;
  toState?: string;
  version: number;
  requestId?: string;
  createdAt: string;
}

export interface PauseDeliveriesInput {
  mode?: DeliveryControlMode;
  reason: string;
  expiresAt?: string;
  /** Caller's last-known control version. Stale submissions return 409. */
  expectedVersion?: number;
}

export interface ResumeDeliveriesInput {
  reason?: string;
  /** Caller's last-known control version. Stale submissions return 409. */
  expectedVersion?: number;
}

export interface HeldDelivery {
  id: string;
  tenantId?: string;
  userId: string;
  type: NotificationChannel;
  status: NotificationStatus;
  priority: NotificationPriority;
  recipientEmail?: string;
  recipientPhone?: string;
  subject?: string;
  body: string;
  provider?: string;
  pauseVersion?: number;
  heldReason?: string;
  heldAt?: string;
  retryCount: number;
  maxRetries: number;
  createdAt: string;
  updatedAt: string;
}

export interface HeldDeliveriesResponse {
  items: HeldDelivery[];
  total: number;
  page: number;
  pageSize: number;
}

// ==================== Provider Balance / Quota / Credit Alerting ====================

export type ProviderHealthLevel =
  | 'healthy' | 'warning' | 'critical' | 'exhausted'
  | 'stale' | 'unavailable' | 'disabled' | 'unsupported';

export type BalanceCapabilityMode =
  | 'automatic_balance' | 'manual_balance' | 'estimated_from_usage'
  | 'status_only' | 'unsupported';

export interface ProviderAccountHealthSummary {
  providerId: string;
  tenantId?: string;
  provider: string;
  channel?: string;
  capabilityMode: BalanceCapabilityMode | string;
  healthLevel: ProviderHealthLevel | string;
  latestAlertLevel?: string;
  balanceValue?: number | null;
  balanceUnit?: string;
  currency?: string;
  quotaRemaining?: number | null;
  quotaLimit?: number | null;
  usagePercent?: number | null;
  isEstimated: boolean;
  isManual: boolean;
  source?: string;
  lastSuccessfulRefreshAt?: string;
  lastRefreshAttemptAt?: string;
  nextScheduledRefreshAt?: string;
  consecutiveFailures: number;
  lastErrorKind?: string;
  lastErrorMessage?: string;
  activeAlertCount: number;
  updatedAt: string;
}

export interface ProviderBalanceSnapshot {
  id: string;
  providerId: string;
  refreshStatus: 'success' | 'failed' | string;
  capabilityMode: string;
  source: string;
  isEstimated: boolean;
  isManual: boolean;
  balanceValue?: number | null;
  balanceUnit?: string;
  currency?: string;
  quotaRemaining?: number | null;
  quotaLimit?: number | null;
  usagePercent?: number | null;
  accountStatus?: string;
  planExpiresAt?: string;
  errorKind?: string;
  errorCode?: string;
  errorMessage?: string;
  fetchedAt: string;
  latencyMs?: number;
}

export interface CreditAlert {
  id: string;
  providerId: string;
  provider: string;
  alertType: string;
  severity: string;
  status: 'active' | 'acknowledged' | 'resolved' | string;
  message?: string;
  balanceValue?: number | null;
  thresholdValue?: number | null;
  firstTriggeredAt: string;
  lastTriggeredAt: string;
  repeatCount: number;
  acknowledgedAt?: string;
  acknowledgedBy?: string;
  resolvedAt?: string;
  resolvedReason?: string;
}

export interface ProviderBalanceSettings {
  enabled: boolean;
  warningThreshold?: number | null;
  criticalThreshold?: number | null;
  refreshIntervalSec?: number | null;
}

export interface ProviderBalanceDetail {
  providerId: string;
  name: string;
  channel: string;
  type?: string;
  health: ProviderAccountHealthSummary | null;
  history: ProviderBalanceSnapshot[];
  alerts: CreditAlert[];
  settings: ProviderBalanceSettings;
}

export interface BalanceRefreshResult {
  providerId: string;
  name: string;
  channel: string;
  capabilityMode: string;
  success: boolean;
  healthLevel?: string;
  balanceValue?: number | null;
  balanceUnit?: string;
  currency?: string;
  errorKind?: string;
  errorCode?: string;
  errorMessage?: string;
  latencyMs?: number;
  checkedAt: string;
}

// ==================== Template ====================

export interface Template {
  id: string;
  key?: string;
  name: string;
  type: NotificationChannel;
  locale: TemplateLocale;
  subject?: string;
  body?: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
  providerTemplates?: any[];
  status?: TemplateStatus;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTemplateInput {
  key?: string;
  name: string;
  type: NotificationChannel;
  locale: TemplateLocale;
  subject?: string;
  body?: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
  providerTemplates?: any[];
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
  displayName?: string;
  description?: string;
  isActive: boolean;
  isDefault?: boolean;
  enabledChannels: string[];
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
