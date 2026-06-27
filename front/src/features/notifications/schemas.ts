import { z } from 'zod';
import type { NotificationChannel, NotificationPriority } from './types';

export const sendNotificationSchema = z.object({
  userId: z.string().min(1, 'forms.required'),
  type: z.string().min(1, 'forms.required') as z.ZodType<NotificationChannel>,
  priority: z.string().optional() as z.ZodType<NotificationPriority | undefined>,
  recipientEmail: z.string().email('forms.invalid_email').optional().or(z.literal('')),
  recipientPhone: z.string().optional(),
  recipientId: z.string().optional(),
  subject: z.string().optional(),
  body: z.string().min(1, 'forms.required'),
  templateId: z.string().optional(),
  locale: z.string().optional(),
  scheduledAt: z.string().optional(),
});

export type SendNotificationFormData = z.infer<typeof sendNotificationSchema>;
