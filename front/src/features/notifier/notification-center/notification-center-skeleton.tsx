'use client';

import { Skeleton } from '@minisource/ui';

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
              <Skeleton className="h-64 w-full" />
              <Skeleton className="h-64 w-full" />
            </div>
            <Skeleton className="h-64 w-full" />
            <Skeleton className="h-64 w-full" />
          </div>
          <Skeleton className="h-64 w-full" />
        </div>
      ))}
    </div>
  );
}
