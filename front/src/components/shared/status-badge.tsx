import { cn } from '@/lib/utils';

interface StatusBadgeProps {
  status: string;
  className?: string;
  showDot?: boolean;
  size?: 'sm' | 'md';
}

const statusConfig: Record<string, { dot: string; bg: string; label?: string }> = {
  // Notification statuses
  pending:    { dot: 'bg-yellow-500', bg: 'bg-yellow-50 text-yellow-700 dark:bg-yellow-950/40 dark:text-yellow-400' },
  queued:     { dot: 'bg-blue-500', bg: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-400' },
  processing: { dot: 'bg-blue-500', bg: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-400' },
  sending:    { dot: 'bg-blue-500', bg: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-400' },
  sent:       { dot: 'bg-green-500', bg: 'bg-green-50 text-green-700 dark:bg-green-950/40 dark:text-green-400' },
  delivered:  { dot: 'bg-green-500', bg: 'bg-green-50 text-green-700 dark:bg-green-950/40 dark:text-green-400' },
  failed:     { dot: 'bg-red-500', bg: 'bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-400' },
  retrying:   { dot: 'bg-amber-500', bg: 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400' },
  canceled:   { dot: 'bg-gray-400', bg: 'bg-gray-50 text-gray-600 dark:bg-gray-900/40 dark:text-gray-400' },
  cancelled:  { dot: 'bg-gray-400', bg: 'bg-gray-50 text-gray-600 dark:bg-gray-900/40 dark:text-gray-400' },
  dead:       { dot: 'bg-red-500', bg: 'bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-400' },
  digested:   { dot: 'bg-purple-500', bg: 'bg-purple-50 text-purple-700 dark:bg-purple-950/40 dark:text-purple-400' },
  scheduled:  { dot: 'bg-indigo-500', bg: 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950/40 dark:text-indigo-400' },
  // Delivery tracking
  seen:       { dot: 'bg-teal-500', bg: 'bg-teal-50 text-teal-700 dark:bg-teal-950/40 dark:text-teal-400' },
  read:       { dot: 'bg-emerald-500', bg: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400' },
  clicked:    { dot: 'bg-cyan-500', bg: 'bg-cyan-50 text-cyan-700 dark:bg-cyan-950/40 dark:text-cyan-400' },
  // Provider statuses
  healthy:    { dot: 'bg-green-500', bg: 'bg-green-50 text-green-700 dark:bg-green-950/40 dark:text-green-400' },
  degraded:   { dot: 'bg-amber-500', bg: 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400' },
  down:       { dot: 'bg-red-500', bg: 'bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-400' },
  disabled:   { dot: 'bg-gray-400', bg: 'bg-gray-50 text-gray-600 dark:bg-gray-900/40 dark:text-gray-400' },
};

export function StatusBadge({ status, className, showDot = true, size = 'sm' }: StatusBadgeProps) {
  const config = statusConfig[status.toLowerCase()] || { dot: 'bg-gray-400', bg: 'bg-gray-50 text-gray-600 dark:bg-gray-900/40 dark:text-gray-400' };
  const sizeClasses = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm';

  return (
    <span className={cn(
      'inline-flex items-center gap-1.5 rounded-full font-medium',
      sizeClasses,
      config.bg,
      className
    )}>
      {showDot && <span className={cn('h-1.5 w-1.5 rounded-full', config.dot)} />}
      {status}
    </span>
  );
}
