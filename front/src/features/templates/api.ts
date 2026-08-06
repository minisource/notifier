import type { NotificationTemplate, CreateTemplateInput, UpdateTemplateInput } from './types';
import { adminTemplatesApi } from '@/features/notifier/api/notifier-api-mode';
import type {
  Template as NotifierTemplate,
  CreateTemplateInput as NotifierCreateTemplateInput,
  UpdateTemplateInput as NotifierUpdateTemplateInput,
} from '@/features/notifier/api/notifier-types';

function parseJSONSafe<T>(val: any, fallback: T): T {
  if (!val) return fallback;
  if (typeof val !== 'string') return val;
  try {
    return JSON.parse(val);
  } catch {
    return fallback;
  }
}

function mapTemplate(t: NotifierTemplate): NotificationTemplate {
  return {
    id: t.id,
    key: t.key,
    name: t.name,
    type: t.type,
    locale: t.locale as NotificationTemplate['locale'],
    subject: t.subject,
    body: t.body,
    description: t.description,
    variables: parseJSONSafe<string[]>(t.variables, []),
    provider: t.provider,
    providerTemplate: t.providerTemplate,
    providerTemplates: parseJSONSafe<any[]>(t.providerTemplates, []),
    isActive: t.isActive,
    createdAt: t.createdAt,
    updatedAt: t.updatedAt,
  };
}

export async function listTemplates(params?: { type?: string; locale?: string }): Promise<NotificationTemplate[]> {
  const result = await adminTemplatesApi.list(params as Record<string, string | number | boolean | undefined>);
  // Backend returns paginated { items: [...], total, ... }
  const items = Array.isArray(result) ? result : (result as any).items || [];
  return items.map(mapTemplate);
}

export async function getTemplate(id: string): Promise<NotificationTemplate> {
  const result = await adminTemplatesApi.get(id);
  return mapTemplate(result);
}

export async function createTemplate(input: CreateTemplateInput): Promise<NotificationTemplate> {
  const result = await adminTemplatesApi.create(input as NotifierCreateTemplateInput);
  return mapTemplate(result);
}

export async function updateTemplate(id: string, input: UpdateTemplateInput): Promise<NotificationTemplate> {
  const result = await adminTemplatesApi.update(id, input as NotifierUpdateTemplateInput);
  return mapTemplate(result);
}

export async function deleteTemplate(id: string): Promise<void> {
  await adminTemplatesApi.delete(id);
}

export async function renderTemplatePreview(templateId: string, variables: Record<string, string>) {
  return adminTemplatesApi.renderPreviewById(templateId, variables);
}
