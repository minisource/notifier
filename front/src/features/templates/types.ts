export type TemplateLocale = 'fa' | 'en';

export interface ProviderTemplateMap {
  provider: string;
  templateKey: string;
  tenantId?: string;
}

export interface NotificationTemplate {
  id: string;
  key?: string;
  name: string;
  type: string;
  locale: TemplateLocale;
  subject?: string;
  body?: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
  providerTemplates?: ProviderTemplateMap[];
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTemplateInput {
  key?: string;
  name: string;
  type: string;
  locale: TemplateLocale;
  subject?: string;
  body?: string;
  description?: string;
  variables?: string[];
  provider?: string;
  providerTemplate?: string;
  providerTemplates?: ProviderTemplateMap[];
}

export interface UpdateTemplateInput extends Partial<CreateTemplateInput> {
  isActive?: boolean;
}
