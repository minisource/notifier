'use client';

import { useState, useEffect } from 'react';
import { useAdminAuth } from './admin-auth-context';
import { notifierRuntimeConfig } from '@/features/notifier/config/notifier-config';
import { Shield, KeyRound, Bug, Loader2, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

export function AdminLoginPage() {
  const { login, devLogin } = useAdminAuth();
  const [token, setToken] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const isDevAuthEnabled =
    notifierRuntimeConfig.mockAuthEnabled ||
    process.env.NEXT_PUBLIC_NOTIFIER_DEV_AUTH_ENABLED === 'true';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = token.trim();
    if (!trimmed) return;

    setIsSubmitting(true);
    setError(null);

    try {
      await login(trimmed);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to authenticate');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!mounted) return null;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-background via-muted/30 to-background p-4">
      {/* Decorative background */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 h-80 w-80 rounded-full bg-primary/5 blur-3xl" />
        <div className="absolute -bottom-40 -left-40 h-80 w-80 rounded-full bg-primary/5 blur-3xl" />
      </div>

      <div className="relative w-full max-w-md">
        {/* Header */}
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-primary/20">
            <Shield className="h-8 w-8 text-primary" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight">Notifier Admin</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Sign in to manage notifications, templates, and providers
          </p>
        </div>

        {/* Login Card */}
        <div className="rounded-xl border bg-card p-6 shadow-sm">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label
                htmlFor="admin-token"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              >
                Admin Token
              </label>
              <div className="relative">
                <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  id="admin-token"
                  type="password"
                  placeholder="Paste your admin JWT token"
                  value={token}
                  onChange={(e) => {
                    setToken(e.target.value);
                    setError(null);
                  }}
                  className={cn(
                    'flex h-10 w-full rounded-md border bg-background px-3 py-2 pl-10 text-sm',
                    'ring-offset-background placeholder:text-muted-foreground',
                    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
                    'disabled:cursor-not-allowed disabled:opacity-50',
                    error && 'border-destructive',
                  )}
                  disabled={isSubmitting}
                  autoFocus
                />
              </div>
              {error && (
                <div className="flex items-center gap-1.5 text-xs text-destructive">
                  <AlertCircle className="h-3.5 w-3.5 flex-shrink-0" />
                  <span>{error}</span>
                </div>
              )}
            </div>

            <button
              type="submit"
              disabled={!token.trim() || isSubmitting}
              className={cn(
                'inline-flex h-10 w-full items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground',
                'ring-offset-background transition-colors',
                'hover:bg-primary/90',
                'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
                'disabled:pointer-events-none disabled:opacity-50',
              )}
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Verifying...
                </>
              ) : (
                'Sign In'
              )}
            </button>
          </form>

          {/* Dev mode bypass */}
          {isDevAuthEnabled && (
            <div className="mt-6">
              <div className="relative mb-4">
                <div className="absolute inset-0 flex items-center">
                  <span className="w-full border-t" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-card px-2 text-muted-foreground">
                    Development Mode
                  </span>
                </div>
              </div>

              <button
                type="button"
                onClick={devLogin}
                className={cn(
                  'inline-flex h-10 w-full items-center justify-center rounded-md border bg-background px-4 py-2 text-sm font-medium',
                  'ring-offset-background transition-colors',
                  'hover:bg-accent hover:text-accent-foreground',
                  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
                )}
              >
                <Bug className="mr-2 h-4 w-4" />
                Continue with Mock Admin
              </button>

              <p className="mt-2 text-center text-xs text-muted-foreground">
                Quick access for local development. Not available in production.
              </p>
            </div>
          )}
        </div>

        {/* Footer */}
        <p className="mt-6 text-center text-xs text-muted-foreground">
          Notifier Service — Admin Panel v1.0.0
        </p>
      </div>
    </div>
  );
}
