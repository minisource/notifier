export function getDirection(locale: string): 'rtl' | 'ltr' {
  return locale === 'fa' ? 'rtl' : 'ltr';
}

export function isRTL(locale: string): boolean {
  return locale === 'fa';
}

export const rtlLocales: string[] = ['fa'];

export type Direction = 'rtl' | 'ltr';
