'use client';

import * as React from 'react';
import { Settings } from 'lucide-react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { UserMenu as DSUserMenu } from '@minisource/app-shell';
import { authAdapter } from '@/shared/auth/auth-adapter';

export function UserMenu() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'en';
  const session = authAdapter.getSession();
  const userName = session.name || session.email?.split('@')[0] || 'User';
  const userEmail = session.email || '';
  const initials = (userName.split(' ').length > 1 ? userName.split(' ').map(n => n[0]).join('') : userName.charAt(0)).toUpperCase().slice(0, 2) || 'U';

  const menuItems = React.useMemo(() => [
    {
      id: 'settings',
      label: t('navigation.settings'),
      icon: Settings,
      onClick: () => router.push(`/settings`),
    },
  ], [locale, router, t]);

  return (
    <DSUserMenu
      name={userName}
      email={userEmail}
      initials={initials}
      items={menuItems}
      dir={locale === 'fa' ? 'rtl' : 'ltr'}
    />
  );
}

