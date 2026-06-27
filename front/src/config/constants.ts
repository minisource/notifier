export const CACHE_TIME = {
  SHORT: 1 * 60 * 1000,
  MEDIUM: 5 * 60 * 1000,
  LONG: 30 * 60 * 1000,
  VERY_LONG: 60 * 60 * 1000,
} as const;

export const QUERY_KEYS = {
  notifications: {
    all: ['notifications'] as const,
    list: (userId: string) => ['notifications', userId] as const,
    unread: (userId: string) => ['notifications', userId, 'unread'] as const,
    detail: (id: string) => ['notifications', id] as const,
    stats: () => ['notifications', 'stats'] as const,
  },
  templates: {
    all: ['templates'] as const,
    list: (page: number) => ['templates', 'list', page] as const,
    detail: (id: string) => ['templates', id] as const,
  },
  preferences: {
    all: ['preferences'] as const,
    byUser: (userId: string) => ['preferences', userId] as const,
  },
  admin: {
    notifications: (page: number) => ['admin', 'notifications', page] as const,
    settings: ['admin', 'settings'] as const,
  },
} as const;

export const NOTIFICATION_TYPES = [
  { value: 'email', label: 'Email', color: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400' },
  { value: 'sms', label: 'SMS', color: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' },
  { value: 'push', label: 'Push', color: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400' },
  { value: 'in_app', label: 'In-App', color: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400' },
] as const;

export const NOTIFICATION_STATUSES = [
  { value: 'pending', label: 'Pending', color: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400' },
  { value: 'sending', label: 'Sending', color: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400' },
  { value: 'sent', label: 'Sent', color: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' },
  { value: 'failed', label: 'Failed', color: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400' },
  { value: 'retrying', label: 'Retrying', color: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400' },
  { value: 'canceled', label: 'Canceled', color: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400' },
] as const;

export const PRIORITIES = [
  { value: 'low', label: 'Low', color: 'bg-gray-100 text-gray-800' },
  { value: 'normal', label: 'Normal', color: 'bg-blue-100 text-blue-800' },
  { value: 'high', label: 'High', color: 'bg-orange-100 text-orange-800' },
  { value: 'urgent', label: 'Urgent', color: 'bg-red-100 text-red-800' },
] as const;
