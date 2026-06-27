import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { templatesKeys } from '../query-keys';
import { listTemplates, getTemplate, createTemplate, updateTemplate, deleteTemplate } from '../api';
import type { CreateTemplateInput, UpdateTemplateInput } from '../types';

export function useTemplates(params?: { type?: string; locale?: string }) {
  return useQuery({
    queryKey: templatesKeys.list(params as Record<string, unknown>),
    queryFn: () => listTemplates(params),
  });
}

export function useTemplate(id: string) {
  return useQuery({
    queryKey: templatesKeys.detail(id),
    queryFn: () => getTemplate(id),
    enabled: !!id,
  });
}

export function useCreateTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateTemplateInput) => createTemplate(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: templatesKeys.lists() });
    },
  });
}

export function useUpdateTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateTemplateInput }) => updateTemplate(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: templatesKeys.lists() });
    },
  });
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteTemplate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: templatesKeys.lists() });
    },
  });
}
