'use client';

import { useTranslations } from 'next-intl';
import { useParams } from 'next/navigation';
import { Badge } from '@minisource/ui';
import { maskEmail, maskPhone, shortId } from '@/lib/utils/format';
import { formatDateTime } from '@/lib/utils/date';
import {
  User, Mail, Phone, Hash, Globe, Clock, Send, CheckCheck,
  Eye, MousePointerClick,   AlertTriangle, RotateCcw,
  FileText, Zap, Key, Fingerprint,
} from 'lucide-react';
import type { Notification } from '../types';
import { ProviderErrorHelp } from '@/features/providers/components/provider-error-help';

interface NotificationDetailRowProps {
  notification: Notification;
}

export function NotificationDetailRow({ notification }: NotificationDetailRowProps) {
  const t = useTranslations();
  const params = useParams();
  const locale = (params?.locale as string) || 'en';

  return (
    <div className="border-t bg-muted/30 px-4 py-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {/* پیام */}
        <div className="space-y-2">
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            {t('notifications.payload')}
          </h4>
          {notification.subject && (
            <div className="flex items-start gap-2 text-sm">
              <FileText className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <div>
                <span className="text-muted-foreground text-xs">{t('notifications.subject')}: </span>
                <span className="font-medium">{notification.subject}</span>
              </div>
            </div>
          )}
          <div className="flex items-start gap-2 text-sm">
            <FileText className="mt-0.5 h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <div className="min-w-0">
              <span className="text-muted-foreground text-xs">{t('notifications.body')}: </span>
              <p className="mt-0.5 whitespace-pre-wrap break-words text-sm">{notification.body}</p>
            </div>
          </div>
        </div>

        {/* گیرنده و قالب */}
        <div className="space-y-2">
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            {t('notifications.recipient')}
          </h4>
          <div className="space-y-1.5 text-sm">
            <div className="flex items-center gap-2">
              <User className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-muted-foreground text-xs">{t('notifications.user_id')}: </span>
              <code className="font-mono text-xs">{shortId(notification.userId)}</code>
            </div>
            {notification.recipientEmail && (
              <div className="flex items-center gap-2">
                <Mail className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('channels.email')}: </span>
                <span className="font-mono text-xs">{maskEmail(notification.recipientEmail)}</span>
              </div>
            )}
            {notification.recipientPhone && (
              <div className="flex items-center gap-2">
                <Phone className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('channels.sms')}: </span>
                <span className="font-mono text-xs">{maskPhone(notification.recipientPhone)}</span>
              </div>
            )}
            {notification.recipientId && (
              <div className="flex items-center gap-2">
                <Fingerprint className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.recipient_id')}: </span>
                <code className="font-mono text-xs">{shortId(notification.recipientId)}</code>
              </div>
            )}
            <div className="flex items-center gap-2">
              <Globe className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-muted-foreground text-xs">{t('notifications.locale')}: </span>
              <Badge variant="outline" className="text-xs">
                {notification.locale === 'fa' ? 'فارسی' : 'English'}
              </Badge>
            </div>
          </div>

          {/* قالب و ارائه‌دهنده */}
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider pt-2">
            {t('notifications.template')}
          </h4>
          <div className="space-y-1.5 text-sm">
            {notification.templateKey && (
              <div className="flex items-center gap-2">
                <FileText className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.template_key')}: </span>
                <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{notification.templateKey}</code>
              </div>
            )}
            {notification.provider && (
              <div className="flex items-center gap-2">
                <Zap className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('deliveries.provider')}: </span>
                <Badge variant="secondary" className="text-xs">{notification.provider}</Badge>
              </div>
            )}
            {notification.providerMsgId && (
              <div className="flex items-center gap-2">
                <Hash className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.provider_msg_id')}: </span>
                <code className="font-mono text-[10px] truncate max-w-[150px]">{notification.providerMsgId}</code>
              </div>
            )}
          </div>
        </div>

        {/* زمان‌ها و وضعیت */}
        <div className="space-y-2">
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            {t('notifications.timeline')}
          </h4>
          <div className="space-y-1.5 text-sm">
            <div className="flex items-center gap-2">
              <Clock className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-muted-foreground text-xs">{t('common.created_at')}: </span>
              <span className="text-xs">{formatDateTime(notification.createdAt, locale)}</span>
            </div>
            {notification.scheduledAt && (
              <div className="flex items-center gap-2">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.scheduled_at')}: </span>
                <span className="text-xs">{formatDateTime(notification.scheduledAt, locale)}</span>
              </div>
            )}
            {notification.sentAt && (
              <div className="flex items-center gap-2">
                <Send className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.sent_at')}: </span>
                <span className="text-xs">{formatDateTime(notification.sentAt, locale)}</span>
              </div>
            )}
            {notification.deliveredAt && (
              <div className="flex items-center gap-2">
                <CheckCheck className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.delivered_at')}: </span>
                <span className="text-xs">{formatDateTime(notification.deliveredAt, locale)}</span>
              </div>
            )}
            {notification.seenAt && (
              <div className="flex items-center gap-2">
                <Eye className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.seen_at')}: </span>
                <span className="text-xs">{formatDateTime(notification.seenAt, locale)}</span>
              </div>
            )}
            {notification.readAt && (
              <div className="flex items-center gap-2">
                <CheckCheck className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.read_at')}: </span>
                <span className="text-xs">{formatDateTime(notification.readAt, locale)}</span>
              </div>
            )}
            {notification.clickedAt && (
              <div className="flex items-center gap-2">
                <MousePointerClick className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.clicked_at')}: </span>
                <span className="text-xs">{formatDateTime(notification.clickedAt, locale)}</span>
              </div>
            )}
            {notification.failedAt && (
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-3.5 w-3.5 text-red-500" />
                <span className="text-muted-foreground text-xs">{t('notifications.failed_at')}: </span>
                <span className="text-xs text-red-600">{formatDateTime(notification.failedAt, locale)}</span>
              </div>
            )}
            {notification.updatedAt && notification.updatedAt !== notification.createdAt && (
              <div className="flex items-center gap-2">
                <RotateCcw className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('common.updated_at')}: </span>
                <span className="text-xs">{formatDateTime(notification.updatedAt, locale)}</span>
              </div>
            )}
          </div>

          {/* اطلاعات تکمیلی */}
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider pt-2">
            {t('notifications.additional_info')}
          </h4>
          <div className="space-y-1.5 text-sm">
            {notification.retryCount > 0 && (
              <div className="flex items-center gap-2">
                <RotateCcw className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.retry_count')}: </span>
                <span className="text-xs">{notification.retryCount}/{notification.maxRetries}</span>
              </div>
            )}
            {notification.idempotencyKey && (
              <div className="flex items-center gap-2">
                <Key className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.idempotency_key')}: </span>
                <code className="font-mono text-[10px] truncate max-w-[150px]">{notification.idempotencyKey}</code>
              </div>
            )}
            {notification.variables && Object.keys(notification.variables).length > 0 && (
              <div className="flex items-start gap-2">
                <Hash className="mt-0.5 h-3.5 w-3.5 text-muted-foreground" />
                <div>
                  <span className="text-muted-foreground text-xs">{t('notifications.variables')}: </span>
                  <div className="mt-1 flex flex-wrap gap-1">
                    {Object.entries(notification.variables).map(([key, value]) => (
                      <Badge key={key} variant="outline" className="text-[10px] font-mono">
                        {key}={value}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>
            )}
            {notification.errorMessage && (
              <div className="flex items-start gap-2 rounded-md bg-red-50 p-2 dark:bg-red-950/20">
                <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-red-500" />
                <div className="min-w-0">
                  <span className="text-xs font-medium text-red-700 dark:text-red-400">{t('notifications.error_message')}: </span>
                  <p className="mt-0.5 text-xs text-red-600 dark:text-red-300">{notification.errorMessage}</p>
                  <ProviderErrorHelp message={notification.errorMessage} provider={notification.provider} notificationId={notification.id} />
                </div>
              </div>
            )}
            {notification.errorCode && (
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground text-xs">{t('notifications.error_code')}: </span>
                <Badge variant="destructive" className="text-xs font-mono">{notification.errorCode}</Badge>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
