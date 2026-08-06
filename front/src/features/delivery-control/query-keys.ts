export const deliveryControlKeys = {
  all: ['delivery-control'] as const,
  status: () => [...deliveryControlKeys.all, 'status'] as const,
  history: () => [...deliveryControlKeys.all, 'history'] as const,
  held: (page = 1, pageSize = 20) => [...deliveryControlKeys.all, 'held', page, pageSize] as const,
};
