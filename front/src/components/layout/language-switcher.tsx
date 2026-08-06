'use client';

import { Button } from '@minisource/ui';
import { Languages } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@minisource/ui';
import { useParams } from 'next/navigation';

const languages = [
  { code: 'fa', label: 'فارسی' },
  { code: 'en', label: 'English' },
];

export function LanguageSwitcher() {
  const params = useParams();
  const locale = (params?.locale as string) || 'en';

  // URLs are locale-free (localePrefix: 'never'), so switching the language
  // only flips the NEXT_LOCALE cookie that next-intl's middleware reads.
  const switchLanguage = (newLocale: string) => {
    if (newLocale === locale) return;
    document.cookie = `NEXT_LOCALE=${newLocale}; path=/; max-age=31536000; samesite=lax`;
    window.location.reload();
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Switch language">
          <Languages className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {languages.map(lang => (
          <DropdownMenuItem
            key={lang.code}
            onClick={() => switchLanguage(lang.code)}
            className={locale === lang.code ? 'bg-accent font-medium' : ''}
          >
            {lang.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
