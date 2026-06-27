'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import {
  Dialog,  DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { AlertTriangle, CheckCircle, Loader2, Server } from 'lucide-react';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { toast } from 'sonner';
import { testProvider } from '../api';

interface Provider {
  id: string;
  name: string;
  channel: string;
  status: string;
}

interface ProviderTestDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  provider: Provider;
}

export function ProviderTestDialog({ open, onOpenChange, provider }: ProviderTestDialogProps) {
  const t = useTranslations();
  const [recipient, setRecipient] = useState('');
  const [body, setBody] = useState('');
  const [dryRun, setDryRun] = useState(true);
  const [testing, setTesting] = useState(false);
  const [result, setResult] = useState<{ success: boolean; message?: string; error?: string } | null>(null);

  const handleTest = async () => {
    if (!recipient.trim() && !dryRun) {
      toast.error(t('forms.required'));
      return;
    }

    setTesting(true);
    setResult(null);

    try {
      const response = await testProvider(provider.id, {
        recipient: recipient || undefined,
        body: body || undefined,
      });

      setResult({
        success: response.success,
        message: response.success
          ? `${t('providers.healthy')} — ${provider.name} ${response.message || 'responded successfully'}`
          : undefined,
        error: !response.success ? (response.message || 'Provider test failed') : undefined,
      });

      if (response.success) {
        toast.success(`${provider.name} ${t('providers.healthy').toLowerCase()}`);
      } else {
        toast.error(`${provider.name} ${t('providers.degraded').toLowerCase()}`);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Provider test failed';
      setResult({ success: false, error: message });
      toast.error(`${provider.name} ${t('providers.degraded').toLowerCase()}`, {
        description: message,
      });
    } finally {
      setTesting(false);
    }
  };

  const handleClose = () => {
    setResult(null);
    setRecipient('');
    setBody('');
    setDryRun(true);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose(); else onOpenChange(o); }}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Server className="h-4 w-4" />
            {t('providers.test')}: {provider.name}
          </DialogTitle>
          <DialogDescription>
            <div className="flex items-center gap-2 mt-1">
              <ChannelBadge channel={provider.channel} size="sm" />
              <Badge variant="outline" className="text-xs">{provider.status}</Badge>
            </div>
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          {/* Dry Run Warning */}
          <div className="flex items-center gap-2 rounded-md border border-amber-200 bg-amber-50 p-3 dark:border-amber-900/50 dark:bg-amber-950/20">
            <AlertTriangle className="h-4 w-4 shrink-0 text-amber-600 dark:text-amber-400" />
            <p className="text-xs text-amber-700 dark:text-amber-300">
              {t('providers.dry_run_warning') || 'Dry-run mode enabled. No real message will be sent.'}
            </p>
          </div>

          {/* Recipient */}
          <div className="space-y-2">
            <Label>{t('notifications.recipient')}</Label>
            <Input
              value={recipient}
              onChange={(e) => setRecipient(e.target.value)}
              placeholder={provider.channel === 'sms' ? '+989121234567' : 'user@example.com'}
            />
          </div>

          {/* Body */}
          <div className="space-y-2">
            <Label>{t('notifications.body')}</Label>
            <Textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Test message content..."
              className="min-h-[80px]"
            />
          </div>

          {/* Dry Run Toggle */}
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div>
              <p className="text-sm font-medium">{t('providers.dry_run') || 'Dry Run'}</p>
              <p className="text-xs text-muted-foreground">
                {dryRun
                  ? (t('providers.dry_run_enabled') || 'Simulate without sending')
                  : (t('providers.dry_run_disabled') || 'Send real test message')}
              </p>
            </div>
            <Switch checked={dryRun} onCheckedChange={setDryRun} />
          </div>

          {/* Result */}
          {result && (
            <div className={`rounded-lg border p-4 ${result.success ? 'border-green-200 bg-green-50 dark:border-green-900/50 dark:bg-green-950/20' : 'border-red-200 bg-red-50 dark:border-red-900/50 dark:bg-red-950/20'}`}>
              <div className="flex items-center gap-2">
                {result.success ? (
                  <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400" />
                ) : (
                  <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400" />
                )}
                <span className={`text-sm font-medium ${result.success ? 'text-green-700 dark:text-green-300' : 'text-red-700 dark:text-red-300'}`}>
                  {result.success ? t('providers.healthy') : t('providers.degraded')}
                </span>
              </div>
              {result.message && <p className="mt-1 text-sm text-green-600 dark:text-green-400">{result.message}</p>}
              {result.error && <p className="mt-1 text-sm text-red-600 dark:text-red-300">{result.error}</p>}
            </div>
          )}
        </div>

        <div className="flex items-center justify-end gap-3 pt-4 border-t">
          <Button variant="outline" onClick={handleClose}>
            {t('common.close')}
          </Button>
          <Button onClick={handleTest} disabled={testing} variant={dryRun ? 'outline' : 'default'}>
            {testing ? <><Loader2 className="ml-1.5 h-4 w-4 animate-spin" /> Testing...</> : t('providers.test')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
