import { http } from '@/shared/api/http-client';
import type {
  PaginatedResponse, Notification, ListNotificationsParams,
  NotificationDelivery, DeliveryAttempt, Provider, ProviderHealthResponse,
  ProviderTestInput, ProviderTestResult, Template, CreateTemplateInput, UpdateTemplateInput,
  RenderPreviewInput, Reminder, CreateReminderInput, UpdateReminderInput,
  CreateNotificationInput,
  DashboardOverview, ObservabilityHealth, ReadinessResult,
  ObservabilityMetrics, QueueOverview, WorkerOverview,
  PreferenceResponse,
  ProviderAttemptSummary, ProviderAttemptDetails, ProviderAttemptEvent,
  ListProviderAttemptsParams, ProviderAttemptListResponse,
  ProviderAccountHealthSummary, ProviderBalanceDetail, ProviderBalanceSettings,
  BalanceRefreshResult, CreditAlert,
  DeliveryControlStatus, DeliveryControlEvent, HeldDeliveriesResponse,
  PauseDeliveriesInput, ResumeDeliveriesInput,
} from './notifier-types';

// ==================== Admin Dashboard ====================

export const adminDashboardApi = {
  getOverview: (params?: { tenantId?: string; projectId?: string; from?: string; to?: string }) =>
    http.get<DashboardOverview>('/admin/notifications/dashboard/overview', { params }),

  getHealth: () =>
    http.get<ObservabilityHealth>('/admin/observability/health'),

  getReadiness: () =>
    http.get<ReadinessResult>('/admin/observability/readiness'),

  getMetrics: () =>
    http.get<ObservabilityMetrics>('/admin/observability/metrics'),

  getQueueOverview: () =>
    http.get<QueueOverview>('/admin/observability/queue'),

  getWorkersOverview: () =>
    http.get<WorkerOverview>('/admin/observability/workers'),
};

// ==================== Admin Notifications ====================

export const adminNotificationsApi = {
  list: (params?: ListNotificationsParams) =>
    http.get<PaginatedResponse<Notification>>('/admin/notifications', { params: params as Record<string, string | number | boolean | undefined> }),

  get: (id: string) =>
    http.get<Notification>(`/admin/notifications/${id}`),

  create: (input: CreateNotificationInput) =>
    http.post<Notification>('/admin/notifications', input),

  retry: (id: string) =>
    http.post<Notification>(`/admin/notifications/${id}/retry`),

  retryMany: (ids: string[]) =>
    http.post<{ retried: number; skipped?: number }>('/admin/notifications/retry', { ids }),

  retryAllFailed: () =>
    http.post<{ retried: number }>('/admin/notifications/retry-failed'),

  cancel: (id: string) =>
    http.post<void>(`/admin/notifications/${id}/cancel`),

  markRead: (id: string) =>
    http.post<void>(`/admin/notifications/${id}/read`),

  markSeen: (id: string) =>
    http.post<void>(`/admin/notifications/${id}/seen`),

  markClicked: (id: string) =>
    http.post<void>(`/admin/notifications/${id}/click`),

  getAttempts: (id: string) =>
    http.get<DeliveryAttempt[]>(`/admin/notifications/${id}/attempts`),

  getDeliveries: (id: string) =>
    http.get<NotificationDelivery[]>(`/admin/notifications/${id}/deliveries`),

  readAll: (userId: string) =>
    http.post<void>(`/admin/notifications/read-all?userId=${userId}`),
};

// ==================== Admin Deliveries ====================

export const adminDeliveriesApi = {
  list: (params?: { status?: string; provider?: string; page?: number; pageSize?: number }) =>
    http.get<PaginatedResponse<NotificationDelivery>>('/admin/deliveries', { params: params as Record<string, string | number | boolean | undefined> }),

  get: (id: string) =>
    http.get<NotificationDelivery>(`/admin/deliveries/${id}`),

  retry: (id: string) =>
    http.post<Notification>(`/admin/deliveries/${id}/retry`),
};

// ==================== Admin Providers ====================

export const adminProvidersApi = {
  list: (params?: { tenantId?: string }) =>
    http.get<Provider[]>('/admin/providers', { params: params as Record<string, string | number | boolean | undefined> | undefined }),

  get: (id: string) =>
    http.get<Provider>(`/admin/providers/${id}`),

  create: (input: { tenantId?: string; name: string; channel: string; type?: string; config?: Record<string, unknown> }) =>
    http.post<Provider>('/admin/providers', input),

  update: (id: string, input: { tenantId?: string; name?: string; channel?: string; type?: string; config?: Record<string, unknown>; priority?: number }) =>
    http.put<Provider>(`/admin/providers/${id}`, input),

  delete: (id: string) =>
    http.delete<void>(`/admin/providers/${id}`),

  toggleStatus: (id: string, isEnabled: boolean) =>
    http.patch<Provider>(`/admin/providers/${id}/status`, { isEnabled }),

  getHealth: () =>
    http.get<ProviderHealthResponse>('/admin/providers/health'),

  healthCheckAll: () =>
    http.post<ProviderHealthResponse>('/admin/providers/health-check'),

  setDefault: (id: string, isDefault: boolean) =>
    http.patch<Provider>(`/admin/providers/${id}/default`, { isDefault }),

  test: (id: string, input?: ProviderTestInput) =>
    http.post<ProviderTestResult>(`/admin/providers/${id}/test`, input),
};

// ==================== Admin Global Delivery Control (Pause / Emergency Freeze) ====================

export const adminDeliveryControlApi = {
  getStatus: () =>
    http.get<DeliveryControlStatus>('/admin/delivery-control/status'),

  // Idempotency-Key makes repeated clicks / browser retries return the
  // original result instead of a duplicate transition. ExpectedVersion is
  // carried in the payload; a stale version returns 409 (no last-write-wins).
  pause: (input: PauseDeliveriesInput, idempotencyKey?: string) =>
    http.post<DeliveryControlStatus>('/admin/delivery-control/pause', input, {
      headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : undefined,
    }),

  resume: (input?: ResumeDeliveriesInput, idempotencyKey?: string) =>
    http.post<DeliveryControlStatus>('/admin/delivery-control/resume', input ?? {}, {
      headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : undefined,
    }),

  getHistory: (params?: { limit?: number }) =>
    http.get<DeliveryControlEvent[]>('/admin/delivery-control/history', { params: params as Record<string, string | number | undefined> | undefined }),

  getHeld: (params?: { page?: number; pageSize?: number }) =>
    http.get<HeldDeliveriesResponse>('/admin/delivery-control/held', { params: params as Record<string, string | number | undefined> | undefined }),
};

// ==================== Admin Provider Balance / Quota ====================

export const adminProviderBalanceApi = {
  listHealth: () =>
    http.get<ProviderAccountHealthSummary[]>('/admin/providers/balance'),

  getDetail: (id: string) =>
    http.get<ProviderBalanceDetail>(`/admin/providers/${id}/balance`),

  refresh: (id: string) =>
    http.post<BalanceRefreshResult>(`/admin/providers/${id}/balance/refresh`),

  updateSettings: (id: string, settings: ProviderBalanceSettings) =>
    http.put<ProviderBalanceSettings>(`/admin/providers/${id}/balance/settings`, settings),

  listAlerts: (params?: { status?: string; providerId?: string }) =>
    http.get<CreditAlert[]>('/admin/providers/balance/alerts', { params: params as Record<string, string | undefined> }),

  acknowledgeAlert: (alertId: string) =>
    http.post<CreditAlert>(`/admin/providers/balance/alerts/${alertId}/acknowledge`),
};

// ==================== Admin Templates ====================

export const adminTemplatesApi = {
  list: (params?: { type?: string; locale?: string; status?: string }) =>
    http.get<Template[]>('/admin/templates', { params: params as Record<string, string | number | boolean | undefined> }),

  get: (id: string) =>
    http.get<Template>(`/admin/templates/${id}`),

  getByKey: (key: string) =>
    http.get<Template>(`/admin/templates/key/${key}`),

  create: (input: CreateTemplateInput) =>
    http.post<Template>('/admin/templates', input),

  update: (id: string, input: UpdateTemplateInput) =>
    http.put<Template>(`/admin/templates/${id}`, input),

  delete: (id: string) =>
    http.delete<void>(`/admin/templates/${id}`),

  renderPreview: (input: RenderPreviewInput) =>
    http.post<{ subject?: string; body: string }>('/admin/templates/render-preview', input),

  renderPreviewById: (id: string, variables: Record<string, string>) =>
    http.post<{ subject?: string; body: string }>(`/admin/templates/${id}/render-preview`, { variables }),

  updateStatus: (id: string, status: string) =>
    http.patch<Template>(`/admin/templates/${id}/status`, { status }),
};

// ==================== Admin Reminders ====================

export const adminRemindersApi = {
  list: (params?: { status?: string; type?: string; page?: number; pageSize?: number }) =>
    http.get<PaginatedResponse<Reminder>>('/admin/reminders', { params: params as Record<string, string | number | boolean | undefined> }),

  get: (id: string) =>
    http.get<Reminder>(`/admin/reminders/${id}`),

  getUserReminders: (userId: string) =>
    http.get<Reminder[]>(`/admin/reminders/user/${userId}`),

  create: (input: CreateReminderInput) =>
    http.post<Reminder>('/admin/reminders', input),

  update: (id: string, input: UpdateReminderInput) =>
    http.put<Reminder>(`/admin/reminders/${id}`, input),

  delete: (id: string) =>
    http.delete<void>(`/admin/reminders/${id}`),

  cancel: (id: string) =>
    http.post<Reminder>(`/admin/reminders/${id}/cancel`),
};

// ==================== Admin Provider Attempts (request lifecycle logs) ====================

export const adminProviderAttemptsApi = {
  list: (params?: ListProviderAttemptsParams) =>
    http.get<ProviderAttemptListResponse>('/admin/attempts', { params: params as Record<string, string | number | boolean | undefined> | undefined }),

  get: (id: string) =>
    http.get<ProviderAttemptDetails>(`/admin/attempts/${id}`),

  getEvents: (id: string) =>
    http.get<ProviderAttemptEvent[]>(`/admin/attempts/${id}/events`),

  listByNotification: (notificationId: string) =>
    http.get<ProviderAttemptSummary[]>(`/admin/notifications/${notificationId}/attempts`),
};

// NOTE: Tenants are owned by the Auth service. The Notifier reads them from
// GET /users/me/tenants (see features/tenants/api.ts) and never manages them.

// ==================== Admin Preferences ====================

export const adminPreferencesApi = {
  list: (userId: string) =>
    http.get<PreferenceResponse[]>(`/admin/preferences/user/${userId}`),

  update: (userId: string, type: string, input: { isEnabled?: boolean; allowInstant?: boolean; allowDigest?: boolean; digestFrequency?: string }) =>
    http.put<PreferenceResponse>(`/admin/preferences/user/${userId}`, { ...input, type }),

  updateChannel: (userId: string, channel: string, input: { isEnabled: boolean; allowInstant?: boolean; allowDigest?: boolean; digestFrequency?: string }) =>
    http.patch<PreferenceResponse>(`/admin/preferences/user/${userId}/channel/${channel}`, input),

  updateCategory: (userId: string, category: string, input: { isEnabled: boolean }) =>
    http.patch<any>(`/admin/preferences/user/${userId}/category/${category}`, input),
};
