const dateFormatter = (locale: string) =>
  new Intl.DateTimeFormat(locale === 'fa' ? 'fa-IR' : 'en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });

const dateTimeFormatter = (locale: string) =>
  new Intl.DateTimeFormat(locale === 'fa' ? 'fa-IR' : 'en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

const timeFormatter = (locale: string) =>
  new Intl.DateTimeFormat(locale === 'fa' ? 'fa-IR' : 'en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });

export function formatDate(date: string | Date, locale = 'fa'): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return dateFormatter(locale).format(d);
}

export function formatDateTime(date: string | Date, locale = 'fa'): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return dateTimeFormatter(locale).format(d);
}

export function formatTime(date: string | Date, locale = 'fa'): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return timeFormatter(locale).format(d);
}

export function formatRelativeTime(date: string | Date, locale = 'fa'): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (locale === 'fa') {
    if (diffSec < 60) return 'همین حالا';
    if (diffMin < 60) return `${diffMin} دقیقه پیش`;
    if (diffHour < 24) return `${diffHour} ساعت پیش`;
    if (diffDay < 7) return `${diffDay} روز پیش`;
    return formatDate(date, locale);
  }

  if (diffSec < 60) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return formatDate(date, locale);
}

export function timeAgo(date: string | Date, locale = 'fa'): string {
  return formatRelativeTime(date, locale);
}
