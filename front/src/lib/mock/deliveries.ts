export interface MockDeliveryAttempt {
  id: string;
  deliveryId: string;
  attemptNumber: number;
  status: string;
  errorMessage?: string;
  providerResponse?: string;
  processingTimeMs: number;
  createdAt: string;
}

export interface MockDelivery {
  id: string;
  notificationId: string;
  provider: string;
  channel: string;
  status: string;
  attemptCount: number;
  maxAttempts: number;
  lastError?: string;
  nextRetryAt?: string;
  createdAt: string;
  updatedAt: string;
  attempts: MockDeliveryAttempt[];
}
