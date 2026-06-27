import { http } from '@/shared/api/http-client';
import type {
  PaginatedResponse, Notification, ListNotificationsParams,
  NotificationDelivery, DeliveryAttempt, Provider, ProviderHealth,
  ProviderTestInput, Template, CreateTemplateInput, UpdateTemplateInput,
  RenderPreviewInput, Reminder, CreateReminderInput, UpdateReminderInput,
  CreateNotificationInput,
  DashboardOverview, ObservabilityHealth, ReadinessResult,
  ObservabilityMetrics, QueueOverview, WorkerOverview,
  Tenant, PreferenceResponse,
} from './notifier-types';

// ==================== Admin Dashboard ====================

export const adminDashboardApi = {
  getOverview: (params?: { tenantId?: string; projectId?: string; from?: string; to?: string }) =>
    http.get<DashboardOverview>('/admin/dashboard/overview', { params }),

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
  list: () =>
    http.get<Provider[]>('/admin/providers'),

  get: (id: string) =>
    http.get<Provider>(`/admin/providers/${id}`),

  create: (input: { name: string; channel: string; type?: string; config?: Record<string, unknown> }) =>
    http.post<Provider>('/admin/providers', input),

  update: (id: string, input: { name?: string; channel?: string; type?: string; config?: Record<string, unknown>; priority?: number }) =>
    http.put<Provider>(`/admin/providers/${id}`, input),

  delete: (id: string) =>
    http.delete<void>(`/admin/providers/${id}`),

  toggleStatus: (id: string, isEnabled: boolean) =>
    http.patch<Provider>(`/admin/providers/${id}/status`, { isEnabled }),

  getHealth: () =>
    http.get<ProviderHealth[]>('/admin/providers/health'),

  setDefault: (id: string, isDefault: boolean) =>
    http.patch<Provider>(`/admin/providers/${id}/default`, { isDefault }),

  test: (id: string, input?: ProviderTestInput) =>
    http.post<{ success: boolean; message?: string }>(`/admin/providers/${id}/test`, input),
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

// ==================== Admin Tenants ====================

export const adminTenantsApi = {
  list: () =>
    http.get<Tenant[]>('/admin/tenants'),
};

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
