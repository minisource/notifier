/**
 * Centralized API Mode Switch
 *
 * This is the ONLY place that decides whether to use mock or real API implementations.
 * Components and hooks must import from here, never directly from mock files.
 *
 * When notifierRuntimeConfig.useMocks is:
 * - true  → mock API implementations are used (no backend needed)
 * - false → real API implementations are used (calls actual backend)
 *
 * Mock auth/session is independent of this switch (controlled separately).
 */

import { notifierRuntimeConfig } from "@/features/notifier/config/notifier-config";

// ==================== Admin API ====================

import {
  adminDashboardApi as realAdminDashboardApi,
  adminNotificationsApi as realAdminNotificationsApi,
  adminDeliveriesApi as realAdminDeliveriesApi,
  adminProvidersApi as realAdminProvidersApi,
  adminTemplatesApi as realAdminTemplatesApi,
  adminRemindersApi as realAdminRemindersApi,
  adminPreferencesApi as realAdminPreferencesApi,
  adminTenantsApi as realAdminTenantsApi,
} from "./notifier-client";

import { notifierMock } from "./notifier-mocks";

// ==================== Me API ====================

import {
  meNotificationsApi as realMeNotificationsApi,
  mePreferencesApi as realMePreferencesApi,
  meRemindersApi as realMeRemindersApi,
} from "./me-client";

// ==================== Switch ====================

const useMocks = notifierRuntimeConfig.useMocks;

// Admin API
export const adminDashboardApi = useMocks
  ? ({
      getOverview: notifierMock.getDashboardOverview,
      getHealth: notifierMock.getHealth,
      getReadiness: notifierMock.getReadiness,
      getMetrics: notifierMock.getMetrics,
      getQueueOverview: notifierMock.getQueueOverview,
      getWorkersOverview: notifierMock.getWorkersOverview,
    } as typeof realAdminDashboardApi)
  : realAdminDashboardApi;

export const adminNotificationsApi = useMocks
  ? ({
      list: notifierMock.listNotifications,
      get: notifierMock.getNotification,
      create: async (input: any) => ({
        id: "mock-" + Date.now(),
        ...input,
        status: "pending",
        retryCount: 0,
        maxRetries: 3,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }),
      retry: async () => {
        throw new Error("Mock not implemented");
      },
      cancel: async () => {
        throw new Error("Mock not implemented");
      },
      markRead: async () => {
        throw new Error("Mock not implemented");
      },
      markSeen: async () => {
        throw new Error("Mock not implemented");
      },
      markClicked: async () => {
        throw new Error("Mock not implemented");
      },
      getAttempts: async () => [],
      getDeliveries: async (_id: string) =>
        (await notifierMock.listDeliveries()).data || [],
    } as unknown as typeof realAdminNotificationsApi)
  : realAdminNotificationsApi;

export const adminDeliveriesApi = useMocks
  ? ({
      list: notifierMock.listDeliveries,
      get: notifierMock.getDelivery,
      retry: async () => {
        throw new Error("Mock not implemented");
      },
    } as unknown as typeof realAdminDeliveriesApi)
  : realAdminDeliveriesApi;

export const adminProvidersApi = useMocks
  ? ({
      list: notifierMock.listProviders,
      get: async (id: string) => {
        const providers = await notifierMock.listProviders();
        const p = providers.find((p: any) => p.id === id);
        if (!p) throw new Error("Provider not found");
        return p;
      },
      create: async (input: any) => ({ id: "mock-" + Date.now(), ...input, status: "healthy", isEnabled: true, successRate: 100, priority: 1, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() }),
      update: async (id: string, input: any) => {
        const providers = await notifierMock.listProviders();
        const p = providers.find((p: any) => p.id === id);
        if (!p) throw new Error("Provider not found");
        return { ...p, ...input, updatedAt: new Date().toISOString() };
      },
      delete: async () => {},
  toggleStatus: async (id: string, isEnabled: boolean) => {
        const providers = await notifierMock.listProviders();
        const p = providers.find((p: any) => p.id === id);
        if (!p) throw new Error("Provider not found");
        return { ...p, isEnabled, updatedAt: new Date().toISOString() };
      },
      setDefault: async (id: string, isDefault: boolean) => {
        const providers = await notifierMock.listProviders();
        const p = providers.find((p: any) => p.id === id);
        if (!p) throw new Error("Provider not found");
        return { ...p, isDefault, updatedAt: new Date().toISOString() };
      },
    } as unknown as typeof realAdminProvidersApi)
  : realAdminProvidersApi;

export const adminTemplatesApi = useMocks
  ? ({
      list: notifierMock.listTemplates,
      get: notifierMock.getTemplate,
      getByKey: async () => {
        throw new Error("Mock not implemented");
      },
      create: async () => {
        throw new Error("Mock not implemented");
      },
      update: async () => {
        throw new Error("Mock not implemented");
      },
      delete: async () => {
        throw new Error("Mock not implemented");
      },
      renderPreview: async () => ({ body: "Mock preview" }),
      renderPreviewById: async () => ({ body: "Mock preview" }),
      updateStatus: async () => {
        throw new Error("Mock not implemented");
      },
    } as unknown as typeof realAdminTemplatesApi)
  : realAdminTemplatesApi;

export const adminRemindersApi = useMocks
  ? ({
      list: notifierMock.listReminders,
      get: notifierMock.getReminder,
      getUserReminders: async () => [],
      create: async () => {
        throw new Error("Mock not implemented");
      },
      update: async () => {
        throw new Error("Mock not implemented");
      },
      delete: async () => {
        throw new Error("Mock not implemented");
      },
      cancel: async () => {
        throw new Error("Mock not implemented");
      },
    } as unknown as typeof realAdminRemindersApi)
  : realAdminRemindersApi;

export const adminTenantsApi = useMocks
  ? ({
      list: notifierMock.listTenants,
    } as unknown as typeof realAdminTenantsApi)
  : realAdminTenantsApi;

export const adminPreferencesApi = useMocks
  ? ({
      list: notifierMock.listPreferences,
      update: notifierMock.updatePreference,
      updateChannel: notifierMock.updatePreference,
      updateCategory: async () => ({ message: "Category preference updated" }),
    } as unknown as typeof realAdminPreferencesApi)
  : realAdminPreferencesApi;

// == Me API ==

export const meNotificationsApi = useMocks
  ? ({
      list: notifierMock.listNotifications,
      listUnread: async () => [],
      getUnreadCount: async () => ({ count: 0 }),
      get: notifierMock.getNotification,
      markRead: async () => {},
      markSeen: async () => {},
      markClicked: async () => {},
      readAll: async () => {},
    } as unknown as typeof realMeNotificationsApi)
  : realMeNotificationsApi;

export const mePreferencesApi = useMocks
  ? ({
      get: notifierMock.getUserPreferences,
      update: async () => {
        throw new Error("Mock not implemented");
      },
      updateChannel: async () => {
        throw new Error("Mock not implemented");
      },
      updateCategory: async () => {
        throw new Error("Mock not implemented");
      },
    } as unknown as typeof realMePreferencesApi)
  : realMePreferencesApi;

export const meRemindersApi = useMocks
  ? ({
      list: notifierMock.listReminders,
      get: notifierMock.getReminder,
      create: async () => {
        throw new Error("Mock not implemented");
      },
      update: async () => {
        throw new Error("Mock not implemented");
      },
      cancel: async () => {
        throw new Error("Mock not implemented");
      },
      delete: async () => {
        throw new Error("Mock not implemented");
      },
    } as unknown as typeof realMeRemindersApi)
  : realMeRemindersApi;
