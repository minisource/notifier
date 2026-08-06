import React, { Suspense } from 'react';
import { TestCenterView } from '@/features/test-center/components/test-center-view';

export const dynamic = 'force-dynamic';

export const metadata = {
  title: 'Notifier Test Center | MiniSource Admin',
  description: 'API Operation Diagnostics & Multi-Step Flow Test Center for MiniSource Notifier Service',
};

export default function TestCenterPage() {
  return (
    <Suspense fallback={
      <div className="flex h-48 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    }>
      <TestCenterView />
    </Suspense>
  );
}
