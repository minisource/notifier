import type { Metadata, Viewport } from 'next';
import { Inter, Vazirmatn } from 'next/font/google';
import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { locales } from '@/i18n';
import { Providers } from '@/components/providers';
import { AppShell } from '@/components/layout/app-shell';
import { getDirection } from '@/lib/utils/direction';
import '@/styles/globals.css';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-sans',
  display: 'swap',
});

const vazirmatn = Vazirmatn({
  subsets: ['arabic'],
  variable: '--font-fa',
  display: 'swap',
});

export const metadata: Metadata = {
  title: {
    default: 'Notifier Admin',
    template: '%s | Notifier Admin',
  },
  description: 'Notifier Service Admin Panel - Manage notifications, templates, and preferences',
  manifest: '/manifest.json',
  appleWebApp: {
    capable: true,
    statusBarStyle: 'default',
    title: 'Notifier Admin',
  },
  formatDetection: {
    telephone: false,
    address: false,
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: 'white' },
    { media: '(prefers-color-scheme: dark)', color: 'black' },
  ],
  width: 'device-width',
  initialScale: 1,
};

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
    <html lang={locale} dir={direction} suppressHydrationWarning>
      <body className={`${fontClass} font-sans antialiased`}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <Providers>
            <AppShell>{children}</AppShell>
          </Providers>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
