import { NextResponse, type NextRequest } from 'next/server';
import createMiddleware from 'next-intl/middleware';
import { locales, defaultLocale } from './i18n';

const intlMiddleware = createMiddleware({
  locales,
  defaultLocale,
  // No locale prefix in URLs — the app is served at clean paths like
  // /dashboard, /notifications (basePath '/notifier' via the gateway).
  // Locale is resolved from the NEXT_LOCALE cookie (default 'en').
  localePrefix: 'never',
  // Default to 'en' regardless of the browser's Accept-Language.
  // Users can still switch via the in-app language switcher (sets NEXT_LOCALE cookie).
  localeDetection: false,
});

// The matcher below runs before this function, but with a basePath the
// incoming pathname keeps the '/notifier' prefix, so the standard
// `api|_next|_vercel` exclusions can't match. Skip non-page requests here so
// next-intl never rewrites them to a locale-prefixed URL (which would 404).
export default function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (
    pathname.includes('/api/') ||
    pathname.includes('/_next/') ||
    /\.\w+$/.test(pathname)
  ) {
    return NextResponse.next();
  }

  return intlMiddleware(request);
}

export const config = {
  matcher: ['/((?!api|_next|_vercel|.*\\..*).*)'],
};
