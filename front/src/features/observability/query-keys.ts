export const observabilityKeys = {
  all: ['observability'] as const,
  health: () => [...observabilityKeys.all, 'health'] as const,
  metrics: () => [...observabilityKeys.all, 'metrics'] as const,
};
