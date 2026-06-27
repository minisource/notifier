import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { remindersKeys } from '../query-keys';
import { listReminders, getReminder, createReminder, cancelReminder } from '../api';
import type { CreateReminderInput } from '../types';

export function useReminders() {
  return useQuery({
    queryKey: remindersKeys.lists(),
    queryFn: listReminders,
  });
}

export function useReminder(id: string) {
  return useQuery({
    queryKey: remindersKeys.detail(id),
    queryFn: () => getReminder(id),
    enabled: !!id,
  });
}

export function useCreateReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateReminderInput) => createReminder(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: remindersKeys.lists() });
    },
  });
}

export function useCancelReminder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => cancelReminder(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: remindersKeys.lists() });
    },
  });
}
