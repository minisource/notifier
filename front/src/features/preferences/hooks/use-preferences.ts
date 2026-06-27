import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { preferencesKeys } from '../query-keys';
import { listPreferences, updatePreference } from '../api';
import type { UpdatePreferenceInput } from '../types';

export function usePreferences(userId?: string) {
  return useQuery({
    queryKey: preferencesKeys.list(userId),
    queryFn: () => listPreferences(userId),
  });
}

export function useUpdatePreference() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, input, type }: { userId: string; input: UpdatePreferenceInput; type: string }) => updatePreference(userId, input, type),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: preferencesKeys.lists() });
    },
  });
}
