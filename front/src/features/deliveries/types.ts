export type DeliveryStatus = 'pending' | 'processing' | 'sent' | 'delivered' | 'failed' | 'retrying' | 'dead' | 'read' | 'seen' | 'clicked';

export interface DeliveryAttempt {
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

export interface Delivery {
  id: string;
  notificationId: string;
  provider: string;
  channel: string;
  status: DeliveryStatus;
  attemptCount: number;
  maxAttempts: number;
  lastError?: string;
  nextRetryAt?: string;
  recipientEmail?: string;
  recipientPhone?: string;
  recipientId?: string;
  subject?: string;
  body?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
  attempts: DeliveryAttempt[];
}

export interface ListDeliveriesParams {
  status?: DeliveryStatus;
  provider?: string;
  failedOnly?: boolean;
}
