import type { Metadata, Viewport } from 'next';

export const metadata: Metadata = {
  title: {
    default: 'Notifier Admin',
    template: '%s | Notifier Admin',
  },
  description: 'Notifier Service Admin Panel - Manage notifications, templates, and preferences',
};

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: 'white' },
    { media: '(prefers-color-scheme: dark)', color: 'black' },
  ],
  width: 'device-width',
  initialScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return children;
}
