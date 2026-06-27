export type ReminderStatus = 'scheduled' | 'processing' | 'sent' | 'cancelled' | 'failed';

export interface Reminder {
  id: string;
  userId: string;
  type: string;
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

export interface CreateReminderInput {
  userId: string;
  type: string;
  recipientEmail?: string;
  recipientPhone?: string;
  templateKey?: string;
  variables?: Record<string, string>;
  scheduledAt: string;
}
