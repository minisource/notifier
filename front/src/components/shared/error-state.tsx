'use client';

import { AlertCircle, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { cn } from '@/lib/utils';
import { useState, useEffect, useRef } from 'react';

interface ErrorStateProps {
  title?: string;
  message?: string;
  onRetry?: () => void;
  /** Auto-retry countdown in seconds (0 = no auto-retry) */
  autoRetrySeconds?: number;
  className?: string;
}

export function ErrorState({
  title = 'Something went wrong',
  message = 'An error occurred while loading data. Please try again.',
  onRetry,
  autoRetrySeconds = 0,
  className,
}: ErrorStateProps) {
  const [countdown, setCountdown] = useState(autoRetrySeconds);
  const hasAutoRetried = useRef(false);

  useEffect(() => {
    if (autoRetrySeconds <= 0 || !onRetry || hasAutoRetried.current) return;
    if (countdown <= 0) {
      hasAutoRetried.current = true;
      onRetry();
      return;
    }
    const timer = setInterval(() => {
      setCountdown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [countdown, autoRetrySeconds, onRetry]);

  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-16 text-center',
        'animate-in fade-in slide-in-from-bottom-3 duration-500',
        className,
      )}
    >
      {/* Animated icon */}
      <div className="relative mb-6">
        <div className="absolute inset-0 animate-pulse rounded-full bg-destructive/10 blur-xl" aria-hidden />
        <div className="relative animate-error-shake">
          <AlertCircle className="h-12 w-12 text-destructive/70" />
        </div>
      </div>

      <Alert variant="destructive" className="max-w-md text-left">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>{title}</AlertTitle>
        <AlertDescription>{message}</AlertDescription>
      </Alert>

      {onRetry && (
        <div className="mt-4 flex items-center gap-3">
          <Button
            onClick={onRetry}
            variant="outline"
            size="sm"
            className="transition-all duration-200 hover:shadow-md active:scale-[0.97]"
          >
            <RefreshCw className="ml-1.5 h-4 w-4" />
            Try Again
          </Button>
          {autoRetrySeconds > 0 && countdown > 0 && (
            <span className="text-xs text-muted-foreground animate-in fade-in duration-300">
              Auto-retry in {countdown}s
            </span>
          )}
        </div>
      )}

      <style jsx>{`
        @keyframes error-shake {
          0%, 100% { transform: translateX(0); }
          10%, 30%, 50%, 70%, 90% { transform: translateX(-2px); }
          20%, 40%, 60%, 80% { transform: translateX(2px); }
        }
        :global(.animate-error-shake) {
          animation: error-shake 0.8s ease-in-out;
        }
      `}</style>
    </div>
  );
}
