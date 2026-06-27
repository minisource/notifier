import { cn } from '@/lib/utils';
import { MessageSquare, Mail, Smartphone, Globe, Webhook } from 'lucide-react';

const channelIcons: Record<string, React.ElementType> = {
  sms: MessageSquare,
  email: Mail,
  push: Smartphone,
  in_app: Globe,
  webhook: Webhook,
};

const channelStyles: Record<string, string> = {
  sms: 'bg-green-50 text-green-700 dark:bg-green-950/40 dark:text-green-400 border-green-200 dark:border-green-900',
  email: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-400 border-blue-200 dark:border-blue-900',
  push: 'bg-purple-50 text-purple-700 dark:bg-purple-950/40 dark:text-purple-400 border-purple-200 dark:border-purple-900',
  in_app: 'bg-orange-50 text-orange-700 dark:bg-orange-950/40 dark:text-orange-400 border-orange-200 dark:border-orange-900',
  webhook: 'bg-gray-50 text-gray-700 dark:bg-gray-900/40 dark:text-gray-400 border-gray-200 dark:border-gray-800',
};

interface ChannelBadgeProps {
  channel: string;
  className?: string;
  showIcon?: boolean;
  size?: 'sm' | 'md';
}

export function ChannelBadge({ channel, className, showIcon = true, size = 'sm' }: ChannelBadgeProps) {
  const Icon = channelIcons[channel.toLowerCase()];
  const style = channelStyles[channel.toLowerCase()] || channelStyles.webhook;
  const sizeClasses = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm';

  return (
    <span className={cn(
      'inline-flex items-center gap-1.5 rounded-full border font-medium',
      sizeClasses,
      style,
      className
    )}>
      {showIcon && Icon && <Icon className="h-3 w-3" />}
      {channel}
    </span>
  );
}
