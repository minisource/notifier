import { cn } from '@/lib/utils';

interface MiniProgressProps {
  value: number; // 0-100
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info';
  size?: 'sm' | 'md';
  className?: string;
  showLabel?: boolean;
  label?: string;
}

const variantStyles = {
  default: 'bg-primary',
  success: 'bg-green-500',
  warning: 'bg-amber-500',
  danger: 'bg-red-500',
  info: 'bg-blue-500',
};

export function MiniProgress({
  value,
  variant = 'default',
  size = 'sm',
  className,
  showLabel = false,
  label,
}: MiniProgressProps) {
  const height = size === 'sm' ? 'h-1.5' : 'h-2';
  const clampedValue = Math.min(100, Math.max(0, value));

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className={cn('flex-1 rounded-full bg-muted', height)}>
        <div
          className={cn('h-full rounded-full transition-all', variantStyles[variant])}
          style={{ width: `${clampedValue}%` }}
        />
      </div>
      {(showLabel || label) && (
        <span className="text-xs font-medium text-muted-foreground whitespace-nowrap">
          {label || `${Math.round(clampedValue)}%`}
        </span>
      )}
    </div>
  );
}
