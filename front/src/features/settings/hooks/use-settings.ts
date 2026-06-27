'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { fetchNotificationSettings, updateNotificationSettings } from '../api';
import type { NotificationSettings } from '../api';

export const settingsKeys = {
  all: ['settings'] as const,
  notifications: () => ['settings', 'notifications'] as const,
};

export function useNotificationSettings() {
  return useQuery({
    queryKey: settingsKeys.notifications(),
    queryFn: fetchNotificationSettings,
    staleTime: 5 * 60 * 1000,
  });
}

export function useUpdateNotificationSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: Partial<NotificationSettings>) =>
      updateNotificationSettings(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.notifications() });
      toast.success('Settings saved');
    },
    onError: () => {
      toast.error('Failed to save settings');
    },
  });
}
