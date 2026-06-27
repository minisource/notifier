export const providersKeys = {
  all: ['providers'] as const,
  lists: () => [...providersKeys.all, 'list'] as const,
  list: () => [...providersKeys.lists()] as const,
  details: () => [...providersKeys.all, 'detail'] as const,
  detail: (id: string) => [...providersKeys.details(), id] as const,
};
