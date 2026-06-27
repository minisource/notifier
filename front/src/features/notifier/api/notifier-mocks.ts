import type {
  PaginatedResponse, Notification, NotificationDelivery, DeliveryAttempt, DeliveryStatus,
  Provider, ProviderHealth, Template, Reminder, UserPreference,
  DashboardOverview, ObservabilityHealth, ReadinessResult,
  ObservabilityMetrics, QueueOverview, WorkerOverview,
  ChannelBreakdownItem, StatusBreakdownItem, DailyTrendItem,
  RecentFailure, Tenant, PreferenceResponse,
} from './notifier-types';

// ==================== Mock Data Generator ====================

const now = new Date();
const hoursAgo = (h: number) => new Date(now.getTime() - h * 3600000).toISOString();
const minsAgo = (m: number) => new Date(now.getTime() - m * 60000).toISOString();
const daysAgo = (d: number) => new Date(now.getTime() - d * 86400000).toISOString();

function generateId(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16);
  });
}

// ==================== Mock Data ====================

const mockNotifications: any[] = [
  { id: generateId(), userId: 'user-mock-001', type: 'email', status: 'sent', priority: 'high', recipientEmail: 'ahmad.rezaei@example.com', subject: 'Your payment receipt #INV-2024-0034', body: 'Your payment of €1,250.00 has been processed.', locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'smtp', sentAt: hoursAgo(2), deliveredAt: hoursAgo(1.9), createdAt: hoursAgo(2.5), updatedAt: hoursAgo(1.9) },
  { id: generateId(), userId: 'user-mock-002', type: 'sms', status: 'sent', priority: 'urgent', recipientPhone: '+989121234567', body: 'Your verification code is: 48291', locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'kavenegar', sentAt: hoursAgo(0.5), createdAt: hoursAgo(1), updatedAt: hoursAgo(0.5) },
  { id: generateId(), userId: 'user-mock-003', type: 'push', status: 'sent', priority: 'low', recipientId: 'device-fcm-001', subject: 'New message from support', body: 'Your ticket #TKT-4521 has been updated.', locale: 'en', retryCount: 0, maxRetries: 3, provider: 'fcm', sentAt: hoursAgo(3), createdAt: hoursAgo(3.5), updatedAt: hoursAgo(2.9) },
  { id: generateId(), userId: 'user-mock-001', type: 'in_app', status: 'sent', priority: 'normal', recipientId: 'user-mock-001', subject: 'Settlement completed', body: 'Your settlement of €3,420.00 has been completed.', locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'in_app_db', sentAt: hoursAgo(5), createdAt: hoursAgo(5.5), updatedAt: hoursAgo(3) },
  { id: generateId(), userId: 'user-mock-001', type: 'email', status: 'failed', priority: 'normal', recipientEmail: 'invalid@example.com', subject: 'Monthly statement', body: 'Your monthly statement for June 2024 is ready.', locale: 'en', retryCount: 3, maxRetries: 3, errorMessage: 'Recipient mailbox full: 552 5.2.2 mailbox full', provider: 'smtp', sentAt: hoursAgo(8), createdAt: hoursAgo(8.5), updatedAt: hoursAgo(4) },
  { id: generateId(), userId: 'user-mock-003', type: 'sms', status: 'failed', priority: 'high', recipientPhone: '+989139999999', body: 'Your OTP code: 73821', locale: 'fa', retryCount: 3, maxRetries: 3, errorMessage: 'Operator network unreachable: SMPP timeout after 30s', provider: 'kavenegar', sentAt: hoursAgo(12), createdAt: hoursAgo(12.5), updatedAt: hoursAgo(10) },
  { id: generateId(), userId: 'user-mock-002', type: 'email', status: 'dead', priority: 'high', recipientEmail: 'bounced@permanent-failure.com', subject: 'Payment confirmation', body: 'Your payment of €250.00 has been confirmed.', locale: 'en', retryCount: 3, maxRetries: 3, errorMessage: 'Permanent failure: 550 5.1.1 The email account does not exist', provider: 'smtp', sentAt: hoursAgo(24), createdAt: hoursAgo(24.5), updatedAt: hoursAgo(18) },
  { id: generateId(), userId: 'user-mock-003', type: 'push', status: 'dead', priority: 'normal', recipientId: 'device-fcm-002', subject: 'New feature update', body: 'We have added new features to your dashboard.', locale: 'en', retryCount: 5, maxRetries: 5, errorMessage: 'Invalid registration token: UNREGISTERED', provider: 'fcm', sentAt: hoursAgo(48), createdAt: hoursAgo(48.5), updatedAt: hoursAgo(36) },
  { id: generateId(), userId: 'user-mock-001', type: 'email', status: 'processing', priority: 'urgent', recipientEmail: 'retry-target@example.com', subject: 'Urgent: Security alert', body: 'A new login was detected from an unrecognized device.', locale: 'fa', retryCount: 1, maxRetries: 3, errorMessage: 'Temporary failure: Connection timed out', provider: 'smtp', sentAt: hoursAgo(0.1), createdAt: hoursAgo(1), updatedAt: minsAgo(5) },
  { id: generateId(), userId: 'user-mock-002', type: 'sms', status: 'queued', priority: 'normal', recipientPhone: '+989141112223', body: 'Your appointment is confirmed for tomorrow at 10:00 AM.', locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'kavenegar', createdAt: minsAgo(2), updatedAt: minsAgo(2) },
  { id: generateId(), userId: 'user-mock-001', type: 'email', status: 'pending', priority: 'low', recipientEmail: 'newsletter@example.com', subject: 'Weekly digest', body: 'Here is your weekly activity summary...', locale: 'en', retryCount: 0, maxRetries: 3, scheduledAt: hoursAgo(0.5), createdAt: hoursAgo(0.5), updatedAt: hoursAgo(0.5) },
  { id: generateId(), userId: 'user-mock-003', type: 'email', status: 'cancelled', priority: 'high', recipientEmail: 'cancelled-user@example.com', subject: 'Password reset request', body: 'Click here to reset your password.', locale: 'en', retryCount: 0, maxRetries: 3, createdAt: hoursAgo(6), updatedAt: hoursAgo(5) },
  { id: generateId(), userId: 'user-mock-001', type: 'in_app', status: 'sent', priority: 'low', recipientId: 'user-mock-001', subject: 'Welcome!', body: 'Thank you for joining our platform.', locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'in_app_db', createdAt: daysAgo(1), updatedAt: daysAgo(1) },
  { id: generateId(), userId: 'user-mock-002', type: 'email', status: 'failed', priority: 'urgent', recipientEmail: 'urgent-fail@example.com', subject: 'Account suspended', body: 'Your account has been suspended due to suspicious activity.', locale: 'fa', retryCount: 2, maxRetries: 3, errorMessage: 'Provider rate limit exceeded: Max 100 emails/hour', provider: 'smtp', sentAt: hoursAgo(3), createdAt: hoursAgo(3.5), updatedAt: hoursAgo(1) },
  { id: generateId(), userId: 'user-mock-003', type: 'push', status: 'sent', priority: 'high', recipientId: 'device-fcm-003', subject: 'Security alert', body: 'New login from Tehran, Iran at 14:32.', locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'fcm', sentAt: hoursAgo(0.25), createdAt: minsAgo(30), updatedAt: minsAgo(14) },
  { id: generateId(), userId: 'user-mock-001', type: 'sms', status: 'queued', priority: 'normal', recipientPhone: '+989154445566', body: 'Your verification code is: 28401', locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'kavenegar', createdAt: minsAgo(1), updatedAt: minsAgo(1) },
].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());

const mockTemplates: Template[] = [
  { id: generateId(), key: 'auth.otp.sms', name: 'OTP via SMS', type: 'sms', locale: 'en', body: 'Your verification code is: {{code}}', variables: ['code'], isActive: true, createdAt: daysAgo(30), updatedAt: daysAgo(5) },
  { id: generateId(), key: 'auth.otp.sms', name: 'پیامک تأیید', type: 'sms', locale: 'fa', body: 'کد تأیید شما: {{code}}', variables: ['code'], isActive: true, createdAt: daysAgo(30), updatedAt: daysAgo(5) },
  { id: generateId(), key: 'auth.otp.email', name: 'OTP via Email', type: 'email', locale: 'en', subject: 'Your verification code', body: '<p>Your verification code is: <strong>{{code}}</strong></p>', variables: ['code'], isActive: true, createdAt: daysAgo(28), updatedAt: daysAgo(4) },
  { id: generateId(), key: 'auth.otp.email', name: 'ایمیل تأیید', type: 'email', locale: 'fa', subject: 'کد تأیید شما', body: '<p>کد تأیید شما: <strong>{{code}}</strong></p>', variables: ['code'], isActive: true, createdAt: daysAgo(28), updatedAt: daysAgo(4) },
  { id: generateId(), key: 'payment.confirmation', name: 'Payment Confirmation', type: 'email', locale: 'en', subject: 'Payment confirmed - {{amount}}', body: '<p>Your payment of {{amount}} has been confirmed.</p>', variables: ['amount', 'reference'], isActive: true, createdAt: daysAgo(25), updatedAt: daysAgo(3) },
  { id: generateId(), key: 'payment.confirmation', name: 'تأیید پرداخت', type: 'email', locale: 'fa', subject: 'پرداخت {{amount}} تأیید شد', body: '<p>پرداخت {{amount}} شما تأیید شد.</p>', variables: ['amount', 'reference'], isActive: true, createdAt: daysAgo(25), updatedAt: daysAgo(3) },
  { id: generateId(), key: 'generic.notification.email', name: 'Generic Email', type: 'email', locale: 'en', subject: '{{subject}}', body: '<p>{{message}}</p>', variables: ['subject', 'message'], isActive: true, createdAt: daysAgo(20), updatedAt: daysAgo(1) },
  { id: generateId(), key: 'generic.notification.sms', name: 'Generic SMS', type: 'sms', locale: 'en', body: '{{message}}', variables: ['message'], isActive: true, createdAt: daysAgo(20), updatedAt: daysAgo(1) },
  { id: generateId(), key: 'welcome.in_app', name: 'Welcome In-App', type: 'in_app', locale: 'fa', subject: 'به ناتیفایر خوش آمدید', body: 'کاربر گرامی {{name}}، به پلتفرم اعلان‌ها خوش آمدید.', variables: ['name'], isActive: true, createdAt: daysAgo(5), updatedAt: daysAgo(0) },
];

const mockReminders: Reminder[] = [
  { id: generateId(), userId: 'user-mock-001', type: 'email', recipientEmail: 'user1@example.com', templateKey: 'auth.otp.email', scheduledAt: hoursAgo(2), status: 'scheduled', createdAt: daysAgo(3), updatedAt: daysAgo(1) },
  { id: generateId(), userId: 'user-mock-002', type: 'sms', recipientPhone: '+989121234567', templateKey: 'auth.otp.sms', scheduledAt: hoursAgo(-4), status: 'scheduled', createdAt: daysAgo(5), updatedAt: daysAgo(2) },
  { id: generateId(), userId: 'user-mock-003', type: 'push', templateKey: 'welcome.in_app', scheduledAt: hoursAgo(24), status: 'scheduled', createdAt: daysAgo(7), updatedAt: daysAgo(3) },
  { id: generateId(), userId: 'user-mock-001', type: 'email', recipientEmail: 'user1@example.com', templateKey: 'payment.confirmation', scheduledAt: hoursAgo(-12), status: 'sent', createdAt: daysAgo(10), updatedAt: daysAgo(1) },
  { id: generateId(), userId: 'user-mock-002', type: 'sms', recipientPhone: '+989121234567', templateKey: 'auth.otp.sms', scheduledAt: hoursAgo(-48), status: 'cancelled', createdAt: daysAgo(14), updatedAt: daysAgo(2) },
];

const mockDeliveries: NotificationDelivery[] = mockNotifications
  .filter(n => n.status !== 'pending' && n.status !== 'queued')
  .slice(0, 20)
  .map(n => {
    const statusMap: Record<string, DeliveryStatus> = { sent: 'delivered', failed: 'failed', dead: 'dead', processing: 'processing', cancelled: 'failed' };
    const attemptCount = Math.max(1, n.retryCount + 1);
    const attempts: DeliveryAttempt[] = Array.from({ length: attemptCount }, (_, i) => ({
      id: generateId(), deliveryId: `del-${n.id}`, attemptNumber: i + 1,
      status: n.status === 'dead' && i < attemptCount - 1 ? 'failed' : n.status === 'dead' ? 'dead' : n.status === 'sent' && i === attemptCount - 1 ? 'delivered' : 'failed' as DeliveryStatus,
      errorMessage: n.status !== 'sent' ? (n.errorMessage || 'Provider error') : undefined,
      errorCode: n.status === 'dead' ? 'DEAD_LETTER' : undefined,
      providerResponse: n.status === 'sent' ? 'Message accepted by provider' : undefined,
      processingTimeMs: Math.floor(Math.random() * 3000) + 150,
      createdAt: new Date(new Date(n.createdAt).getTime() + i * 60000).toISOString(),
      completedAt: i === attemptCount - 1 ? new Date(new Date(n.createdAt).getTime() + i * 60000 + 30000).toISOString() : undefined,
    }));
    return {
      id: `del-${n.id}`, notificationId: n.id, provider: n.provider || 'mock', channel: n.type,
      status: statusMap[n.status] || 'pending', attemptCount, maxAttempts: n.maxRetries + 1,
      lastError: n.errorMessage,
      nextRetryAt: (n.status === 'processing' || n.status === 'failed') && n.retryCount < n.maxRetries ? new Date(Date.now() + 3600000).toISOString() : undefined,
      createdAt: n.createdAt, updatedAt: n.updatedAt, attempts,
    };
  });

const mockProviders: Provider[] = [
  { id: generateId(), name: 'Kavenegar', channel: 'sms', status: 'healthy', successRate: 98.5, latencyMs: 450, isEnabled: true, priority: 1 },
  { id: generateId(), name: 'SMTP Server', channel: 'email', status: 'healthy', successRate: 99.1, latencyMs: 320, isEnabled: true, priority: 1 },
  { id: generateId(), name: 'SendGrid', channel: 'email', status: 'healthy', successRate: 99.8, latencyMs: 280, isEnabled: true, priority: 2 },
  { id: generateId(), name: 'FCM', channel: 'push', status: 'down', successRate: 45.0, latencyMs: 5000, isEnabled: false, priority: 1 },
  { id: generateId(), name: 'APNs', channel: 'push', status: 'healthy', successRate: 97.3, latencyMs: 380, isEnabled: true, priority: 2 },
  { id: generateId(), name: 'In-App DB', channel: 'in_app', status: 'healthy', successRate: 100.0, latencyMs: 50, isEnabled: true, priority: 1 },
  { id: generateId(), name: 'Twilio', channel: 'sms', status: 'degraded', successRate: 82.1, latencyMs: 1200, isEnabled: false, priority: 3 },
];

const mockProviderHealth: ProviderHealth[] = [
  { provider: 'Kavenegar', channel: 'sms', status: 'healthy', successRate: 98.5, latencyMs: 450, lastChecked: minsAgo(1) },
  { provider: 'SMTP Server', channel: 'email', status: 'healthy', successRate: 99.1, latencyMs: 320, lastChecked: minsAgo(1) },
  { provider: 'SendGrid', channel: 'email', status: 'healthy', successRate: 99.8, latencyMs: 280, lastChecked: minsAgo(1) },
  { provider: 'FCM', channel: 'push', status: 'down', successRate: 45.0, latencyMs: 5000, lastChecked: minsAgo(5), error: 'Connection refused' },
  { provider: 'APNs', channel: 'push', status: 'healthy', successRate: 97.3, latencyMs: 380, lastChecked: minsAgo(1) },
  { provider: 'In-App DB', channel: 'in_app', status: 'healthy', successRate: 100.0, latencyMs: 50, lastChecked: minsAgo(1) },
  { provider: 'Twilio', channel: 'sms', status: 'degraded', successRate: 82.1, latencyMs: 1200, lastChecked: minsAgo(10), error: 'High latency detected' },
];

// ==================== Mock API Responses ====================

function delay(ms: number = 300): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}



function paginate<T>(data: T[], page: number = 1, pageSize: number = 20): PaginatedResponse<T> {
  const start = (page - 1) * pageSize;
  const items = data.slice(start, start + pageSize);
  return { data: items, total: data.length, page, pageSize, totalPages: Math.ceil(data.length / pageSize) };
}

export const notifierMock = {
  // Dashboard
  getDashboardOverview: async (): Promise<DashboardOverview> => {
    await delay();
    const sent = mockNotifications.filter(n => n.status === 'sent').length;
    const failed = mockNotifications.filter(n => n.status === 'failed' || n.status === 'dead').length;
    const total = mockNotifications.length;

    const channelBreakdown: ChannelBreakdownItem[] = [
      { channel: 'sms', count: 5, sent: 3, failed: 1, successRate: 75 },
      { channel: 'email', count: 8, sent: 4, failed: 3, successRate: 57.1 },
      { channel: 'push', count: 3, sent: 3, failed: 0, successRate: 100 },
      { channel: 'in_app', count: 2, sent: 2, failed: 0, successRate: 100 },
    ];

    const statusBreakdown: StatusBreakdownItem[] = [
      { status: 'sent', count: sent },
      { status: 'failed', count: failed },
      { status: 'queued', count: 2 },
      { status: 'processing', count: 1 },
      { status: 'dead', count: 2 },
      { status: 'cancelled', count: 1 },
    ];

    const dailyTrend: DailyTrendItem[] = Array.from({ length: 7 }, (_, i) => ({
      date: daysAgo(i).split('T')[0],
      total: Math.floor(Math.random() * 200) + 50,
      sent: Math.floor(Math.random() * 150) + 30,
      failed: Math.floor(Math.random() * 10),
      dead: Math.floor(Math.random() * 3),
    }));

    const recentFailures: RecentFailure[] = mockNotifications
      .filter(n => n.status === 'failed' || n.status === 'dead')
      .slice(0, 5)
      .map(n => ({
        id: generateId(), notificationId: n.id, channel: n.type,
        provider: n.provider, status: n.status,
        errorCode: n.status === 'dead' ? 'DEAD_LETTER' : 'TEMP_FAILURE',
        errorMessage: n.errorMessage, createdAt: n.createdAt, lastAttemptAt: n.updatedAt,
      }));

    return {
      totalNotifications: total + 124500,
      notificationsToday: 125,
      sentToday: 110,
      failedToday: 3,
      queuedCount: 2,
      processingCount: 1,
      retryingCount: 1,
      deadLetterCount: 2,
      cancelledCount: 1,
      successRate: 92.5,
      failureRate: 7.5,
      averageDeliveryMs: 1240,
      activeReminders: 3,
      scheduledReminders: 5,
      failedReminders: 0,
      providers: { healthyCount: 5, degradedCount: 1, downCount: 1, disabledCount: 0, unknownCount: 0 },
      channelBreakdown,
      statusBreakdown,
      dailyTrend,
      recentNotifications: mockNotifications.slice(0, 10),
      recentFailures,
      recentDeadLetters: recentFailures.filter(r => r.status === 'dead'),
      queue: { pendingCount: 3, queuedCount: 2, processingCount: 1, retryingCount: 1, deadCount: 2 },
      generatedAt: new Date().toISOString(),
    };
  },

  // Notifications
  listNotifications: async (params?: { page?: number; pageSize?: number; status?: string; type?: string }): Promise<PaginatedResponse<Notification>> => {
    await delay();
    let filtered = [...mockNotifications];
    if (params?.status) filtered = filtered.filter(n => n.status === params.status);
    if (params?.type) filtered = filtered.filter(n => n.type === params.type);
    return paginate(filtered, params?.page, params?.pageSize);
  },

  getNotification: async (id: string): Promise<Notification> => {
    await delay();
    const n = mockNotifications.find(n => n.id === id);
    if (!n) throw new Error('Notification not found');
    return n;
  },

  // Templates
  listTemplates: async (params?: { type?: string; locale?: string }): Promise<Template[]> => {
    await delay();
    let filtered = [...mockTemplates];
    if (params?.type) filtered = filtered.filter(t => t.type === params.type);
    if (params?.locale) filtered = filtered.filter(t => t.locale === params.locale);
    return filtered;
  },

  getTemplate: async (id: string): Promise<Template> => {
    await delay();
    const t = mockTemplates.find(t => t.id === id);
    if (!t) throw new Error('Template not found');
    return t;
  },

  // Reminders
  listReminders: async (params?: { status?: string; type?: string }): Promise<PaginatedResponse<Reminder>> => {
    await delay();
    let filtered = [...mockReminders];
    if (params?.status) filtered = filtered.filter(r => r.status === params.status);
    return paginate(filtered);
  },

  getReminder: async (id: string): Promise<Reminder> => {
    await delay();
    const r = mockReminders.find(r => r.id === id);
    if (!r) throw new Error('Reminder not found');
    return r;
  },

  // Deliveries
  listDeliveries: async (params?: { status?: string; provider?: string; page?: number; pageSize?: number }): Promise<PaginatedResponse<NotificationDelivery>> => {
    await delay();
    let filtered = [...mockDeliveries];
    if (params?.status) filtered = filtered.filter(d => d.status === params.status);
    if (params?.provider) filtered = filtered.filter(d => d.provider === params.provider);
    return paginate(filtered, params?.page, params?.pageSize);
  },

  getDelivery: async (id: string): Promise<NotificationDelivery> => {
    await delay();
    const d = mockDeliveries.find(d => d.id === id);
    if (!d) throw new Error('Delivery not found');
    return d;
  },

  // Providers
  listProviders: async (): Promise<Provider[]> => {
    await delay();
    return [...mockProviders];
  },

  getProviderHealth: async (): Promise<ProviderHealth[]> => {
    await delay();
    return [...mockProviderHealth];
  },

  // Tenants
  listTenants: async (): Promise<Tenant[]> => {
    await delay();
    return [
      { id: 'tenant-default', name: 'Default Project', slug: 'default', isActive: true, enabledChannels: ['email', 'sms', 'push', 'in_app', 'webhook'], monthlyQuota: 100000, usedThisMonth: 28450, createdAt: daysAgo(90) },
      { id: 'tenant-divipay', name: 'DiviPay', slug: 'divipay', isActive: true, enabledChannels: ['email', 'sms', 'push'], monthlyQuota: 50000, usedThisMonth: 12300, createdAt: daysAgo(45) },
      { id: 'tenant-auth', name: 'Auth Service', slug: 'auth', isActive: true, enabledChannels: ['email', 'sms', 'in_app'], monthlyQuota: 200000, usedThisMonth: 56700, createdAt: daysAgo(60) },
    ];
  },

  // Preferences
  listPreferences: async (_userId: string): Promise<PreferenceResponse[]> => {
    await delay();
    return [
      { id: generateId(), userId: _userId, type: 'sms', isEnabled: true, allowInstant: true, allowDigest: false, digestFrequency: 'daily', updatedAt: daysAgo(1) },
      { id: generateId(), userId: _userId, type: 'email', isEnabled: true, allowInstant: true, allowDigest: true, digestFrequency: 'daily', updatedAt: daysAgo(1) },
      { id: generateId(), userId: _userId, type: 'push', isEnabled: true, allowInstant: true, allowDigest: false, digestFrequency: 'daily', updatedAt: daysAgo(1) },
      { id: generateId(), userId: _userId, type: 'in_app', isEnabled: true, allowInstant: true, allowDigest: false, digestFrequency: 'daily', updatedAt: daysAgo(1) },
    ];
  },

  updatePreference: async (_userId: string, _type: string, input: any): Promise<PreferenceResponse> => {
    await delay(200);
    return { id: generateId(), userId: _userId, type: _type, isEnabled: input.isEnabled ?? true, allowInstant: input.allowInstant ?? true, allowDigest: input.allowDigest ?? false, digestFrequency: input.digestFrequency || 'daily', updatedAt: new Date().toISOString() };
  },

  // Preferences (legacy)
  getUserPreferences: async (_userId: string): Promise<UserPreference> => {
    await delay();
    return {
      id: generateId(),
      userId: _userId,
      channels: [
        { channel: 'sms', isEnabled: true, allowInstant: true, allowDigest: false, digestFrequency: 'daily' },
        { channel: 'email', isEnabled: true, allowInstant: true, allowDigest: true, digestFrequency: 'daily' },
        { channel: 'push', isEnabled: true, allowInstant: true, allowDigest: false, digestFrequency: 'daily' },
        { channel: 'in_app', isEnabled: true, allowInstant: true, allowDigest: false, digestFrequency: 'daily' },
      ],
      updatedAt: daysAgo(1),
    };
  },

  // Observability
  getHealth: async (): Promise<ObservabilityHealth> => {
    await delay();
    return {
      status: 'healthy', service: 'notifier', version: '1.0.0', environment: 'development',
      uptimeSeconds: 86400 * 14 + 3600 * 6,
      dependencies: [
        { name: 'database', status: 'healthy', latencyMs: 5 },
        { name: 'redis', status: 'healthy', latencyMs: 2 },
        { name: 'worker', status: 'healthy' },
      ],
      generatedAt: new Date().toISOString(),
    };
  },

  getReadiness: async (): Promise<ReadinessResult> => {
    await delay();
    return {
      ready: true, overall: 'ready',
      checks: [
        { name: 'database', status: 'ready' },
        { name: 'redis', status: 'ready' },
        { name: 'config', status: 'ready' },
      ],
      generatedAt: new Date().toISOString(),
    };
  },

  getMetrics: async (): Promise<ObservabilityMetrics> => {
    await delay();
    return {
      notifications: { total: 124530, createdToday: 125, sentToday: 110, failedToday: 3, deadToday: 1, successRate: 92.5, failureRate: 7.5, averageDeliveryMs: 1240 },
      deliveries: { totalAttempts: 145, failedAttempts: 12, retrying: 1, dead: 2, averageLatencyMs: 1240, p95LatencyMs: 3200 },
      providers: {
        smtp: { sent: 4500, failed: 45, successRate: 99.0, averageLatencyMs: 320, health: 'healthy' },
        kavenegar: { sent: 8900, failed: 23, successRate: 99.7, averageLatencyMs: 450, health: 'healthy' },
        fcm: { sent: 1200, failed: 280, successRate: 76.7, averageLatencyMs: 5000, health: 'down' },
      },
      queue: { pendingCount: 3, queuedCount: 2, processingCount: 1, retryingCount: 1, deadCount: 2, scheduledCount: 5, throughputPerMinute: 12.5, averageLatencyMs: 1240, generatedAt: new Date().toISOString() },
      workers: { workers: [], activeCount: 4, idleCount: 0, failedCount: 0, generatedAt: new Date().toISOString() },
      generatedAt: new Date().toISOString(),
    };
  },

  getQueueOverview: async (): Promise<QueueOverview> => {
    await delay();
    return {
      pendingCount: 3, queuedCount: 2, processingCount: 1, retryingCount: 1, deadCount: 2, scheduledCount: 5,
      oldestPendingAt: hoursAgo(2), nextRetryAt: hoursAgo(1), throughputPerMinute: 12.5, averageLatencyMs: 1240,
      generatedAt: new Date().toISOString(),
    };
  },

  getWorkersOverview: async (): Promise<WorkerOverview> => {
    await delay();
    return {
      workers: [
        { workerName: 'notification-worker', enabled: true, status: 'running', lastRunAt: minsAgo(0.5), pollInterval: '5s', batchSize: 10 },
        { workerName: 'reminder-worker', enabled: true, status: 'idle', lastRunAt: minsAgo(2), pollInterval: '30s', batchSize: 50 },
        { workerName: 'dead-letter-worker', enabled: true, status: 'running', lastRunAt: minsAgo(1), pollInterval: '60s', batchSize: 20 },
      ],
      activeCount: 2, idleCount: 1, failedCount: 0, generatedAt: new Date().toISOString(),
    };
  },
};
