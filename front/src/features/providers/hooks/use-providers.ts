import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { providersKeys } from '../query-keys';
import { listProviders, testProvider, createProvider, updateProvider, deleteProvider, toggleProviderStatus, getProvider, getProviderHealth, setDefaultProvider } from '../api';

export function useProviders() {
  return useQuery({
    queryKey: providersKeys.list(),
    queryFn: listProviders,
  });
}

export function useProvider(id: string) {
  return useQuery({
    queryKey: providersKeys.detail(id),
    queryFn: () => getProvider(id),
    enabled: !!id,
  });
}

export function useProviderHealth() {
  return useQuery({
    queryKey: ['providers', 'health'],
    queryFn: getProviderHealth,
  });
}

export function useCreateProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { name: string; channel: string; type?: string; status?: string; priority?: number; isDefault?: boolean; description?: string; config?: Record<string, unknown>; secretConfig?: Record<string, unknown> }) => createProvider(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: providersKeys.lists() });
      toast.success('Provider created');
    },
    onError: (err: Error) => {
      toast.error('Failed to create provider', { description: err.message });
    },
  });
}

export function useUpdateProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: { name?: string; channel?: string; type?: string; config?: Record<string, unknown>; priority?: number; isEnabled?: boolean; isDefault?: boolean; isPrimary?: boolean; status?: string; description?: string; secretConfig?: Record<string, unknown> } }) => updateProvider(id, input),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: providersKeys.lists() });
      queryClient.invalidateQueries({ queryKey: providersKeys.detail(variables.id) });
      queryClient.invalidateQueries({ queryKey: ['providers', 'health'] });
      toast.success('Provider updated');
    },
    onError: (err: Error) => {
      toast.error('Failed to update provider', { description: err.message });
    },
  });
}

export function useSetDefaultProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, isDefault }: { id: string; isDefault: boolean }) => setDefaultProvider(id, isDefault),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: providersKeys.lists() });
      queryClient.invalidateQueries({ queryKey: providersKeys.detail(variables.id) });
      queryClient.invalidateQueries({ queryKey: ['providers', 'health'] });
      toast.success(variables.isDefault ? 'Set as default provider' : 'Removed default status');
    },
    onError: (err: Error) => {
      toast.error('Failed to update default provider', { description: err.message });
    },
  });
}

export function useDeleteProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: providersKeys.lists() });
      toast.success('Provider deleted');
    },
    onError: (err: Error) => {
      toast.error('Failed to delete provider', { description: err.message });
    },
  });
}

export function useToggleProviderStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, isEnabled }: { id: string; isEnabled: boolean }) => toggleProviderStatus(id, isEnabled),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: providersKeys.lists() });
      queryClient.invalidateQueries({ queryKey: providersKeys.detail(variables.id) });
      queryClient.invalidateQueries({ queryKey: ['providers', 'health'] });
      toast.success(variables.isEnabled ? 'Provider enabled' : 'Provider disabled');
    },
    onError: (err: Error) => {
      toast.error('Failed to toggle provider status', { description: err.message });
    },
  });
}

export function useTestProvider() {
  return useMutation({
    mutationFn: (id: string) => testProvider(id),
    onSuccess: () => {
      toast.success('Provider test passed');
    },
    onError: (err: Error) => {
      toast.error('Provider test failed', { description: err.message });
    },
  });
}
