'use client';

import { useTranslations } from 'next-intl';
import { useParams, useRouter } from 'next/navigation';
import { useState } from 'react';
import { PageHeader, Button, Card, CardHeader, CardTitle, CardContent, Badge, ErrorState, Skeleton } from '@minisource/ui';
import { StatusBadge } from '@/components/shared/status-badge';
import { ChannelBadge } from '@/components/shared/channel-badge';
import { useAdminProviderAttempt, useAdminProviderAttemptEvents } from '@/features/notifier/api/notifier-queries';
import { ArrowLeft, Copy, Check, ExternalLink, ShieldAlert, ShieldCheck, Scissors, Play, Loader2, TerminalSquare } from 'lucide-react';
import { formatDateTime } from '@/lib/utils/date';
import { formatMilliseconds } from '@/lib/utils/format';
import { buildCurlCommand } from '@/lib/utils/curl';
import { toast } from 'sonner';
import { ProviderErrorHelp } from '@/features/providers/components/provider-error-help';
import type { ProviderAttemptDetails, ProviderAttemptEvent } from '@/features/notifier/api/notifier-types';

// ==================== Sanitized payload viewer ====================
// Renders only already-sanitized values. Never executes HTML. Offers copy that
// copies only the sanitized string.

function PayloadViewer({ value, empty }: { value?: string; empty: string }) {
  const t = useTranslations();
  const [copied, setCopied] = useState(false);

  if (!value) {
    return (
      <div className="rounded-lg border border-dashed p-3 text-xs text-muted-foreground">
        {empty}
        <span className="ml-1 text-amber-600 dark:text-amber-400">{t('providerLogs.redaction.not_captured')}</span>
      </div>
    );
  }

  // Truncated bodies (a core feature of this logging) may be invalid JSON —
  // never crash the viewer; fall back to plain text rendering.
  let display = value;
  const trimmed = value.trim();
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try {
      display = JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      display = value;
    }
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      toast.success(t('providerLogs.copy.copied'));
      setTimeout(() => setCopied(false), 1500);
    } catch {
      toast.error(t('errors.generic'));
    }
  };

  return (
    <div className="relative">
      <Button variant="ghost" size="icon" className="absolute right-2 top-2 h-7 w-7" onClick={handleCopy} title={t('providerLogs.copy.copy_sanitized_hint') as string}>
        {copied ? <Check className="h-3.5 w-3.5 text-green-600" /> : <Copy className="h-3.5 w-3.5" />}
      </Button>
      <pre className="max-h-96 overflow-auto rounded-lg bg-muted/40 p-3 font-mono text-xs whitespace-pre-wrap break-all">
        {display}
      </pre>
    </div>
  );
}

// ==================== cURL sample builder ====================
// Renders a copyable curl command built from the SANITIZED request data
// (method, URL, headers, body). Secrets are already [REDACTED] server-side,
// so copying this can never leak credentials. A "Run test request" button
// fires the same request from the browser so the operator can validate the
// endpoint + payload inline; cross-origin providers may block it (CORS),
// in which case the copied curl run in a terminal is the full-fidelity path.

function CurlBlock({ attempt }: { attempt: ProviderAttemptDetails }) {
  const t = useTranslations();
  const [copied, setCopied] = useState(false);
  const [running, setRunning] = useState(false);
  const [runResult, setRunResult] = useState<{ ok: boolean; text: string } | null>(null);

  const curl = buildCurlCommand({
    method: attempt.requestMethod,
    url: attempt.requestUrlSanitized,
    headers: attempt.requestHeadersSanitized,
    body: attempt.requestBodySanitized,
    baseUrl: typeof window !== 'undefined' ? window.location.origin : undefined,
  });

  if (!curl) {
    // No HTTP request was captured (e.g. SMTP/push pseudo-URLs) — nothing to
    // reproduce with curl.
    return null;
  }

  // The sanitized request may still carry [REDACTED] placeholders (API key in
  // the URL path, OTP tokens in the body). Running it inline would always fail,
  // so disable the Run button until the operator edits a copy in their terminal.
  const hasPlaceholders = (attempt.requestUrlSanitized || '').includes('[REDACTED]');

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(curl);
      setCopied(true);
      toast.success(t('providerLogs.curl.copied'));
      setTimeout(() => setCopied(false), 1500);
    } catch {
      toast.error(t('errors.generic'));
    }
  };

  const handleRun = async () => {
    setRunning(true);
    setRunResult(null);
    try {
      const method = (attempt.requestMethod || 'POST').toUpperCase();
      const headers: Record<string, string> = { ...(attempt.requestHeadersSanitized || {}) };
      const body = attempt.requestBodySanitized || undefined;
      if (body && !Object.keys(headers).some((k) => k.toLowerCase() === 'content-type')) {
        const trimmed = body.trim();
        headers['Content-Type'] = trimmed.startsWith('{') || trimmed.startsWith('[')
          ? 'application/json'
          : 'application/x-www-form-urlencoded';
      }
      const res = await fetch(attempt.requestUrlSanitized!, {
        method,
        headers,
        body: method === 'GET' || method === 'HEAD' ? undefined : body,
      });
      const text = await res.text();
      setRunResult({ ok: res.ok, text: text.slice(0, 4000) });
    } catch {
      setRunResult({ ok: false, text: t('providerLogs.curl.cors_hint') });
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
          <TerminalSquare className="h-3.5 w-3.5" />
          {t('providerLogs.curl.title')}
        </span>
        <div className="ltr:ml-auto rtl:mr-auto flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className="h-7 text-[11px] gap-1"
            onClick={handleRun}
            disabled={running || hasPlaceholders}
            title={hasPlaceholders ? (t('providerLogs.curl.redacted_hint') as string) : undefined}
          >
            {running ? <Loader2 className="h-3 w-3 animate-spin" /> : <Play className="h-3 w-3" />}
            {running ? t('providerLogs.curl.running') : t('providerLogs.curl.run')}
          </Button>
          <Button variant="outline" size="sm" className="h-7 text-[11px] gap-1" onClick={handleCopy} title={t('providerLogs.copy.copy_sanitized_hint') as string}>
            {copied ? <Check className="h-3 w-3 text-green-600" /> : <Copy className="h-3 w-3" />}
            {t('providerLogs.curl.copy')}
          </Button>
        </div>
      </div>

      <pre className="max-h-64 overflow-auto rounded-lg bg-muted/40 p-3 font-mono text-xs whitespace-pre-wrap break-all">
        {curl}
      </pre>

      <p className="flex items-start gap-1.5 text-[11px] text-amber-600 dark:text-amber-400">
        <ShieldAlert className="mt-0.5 h-3 w-3 shrink-0" />
        {t('providerLogs.curl.redacted_hint')}
      </p>

      {runResult && (
        <div className={`rounded-lg border p-2.5 text-xs ${runResult.ok ? 'border-emerald-500/30 bg-emerald-500/5' : 'border-red-500/30 bg-red-500/5'}`}>
          <span className={`font-medium ${runResult.ok ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'}`}>
            {t('providerLogs.curl.result')}
          </span>
          <pre className="mt-1 max-h-40 overflow-auto font-mono text-[11px] whitespace-pre-wrap break-all text-muted-foreground">
            {runResult.text}
          </pre>
        </div>
      )}
    </div>
  );
}

// ==================== Lifecycle timeline ====================

function Timeline({ events, locale }: { events?: ProviderAttemptEvent[]; locale: string }) {
  const t = useTranslations();
  if (!events || events.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">{t('providerLogs.no_attempts')}</p>
    );
  }

  return (
    <ol className="relative space-y-4 border-l-2 border-border pl-5">
      {events.map((e) => (
        <li key={e.id} className="relative">
          <span className="absolute -left-[26px] top-1 h-3 w-3 rounded-full border-2 border-background bg-primary" />
          <div className="flex flex-col gap-0.5">
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-sm font-medium">{t(`providerLogs.timeline.${e.eventType}`, { fallback: e.eventType })}</span>
              {e.newStatus && <StatusBadge status={e.newStatus} size="sm" />}
            </div>
            <span className="text-xs text-muted-foreground">{formatDateTime(e.occurredAt, locale)}</span>
            {e.eventPayloadSanitized && Object.keys(e.eventPayloadSanitized).length > 0 && (
              <pre className="mt-1 max-h-32 overflow-auto rounded bg-muted/40 p-2 font-mono text-[11px] whitespace-pre-wrap break-all">
                {JSON.stringify(e.eventPayloadSanitized, null, 2)}
              </pre>
            )}
          </div>
        </li>
      ))}
    </ol>
  );
}

// ==================== Detail fields ====================

function Field({ label, value, mono }: { label: string; value?: string; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-muted-foreground">{label}</span>
      {value ? (
        <span className={`text-sm ${mono ? 'font-mono text-xs break-all' : ''}`}>{value}</span>
      ) : (
        <span className="text-sm text-muted-foreground">—</span>
      )}
    </div>
  );
}

// ==================== Page ====================

export default function ProviderAttemptDetailPage() {
  const t = useTranslations();
  const router = useRouter();
  const params = useParams();
  const locale = (params?.locale as string) || 'en';
  const id = params?.attemptId as string;

  const { data: attempt, isLoading, isError, error, refetch } = useAdminProviderAttempt(id);
  const { data: events } = useAdminProviderAttemptEvents(id);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('providerLogs.detail_title')}>
          <Button variant="ghost" onClick={() => router.push('/provider-logs')} disabled>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (isError || !attempt) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('providerLogs.detail_title')}>
          <Button variant="ghost" onClick={() => router.push('/provider-logs')}>
            <ArrowLeft className="ml-2 h-4 w-4" />
            {t('common.back')}
          </Button>
        </PageHeader>
        <ErrorState
          title={t('errors.not_found')}
          message={(error as Error)?.message || t('providerLogs.no_attempts')}
          onRetry={() => refetch()}
        />
      </div>
    );
  }

  const a: ProviderAttemptDetails = attempt;
  const showTruncation = a.bodyTruncated || (a.originalSizeBytes ?? 0) > (a.capturedSizeBytes ?? 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('providerLogs.detail_title')}
        description={`${a.provider} · ${a.channel} · attempt #${a.attemptNumber}${a.fallbackSequence > 0 ? ` +${a.fallbackSequence}` : ''}`}
      >
        <Button variant="ghost" size="sm" onClick={() => router.push('/provider-logs')}>
          <ArrowLeft className="ml-1.5 h-4 w-4" />
          {t('providerLogs.actions.back_to_logs')}
        </Button>
        <Button variant="outline" size="sm" onClick={() => router.push(`/notifications/${a.notificationId}`)}>
          <ExternalLink className="ml-1.5 h-4 w-4" />
          {t('providerLogs.actions.view_notification')}
        </Button>
      </PageHeader>

      <div className="space-y-6">
        {/* Overview */}
        <Card className="overflow-hidden">
          <CardContent className="p-5">
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status={a.status} size="sm" />
              <Badge variant="outline">{a.provider}</Badge>
              <ChannelBadge channel={a.channel} size="sm" />
              {a.retryable && <Badge variant="secondary">{t('providerLogs.retryable')}</Badge>}
            </div>
            <div className="mt-4 grid grid-cols-1 gap-x-6 gap-y-3 text-sm sm:grid-cols-2 lg:grid-cols-3">
              <Field label={t('providerLogs.notification_id')} value={a.notificationId} mono />
              <Field label={t('providerLogs.attempt')} value={`#${a.attemptNumber}`} />
              <Field label={t('providerLogs.recipient')} value={a.recipientMasked} mono />
              <Field label={t('providerLogs.provider_status')} value={a.providerStatus} />
              <Field label={t('providerLogs.provider_message_id')} value={a.providerMessageId} mono />
              <Field label={t('providerLogs.duration')} value={a.durationMs ? formatMilliseconds(a.durationMs) : undefined} />
              <Field label={t('providerLogs.http_status')} value={a.responseStatusCode ? String(a.responseStatusCode) : undefined} mono />
              <Field label={t('providerLogs.queued_at')} value={formatDateTime(a.queuedAt, locale)} />
              <Field label={t('providerLogs.completed_at')} value={a.completedAt ? formatDateTime(a.completedAt, locale) : undefined} />
            </div>

            {a.bodyPreview && (
              <div className="mt-4 rounded-lg bg-muted/40 p-3">
                <div className="flex items-center gap-1.5 text-sm font-medium">
                  <ShieldCheck className="h-4 w-4" />
                  {t('providerLogs.redaction.masked')} — {t('notifications.body')}
                </div>
                <p className="mt-1 whitespace-pre-wrap text-sm text-muted-foreground">{a.bodyPreview}</p>
              </div>
            )}

            {a.normalizedErrorMessage && (
              <div className="mt-4 rounded-lg bg-red-50 p-3 dark:bg-red-950/20">
                <div className="flex items-center gap-1.5 text-sm font-medium text-red-700 dark:text-red-400">
                  <ShieldAlert className="h-4 w-4" />
                  {t('providerLogs.error_message')}
                </div>
                <p className="mt-1 break-words text-sm text-red-600 dark:text-red-300">{a.normalizedErrorMessage}</p>
                <ProviderErrorHelp message={a.normalizedErrorMessage} provider={a.provider} notificationId={a.notificationId} />
                <div className="mt-2 flex flex-wrap gap-2 text-xs">
                  {a.normalizedErrorKind && <Badge variant="destructive">{t(`providerLogs.error_kinds.${a.normalizedErrorKind}`, { fallback: a.normalizedErrorKind })}</Badge>}
                  {a.normalizedErrorCode && <Badge variant="outline">{a.normalizedErrorCode}</Badge>}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Timeline */}
        <Card>
          <CardHeader><CardTitle>{t('providerLogs.sections.timeline')}</CardTitle></CardHeader>
          <CardContent>
            <Timeline events={a.events?.length ? a.events : events} locale={locale} />
          </CardContent>
        </Card>

        {/* Request */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              {t('providerLogs.sections.request')}
              {showTruncation && (
                <Badge variant="secondary" className="gap-1">
                  <Scissors className="h-3 w-3" /> {t('providerLogs.redaction.truncated')}
                </Badge>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {a.requestMethod && (
              <div className="flex flex-wrap gap-2 text-xs">
                <Badge variant="outline">{a.requestMethod}</Badge>
                {a.requestUrlSanitized && <code className="text-muted-foreground break-all">{a.requestUrlSanitized}</code>}
                {a.requestSizeBytes ? <span className="text-muted-foreground">{a.requestSizeBytes} B</span> : null}
              </div>
            )}
            <CurlBlock attempt={a} />
            {a.requestHeadersSanitized && Object.keys(a.requestHeadersSanitized).length > 0 && (
              <div>
                <span className="text-xs text-muted-foreground">Headers</span>
                <PayloadViewer value={JSON.stringify(a.requestHeadersSanitized)} empty={t('providerLogs.redaction.not_captured')} />
              </div>
            )}
            <div>
              <span className="text-xs text-muted-foreground">Body</span>
              <PayloadViewer value={a.requestBodySanitized} empty={t('providerLogs.redaction.not_captured')} />
            </div>
            {showTruncation && (
              <p className="text-xs text-muted-foreground">
                {t('providerLogs.redaction.truncated_hint', { size: formatBytes(a.originalSizeBytes ?? 0) })}
              </p>
            )}
          </CardContent>
        </Card>

        {/* Response */}
        <Card>
          <CardHeader><CardTitle>{t('providerLogs.sections.response')}</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <div>
              <span className="text-xs text-muted-foreground">Body</span>
              <PayloadViewer value={a.responseBodySanitized} empty={t('providerLogs.redaction.not_captured')} />
            </div>
            {a.responseHeadersSanitized && Object.keys(a.responseHeadersSanitized).length > 0 && (
              <div>
                <span className="text-xs text-muted-foreground">Headers</span>
                <PayloadViewer value={JSON.stringify(a.responseHeadersSanitized)} empty={t('providerLogs.redaction.not_captured')} />
              </div>
            )}
          </CardContent>
        </Card>

        {/* Correlation */}
        <Card>
          <CardHeader><CardTitle>{t('providerLogs.sections.correlation')}</CardTitle></CardHeader>
          <CardContent className="grid grid-cols-1 gap-x-6 gap-y-3 sm:grid-cols-2">
            <Field label={t('providerLogs.request_id')} value={a.requestId} mono />
            <Field label={t('providerLogs.correlation_id')} value={a.correlationId} mono />
            <Field label={t('providerLogs.trace_id')} value={a.traceId || t('providerLogs.redaction.not_captured')} mono />
            <Field label={t('providerLogs.attempt_number')} value={a.parentAttemptId ? a.parentAttemptId : undefined} mono />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}
