export const preferencesKeys = {
  all: ['preferences'] as const,
  lists: () => [...preferencesKeys.all, 'list'] as const,
  list: (userId?: string) => [...preferencesKeys.lists(), userId] as const,
};
