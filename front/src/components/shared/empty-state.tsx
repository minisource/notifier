'use client';

import { Inbox } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface EmptyStateAction {
  label: string;
  onClick: () => void;
  variant?: 'default' | 'outline' | 'secondary' | 'ghost';
}

interface EmptyStateProps {
  title?: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
  actions?: EmptyStateAction[];
  icon?: React.ElementType;
  /** Optional tips/suggestions shown below the main content */
  tips?: string[];
  className?: string;
}

export function EmptyState({
  title = 'No data found',
  description = 'There are no items to display.',
  actionLabel,
  onAction,
  actions = [],
  icon: Icon = Inbox,
  tips,
  className,
}: EmptyStateProps) {
  const allActions: EmptyStateAction[] = [
    ...(actionLabel && onAction ? [{ label: actionLabel, onClick: onAction, variant: 'outline' as const }] : []),
    ...actions,
  ];

  return (
    <div
      className={cn(
        'relative flex flex-col items-center justify-center py-20 px-6 text-center overflow-hidden',
        'animate-in fade-in slide-in-from-bottom-4 duration-500',
        className,
      )}
    >
      {/* Decorative background pattern */}
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.03] dark:opacity-[0.05]"
        style={{
          backgroundImage: `radial-gradient(circle at 1px 1px, currentColor 1px, transparent 0)`,
          backgroundSize: '24px 24px',
        }}
        aria-hidden
      />

      {/* Icon */}
      <div className="relative mb-6">
        <div className="absolute inset-0 animate-pulse rounded-full bg-primary/5 blur-xl" aria-hidden />
        <div className="relative animate-empty-float">
          <Icon className="h-16 w-16 text-muted-foreground/30 transition-colors duration-300 group-hover:text-muted-foreground/50" />
        </div>
      </div>

      {/* Title */}
      <h3 className="relative max-w-sm text-xl font-semibold tracking-tight">
        {title}
      </h3>

      {/* Description */}
      {description && (
        <p className="relative mt-2 max-w-md text-sm text-muted-foreground leading-relaxed">
          {description}
        </p>
      )}

      {/* Action buttons */}
      {allActions.length > 0 && (
        <div className="relative mt-6 flex flex-wrap items-center justify-center gap-3">
          {allActions.map((action, idx) => (
            <Button
              key={idx}
              onClick={action.onClick}
              variant={action.variant || 'default'}
              size={idx === 0 ? 'default' : 'sm'}
              className={cn(
                'transition-all duration-200',
                'hover:shadow-md active:scale-[0.97]',
                idx === 0 && 'font-medium',
              )}
            >
              {action.label}
            </Button>
          ))}
        </div>
      )}

      {/* Tips / suggestions */}
      {tips && tips.length > 0 && (
        <div className="relative mt-8 w-full max-w-sm rounded-lg border border-dashed bg-muted/30 p-4 text-left">
          <p className="mb-2 text-xs font-medium text-muted-foreground uppercase tracking-wider">
            <span role="img" aria-label="lightbulb">💡</span> Suggestions
          </p>
          <ul className="space-y-1.5">
            {tips.map((tip, idx) => (
              <li key={idx} className="flex items-start gap-2 text-xs text-muted-foreground">
                <span className="mt-0.5 block h-1.5 w-1.5 shrink-0 rounded-full bg-primary/40" />
                {tip}
              </li>
            ))}
          </ul>
        </div>
      )}

      <style jsx>{`
        @keyframes empty-float {
          0%, 100% { transform: translateY(0px); }
          50% { transform: translateY(-6px); }
        }
        :global(.animate-empty-float) {
          animation: empty-float 3s ease-in-out infinite;
        }
      `}</style>
    </div>
  );
}
