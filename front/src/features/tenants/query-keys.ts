export const tenantsKeys = {
  all: ['tenants'] as const,
  lists: () => [...tenantsKeys.all, 'list'] as const,
  list: () => [...tenantsKeys.lists()] as const,
};
