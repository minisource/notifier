'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';

export default function LocaleHome() {
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';

  useEffect(() => {
    router.replace(`/${locale}/dashboard`);
  }, [locale, router]);

  return null;
}
