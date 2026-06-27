export const templatesKeys = {
  all: ['templates'] as const,
  lists: () => [...templatesKeys.all, 'list'] as const,
  list: (params?: Record<string, unknown>) => [...templatesKeys.lists(), params] as const,
  details: () => [...templatesKeys.all, 'detail'] as const,
  detail: (id: string) => [...templatesKeys.details(), id] as const,
};
