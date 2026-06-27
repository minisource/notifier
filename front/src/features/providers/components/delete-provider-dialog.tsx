'use client';

import { useTranslations } from 'next-intl';
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { AlertTriangle, Loader2, Trash2 } from 'lucide-react';
import { useDeleteProvider } from '@/features/providers/hooks/use-providers';
import { toast } from 'sonner';

interface DeleteProviderDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  providerId: string;
  providerName: string;
  channel: string;
}

export function DeleteProviderDialog({ open, onOpenChange, providerId, providerName, channel }: DeleteProviderDialogProps) {
  const t = useTranslations();
  const deleteMutation = useDeleteProvider();

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(providerId);
      toast.success(`${providerName} ${t('providers.deleted') || 'deleted'}`);
      onOpenChange(false);
    } catch (err: any) {
      toast.error(err?.message || 'Failed to delete provider');
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-red-600 dark:text-red-400">
            <Trash2 className="h-5 w-5" />
            {t('providers.delete_title') || 'Delete Provider'}
          </DialogTitle>
          <DialogDescription>
            <div className="mt-2 space-y-4">
              <div className="flex items-start gap-3 rounded-md border border-red-200 bg-red-50 p-3 dark:border-red-900/50 dark:bg-red-950/20">
                <AlertTriangle className="h-5 w-5 shrink-0 text-red-600 dark:text-red-400" />
                <div className="text-sm text-red-700 dark:text-red-300">
                  <p className="font-medium">{t('providers.delete_warning') || 'This action cannot be undone.'}</p>
                  <p className="mt-1">
                    {t('providers.delete_warning_detail') || 'Deleting this provider may affect notification delivery for the'} <strong>{channel}</strong> {t('providers.channel') || 'channel'}.
                  </p>
                </div>
              </div>

              <div className="rounded-lg border p-3">
                <p className="text-sm font-medium">{t('providers.provider') || 'Provider'}: {providerName}</p>
                <p className="text-xs text-muted-foreground mt-1">{t('providers.channel') || 'Channel'}: {channel}</p>
              </div>
            </div>
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center justify-end gap-3 pt-4 border-t">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={deleteMutation.isPending}>
            {t('common.cancel')}
          </Button>
          <Button variant="destructive" onClick={handleDelete} disabled={deleteMutation.isPending}>
            {deleteMutation.isPending ? (
              <><Loader2 className="ml-1.5 h-4 w-4 animate-spin" /> {t('common.deleting')}</>
            ) : (
              <><Trash2 className="ml-1.5 h-4 w-4" /> {t('common.delete')}</>
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
