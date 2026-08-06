'use client';

import { useTranslations } from 'next-intl';
import { ExternalLink, FileSearch, HelpCircle } from 'lucide-react';

// Kavenegar REST API reference shown when a Kavenegar send fails.
export const KAVENEGAR_DOCS_URL = 'https://kavenegar.com/rest.html';

// Kavenegar API error codes we can explain proactively (see kavenegar.com/rest.html).
const KAVENEGAR_KNOWN_CODES = new Set([401, 403, 406, 412, 413]);

export interface ParsedProviderError {
  provider: 'kavenegar';
  code: number;
}

/**
 * Parses a raw provider error string into a structured, explainable error.
 * Currently recognises Kavenegar APIError responses, e.g.
 *   "kavenegar send failed for +989...: APIError[412] : ..."
 */
export function parseProviderError(message?: string): ParsedProviderError | null {
  if (!message) return null;
  const match = message.match(/APIError\[(\d+)\]/);
  if (!match) return null;
  const code = Number(match[1]);
  if (!Number.isFinite(code)) return null;
  return { provider: 'kavenegar', code };
}

interface ProviderErrorHelpProps {
  /** Raw error message returned by the provider/backend. */
  message?: string;
  /** Provider type/name (e.g. "kavenegar") — used as a fallback gate. */
  provider?: string;
  /** When present, adds a link to the attempt logs for this notification. */
  notificationId?: string;
}

/**
 * Shows an actionable explanation for a known provider send failure, with a
 * link to the provider's REST documentation and (optionally) to the attempt
 * logs for the affected notification. Renders nothing for unknown errors so
 * generic error text keeps its existing presentation.
 */
export function ProviderErrorHelp({ message, provider, notificationId }: ProviderErrorHelpProps) {
  const t = useTranslations();
  if (!message) return null;

  const parsed = parseProviderError(message);
  // The `provider` prop is a user-defined display name (e.g. "Main SMS"), so
  // also match the message itself — every backend Kavenegar error string
  // starts with "kavenegar …".
  const isKavenegar =
    (provider ?? '').toLowerCase().includes('kavenegar') ||
    /kavenegar/i.test(message);
  if (!parsed || parsed.provider !== 'kavenegar' || !isKavenegar) return null;

  const { code } = parsed;
  if (!KAVENEGAR_KNOWN_CODES.has(code)) return null;

  const hintKey = `providerErrors.kavenegar.code_${code}`;
  const hint = t(hintKey, { fallback: '' });

  return (
    <div className="mt-2 rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-900/50 dark:bg-amber-950/20">
      <div className="flex items-start gap-2">
        <HelpCircle className="mt-0.5 h-4 w-4 shrink-0 text-amber-600 dark:text-amber-400" />
        <div className="min-w-0 space-y-1">
          <p className="text-xs font-medium text-amber-800 dark:text-amber-300">
            {t('providerErrors.title')} · Kavenegar APIError[{code}]
          </p>
          {hint && (
            <p className="text-xs leading-relaxed text-amber-700 dark:text-amber-400">{hint}</p>
          )}
          <div className="flex flex-wrap items-center gap-3 pt-0.5">
            <a
              href={KAVENEGAR_DOCS_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-xs font-medium text-amber-800 underline underline-offset-2 hover:text-amber-950 dark:text-amber-300 dark:hover:text-amber-100"
            >
              <ExternalLink className="h-3 w-3" />
              {t('providerErrors.docs_link')}
            </a>
            {notificationId && (
              <a
                href={`/provider-logs?notificationId=${notificationId}`}
                className="inline-flex items-center gap-1 text-xs font-medium text-amber-800 underline underline-offset-2 hover:text-amber-950 dark:text-amber-300 dark:hover:text-amber-100"
              >
                <FileSearch className="h-3 w-3" />
                {t('providerErrors.details_link')}
              </a>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
