import { getRequestConfig } from 'next-intl/server';
import { hasLocale } from 'next-intl';

export const locales = ['fa', 'en'] as const;
export const defaultLocale = 'en' as const;

export type Locale = (typeof locales)[number];

export default getRequestConfig(async ({ requestLocale }) => {
  const requested = await requestLocale;
  const locale = hasLocale(locales, requested) ? requested : defaultLocale;

  return {
    locale,
    messages: (await import(`./messages/${locale}.json`)).default,
    // Never render a raw i18n key like "common.active" in the UI. Return an
    // empty string for missing keys so callers using `t('key') || 'Fallback'`
    // fall back to their English/translated default.
    getMessageFallback: () => '',
  };
});
