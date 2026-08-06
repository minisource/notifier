'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function LocaleHome() {
  const router = useRouter();

  useEffect(() => {
    // No locale prefix in URLs anymore — root goes straight to the dashboard.
    router.replace('/dashboard');
  }, [router]);

  return null;
}
