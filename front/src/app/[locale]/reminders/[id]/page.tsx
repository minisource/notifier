'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { PageHeader } from '@/components/shared/page-header';
import { PageContainer } from '@/components/shared/page-container';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ErrorState } from '@/components/shared/error-state';
import { PageSkeleton } from '@/components/shared/loading-state';
import { ConfirmDialog } from '@/components/shared/confirm-dialog';
import { useState } from 'react';
import { useReminder, useCancelReminder } from '@/features/reminders/hooks/use-reminders';
import { ArrowLeft, Calendar, CalendarX, User, Send, Clock } from 'lucide-react';
import { formatDateTime } from '@/lib/utils/date';
import { maskEmail, maskPhone } from '@/lib/utils/format';

export default function ReminderDetailPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'fa';
  const id = params?.id as string;
  const [showCancelDialog, setShowCancelDialog] = useState(false);

  const { data: reminder, isLoading, isError, error, refetch } = useReminder(id);
  const cancelMutation = useCancelReminder();

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title={t('reminders.title')}>
          <Button variant="ghost" onClick={() => router.push(`/${locale}/reminders`)} disabled>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <PageSkeleton context="reminders" layout="detail" />
      </PageContainer>
    );
  }

  if (isError || !reminder) {
    return (
      <PageContainer>
        <PageHeader title={t('reminders.title')}>
          <Button variant="ghost" onClick={() => router.push(`/${locale}/reminders`)}>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.not_found')}
          message={(error as Error)?.message || t('reminders.no_reminders')}
          onRetry={() => refetch()}
        />
      </PageContainer>
    );
  }

  const canCancel = reminder.status === 'scheduled';
  const recipientDisplay = reminder.recipientEmail
    ? maskEmail(reminder.recipientEmail)
    : reminder.recipientPhone
      ? maskPhone(reminder.recipientPhone)
      : reminder.userId;

  return (
    <PageContainer>
      <PageHeader
        title={`${t('reminders.title')} — ${formatDateTime(reminder.scheduledAt, locale)}`}
        subtitle={`${reminder.id.slice(0, 8)}... · ${t(`statuses.${reminder.status}`)}`}
      >
        <Button variant="ghost" size="sm" onClick={() => router.push(`/${locale}/reminders`)}>
          <ArrowLeft className="ml-1.5 h-4 w-4" />
          {t('common.back')}
        </Button>
      </PageHeader>

      <div className="space-y-6">
        {/* Summary */}
        <Card className="overflow-hidden">
          <CardContent className="p-5">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div className="space-y-3">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant={reminder.status === 'sent' ? 'default' : reminder.status === 'cancelled' ? 'secondary' : reminder.status === 'failed' ? 'destructive' : 'outline'}>
                    {t(`statuses.${reminder.status}`)}
                  </Badge>
                  <Badge variant="outline">{reminder.type}</Badge>
                </div>

                <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <User className="h-3.5 w-3.5" />
                    <span>{reminder.userId}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <Send className="h-3.5 w-3.5" />
                    <span>{recipientDisplay}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <Calendar className="h-3.5 w-3.5" />
                    <span>{formatDateTime(reminder.scheduledAt, locale)}</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-muted-foreground">
                    <Clock className="h-3.5 w-3.5" />
                    <span>{formatDateTime(reminder.createdAt, locale)}</span>
                  </div>
                </div>

                {reminder.templateKey && (
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <span>{t('templates.key')}:</span>
                    <code className="rounded bg-muted px-1.5 py-0.5 font-mono">{reminder.templateKey}</code>
                  </div>
                )}
              </div>

              {/* Actions */}
              {canCancel && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setShowCancelDialog(true)}
                  disabled={cancelMutation.isPending}
                >
                  <CalendarX className="ml-1.5 h-4 w-4" />
                  {t('reminders.cancel')}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <ConfirmDialog
        open={showCancelDialog}
        onOpenChange={setShowCancelDialog}
        onConfirm={() => {
          cancelMutation.mutate(reminder.id);
          setShowCancelDialog(false);
        }}
        title={t('common.confirm_action')}
        description={t('reminders.cancel')}
        confirmLabel={t('reminders.cancel')}
        cancelLabel={t('common.no')}
        destructive
      />
    </PageContainer>
  );
}
