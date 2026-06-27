'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from 'next-themes';
import { Toaster } from '@/components/ui/sonner';
import { useState } from 'react';
import { AdminAuthProvider } from '@/features/auth/admin-auth-context';

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 60_000,
        gcTime: 5 * 60_000,
        retry: (failureCount, error) => {
          // Don't retry client errors (4xx) — they won't succeed on retry
          const status = (error as any)?.status ?? (error as any)?.response?.status;
          if (status && [400, 401, 403, 404, 422].includes(status)) return false;
          return failureCount < 1;
        },
        refetchOnWindowFocus: false,
        refetchOnReconnect: true,
        refetchOnMount: true,
      },
      mutations: {
        retry: false,
      },
    },
  }));

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
        <AdminAuthProvider>
          {children}
        </AdminAuthProvider>
        <Toaster />
      </ThemeProvider>
    </QueryClientProvider>
  );
}
