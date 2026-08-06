'use client';

import { useLayoutEffect } from 'react';

interface FontSetterProps {
  /** The next/font variable class for the active locale (e.g. Inter or Vazirmatn) */
  fontClass: string;
}

/** The class currently applied to <html> — lets us remove only what we added. */
let lastAppliedClass: string | null = null;

/**
 * Applies the active next/font variable class (e.g. Inter for en, Vazirmatn for
 * fa — both define `--font-sans`) to <html> so that ALL content — including
 * Radix portals that render directly under <body> (dropdowns, dialogs, sheets,
 * toasts) — inherits the same font as the rest of the app. Mirrors auth/front's
 * DirectionSetter pattern.
 */
export function FontSetter({ fontClass }: FontSetterProps) {
  useLayoutEffect(() => {
    const html = document.documentElement;
    if (lastAppliedClass) {
      html.classList.remove(lastAppliedClass);
      lastAppliedClass = null;
    }
    if (fontClass) {
      html.classList.add(fontClass);
      lastAppliedClass = fontClass;
    }
  }, [fontClass]);

  return null;
}
