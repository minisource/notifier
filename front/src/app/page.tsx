import { redirect } from 'next/navigation';

export default function RootPage() {
  // Locale prefixes were removed — the app root goes straight to the dashboard.
  // Next.js automatically applies basePath ('/notifier') to this internal target.
  redirect('/dashboard');
}
