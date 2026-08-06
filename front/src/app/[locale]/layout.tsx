import { Inter, Vazirmatn } from 'next/font/google';
import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { locales } from '@/i18n';
import { Providers } from '@/components/providers';
import { AppShell } from '@/components/layout/app-shell';
import { FontSetter } from '@/components/shared/font-setter';
import { getDirection } from '@/lib/utils/direction';
import '@/styles/globals.css';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-sans',
  display: 'swap',
});

const vazirmatn = Vazirmatn({
  subsets: ['arabic'],
  variable: '--font-sans',
  display: 'swap',
});

/**
 * Locale layout — provides locale context, fonts, and app shell.
 * Note: <html> and <body> are now in the root layout (app/layout.tsx).
 * Font CSS variables are applied via className on the wrapper div.
 */
export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;

  if (!locales.includes(locale as any)) {
    notFound();
  }

  const messages = await getMessages();
  const direction = getDirection(locale);
  const fontClass = locale === 'fa' ? vazirmatn.variable : inter.variable;

  return (
    <div className={`${fontClass} font-sans antialiased`} dir={direction}>
      {/* Keeps the font on <html> so Radix portals (dropdowns/dialogs) inherit it too */}
      <FontSetter fontClass={fontClass} />
      <NextIntlClientProvider locale={locale} messages={messages}>
        <Providers>
          <AppShell>{children}</AppShell>
        </Providers>
      </NextIntlClientProvider>
    </div>
  );
}
