// Minimal UUID-like generator without external dependency
function generateId(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16);
  });
}

export type NotificationChannel = 'sms' | 'email' | 'push' | 'in_app' | 'webhook';
export type NotificationStatus = 'pending' | 'queued' | 'processing' | 'sent' | 'failed' | 'dead' | 'cancelled';
export type NotificationPriority = 'low' | 'normal' | 'high' | 'urgent';
export type DeliveryStatus = 'pending' | 'processing' | 'sent' | 'delivered' | 'failed' | 'retrying' | 'dead' | 'read' | 'seen' | 'clicked';
export type ProviderStatus = 'healthy' | 'degraded' | 'down' | 'disabled';
export type ReminderStatus = 'scheduled' | 'processing' | 'sent' | 'cancelled' | 'failed';
export type TemplateLocale = 'fa' | 'en';
export type MockRole = 'admin' | 'operator' | 'viewer';

export interface MockNotification {
  id: string;
  userId: string;
  type: NotificationChannel;
  status: NotificationStatus;
  priority: NotificationPriority;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  subject?: string;
  body: string;
  metadata?: Record<string, unknown>;
  templateId?: string;
  templateKey?: string;
  locale: string;
  scheduledAt?: string;
  sentAt?: string;
  deliveredAt?: string;
  seenAt?: string;
  readAt?: string;
  clickedAt?: string;
  retryCount: number;
  maxRetries: number;
  errorMessage?: string;
  provider?: string;
  providerMsgId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MockTemplate {
  id: string;
  key?: string;
  name: string;
  type: NotificationChannel;
  locale: TemplateLocale;
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

export interface MockReminder {
  id: string;
  userId: string;
  type: NotificationChannel;
  recipientEmail?: string;
  recipientPhone?: string;
  templateKey?: string;
  variables?: Record<string, string>;
  scheduledAt: string;
  status: ReminderStatus;
  notificationId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MockDelivery {
  id: string;
  notificationId: string;
  provider: string;
  channel: NotificationChannel;
  status: DeliveryStatus;
  attemptCount: number;
  maxAttempts: number;
  lastError?: string;
  nextRetryAt?: string;
  createdAt: string;
  updatedAt: string;
  attempts: MockDeliveryAttempt[];
}

export interface MockDeliveryAttempt {
  id: string;
  deliveryId: string;
  attemptNumber: number;
  status: DeliveryStatus;
  errorMessage?: string;
  errorCode?: string;
  providerResponse?: string;
  processingTimeMs: number;
  createdAt: string;
  completedAt?: string;
}

export interface MockProvider {
  id: string;
  name: string;
  channel: NotificationChannel;
  status: ProviderStatus;
  successRate: number;
  latencyMs?: number;
  lastFailure?: string;
  isEnabled: boolean;
  priority: number;
}

export interface MockPreference {
  id: string;
  userId: string;
  type: NotificationChannel;
  isEnabled: boolean;
  allowInstant: boolean;
  allowDigest: boolean;
  digestFrequency: 'daily' | 'weekly' | 'monthly';
  quietHours?: { start: string; end: string; timezone: string };
  categorySettings?: Record<string, boolean>;
  updatedAt: string;
}

export interface MockTenant {
  id: string;
  name: string;
  slug: string;
  isActive: boolean;
  enabledChannels: NotificationChannel[];
  monthlyQuota: number;
  usedThisMonth: number;
  createdAt: string;
}

export interface MockMetric {
  totalNotifications: number;
  sentToday: number;
  failedToday: number;
  queued: number;
  deadLetter: number;
  deliverySuccessRate: number;
  avgDeliveryTimeMs: number;
  activeReminders: number;
  queueDepth: number;
  channelBreakdown: Record<string, number>;
}

// Generate mock data
const now = new Date();
const daysAgo = (d: number) => new Date(now.getTime() - d * 86400000).toISOString();
const hoursAgo = (h: number) => new Date(now.getTime() - h * 3600000).toISOString();
const minsAgo = (m: number) => new Date(now.getTime() - m * 60000).toISOString();

const mockUserIds = ['user-mock-001', 'user-mock-002', 'user-mock-003'];

export const mockNotifications: MockNotification[] = [
  // --- Sent / Delivered Examples ---
  {
    id: generateId(), userId: 'user-mock-001', type: 'email', status: 'sent', priority: 'high',
    recipientEmail: 'ahmad.rezaei@example.com', subject: 'Your payment receipt #INV-2024-0034',
    body: 'Dear Ahmad, your payment of €1,250.00 has been processed successfully.',
    locale: 'fa', templateId: generateId(), templateKey: 'payment.confirmation',
    retryCount: 0, maxRetries: 3, provider: 'smtp', providerMsgId: 'smtp-msg-001',
    sentAt: hoursAgo(2), deliveredAt: hoursAgo(1.9), createdAt: hoursAgo(2.5), updatedAt: hoursAgo(1.9),
  },
  {
    id: generateId(), userId: 'user-mock-002', type: 'sms', status: 'sent', priority: 'urgent',
    recipientPhone: '+989121234567', body: 'Your verification code is: 48291. Valid for 5 minutes.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'kavenegar', providerMsgId: 'kv-msg-002',
    sentAt: hoursAgo(0.5), deliveredAt: hoursAgo(0.45), createdAt: hoursAgo(1), updatedAt: hoursAgo(0.45),
  },
  {
    id: generateId(), userId: 'user-mock-003', type: 'push', status: 'sent', priority: 'low',
    recipientId: 'user-device-fcm-001', subject: 'New message from support',
    body: 'Your ticket #TKT-4521 has been updated by the support team.',
    locale: 'en', retryCount: 0, maxRetries: 3, provider: 'fcm', providerMsgId: 'fcm-msg-003',
    sentAt: hoursAgo(3), deliveredAt: hoursAgo(2.9), createdAt: hoursAgo(3.5), updatedAt: hoursAgo(2.9),
  },
  // --- In-App with Read/Seen/Clicked sequence ---
  {
    id: generateId(), userId: 'user-mock-001', type: 'in_app', status: 'sent', priority: 'normal',
    recipientId: 'user-mock-001', subject: 'Settlement completed',
    body: 'Your settlement of €3,420.00 has been completed and is available in your account.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'in_app_db',
    sentAt: hoursAgo(5), deliveredAt: hoursAgo(4.9), seenAt: hoursAgo(4), readAt: hoursAgo(3.5), clickedAt: hoursAgo(3),
    createdAt: hoursAgo(5.5), updatedAt: hoursAgo(3),
  },
  {
    id: generateId(), userId: 'user-mock-002', type: 'in_app', status: 'sent', priority: 'high',
    recipientId: 'user-mock-002', subject: 'Payment failed',
    body: 'Your payment of €89.00 has failed due to insufficient funds. Please update your payment method.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'in_app_db',
    sentAt: hoursAgo(1), deliveredAt: hoursAgo(0.9), seenAt: hoursAgo(0.5), readAt: minsAgo(20),
    createdAt: hoursAgo(1.5), updatedAt: minsAgo(20),
  },
  // --- Failed Examples ---
  {
    id: generateId(), userId: 'user-mock-001', type: 'email', status: 'failed', priority: 'normal',
    recipientEmail: 'invalid@example.com', subject: 'Monthly statement',
    body: 'Your monthly statement for June 2024 is ready.',
    locale: 'en', retryCount: 3, maxRetries: 3,
    errorMessage: 'Recipient mailbox full: 552 5.2.2 mailbox full',
    provider: 'smtp', providerMsgId: 'smtp-msg-004',
    sentAt: hoursAgo(8), createdAt: hoursAgo(8.5), updatedAt: hoursAgo(4),
  },
  {
    id: generateId(), userId: 'user-mock-003', type: 'sms', status: 'failed', priority: 'high',
    recipientPhone: '+989139999999', body: 'Your OTP code: 73821',
    locale: 'fa', retryCount: 3, maxRetries: 3,
    errorMessage: 'Operator network unreachable: SMPP timeout after 30s',
    provider: 'kavenegar',
    sentAt: hoursAgo(12), createdAt: hoursAgo(12.5), updatedAt: hoursAgo(10),
  },
  // --- Dead Letter Example ---
  {
    id: generateId(), userId: 'user-mock-002', type: 'email', status: 'dead', priority: 'high',
    recipientEmail: 'bounced@permanent-failure.com', subject: 'Payment confirmation',
    body: 'Your payment of €250.00 has been confirmed.',
    locale: 'en', retryCount: 3, maxRetries: 3,
    errorMessage: 'Permanent failure: 550 5.1.1 The email account does not exist',
    provider: 'smtp',
    sentAt: hoursAgo(24), createdAt: hoursAgo(24.5), updatedAt: hoursAgo(18),
  },
  {
    id: generateId(), userId: 'user-mock-003', type: 'push', status: 'dead', priority: 'normal',
    recipientId: 'user-device-fcm-002', subject: 'New feature update',
    body: 'We have added new features to your dashboard.',
    locale: 'en', retryCount: 5, maxRetries: 5,
    errorMessage: 'Invalid registration token: UNREGISTERED device',
    provider: 'fcm',
    sentAt: hoursAgo(48), createdAt: hoursAgo(48.5), updatedAt: hoursAgo(36),
  },
  // --- Retrying Example ---
  {
    id: generateId(), userId: 'user-mock-001', type: 'email', status: 'processing', priority: 'urgent',
    recipientEmail: 'retry-target@example.com', subject: 'Urgent: Security alert',
    body: 'A new login was detected from an unrecognized device.',
    locale: 'fa', retryCount: 1, maxRetries: 3,
    errorMessage: 'Temporary failure: Connection timed out',
    provider: 'smtp', providerMsgId: 'smtp-msg-005',
    sentAt: hoursAgo(0.1), createdAt: hoursAgo(1), updatedAt: minsAgo(5),
  },
  // --- Queued / Pending Examples ---
  {
    id: generateId(), userId: 'user-mock-002', type: 'sms', status: 'queued', priority: 'normal',
    recipientPhone: '+989141112223', body: 'Your appointment is confirmed for tomorrow at 10:00 AM.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'kavenegar',
    createdAt: minsAgo(2), updatedAt: minsAgo(2),
  },
  {
    id: generateId(), userId: 'user-mock-001', type: 'email', status: 'pending', priority: 'low',
    recipientEmail: 'newsletter@example.com', subject: 'Weekly digest - July 2024',
    body: 'Here is your weekly activity summary...',
    locale: 'en', retryCount: 0, maxRetries: 3,
    scheduledAt: hoursAgo(0.5), createdAt: hoursAgo(0.5), updatedAt: hoursAgo(0.5),
  },
  // --- Cancelled Example ---
  {
    id: generateId(), userId: 'user-mock-003', type: 'email', status: 'cancelled', priority: 'high',
    recipientEmail: 'cancelled-user@example.com', subject: 'Password reset request',
    body: 'Click here to reset your password.',
    locale: 'en', retryCount: 0, maxRetries: 3,
    errorMessage: 'Cancelled by user request',
    createdAt: hoursAgo(6), updatedAt: hoursAgo(5),
  },
  // --- Webhook Example ---
  {
    id: generateId(), userId: 'user-mock-001', type: 'webhook', status: 'sent', priority: 'normal',
    recipientId: 'wh-tenant-001', subject: 'Order status update',
    body: JSON.stringify({ event: 'order.shipped', orderId: 'ORD-00234', status: 'shipped' }),
    locale: 'en', retryCount: 0, maxRetries: 3, provider: 'webhook-gateway',
    metadata: { endpoint: 'https://api.example.com/webhooks/notifications', statusCode: 200 },
    sentAt: hoursAgo(1.5), createdAt: hoursAgo(2), updatedAt: hoursAgo(1.5),
  },
  {
    id: generateId(), userId: 'user-mock-002', type: 'webhook', status: 'failed', priority: 'high',
    recipientId: 'wh-tenant-002', subject: 'Payment webhook',
    body: JSON.stringify({ event: 'payment.completed', amount: 1250, currency: 'EUR' }),
    locale: 'en', retryCount: 3, maxRetries: 3,
    errorMessage: 'HTTP 503: Upstream service unavailable',
    provider: 'webhook-gateway',
    metadata: { endpoint: 'https://partner.example.com/hooks', statusCode: 503, retryAfter: 30 },
    sentAt: hoursAgo(4), createdAt: hoursAgo(4.5), updatedAt: hoursAgo(2),
  },
  // --- More diverse examples ---
  {
    id: generateId(), userId: 'user-mock-001', type: 'sms', status: 'sent', priority: 'high',
    recipientPhone: '+989123456789', subject: 'Payment reminder',
    body: 'Dear customer, payment of €49.99 is due tomorrow. Late fee applies after due date.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'kavenegar',
    sentAt: hoursAgo(6), deliveredAt: hoursAgo(5.9), createdAt: hoursAgo(6.5), updatedAt: hoursAgo(5.9),
  },
  {
    id: generateId(), userId: 'user-mock-003', type: 'push', status: 'queued', priority: 'normal',
    recipientId: 'user-device-apns-001', subject: 'New message',
    body: 'You have a new message from the support team.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'apns',
    createdAt: minsAgo(10), updatedAt: minsAgo(10),
  },
  {
    id: generateId(), userId: 'user-mock-002', type: 'in_app', status: 'sent', priority: 'low',
    recipientId: 'user-mock-002', subject: 'Welcome to Notifier!',
    body: 'Thank you for joining our platform. Here are some tips to get started.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'in_app_db',
    seenAt: hoursAgo(20), readAt: hoursAgo(19), createdAt: hoursAgo(24), updatedAt: hoursAgo(19),
  },
  {
    id: generateId(), userId: 'user-mock-001', type: 'email', status: 'failed', priority: 'urgent',
    recipientEmail: 'urgent-fail@example.com', subject: 'Account suspended',
    body: 'Your account has been suspended due to suspicious activity.',
    locale: 'fa', retryCount: 2, maxRetries: 3,
    errorMessage: 'Provider rate limit exceeded: Max 100 emails/hour',
    provider: 'smtp',
    sentAt: hoursAgo(3), createdAt: hoursAgo(3.5), updatedAt: hoursAgo(1),
  },
  {
    id: generateId(), userId: 'user-mock-003', type: 'push', status: 'sent', priority: 'high',
    recipientId: 'user-device-fcm-003', subject: 'Security alert',
    body: 'New login from Tehran, Iran at 14:32. If this was you, ignore this message.',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'fcm',
    sentAt: hoursAgo(0.25), deliveredAt: minsAgo(14), createdAt: minsAgo(30), updatedAt: minsAgo(14),
  },
  {
    id: generateId(), userId: 'user-mock-001', type: 'sms', status: 'queued', priority: 'normal',
    recipientPhone: '+989154445566', body: 'Your verification code is: 28401',
    locale: 'fa', retryCount: 0, maxRetries: 3, provider: 'kavenegar',
    createdAt: minsAgo(1), updatedAt: minsAgo(1),
  },
  {
    id: generateId(), userId: 'user-mock-002', type: 'email', status: 'processing', priority: 'normal',
    recipientEmail: 'processing@example.com', subject: 'Invoice #INV-00521',
    body: 'Please find attached invoice #INV-00521 for the amount of €340.00.',
    locale: 'en', retryCount: 0, maxRetries: 3, provider: 'smtp',
    sentAt: minsAgo(5), createdAt: minsAgo(10), updatedAt: minsAgo(5),
  },
];

// Sort mock notifications by createdAt desc
mockNotifications.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());

// Generate mock delivery attempts for each notification with status sent/failed/dead/processing
export function getMockDeliveries(): MockDelivery[] {
  return mockNotifications
    .filter(n => n.status !== 'pending' && n.status !== 'queued')
    .slice(0, 20)
    .map(n => {
      const statusMap: Record<string, DeliveryStatus> = {
        sent: 'delivered', failed: 'failed', dead: 'dead',
        processing: 'processing', cancelled: 'failed',
      };
      const attempts: MockDeliveryAttempt[] = [];
      const attemptCount = Math.max(1, n.retryCount + 1);
      for (let i = 0; i < attemptCount; i++) {
        const isLast = i === attemptCount - 1;
        const isSuccess = n.status === 'sent';
        const isDead = n.status === 'dead';
        attempts.push({
          id: generateId(),
          deliveryId: `del-${n.id}`,
          attemptNumber: i + 1,
          status: isDead && !isLast ? 'failed' : isDead && isLast ? 'dead' : isSuccess && isLast ? 'delivered' : !isLast ? 'failed' : 'delivered',
          errorMessage: !isSuccess ? (n.errorMessage || 'Provider error') : undefined,
          errorCode: !isSuccess ? (n.status === 'dead' ? 'DEAD_LETTER' : 'TEMP_FAILURE') : undefined,
          providerResponse: isSuccess ? 'Message accepted by provider' : undefined,
          processingTimeMs: Math.floor(Math.random() * 3000) + 150,
          createdAt: new Date(new Date(n.createdAt).getTime() + i * 60000).toISOString(),
          completedAt: isLast ? new Date(new Date(n.createdAt).getTime() + i * 60000 + 30000).toISOString() : undefined,
        });
      }
      return {
        id: `del-${n.id}`,
        notificationId: n.id,
        provider: n.provider || 'mock',
        channel: n.type,
        status: statusMap[n.status] || 'pending',
        attemptCount,
        maxAttempts: n.maxRetries + 1,
        lastError: n.errorMessage,
        nextRetryAt: (n.status === 'processing' || n.status === 'failed') && n.retryCount < n.maxRetries
          ? new Date(Date.now() + 3600000).toISOString() : undefined,
        createdAt: n.createdAt,
        updatedAt: n.updatedAt,
        attempts,
      };
    });
}

export const mockTemplates: MockTemplate[] = [
  { id: generateId(), key: 'auth.otp.sms', name: 'OTP via SMS', type: 'sms', locale: 'en', body: 'Your verification code is: {{code}}', variables: ['code'], isActive: true, createdAt: daysAgo(30), updatedAt: daysAgo(5) },
  { id: generateId(), key: 'auth.otp.sms', name: 'پیامک تأیید', type: 'sms', locale: 'fa', body: 'کد تأیید شما: {{code}}', variables: ['code'], isActive: true, createdAt: daysAgo(30), updatedAt: daysAgo(5) },
  { id: generateId(), key: 'auth.otp.email', name: 'OTP via Email', type: 'email', locale: 'en', subject: 'Your verification code', body: '<p>Your verification code is: <strong>{{code}}</strong></p>', variables: ['code'], isActive: true, createdAt: daysAgo(28), updatedAt: daysAgo(4) },
  { id: generateId(), key: 'auth.otp.email', name: 'ایمیل تأیید', type: 'email', locale: 'fa', subject: 'کد تأیید شما', body: '<p>کد تأیید شما: <strong>{{code}}</strong></p>', variables: ['code'], isActive: true, createdAt: daysAgo(28), updatedAt: daysAgo(4) },
  { id: generateId(), key: 'payment.confirmation', name: 'Payment Confirmation', type: 'email', locale: 'en', subject: 'Payment confirmed - {{amount}}', body: '<p>Your payment of {{amount}} has been confirmed. Reference: {{reference}}</p>', variables: ['amount', 'reference'], isActive: true, createdAt: daysAgo(25), updatedAt: daysAgo(3) },
  { id: generateId(), key: 'payment.confirmation', name: 'تأیید پرداخت', type: 'email', locale: 'fa', subject: 'پرداخت {{amount}} تأیید شد', body: '<p>پرداخت {{amount}} شما تأیید شد. کد پیگیری: {{reference}}</p>', variables: ['amount', 'reference'], isActive: true, createdAt: daysAgo(25), updatedAt: daysAgo(3) },
  { id: generateId(), key: 'generic.notification.email', name: 'Generic Email', type: 'email', locale: 'en', subject: '{{subject}}', body: '<p>{{message}}</p>', variables: ['subject', 'message'], isActive: true, createdAt: daysAgo(20) as unknown as string, updatedAt: daysAgo(1) },
  { id: generateId(), key: 'generic.notification.sms', name: 'Generic SMS', type: 'sms', locale: 'en', body: '{{message}}', variables: ['message'], isActive: true, createdAt: daysAgo(20), updatedAt: daysAgo(1) },
  { id: generateId(), key: 'security.alert', name: 'Security Alert', type: 'email', locale: 'en', subject: 'Security alert: {{action}}', body: '<p>A {{action}} was detected on your account at {{time}}.</p>', variables: ['action', 'time'], isActive: true, createdAt: daysAgo(15), updatedAt: daysAgo(1) },
  { id: generateId(), key: 'security.alert', name: 'هشدار امنیتی', type: 'email', locale: 'fa', subject: 'هشدار امنیتی: {{action}}', body: '<p>{{action}} در حساب شما در ساعت {{time}} تشخیص داده شد.</p>', variables: ['action', 'time'], isActive: true, createdAt: daysAgo(15), updatedAt: daysAgo(1) },
  { id: generateId(), key: 'reminder.push', name: 'Push Reminder', type: 'push', locale: 'en', subject: 'Reminder: {{title}}', body: 'Don\'t forget about {{title}}', variables: ['title'], isActive: true, createdAt: daysAgo(10), updatedAt: daysAgo(1) },
  { id: generateId(), key: 'welcome.in_app', name: 'Welcome In-App', type: 'in_app', locale: 'fa', subject: 'به ناتیفایر خوش آمدید', body: 'کاربر گرامی {{name}}، به پلتفرم اعلان‌ها خوش آمدید.', variables: ['name'], isActive: true, createdAt: daysAgo(5), updatedAt: daysAgo(0) },
];

export const mockReminders: MockReminder[] = Array.from({ length: 8 }, (_, i) => ({
  id: generateId(),
  userId: mockUserIds[Math.floor(Math.random() * mockUserIds.length)],
  type: (['email', 'sms', 'push', 'in_app'] as NotificationChannel[])[Math.floor(Math.random() * 4)],
  recipientEmail: `user${i}@example.com`,
  templateKey: ['auth.otp.sms', 'payment.confirmation', 'reminder.push'][Math.floor(Math.random() * 3)],
  variables: { name: `User ${i}` },
  scheduledAt: hoursAgo(Math.floor(Math.random() * 48) - 24),
  status: (['scheduled', 'scheduled', 'scheduled', 'sent', 'cancelled'] as ReminderStatus[])[Math.floor(Math.random() * 5)],
  createdAt: daysAgo(Math.floor(Math.random() * 14)),
  updatedAt: hoursAgo(Math.floor(Math.random() * 24)),
}));

export const mockProviders: MockProvider[] = [
  { id: generateId(), name: 'Kavenegar', channel: 'sms', status: 'healthy', successRate: 98.5, latencyMs: 450, isEnabled: true, priority: 1 },
  { id: generateId(), name: 'Twilio', channel: 'sms', status: 'degraded', successRate: 82.1, latencyMs: 1200, lastFailure: hoursAgo(2), isEnabled: false, priority: 3 },
  { id: generateId(), name: 'SMTP Server', channel: 'email', status: 'healthy', successRate: 99.1, latencyMs: 320, isEnabled: true, priority: 1 },
  { id: generateId(), name: 'SendGrid', channel: 'email', status: 'healthy', successRate: 99.8, latencyMs: 280, isEnabled: true, priority: 2 },
  { id: generateId(), name: 'FCM', channel: 'push', status: 'down', successRate: 45.0, latencyMs: 5000, lastFailure: hoursAgo(1), isEnabled: false, priority: 1 },
  { id: generateId(), name: 'APNs', channel: 'push', status: 'healthy', successRate: 97.3, latencyMs: 380, isEnabled: true, priority: 2 },
  { id: generateId(), name: 'In-App DB', channel: 'in_app', status: 'healthy', successRate: 100.0, latencyMs: 50, isEnabled: true, priority: 1 },
  { id: generateId(), name: 'Webhook Gateway', channel: 'webhook', status: 'healthy', successRate: 95.2, latencyMs: 650, isEnabled: true, priority: 1 },
];

export const mockPreferences: MockPreference[] = mockUserIds.flatMap(userId =>
  (['sms', 'email', 'push', 'in_app'] as NotificationChannel[]).map(channel => ({
    id: generateId(),
    userId,
    type: channel,
    isEnabled: Math.random() > 0.2,
    allowInstant: Math.random() > 0.3,
    allowDigest: Math.random() > 0.5,
    digestFrequency: (['daily', 'weekly', 'monthly'] as const)[Math.floor(Math.random() * 3)],
    quietHours: Math.random() > 0.7 ? { start: '22:00', end: '08:00', timezone: 'Asia/Tehran' } : undefined,
    categorySettings: Math.random() > 0.6 ? { marketing: false, alerts: true, updates: true } : undefined,
    updatedAt: daysAgo(Math.floor(Math.random() * 7)),
  }))
);

export const mockTenants: MockTenant[] = [
  { id: 'tenant-default', name: 'Default Project', slug: 'default', isActive: true, enabledChannels: ['email', 'sms', 'push', 'in_app', 'webhook'], monthlyQuota: 100000, usedThisMonth: 28450, createdAt: daysAgo(90) },
  { id: 'tenant-divipay', name: 'DiviPay', slug: 'divipay', isActive: true, enabledChannels: ['email', 'sms', 'push'], monthlyQuota: 50000, usedThisMonth: 12300, createdAt: daysAgo(45) },
  { id: 'tenant-auth', name: 'Auth Service', slug: 'auth', isActive: true, enabledChannels: ['email', 'sms', 'in_app'], monthlyQuota: 200000, usedThisMonth: 56700, createdAt: daysAgo(60) },
];

export const mockMetrics: MockMetric = {
  totalNotifications: 125430,
  sentToday: 1234,
  failedToday: 23,
  queued: 45,
  deadLetter: 12,
  deliverySuccessRate: 98.3,
  avgDeliveryTimeMs: 1240,
  activeReminders: 8,
  queueDepth: 45,
  channelBreakdown: { sms: 45200, email: 52100, push: 18130, in_app: 10000, webhook: 5000 },
};
