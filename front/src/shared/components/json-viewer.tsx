'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Eye, EyeOff, Copy, Check } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface JsonViewerProps {
  data: Record<string, unknown>;
  className?: string;
  sensitiveKeys?: string[];
  compact?: boolean;
}

const DEFAULT_SENSITIVE_KEYS = [
  'password', 'token', 'secret', 'apiKey', 'apikey', 'authorization',
  'auth', 'otp', 'code', 'refreshToken', 'accessToken', 'privateKey',
  'webhookSecret', 'jwt', 'session',
];

export function JsonViewer({ data, className, sensitiveKeys = DEFAULT_SENSITIVE_KEYS, compact }: JsonViewerProps) {
  const t = useTranslations();
  const [showSecrets, setShowSecrets] = useState(false);
  const [copied, setCopied] = useState(false);

  const isSensitiveKey = (key: string) =>
    sensitiveKeys.some(sk => key.toLowerCase().includes(sk.toLowerCase()));

  const sanitizeValue = (key: string, value: unknown): unknown => {
    if (showSecrets) return value;
    if (isSensitiveKey(key)) return '••••••••';
    return value;
  };

  const sanitizedData: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(data)) {
    sanitizedData[key] = sanitizeValue(key, value);
  }

  const formatted = JSON.stringify(sanitizedData, null, compact ? undefined : 2);
  const hasSensitive = Object.keys(data).some(k => isSensitiveKey(k));

  const handleCopy = () => {
    navigator.clipboard.writeText(JSON.stringify(data, null, 2));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between">
        {hasSensitive && (
          <Button
            variant="ghost"
            size="sm"
            className="h-6 text-xs gap-1"
            onClick={() => setShowSecrets(!showSecrets)}
          >
            {showSecrets ? (
              <><EyeOff className="h-3 w-3" /> {t('notifications.metadata_sanitized')}</>
            ) : (
              <><Eye className="h-3 w-3" /> {t('notifications.metadata_sensitive')}</>
            )}
          </Button>
        )}
        <Button
          variant="ghost"
          size="sm"
          className="h-6 text-xs gap-1"
          onClick={handleCopy}
        >
          {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
        </Button>
      </div>
      <pre className="rounded-md bg-muted/50 p-3 text-xs font-mono overflow-x-auto whitespace-pre">
        {formatted}
      </pre>
    </div>
  );
}
