import { z } from 'zod';

export const createTemplateSchema = z.object({
  name: z.string().min(1, 'forms.required'),
  type: z.string().min(1, 'forms.required'),
  locale: z.string().min(1, 'forms.required'),
  subject: z.string().optional(),
  body: z.string().min(1, 'forms.required'),
  description: z.string().optional(),
  variables: z.string().optional(),
});

export type CreateTemplateFormData = z.infer<typeof createTemplateSchema>;
