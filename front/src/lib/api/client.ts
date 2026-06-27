import type {
  MockNotification, MockTemplate, MockReminder,
  MockDelivery, MockProvider, MockPreference, MockTenant, MockMetric,
} from '@/lib/mock/db';
import {
  mockNotifications, mockTemplates, mockReminders,
  mockProviders, mockPreferences, mockTenants, mockMetrics,
} from '@/lib/mock/db';

// Types
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface ApiError {
  message: string;
  code?: string;
  status?: number;
}

function delay(ms: number = 300): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// === Notifications API ===

export interface ListNotificationsParams {
  page?: number;
  pageSize?: number;
  type?: string;
  status?: string;
  userId?: string;
  search?: string;
}

export async function listNotifications(params?: ListNotificationsParams): Promise<PaginatedResponse<MockNotification>> {
  await delay();
  const page = params?.page || 1;
  const pageSize = params?.pageSize || 20;
  let filtered = [...mockNotifications];
  if (params?.type) filtered = filtered.filter(n => n.type === params.type);
  if (params?.status) filtered = filtered.filter(n => n.status === params.status);
  if (params?.search) filtered = filtered.filter(n =>
    n.recipientEmail?.includes(params.search!) ||
    n.recipientPhone?.includes(params.search!) ||
    n.subject?.includes(params.search!)
  );
  filtered.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
  const start = (page - 1) * pageSize;
  return {
    data: filtered.slice(start, start + pageSize),
    total: filtered.length,
    page,
    pageSize,
    totalPages: Math.ceil(filtered.length / pageSize),
  };
}

export async function getNotification(id: string): Promise<MockNotification> {
  await delay();
  const notif = mockNotifications.find(n => n.id === id);
  if (!notif) throw { message: 'Notification not found', status: 404 } as ApiError;
  return notif;
}

export async function sendNotification(data: Partial<MockNotification>): Promise<MockNotification> {
  await delay(500);
  const notification: MockNotification = {
    id: `mock-${Date.now()}`,
    userId: data.userId || 'user-mock-001',
    type: data.type || 'email',
    status: 'pending',
    priority: data.priority || 'normal',
    body: data.body || '',
    locale: data.locale || 'en',
    retryCount: 0,
    maxRetries: 3,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    ...data,
  };
  mockNotifications.unshift(notification);
  return notification;
}

// === Templates API ===

export async function listTemplates(params?: { type?: string; locale?: string }): Promise<MockTemplate[]> {
  await delay();
  let filtered = [...mockTemplates];
  if (params?.type) filtered = filtered.filter(t => t.type === params.type);
  if (params?.locale) filtered = filtered.filter(t => t.locale === params.locale);
  return filtered;
}

export async function getTemplate(id: string): Promise<MockTemplate> {
  await delay();
  const template = mockTemplates.find(t => t.id === id);
  if (!template) throw { message: 'Template not found', status: 404 } as ApiError;
  return template;
}

export async function createTemplate(data: Partial<MockTemplate>): Promise<MockTemplate> {
  await delay(500);
  const template: MockTemplate = {
    id: `mock-${Date.now()}`,
    name: data.name || '',
    type: data.type || 'email',
    locale: data.locale || 'en',
    body: data.body || '',
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    ...data,
  };
  mockTemplates.push(template);
  return template;
}

export async function renderTemplatePreview(templateId: string, variables: Record<string, string>): Promise<{ subject: string; body: string }> {
  await delay();
  const template = await getTemplate(templateId);
  let subject = template.subject || '';
  let body = template.body;
  for (const [key, value] of Object.entries(variables)) {
    subject = subject.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), value);
    body = body.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), value);
  }
  return { subject, body };
}

// === Reminders API ===

export async function listReminders(): Promise<MockReminder[]> {
  await delay();
  return [...mockReminders].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
}

export async function createReminder(data: Partial<MockReminder>): Promise<MockReminder> {
  await delay(500);
  const reminder: MockReminder = {
    id: `mock-${Date.now()}`,
    userId: data.userId || 'user-mock-001',
    type: data.type || 'email',
    scheduledAt: data.scheduledAt || new Date().toISOString(),
    status: 'scheduled',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    ...data,
  };
  mockReminders.unshift(reminder);
  return reminder;
}

// === Deliveries API ===

export async function listDeliveries(params?: { status?: string; provider?: string; failedOnly?: boolean }): Promise<MockDelivery[]> {
  await delay();
  // Generate mock deliveries from failed/retrying notifications
  const deliveries: MockDelivery[] = mockNotifications
    .filter(n => n.status === 'failed' || n.status === 'processing' || n.status === 'dead' || n.status === 'sent')
    .slice(0, 30)
    .map(n => ({
      id: `del-${n.id}`,
      notificationId: n.id,
      provider: n.provider || 'mock',
      channel: n.type,
      status: (n.status === 'sent' ? 'delivered' : n.status === 'processing' ? 'retrying' : n.status === 'dead' ? 'dead' : 'failed') as any,
      attemptCount: n.retryCount + 1,
      maxAttempts: n.maxRetries + 1,
      lastError: n.errorMessage,
      nextRetryAt: n.status === 'processing' ? new Date(Date.now() + 3600000).toISOString() : undefined,
      createdAt: n.createdAt,
      updatedAt: n.updatedAt,
      attempts: Array.from({ length: n.retryCount + 1 }, (_, i) => ({
        id: `att-${n.id}-${i}`,
        deliveryId: `del-${n.id}`,
        attemptNumber: i + 1,
        status: i < n.retryCount ? 'failed' : (n.status === 'sent' ? 'delivered' : n.status === 'dead' ? 'dead' : 'failed') as any,
        errorMessage: i < n.retryCount ? 'Provider timeout' : undefined,
        providerResponse: i < n.retryCount ? undefined : 'Message accepted',
        processingTimeMs: Math.floor(Math.random() * 2000) + 200,
        createdAt: new Date(new Date(n.createdAt).getTime() + i * 3600000).toISOString(),
      })),
    }));
  if (params?.failedOnly) return deliveries.filter(d => d.status === 'failed' || d.status === 'dead');
  return deliveries;
}

// === Providers API ===

export async function listProviders(): Promise<MockProvider[]> {
  await delay();
  return mockProviders;
}

// === Preferences API ===

export async function listPreferences(userId?: string): Promise<MockPreference[]> {
  await delay();
  if (userId) return mockPreferences.filter(p => p.userId === userId);
  return mockPreferences;
}

// === Tenants API ===

export async function listTenants(): Promise<MockTenant[]> {
  await delay();
  return mockTenants;
}

// === Observability / Metrics API ===

export async function getMetrics(): Promise<MockMetric> {
  await delay();
  return { ...mockMetrics };
}

export async function getHealthStatus(): Promise<{ status: string; uptime: string; workers: number; queueDepth: number }> {
  await delay();
  return {
    status: 'healthy',
    uptime: '14d 6h 32m',
    workers: 10,
    queueDepth: 45,
  };
}
