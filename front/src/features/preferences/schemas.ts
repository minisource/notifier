import { z } from 'zod';

export const updatePreferenceSchema = z.object({
  isEnabled: z.boolean().optional(),
  allowInstant: z.boolean().optional(),
  allowDigest: z.boolean().optional(),
  digestFrequency: z.enum(['daily', 'weekly', 'monthly']).optional(),
});

export type UpdatePreferenceFormData = z.infer<typeof updatePreferenceSchema>;
