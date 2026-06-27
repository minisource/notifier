import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { adminDashboardApi, adminNotificationsApi, adminDeliveriesApi, adminProvidersApi, adminTemplatesApi, adminRemindersApi } from './notifier-api-mode';
import { meNotificationsApi, mePreferencesApi, meRemindersApi } from './notifier-api-mode';
import type {
  ListNotificationsParams,
  CreateTemplateInput, UpdateTemplateInput, RenderPreviewInput,
  CreateReminderInput, UpdateReminderInput,
  UpdatePreferenceInput,
} from './notifier-types';

// ==================== Query Keys ====================

export const notifierKeys = {
  all: ['notifier'] as const,
  dashboard: {
    all: ['notifier', 'dashboard'] as const,
    overview: () => [...notifierKeys.dashboard.all, 'overview'] as const,
    health: () => [...notifierKeys.dashboard.all, 'health'] as const,
    readiness: () => [...notifierKeys.dashboard.all, 'readiness'] as const,
    metrics: () => [...notifierKeys.dashboard.all, 'metrics'] as const,
    queue: () => [...notifierKeys.dashboard.all, 'queue'] as const,
    workers: () => [...notifierKeys.dashboard.all, 'workers'] as const,
  },
  notifications: {
    all: ['notifier', 'notifications'] as const,
    lists: () => [...notifierKeys.notifications.all, 'list'] as const,
    list: (params?: ListNotificationsParams) => [...notifierKeys.notifications.lists(), params] as const,
    details: () => [...notifierKeys.notifications.all, 'detail'] as const,
    detail: (id: string) => [...notifierKeys.notifications.details(), id] as const,
    attempts: (id: string) => [...notifierKeys.notifications.detail(id), 'attempts'] as const,
    deliveries: (id: string) => [...notifierKeys.notifications.detail(id), 'deliveries'] as const,
  },
  deliveries: {
    all: ['notifier', 'deliveries'] as const,
    lists: () => [...notifierKeys.deliveries.all, 'list'] as const,
    list: (params?: Record<string, unknown>) => [...notifierKeys.deliveries.lists(), params] as const,
    detail: (id: string) => [...notifierKeys.deliveries.all, id] as const,
  },
  providers: {
    all: ['notifier', 'providers'] as const,
    list: () => [...notifierKeys.providers.all, 'list'] as const,
    health: () => [...notifierKeys.providers.all, 'health'] as const,
  },
  templates: {
    all: ['notifier', 'templates'] as const,
    lists: () => [...notifierKeys.templates.all, 'list'] as const,
    list: (params?: Record<string, unknown>) => [...notifierKeys.templates.lists(), params] as const,
    details: () => [...notifierKeys.templates.all, 'detail'] as const,
    detail: (id: string) => [...notifierKeys.templates.details(), id] as const,
  },
  reminders: {
    all: ['notifier', 'reminders'] as const,
    lists: () => [...notifierKeys.reminders.all, 'list'] as const,
    list: (params?: Record<string, unknown>) => [...notifierKeys.reminders.lists(), params] as const,
    details: () => [...notifierKeys.reminders.all, 'detail'] as const,
    detail: (id: string) => [...notifierKeys.reminders.details(), id] as const,
  },
  preferences: {
    all: ['notifier', 'preferences'] as const,
    get: (userId?: string) => [...notifierKeys.preferences.all, userId] as const,
  },
  me: {
    notifications: {
      all: ['notifier', 'me', 'notifications'] as const,
      lists: () => [...notifierKeys.me.notifications.all, 'list'] as const,
      list: (params?: ListNotificationsParams) => [...notifierKeys.me.notifications.lists(), params] as const,
      unread: () => [...notifierKeys.me.notifications.all, 'unread'] as const,
      unreadCount: () => [...notifierKeys.me.notifications.all, 'unread-count'] as const,
      detail: (id: string) => [...notifierKeys.me.notifications.all, id] as const,
    },
    preferences: {
      all: ['notifier', 'me', 'preferences'] as const,
      get: () => [...notifierKeys.me.preferences.all, 'get'] as const,
    },
    reminders: {
      all: ['notifier', 'me', 'reminders'] as const,
      lists: () => [...notifierKeys.me.reminders.all, 'list'] as const,
      detail: (id: string) => [...notifierKeys.me.reminders.all, id] as const,
    },
  },
};

// ==================== Dashboard Hooks ====================

export function useAdminDashboardOverview(params?: { tenantId?: string; projectId?: string; from?: string; to?: string }) {
  return useQuery({
    queryKey: notifierKeys.dashboard.overview(),
    queryFn: () => adminDashboardApi.getOverview(params),
    refetchInterval: 30000,
  });
}

export function useAdminHealth() {
  return useQuery({
    queryKey: notifierKeys.dashboard.health(),
    queryFn: () => adminDashboardApi.getHealth(),
    refetchInterval: 30000,
  });
}

export function useAdminReadiness() {
  return useQuery({
    queryKey: notifierKeys.dashboard.readiness(),
    queryFn: () => adminDashboardApi.getReadiness(),
    refetchInterval: 30000,
  });
}

export function useAdminMetrics() {
  return useQuery({
    queryKey: notifierKeys.dashboard.metrics(),
    queryFn: () => adminDashboardApi.getMetrics(),
    refetchInterval: 30000,
  });
}

export function useAdminQueueOverview() {
  return useQuery({
    queryKey: notifierKeys.dashboard.queue(),
    queryFn: () => adminDashboardApi.getQueueOverview(),
    refetchInterval: 15000,
  });
}

export function useAdminWorkersOverview() {
  return useQuery({
    queryKey: notifierKeys.dashboard.workers(),
    queryFn: () => adminDashboardApi.getWorkersOverview(),
    refetchInterval: 30000,
  });
}

// ==================== Notification Hooks ====================

export function useAdminNotifications(params?: ListNotificationsParams) {
  return useQuery({
    queryKey: notifierKeys.notifications.list(params),
    queryFn: () => adminNotificationsApi.list(params),
  });
}

export function useAdminNotification(id: string) {
  return useQuery({
    queryKey: notifierKeys.notifications.detail(id),
    queryFn: () => adminNotificationsApi.get(id),
    enabled: !!id,
  });
}

export function useAdminNotificationAttempts(id: string) {
  return useQuery({
    queryKey: notifierKeys.notifications.attempts(id),
    queryFn: () => adminNotificationsApi.getAttempts(id),
    enabled: !!id,
  });
}

export function useAdminNotificationDeliveries(id: string) {
  return useQuery({
    queryKey: notifierKeys.notifications.deliveries(id),
    queryFn: () => adminNotificationsApi.getDeliveries(id),
    enabled: !!id,
  });
}

export function useRetryNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminNotificationsApi.retry(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.notifications.lists() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.notifications.detail(id) });
      toast.success('Retry initiated', { description: 'Notification has been requeued' });
    },
    onError: () => toast.error('Retry failed'),
  });
}

export function useCancelNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminNotificationsApi.cancel(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.notifications.lists() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.notifications.detail(id) });
      toast.success('Notification cancelled');
    },
    onError: () => toast.error('Failed to cancel notification'),
  });
}

// ==================== Delivery Hooks ====================

export function useAdminDeliveries(params?: { status?: string; provider?: string; page?: number; pageSize?: number }) {
  return useQuery({
    queryKey: notifierKeys.deliveries.list(params as Record<string, unknown>),
    queryFn: () => adminDeliveriesApi.list(params),
  });
}

export function useAdminDelivery(id: string) {
  return useQuery({
    queryKey: notifierKeys.deliveries.detail(id),
    queryFn: () => adminDeliveriesApi.get(id),
    enabled: !!id,
  });
}

export function useRetryDelivery() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminDeliveriesApi.retry(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.deliveries.lists() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.deliveries.detail(id) });
      toast.success('Delivery retry initiated');
    },
    onError: () => toast.error('Delivery retry failed'),
  });
}

// ==================== Provider Hooks ====================

export function useAdminProviders() {
  return useQuery({
    queryKey: notifierKeys.providers.list(),
    queryFn: () => adminProvidersApi.list(),
  });
}

export function useAdminProviderHealth() {
  return useQuery({
    queryKey: notifierKeys.providers.health(),
    queryFn: () => adminProvidersApi.getHealth(),
    refetchInterval: 30000,
  });
}

export function useTestProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input?: { recipient?: string; body?: string } }) =>
      adminProvidersApi.test(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.providers.health() });
      toast.success('Provider test completed');
    },
    onError: () => toast.error('Provider test failed'),
  });
}

// ==================== Template Hooks ====================

export function useAdminTemplates(params?: { type?: string; locale?: string; status?: string }) {
  return useQuery({
    queryKey: notifierKeys.templates.list(params as Record<string, unknown>),
    queryFn: () => adminTemplatesApi.list(params),
  });
}

export function useAdminTemplate(id: string) {
  return useQuery({
    queryKey: notifierKeys.templates.detail(id),
    queryFn: () => adminTemplatesApi.get(id),
    enabled: !!id,
  });
}

export function useCreateTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateTemplateInput) => adminTemplatesApi.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.templates.lists() });
      toast.success('Template created');
    },
    onError: () => toast.error('Failed to create template'),
  });
}

export function useUpdateTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateTemplateInput }) =>
      adminTemplatesApi.update(id, input),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.templates.lists() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.templates.detail(id) });
      toast.success('Template updated');
    },
    onError: () => toast.error('Failed to update template'),
  });
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminTemplatesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.templates.lists() });
      toast.success('Template deleted');
    },
    onError: () => toast.error('Failed to delete template'),
  });
}

export function useRenderTemplatePreview() {
  return useMutation({
    mutationFn: (input: RenderPreviewInput) =>
      adminTemplatesApi.renderPreview(input),
  });
}

// ==================== Reminder Hooks ====================

export function useAdminReminders(params?: { status?: string; type?: string; page?: number; pageSize?: number }) {
  return useQuery({
    queryKey: notifierKeys.reminders.list(params as Record<string, unknown>),
    queryFn: () => adminRemindersApi.list(params),
  });
}

export function useAdminReminder(id: string) {
  return useQuery({
    queryKey: notifierKeys.reminders.detail(id),
    queryFn: () => adminRemindersApi.get(id),
    enabled: !!id,
  });
}

export function useCreateReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateReminderInput) => adminRemindersApi.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.reminders.lists() });
      toast.success('Reminder created');
    },
    onError: () => toast.error('Failed to create reminder'),
  });
}

export function useUpdateReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateReminderInput }) =>
      adminRemindersApi.update(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.reminders.lists() });
      toast.success('Reminder updated');
    },
    onError: () => toast.error('Failed to update reminder'),
  });
}

export function useCancelReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminRemindersApi.cancel(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.reminders.lists() });
      queryClient.invalidateQueries({ queryKey: notifierKeys.reminders.detail(id) });
      toast.success('Reminder cancelled');
    },
    onError: () => toast.error('Failed to cancel reminder'),
  });
}

export function useDeleteReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminRemindersApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.reminders.lists() });
      toast.success('Reminder deleted');
    },
    onError: () => toast.error('Failed to delete reminder'),
  });
}

// ==================== Me Hooks ====================

export function useMeNotifications(params?: ListNotificationsParams) {
  return useQuery({
    queryKey: notifierKeys.me.notifications.list(params),
    queryFn: () => meNotificationsApi.list(params),
  });
}

export function useMeUnreadCount() {
  return useQuery({
    queryKey: notifierKeys.me.notifications.unreadCount(),
    queryFn: () => meNotificationsApi.getUnreadCount(),
    refetchInterval: 15000,
  });
}

export function useMeNotification(id: string) {
  return useQuery({
    queryKey: notifierKeys.me.notifications.detail(id),
    queryFn: () => meNotificationsApi.get(id),
    enabled: !!id,
  });
}

export function useMeReadAllNotifications() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => meNotificationsApi.readAll(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.notifications.all });
      toast.success('All notifications marked as read');
    },
    onError: () => toast.error('Failed to mark all as read'),
  });
}

export function useMePreferences() {
  return useQuery({
    queryKey: notifierKeys.me.preferences.get(),
    queryFn: () => mePreferencesApi.get(),
  });
}

export function useMeUpdatePreference() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdatePreferenceInput) => mePreferencesApi.update(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.preferences.all });
      toast.success('Preferences updated');
    },
    onError: () => toast.error('Failed to update preferences'),
  });
}

export function useMeReminders(params?: { status?: string; type?: string }) {
  return useQuery({
    queryKey: notifierKeys.me.reminders.lists(),
    queryFn: () => meRemindersApi.list(params),
  });
}

export function useMeReminder(id: string) {
  return useQuery({
    queryKey: notifierKeys.me.reminders.detail(id),
    queryFn: () => meRemindersApi.get(id),
    enabled: !!id,
  });
}

export function useMeCreateReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateReminderInput) => meRemindersApi.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.reminders.all });
      toast.success('Reminder created');
    },
    onError: () => toast.error('Failed to create reminder'),
  });
}

export function useMeCancelReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => meRemindersApi.cancel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifierKeys.me.reminders.all });
      toast.success('Reminder cancelled');
    },
    onError: () => toast.error('Failed to cancel reminder'),
  });
}
