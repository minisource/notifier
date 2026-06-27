import type { Reminder, CreateReminderInput } from './types';
import { adminRemindersApi } from '@/features/notifier/api/notifier-api-mode';
import type { Reminder as NotifierReminder, CreateReminderInput as NotifierCreateReminderInput } from '@/features/notifier/api/notifier-types';

function mapReminder(r: NotifierReminder): Reminder {
  return {
    id: r.id,
    userId: r.userId,
    type: r.type,
    recipientEmail: r.recipientEmail,
    recipientPhone: r.recipientPhone,
    templateKey: r.templateKey,
    variables: r.variables,
    scheduledAt: r.scheduledAt,
    status: r.status as Reminder['status'],
    notificationId: r.notificationId,
    createdAt: r.createdAt,
    updatedAt: r.updatedAt,
  };
}

export async function listReminders(): Promise<Reminder[]> {
  const result = await adminRemindersApi.list();
  // Backend returns paginated { items: [...], total, ... } (uses dto.PaginatedResponse with Items field)
  // Mock returns { data: [...], total, ... } (PaginatedResponse with data field from notifier-types)
  const items = (result as any).items || (result as any).data || [];
  return items.map(mapReminder);
}

export async function getReminder(id: string): Promise<Reminder> {
  const result = await adminRemindersApi.get(id);
  return mapReminder(result);
}

export async function createReminder(input: CreateReminderInput): Promise<Reminder> {
  const notifierInput: NotifierCreateReminderInput = {
    userId: input.userId,
    type: input.type as NotifierReminder['type'],
    recipientEmail: input.recipientEmail,
    recipientPhone: input.recipientPhone,
    templateKey: input.templateKey,
    variables: input.variables,
    scheduledAt: input.scheduledAt,
  };
  const result = await adminRemindersApi.create(notifierInput);
  return mapReminder(result);
}

export async function cancelReminder(id: string): Promise<void> {
  await adminRemindersApi.cancel(id);
}

export async function updateReminder(id: string, input: Partial<CreateReminderInput>): Promise<Reminder> {
  const result = await adminRemindersApi.update(id, input as Record<string, unknown>);
  return mapReminder(result);
}

export async function deleteReminder(id: string): Promise<void> {
  await adminRemindersApi.delete(id);
}
