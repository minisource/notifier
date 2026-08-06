'use client';

import { useState, useEffect } from 'react';

/**
 * Check if the viewport matches a CSS media query.
 * Returns false during SSR / before hydration.
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(false);

  useEffect(() => {
    const media = window.matchMedia(query);

    if (media.matches !== matches) {
      setMatches(media.matches);
    }

    const listener = (event: MediaQueryListEvent) => {
      setMatches(event.matches);
    };

    media.addEventListener('change', listener);
    return () => media.removeEventListener('change', listener);
  }, [matches, query]);

  return matches;
}

/** True on screens narrower than 640px (below Tailwind's `sm` breakpoint). */
export function useIsMobile(): boolean {
  return useMediaQuery('(max-width: 639px)');
}
