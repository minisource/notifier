'use client';

import { useTranslations } from 'next-intl';
import { useState } from 'react';
import {
  PageHeader,
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  Badge,
  Button,
  ConfirmActionDialog,
  Skeleton,
  ErrorState,
} from '@minisource/ui';
import { AlertTriangle, PauseCircle, PlayCircle, History, Inbox } from 'lucide-react';
import { RoleGuard } from '@/shared/components/role-guard';
import { useDeliveryControlStatus, usePauseDeliveries, useResumeDeliveries, useDeliveryControlHistory, useHeldDeliveries } from '@/features/delivery-control/hooks/use-delivery-control';
import type { DeliveryControlStatus as DCStatus } from '@/features/notifier/api/notifier-types';

function stateBadgeVariant(state: DCStatus['state']): 'default' | 'destructive' | 'yellow' | 'green' {
  switch (state) {
    case 'active':
      return 'green';
    case 'paused':
    case 'pause_requested':
      return 'destructive';
    case 'resume_requested':
    case 'active_with_uncertain_attempts':
      return 'yellow';
    default:
      return 'default';
  }
}

function formatDateTime(value?: string): string {
  if (!value) return '—';
  try {
    return new Date(value).toLocaleString();
  } catch {
    return value;
  }
}

export default function DeliveryControlPage() {
  const t = useTranslations();
  const { data: status, isLoading, isError, error, refetch } = useDeliveryControlStatus();
  const { data: history } = useDeliveryControlHistory(50);
  const { data: held } = useHeldDeliveries(1, 20);
  const pauseMutation = usePauseDeliveries();
  const resumeMutation = useResumeDeliveries();

  const [pauseOpen, setPauseOpen] = useState(false);
  const [resumeOpen, setResumeOpen] = useState(false);
  const [reason, setReason] = useState('');
  const [mode, setMode] = useState<'immediate' | 'drain'>('immediate');
  const [resumeReason, setResumeReason] = useState('');

  const isPaused =
    status?.state === 'paused' ||
    status?.state === 'pause_requested' ||
    status?.state === 'resume_requested';

  if (isLoading) return <Skeleton className="h-64 w-full" />;

  if (isError || !status) {
    return (
      <RoleGuard>
        <PageHeader title={t('deliveryControl.title')} description={t('deliveryControl.subtitle')} />
        <ErrorState
          message={error instanceof Error ? error.message : t('common.error_occurred')}
          onRetry={() => refetch()}
        />
      </RoleGuard>
    );
  }

  const handlePause = async () => {
    if (pauseMutation.isPending || resumeMutation.isPending) return;
    try {
      await pauseMutation.mutateAsync({
        mode,
        reason,
        expiresAt: undefined,
        expectedVersion: status.version,
      });
      setPauseOpen(false);
      setReason('');
      setMode('immediate');
    } catch {
      // toast handled by hook
    }
  };

  const handleResume = async () => {
    if (pauseMutation.isPending || resumeMutation.isPending) return;
    try {
      await resumeMutation.mutateAsync({
        reason: resumeReason || undefined,
        expectedVersion: status.version,
      });
      setResumeOpen(false);
      setResumeReason('');
    } catch {
      // toast handled by hook
    }
  };

  const stateKey = `state_${status.state}`;
  const stateLabel = t(`deliveryControl.${stateKey}`) || status.state;
  const modeLabel = status.mode === 'drain'
    ? t('deliveryControl.mode_drain')
    : t('deliveryControl.mode_immediate');

  return (
    <RoleGuard>
      <div className="space-y-5">
        <PageHeader title={t('deliveryControl.title')} description={t('deliveryControl.subtitle')} />

        {/* Current status */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>{t('deliveryControl.status_title')}</CardTitle>
            <Badge variant={stateBadgeVariant(status.state)}>{stateLabel}</Badge>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <StatusItem label={t('deliveryControl.reason')} value={status.reason || '—'} />
              <StatusItem label={t('deliveryControl.actor')} value={status.pausedBy || status.resumedBy || '—'} />
              <StatusItem label={t('deliveryControl.mode_label')} value={modeLabel} />
              <StatusItem label={t('deliveryControl.version')} value={String(status.version)} />
              <StatusItem label={t('deliveryControl.paused_at')} value={formatDateTime(status.pausedAt)} />
              <StatusItem label={t('deliveryControl.resumed_at')} value={formatDateTime(status.resumedAt)} />
              <StatusItem label={t('deliveryControl.expires_at')} value={formatDateTime(status.expiresAt)} />
              <StatusItem label={t('deliveryControl.last_updated')} value={formatDateTime(status.lastUpdatedAt)} />
            </div>

            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <CountTile label={t('deliveryControl.held_count')} value={status.heldCount} tone="destructive" />
              <CountTile label={t('deliveryControl.retrying_held')} value={status.retryingHeld} />
              <CountTile label={t('deliveryControl.uncertain_count')} value={status.uncertainCount} tone="warning" />
              <CountTile label={t('deliveryControl.active_attempt_count')} value={status.activeAttemptCount} />
            </div>

            {isPaused && (
              <div className="flex items-start gap-2 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/20 dark:text-red-300">
                <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
                <span>{t('deliveryControl.in_flight_warning')}</span>
              </div>
            )}

            <div className="flex flex-wrap gap-3">
              {status.canPause && (
                <Button
                  variant="destructive"
                  onClick={() => setPauseOpen(true)}
                  disabled={pauseMutation.isPending || resumeMutation.isPending}
                >
                  <PauseCircle className="mr-2 h-4 w-4" />
                  {t('deliveryControl.pause_button')}
                </Button>
              )}
              {status.canResume && (
                <Button
                  onClick={() => setResumeOpen(true)}
                  disabled={pauseMutation.isPending || resumeMutation.isPending}
                >
                  <PlayCircle className="mr-2 h-4 w-4" />
                  {t('deliveryControl.resume_button')}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Held deliveries */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Inbox className="h-4 w-4" />
              {t('deliveryControl.held_title')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {!held || held.items.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('deliveryControl.no_held')}</p>
            ) : (
              <div className="divide-y rounded-md border">
                {held.items.slice(0, 10).map((item) => (
                  <div key={item.id} className="flex flex-col gap-1 p-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">
                        {item.subject || item.type} · {item.recipientPhone || item.recipientEmail || '—'}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {item.heldReason || '—'} · {formatDateTime(item.heldAt)}
                      </p>
                    </div>
                    <Badge variant="yellow" className="shrink-0">
                      {t('deliveryControl.held_count_label', { count: item.retryCount })}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
            {held && held.total > 10 && (
              <p className="mt-2 text-xs text-muted-foreground">
                {t('deliveryControl.held_count_label', { count: held.total })}
              </p>
            )}
          </CardContent>
        </Card>

        {/* Audit history */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <History className="h-4 w-4" />
              {t('deliveryControl.history_title')}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {!history || history.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('deliveryControl.no_history')}</p>
            ) : (
              <div className="space-y-2">
                {history.map((event) => (
                  <div key={event.id} className="flex flex-col gap-1 rounded-md border p-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex items-center gap-2">
                      <Badge variant={event.action.startsWith('pause') || event.action === 'pause_effective' ? 'destructive' : 'default'}>
                        {event.action}
                      </Badge>
                      <span className="text-sm text-muted-foreground">
                        {event.actor || '—'} · v{event.version}
                      </span>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {event.reason || '—'} · {formatDateTime(event.createdAt)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Pause confirmation */}
        <ConfirmActionDialog
          open={pauseOpen}
          onOpenChange={setPauseOpen}
          title={t('deliveryControl.pause_dialog_title')}
          tone="destructive"
          confirmIcon={<PauseCircle className="h-5 w-5" />}
          confirmLabel={t('deliveryControl.confirm_pause')}
          cancelLabel={t('common.cancel')}
          pending={pauseMutation.isPending}
          onConfirm={handlePause}
          description={
            <div className="mt-2 space-y-4 text-left">
              <div className="flex items-start gap-3 rounded-md border border-red-200 bg-red-50 p-3 dark:border-red-900/50 dark:bg-red-950/20">
                <AlertTriangle className="h-5 w-5 shrink-0 text-red-600 dark:text-red-400" />
                <p className="text-sm text-red-700 dark:text-red-300">
                  {t('deliveryControl.pause_dialog_description')}
                </p>
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium">{t('deliveryControl.mode_label')}</label>
                <div className="flex gap-3">
                  {(['immediate', 'drain'] as const).map((m) => (
                    <label key={m} className="flex items-center gap-2 text-sm">
                      <input
                        type="radio"
                        name="pauseMode"
                        value={m}
                        checked={mode === m}
                        onChange={() => setMode(m)}
                        className="h-4 w-4 accent-red-600"
                      />
                      {m === 'immediate' ? t('deliveryControl.mode_immediate') : t('deliveryControl.mode_drain')}
                    </label>
                  ))}
                </div>
                <label className="block text-sm font-medium">{t('deliveryControl.reason_label')}</label>
                <textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder={t('deliveryControl.reason_placeholder')}
                  rows={3}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                />
              </div>
            </div>
          }
        />

        {/* Resume confirmation */}
        <ConfirmActionDialog
          open={resumeOpen}
          onOpenChange={setResumeOpen}
          title={t('deliveryControl.resume_dialog_title')}
          confirmIcon={<PlayCircle className="h-5 w-5" />}
          confirmLabel={t('deliveryControl.confirm_resume')}
          cancelLabel={t('common.cancel')}
          pending={resumeMutation.isPending}
          onConfirm={handleResume}
          description={
            <div className="mt-2 space-y-3 text-left">
              <p className="text-sm text-muted-foreground">{t('deliveryControl.resume_dialog_description')}</p>
              <div className="rounded-md border p-3">
                <p className="text-sm font-medium">
                  {t('deliveryControl.held_count')}: <strong>{status.heldCount}</strong>
                </p>
                <p className="text-sm font-medium">
                  {t('deliveryControl.uncertain_count')}: <strong>{status.uncertainCount}</strong>
                </p>
              </div>
              <label className="block text-sm font-medium">{t('deliveryControl.reason_label')}</label>
              <input
                value={resumeReason}
                onChange={(e) => setResumeReason(e.target.value)}
                placeholder={t('deliveryControl.resume_reason_placeholder')}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
            </div>
          }
        />
      </div>
    </RoleGuard>
  );
}

function StatusItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 break-words text-sm font-medium">{value}</p>
    </div>
  );
}

function CountTile({ label, value, tone }: { label: string; value: number; tone?: 'destructive' | 'warning' }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={`mt-1 text-2xl font-bold ${tone === 'destructive' ? 'text-red-600 dark:text-red-400' : tone === 'warning' ? 'text-amber-600 dark:text-amber-400' : ''}`}>
        {value}
      </p>
    </div>
  );
}
