'use client';

import { useEffect, useState } from 'react';
import { useTheme } from 'next-themes';
import { ModeToggle } from '@minisource/ui';

/**
 * Client-only gate: renders nothing on the server so SSR and first client
 * render stay identical. Prevents a hydration mismatch flash from the theme
 * icons (next-themes reports the theme only after mount).
 */
function ClientOnly({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return <div className="h-9 w-9" aria-hidden="true" />;
  }

  return children;
}

/**
 * Theme toggle that drives next-themes from THIS app's own context.
 *
 * Important: `@minisource/ui` bundles its own `next-themes` copy (resolved
 * from design-system/node_modules/.pnpm), so calling `useTheme()` inside
 * ModeToggle reads a DIFFERENT React context than the app's ThemeProvider and
 * `setTheme` silently no-ops. We therefore pass `theme`/`onToggle` from the
 * app's own `useTheme` as props — same pattern as auth/front header-controls.
 */
export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <ClientOnly>
      <ModeToggle theme={theme} onToggle={(next) => setTheme(next)} />
    </ClientOnly>
  );
}
