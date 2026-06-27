'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import { useEffect, useState } from 'react';

// ==================== Loading messages by page context ====================

const LOADING_MESSAGES: Record<string, string[]> = {
  default: [
    'Loading data...',
    'Almost there...',
    'Fetching latest information...',
  ],
  notifications: [
    'Loading notifications...',
    'Fetching latest messages...',
    'Preparing notification list...',
  ],
  templates: [
    'Loading templates...',
    'Preparing template editor...',
    'Fetching template library...',
  ],
  providers: [
    'Checking provider status...',
    'Loading provider configurations...',
    'Measuring provider health...',
  ],
  dashboard: [
    'Building dashboard overview...',
    'Aggregating metrics...',
    'Loading real-time data...',
  ],
  reminders: [
    'Loading scheduled reminders...',
    'Fetching reminder schedules...',
    'Preparing reminder list...',
  ],
  deliveries: [
    'Tracking delivery status...',
    'Loading delivery logs...',
    'Fetching delivery attempts...',
  ],
  tenants: [
    'Loading project data...',
    'Fetching tenant configurations...',
    'Preparing project overview...',
  ],
  observability: [
    'Gathering health metrics...',
    'Checking all dependencies...',
    'Preparing observability dashboard...',
  ],
  preferences: [
    'Loading user preferences...',
    'Fetching notification settings...',
    'Preparing preferences...',
  ],
};

// ==================== LoadingMessage ====================

function LoadingMessage({ context = 'default' }: { context?: string }) {
  const messages = LOADING_MESSAGES[context] || LOADING_MESSAGES.default;
  const [index, setIndex] = useState(0);
  const [visible, setVisible] = useState(true);

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;
    const interval = setInterval(() => {
      setVisible(false);
      timeoutId = setTimeout(() => {
        setIndex((prev) => (prev + 1) % messages.length);
        setVisible(true);
      }, 300);
    }, 4000);
    return () => {
      clearInterval(interval);
      if (timeoutId !== null) clearTimeout(timeoutId);
    };
  }, [messages.length]);

  return (
    <p
      className={cn(
        'text-xs text-muted-foreground transition-opacity duration-300',
        visible ? 'opacity-100' : 'opacity-0',
      )}
    >
      {messages[index]}
    </p>
  );
}

// ==================== Skeleton helpers ====================

function Shimmer({ className, maxWidth }: { className?: string; maxWidth?: string }) {
  return (
    <div style={maxWidth ? { maxWidth } : undefined} className={cn('flex-1', maxWidth && 'shrink-0')}>
      <Skeleton className={cn('h-full w-full animate-pulse rounded', className)} />
    </div>
  );
}

function HeaderSkeleton({ columns = 4 }: { columns?: number }) {
  return (
    <div className="flex gap-4">
      {Array.from({ length: columns }).map((_, i) => (
        <Shimmer key={i} className="h-24 rounded-lg" />
      ))}
    </div>
  );
}

function TableRowSkeleton({ columns = 6 }: { columns?: number }) {
  return (
    <div className="flex gap-4 py-2.5">
      {Array.from({ length: columns }).map((_, i) => (
        <Shimmer
          key={i}
          className="h-5 rounded"
          maxWidth={i === 0 ? '220px' : i === columns - 1 ? '48px' : undefined}
        />
      ))}
    </div>
  );
}

// ==================== LoadingState (deprecated — kept for compatibility) ====================

interface LoadingStateProps {
  rows?: number;
  columns?: number;
}

export function LoadingState({ rows = 5, columns = 4 }: LoadingStateProps) {
  return (
    <div className="space-y-4 animate-in fade-in duration-300">
      <HeaderSkeleton columns={columns} />
      <div className="space-y-2">
        {Array.from({ length: rows }).map((_, i) => (
          <TableRowSkeleton key={i} columns={columns} />
        ))}
      </div>
    </div>
  );
}

// ==================== TableSkeleton ====================

interface TableSkeletonProps {
  rows?: number;
  columns?: number;
  context?: string;
}

export function TableSkeleton({ rows = 8, columns = 6, context }: TableSkeletonProps) {
  return (
    <div className="space-y-4 animate-in fade-in duration-300">
      {context && (
        <div className="flex items-center gap-2 pb-1">
          <span className="relative flex h-2 w-2">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary/40" />
            <span className="relative inline-flex h-2 w-2 rounded-full bg-primary/60" />
          </span>
          <LoadingMessage context={context} />
        </div>
      )}
      <div className="flex gap-4 pb-2 border-b border-border/50">
        {Array.from({ length: columns }).map((_, i) => (
          <Shimmer key={i} className="h-5 flex-1 rounded" />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <TableRowSkeleton key={i} columns={columns} />
      ))}
    </div>
  );
}

// ==================== CardSkeleton ====================

interface CardSkeletonProps {
  cards?: number;
  context?: string;
}

export function CardSkeleton({ cards = 6, context }: CardSkeletonProps) {
  return (
    <div className="space-y-4 animate-in fade-in duration-300">
      {context && (
        <div className="flex items-center gap-2">
          <span className="relative flex h-2 w-2">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary/40" />
            <span className="relative inline-flex h-2 w-2 rounded-full bg-primary/60" />
          </span>
          <LoadingMessage context={context} />
        </div>
      )}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: cards }).map((_, i) => (
          <div key={i} className="rounded-lg border border-border/70 p-4 space-y-3">
            <div className="flex items-center justify-between">
              <Shimmer className="h-5 w-32 rounded" />
              <Shimmer className="h-6 w-20 rounded-full" />
            </div>
            <div className="space-y-2">
              <Shimmer className="h-4 w-full rounded" />
              <Shimmer className="h-4 w-3/4 rounded" />
              <Shimmer className="h-4 w-1/2 rounded" />
            </div>
            <Shimmer className="h-2 w-full rounded-full" />
            <Shimmer className="h-8 w-full rounded-md" />
          </div>
        ))}
      </div>
    </div>
  );
}

// ==================== PageSkeleton ====================

interface PageSkeletonProps {
  context?: string;
  /** 'table' | 'cards' | 'detail' */
  layout?: 'table' | 'cards' | 'detail';
  rows?: number;
  columns?: number;
  cards?: number;
}

export function PageSkeleton({
  context,
  layout = 'table',
  rows = 8,
  columns = 6,
  cards = 6,
}: PageSkeletonProps) {
  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      {/* Header skeleton */}
      <div className="flex items-center justify-between">
        <div className="space-y-1.5">
          <Shimmer className="h-7 w-48 rounded" />
          <Shimmer className="h-4 w-72 rounded" />
        </div>
        <div className="flex items-center gap-2">
          <Shimmer className="h-9 w-24 rounded-md" />
          <Shimmer className="h-9 w-32 rounded-md" />
        </div>
      </div>

      {/* Content skeleton */}
      {layout === 'table' && <TableSkeleton rows={rows} columns={columns} context={context} />}
      {layout === 'cards' && <CardSkeleton cards={cards} context={context} />}
      {layout === 'detail' && (
        <div className="space-y-4">
          {context && (
            <div className="flex items-center gap-2">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary/40" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-primary/60" />
              </span>
              <LoadingMessage context={context} />
            </div>
          )}
          <div className="grid gap-4 sm:grid-cols-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="space-y-2 rounded-lg border p-4">
                <Shimmer className="h-4 w-24 rounded" />
                <Shimmer className="h-5 w-40 rounded" />
              </div>
            ))}
          </div>
          <Shimmer className="h-48 w-full rounded-lg" />
        </div>
      )}
    </div>
  );
}
