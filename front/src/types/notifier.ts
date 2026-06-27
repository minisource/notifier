export type NotificationType = 'sms' | 'email' | 'push' | 'in_app';
export type NotificationStatus = 'pending' | 'sending' | 'sent' | 'failed' | 'retrying' | 'canceled';
export type NotificationPriority = 'low' | 'normal' | 'high' | 'urgent';
export type DigestFrequency = 'daily' | 'weekly' | 'monthly';

export interface Notification {
  id: string;
  userId: string;
  type: NotificationType;
  status: NotificationStatus;
  priority: NotificationPriority;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  subject?: string;
  body: string;
  metadata?: Record<string, unknown>;
  templateId?: string;
  scheduledAt?: string;
  sentAt?: string;
  readAt?: string;
  retryCount?: number;
  errorMessage?: string;
  provider?: string;
  providerMsgId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface NotificationTemplate {
  id: string;
  name: string;
  type: NotificationType;
  subject?: string;
  body: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface NotificationPreference {
  id: string;
  userId: string;
  type: NotificationType;
  isEnabled: boolean;
  allowInstant: boolean;
  allowDigest: boolean;
  digestFrequency: DigestFrequency;
  quietHours?: string;
  categorySettings?: Record<string, boolean>;
  createdAt: string;
  updatedAt: string;
}

export interface NotificationLog {
  id: string;
  notificationId: string;
  action: string;
  status: NotificationStatus;
  message?: string;
  errorDetails?: string;
  providerResponse?: string;
  processingTimeMs?: number;
  createdAt: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface DashboardStats {
  totalSent: number;
  totalFailed: number;
  totalPending: number;
  totalToday: number;
  byType: Record<string, number>;
  recentNotifications: Notification[];
}

export interface CreateNotificationDto {
  userId: string;
  type: NotificationType;
  priority?: NotificationPriority;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  subject?: string;
  body: string;
  metadata?: Record<string, unknown>;
  templateId?: string;
  scheduledAt?: string;
}

export interface BatchNotificationDto {
  notifications: CreateNotificationDto[];
}

export interface UpdatePreferenceDto {
  type: NotificationType;
  isEnabled?: boolean;
  allowInstant?: boolean;
  allowDigest?: boolean;
  digestFrequency?: DigestFrequency;
  categorySettings?: Record<string, boolean>;
}

export interface CreateTemplateDto {
  name: string;
  type: NotificationType;
  subject?: string;
  body: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
}

export interface Setting {
  id: string;
  key: string;
  value: string;
  category: string;
  description?: string;
  isEncrypted: boolean;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}
