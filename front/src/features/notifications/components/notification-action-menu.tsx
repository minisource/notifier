'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { MoreHorizontal, Eye, Copy, RotateCcw, XCircle, CheckCheck, ExternalLink } from 'lucide-react';
import { toast } from 'sonner';
import { useRetryNotification, useCancelNotification, useMarkNotificationRead } from '../hooks/use-notifications';
import type { Notification, NotificationStatus } from '../types';

interface NotificationActionMenuProps {
  notification: Notification;
  showView?: boolean;
  onView?: () => void;
}

function canRetry(status: NotificationStatus): boolean {
  return status === 'failed' || status === 'dead';
}

function canCancel(status: NotificationStatus): boolean {
  return status === 'pending' || status === 'queued' || status === 'processing';
}

function canMarkRead(notification: Notification): boolean {
  return notification.type === 'in_app' && !notification.readAt;
}

export function NotificationActionMenu({ notification, showView = true, onView }: NotificationActionMenuProps) {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';
  const [confirmAction, setConfirmAction] = useState<'retry' | 'cancel' | null>(null);

  const retryMutation = useRetryNotification();
  const cancelMutation = useCancelNotification();
  const markReadMutation = useMarkNotificationRead();

  const handleCopyId = () => {
    navigator.clipboard.writeText(notification.id);
    toast.success(t('common.copied'));
  };

  const handleView = () => {
    if (onView) {
      onView();
    } else {
      router.push(`/${locale}/notifications/${notification.id}`);
    }
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <MoreHorizontal className="h-4 w-4" />
            <span className="sr-only">{t('common.actions')}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-48">
          <DropdownMenuLabel>{t('common.actions')}</DropdownMenuLabel>
          <DropdownMenuSeparator />

          {showView && (
            <DropdownMenuItem onClick={handleView}>
              <Eye className="ml-2 h-4 w-4" />
              {t('common.view_details')}
            </DropdownMenuItem>
          )}

          <DropdownMenuItem onClick={() => router.push(`/${locale}/notifications/${notification.id}`)}>
            <ExternalLink className="ml-2 h-4 w-4" />
            {t('notifications.actions.open_detail')}
          </DropdownMenuItem>

          <DropdownMenuItem onClick={handleCopyId}>
            <Copy className="ml-2 h-4 w-4" />
            {t('common.copy_id')}
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          {canRetry(notification.status) && (
            <DropdownMenuItem
              onClick={() => setConfirmAction('retry')}
              disabled={retryMutation.isPending}
            >
              <RotateCcw className="ml-2 h-4 w-4" />
              {t('notifications.actions.retry')}
            </DropdownMenuItem>
          )}

          {canCancel(notification.status) && (
            <DropdownMenuItem
              onClick={() => setConfirmAction('cancel')}
              disabled={cancelMutation.isPending}
            >
              <XCircle className="ml-2 h-4 w-4" />
              {t('notifications.actions.cancel')}
            </DropdownMenuItem>
          )}

          {canMarkRead(notification) && (
            <DropdownMenuItem
              onClick={() => markReadMutation.mutate(notification.id)}
              disabled={markReadMutation.isPending}
            >
              <CheckCheck className="ml-2 h-4 w-4" />
              {t('notifications.actions.mark_read')}
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      <ConfirmDialog
        open={confirmAction === 'retry'}
        onOpenChange={(open) => { if (!open) setConfirmAction(null); }}
        onConfirm={() => {
          retryMutation.mutate(notification.id);
          setConfirmAction(null);
        }}
        title={t('notifications.actions.confirm_retry_title')}
        description={t('notifications.actions.confirm_retry_desc')}
        confirmLabel={t('notifications.actions.retry')}
        cancelLabel={t('common.cancel')}
        destructive={notification.status === 'dead'}
      />

      <ConfirmDialog
        open={confirmAction === 'cancel'}
        onOpenChange={(open) => { if (!open) setConfirmAction(null); }}
        onConfirm={() => {
          cancelMutation.mutate(notification.id);
          setConfirmAction(null);
        }}
        title={t('notifications.actions.confirm_cancel_title')}
        description={t('notifications.actions.confirm_cancel_desc')}
        confirmLabel={t('notifications.actions.cancel')}
        cancelLabel={t('common.no')}
        destructive
      />
    </>
  );
}
