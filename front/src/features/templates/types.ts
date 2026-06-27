export type TemplateLocale = 'fa' | 'en';

export interface NotificationTemplate {
  id: string;
  key?: string;
  name: string;
  type: string;
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

export interface CreateTemplateInput {
  name: string;
  type: string;
  locale: TemplateLocale;
  subject?: string;
  body: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
}

export interface UpdateTemplateInput extends Partial<CreateTemplateInput> {
  isActive?: boolean;
}
