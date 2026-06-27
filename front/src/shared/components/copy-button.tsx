'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Copy, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface CopyButtonProps {
  value: string;
  label?: string;
  className?: string;
  variant?: 'ghost' | 'outline' | 'default';
  size?: 'sm' | 'default' | 'icon';
  showLabel?: boolean;
}

export function CopyButton({ value, label, className, variant = 'ghost', size = 'icon', showLabel }: CopyButtonProps) {
  const t = useTranslations();
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Button
      variant={variant}
      size={size}
      className={cn('gap-1', className)}
      onClick={handleCopy}
      aria-label={t('common.copy_id')}
    >
      {copied ? <Check className="h-3.5 w-3.5 text-green-500" /> : <Copy className="h-3.5 w-3.5" />}
      {showLabel && (label || t('common.copy_id'))}
    </Button>
  );
}
