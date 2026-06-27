'use client';

import { useTranslations } from 'next-intl';
import { cn } from '@/lib/utils';
import { Shield, EyeOff } from 'lucide-react';

interface NotificationMetadataViewerProps {
  metadata?: Record<string, unknown>;
  className?: string;
}

const SENSITIVE_KEYS = ['secret', 'token', 'password', 'api_key', 'apiKey', 'secretKey', 'privateKey', 'auth'];

function isSensitiveKey(key: string): boolean {
  return SENSITIVE_KEYS.some(s => key.toLowerCase().includes(s.toLowerCase()));
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return '—';
  if (typeof value === 'string') {
    if (value.length > 200) return value.slice(0, 200) + '...';
    return value;
  }
  if (typeof value === 'object') {
    try {
      const str = JSON.stringify(value, null, 2);
      return str.length > 500 ? str.slice(0, 500) + '...' : str;
    } catch {
      return String(value);
    }
  }
  return String(value);
}

export function NotificationMetadataViewer({ metadata, className }: NotificationMetadataViewerProps) {
  const t = useTranslations();

  if (!metadata || Object.keys(metadata).length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-center">
        <Shield className="h-8 w-8 text-muted-foreground/50" />
        <p className="mt-2 text-sm text-muted-foreground">{t('notifications.metadata_no_data')}</p>
      </div>
    );
  }

  const entries = Object.entries(metadata);

  return (
    <div className={cn('space-y-1', className)}>
      <div className="mb-2 flex items-center gap-1.5 text-xs text-muted-foreground">
        <Shield className="h-3 w-3" />
        {t('notifications.metadata_sanitized')}
      </div>
      {entries.map(([key, value]) => {
        const isSensitive = isSensitiveKey(key);
        return (
          <div
            key={key}
            className="flex items-start gap-2 rounded-md border border-border/50 bg-muted/20 px-3 py-2 text-sm"
          >
            <code className="min-w-[100px] shrink-0 text-xs font-medium text-foreground">
              {key}
            </code>
            <span className="min-w-0 break-all font-mono text-xs text-muted-foreground">
              {isSensitive ? (
                <span className="flex items-center gap-1 text-muted-foreground/60">
                  <EyeOff className="h-3 w-3" />
                  {t('notifications.metadata_sensitive')}
                </span>
              ) : (
                formatValue(value)
              )}
            </span>
          </div>
        );
      })}
    </div>
  );
}
