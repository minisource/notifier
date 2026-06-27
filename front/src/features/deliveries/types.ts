export type DeliveryStatus = 'pending' | 'processing' | 'sent' | 'delivered' | 'failed' | 'retrying' | 'dead' | 'read' | 'seen' | 'clicked';

export interface DeliveryAttempt {
  id: string;
  deliveryId: string;
  attemptNumber: number;
  status: DeliveryStatus;
  errorMessage?: string;
  providerResponse?: string;
  processingTimeMs: number;
  createdAt: string;
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
  createdAt: string;
  updatedAt: string;
  attempts: DeliveryAttempt[];
}

export interface ListDeliveriesParams {
  status?: DeliveryStatus;
  provider?: string;
  failedOnly?: boolean;
}
