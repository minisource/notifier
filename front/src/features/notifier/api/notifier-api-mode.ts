/**
 * Notifier API exports.
 *
 * Mock API implementations have been removed. These exports always point to
 * the real backend clients — there is no mock/real switch anymore.
 *
 * Components and hooks import from here (not from notifier-client directly)
 * to keep a single, stable surface for the API layer.
 */

import {
  adminDashboardApi,
  adminNotificationsApi,
  adminDeliveriesApi,
  adminProvidersApi,
  adminProviderBalanceApi,
  adminTemplatesApi,
  adminRemindersApi,
  adminPreferencesApi,
  adminProviderAttemptsApi,
  adminDeliveryControlApi,
} from "./notifier-client";

import {
  meNotificationsApi,
  mePreferencesApi,
  meRemindersApi,
} from "./me-client";

export {
  adminDashboardApi,
  adminNotificationsApi,
  adminDeliveriesApi,
  adminProvidersApi,
  adminProviderBalanceApi,
  adminTemplatesApi,
  adminRemindersApi,
  adminPreferencesApi,
  adminProviderAttemptsApi,
  adminDeliveryControlApi,
  meNotificationsApi,
  mePreferencesApi,
  meRemindersApi,
};
