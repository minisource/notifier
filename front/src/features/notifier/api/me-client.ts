import { http } from '@/shared/api/http-client';
import type {
  PaginatedResponse, Notification, ListNotificationsParams,
  Reminder, CreateReminderInput, UpdateReminderInput,
  UserPreference,  UpdatePreferenceInput,
} from './notifier-types';

// ==================== Me Notifications ====================

export const meNotificationsApi = {
  list: (params?: ListNotificationsParams) =>
    http.get<PaginatedResponse<Notification>>('/me/notifications', { params: params as Record<string, string | number | boolean | undefined> }),

  listUnread: () =>
    http.get<Notification[]>('/me/notifications/unread'),

  getUnreadCount: () =>
    http.get<{ count: number }>('/me/notifications/unread-count'),

  get: (id: string) =>
    http.get<Notification>(`/me/notifications/${id}`),

  markRead: (id: string) =>
    http.post<void>(`/me/notifications/${id}/read`),

  markSeen: (id: string) =>
    http.post<void>(`/me/notifications/${id}/seen`),

  markClicked: (id: string) =>
    http.post<void>(`/me/notifications/${id}/click`),

  readAll: () =>
    http.post<void>('/me/notifications/read-all'),
};

// ==================== Me Preferences ====================

export const mePreferencesApi = {
  get: () =>
    http.get<UserPreference>('/me/preferences'),

  update: (input: UpdatePreferenceInput) =>
    http.put<UserPreference>('/me/preferences', input),

  updateChannel: (channel: string, input: Partial<{ isEnabled: boolean; allowInstant: boolean; allowDigest: boolean; digestFrequency: string }>) =>
    http.patch<UserPreference>(`/me/preferences/channel/${channel}`, input),

  updateCategory: (category: string, input: { isEnabled: boolean }) =>
    http.patch<UserPreference>(`/me/preferences/category/${category}`, input),
};

// ==================== Me Reminders ====================

export const meRemindersApi = {
  list: (params?: { status?: string; type?: string }) =>
    http.get<Reminder[]>('/me/reminders', { params: params as Record<string, string | number | boolean | undefined> }),

  get: (id: string) =>
    http.get<Reminder>(`/me/reminders/${id}`),

  create: (input: CreateReminderInput) =>
    http.post<Reminder>('/me/reminders', input),

  update: (id: string, input: UpdateReminderInput) =>
    http.put<Reminder>(`/me/reminders/${id}`, input),

  cancel: (id: string) =>
    http.post<Reminder>(`/me/reminders/${id}/cancel`),

  delete: (id: string) =>
    http.delete<void>(`/me/reminders/${id}`),
};
