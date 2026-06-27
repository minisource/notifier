export const remindersKeys = {
  all: ['reminders'] as const,
  lists: () => [...remindersKeys.all, 'list'] as const,
  list: (params?: Record<string, unknown>) => [...remindersKeys.lists(), params] as const,
  details: () => [...remindersKeys.all, 'detail'] as const,
  detail: (id: string) => [...remindersKeys.details(), id] as const,
};
