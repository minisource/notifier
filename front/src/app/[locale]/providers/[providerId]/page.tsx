'use client';

import { useParams } from 'next/navigation';
import { ProviderDetail } from '@/features/providers/components/provider-detail';

export default function ProviderDetailPage() {
  const params = useParams();
  const providerId = params?.providerId as string;

  if (!providerId) return null;

  return <ProviderDetail providerId={providerId} />;
}
