import { z } from 'zod';

export const createReminderSchema = z.object({
  userId: z.string().min(1, 'forms.required'),
  type: z.string().min(1, 'forms.required'),
  recipientEmail: z.string().email('forms.invalid_email').optional().or(z.literal('')),
  recipientPhone: z.string().optional(),
  templateKey: z.string().optional(),
  scheduledAt: z.string().min(1, 'forms.required'),
});

export type CreateReminderFormData = z.infer<typeof createReminderSchema>;
