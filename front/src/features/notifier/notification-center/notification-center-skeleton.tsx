'use client';

import { Skeleton } from '@/components/ui/skeleton';

interface NotificationCenterSkeletonProps {
  count?: number;
}

export function NotificationCenterSkeleton({ count = 5 }: NotificationCenterSkeletonProps) {
  return (
    <div className="p-4 space-y-3">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex items-start gap-3">
          <div className="flex-1 space-y-2">
            <div className="flex gap-2">
              <Skeleton className="h-5 w-12 rounded-full" />
              <Skeleton className="h-5 w-16 rounded-full" />
            </div>
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-1/2" />
          </div>
          <Skeleton className="h-7 w-7 rounded" />
        </div>
      ))}
    </div>
  );
}
