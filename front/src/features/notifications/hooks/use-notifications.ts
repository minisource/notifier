import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { notificationsKeys } from '../query-keys';
import {
  listNotifications, getNotification, sendNotification,
  sendBatchNotifications, retryNotification, cancelNotification,
  markNotificationRead, markAllNotificationsRead,
  getNotificationDeliveries, getTemplatesForSelect,
} from '../api';
import type { ListNotificationsParams, SendNotificationInput, SendBatchNotificationInput } from '../types';
import { ApiError } from '@/shared/api/api-error';

export function useNotifications(params?: ListNotificationsParams) {
  return useQuery({
    queryKey: notificationsKeys.list(params as Record<string, unknown>),
    queryFn: () => listNotifications(params),
  });
}

export function useNotification(id: string) {
  return useQuery({
    queryKey: notificationsKeys.detail(id),
    queryFn: () => getNotification(id),
    enabled: !!id,
  });
}

export function useNotificationDeliveries(notificationId: string) {
  return useQuery({
    queryKey: [...notificationsKeys.detail(notificationId), 'deliveries'],
    queryFn: () => getNotificationDeliveries(notificationId),
    enabled: !!notificationId,
  });
}

export function useTemplatesForSelect() {
  return useQuery({
    queryKey: ['templates', 'select'],
    queryFn: () => getTemplatesForSelect(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useSendNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: SendNotificationInput) => sendNotification(input),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: notificationsKeys.lists() });
      toast.success('Notification sent', {
        description: `ID: ${result.id.slice(0, 8)}...`,
      });
    },
    onError: (error) => {
      if (error instanceof ApiError) {
        toast.error(error.message, {
          description: error.details ? JSON.stringify(error.details) : `Code: ${error.code}`,
        });
      } else {
        toast.error('Failed to send notification');
      }
    },
  });
}

export function useSendBatchNotifications() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: SendBatchNotificationInput) => sendBatchNotifications(input),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: notificationsKeys.lists() });
      toast.success('Batch sent', {
        description: `${result.length} notifications queued`,
      });
    },
    onError: () => {
      toast.error('Batch send failed');
    },
  });
}

export function useRetryNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => retryNotification(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: notificationsKeys.lists() });
      queryClient.invalidateQueries({ queryKey: notificationsKeys.detail(id) });
      toast.success('Retry initiated', {
        description: 'Notification has been requeued',
      });
    },
    onError: () => {
      toast.error('Retry failed');
    },
  });
}

export function useCancelNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => cancelNotification(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: notificationsKeys.lists() });
      queryClient.invalidateQueries({ queryKey: notificationsKeys.detail(id) });
      toast.success('Notification cancelled');
    },
    onError: () => {
      toast.error('Failed to cancel notification');
    },
  });
}

export function useMarkNotificationRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => markNotificationRead(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: notificationsKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: notificationsKeys.lists() });
      toast.success('Marked as read');
    },
    onError: () => {
      toast.error('Failed to mark as read');
    },
  });
}

export function useMarkAllNotificationsRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) => markAllNotificationsRead(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationsKeys.lists() });
      toast.success('All notifications marked as read');
    },
    onError: () => {
      toast.error('Failed to mark all as read');
    },
  });
}
