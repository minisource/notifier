const numberFormatter = (locale: string) =>
  new Intl.NumberFormat(locale === 'fa' ? 'fa-IR' : 'en-US');

const percentFormatter = (locale: string) =>
  new Intl.NumberFormat(locale === 'fa' ? 'fa-IR' : 'en-US', {
    style: 'percent',
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  });

export function formatNumber(num: number, locale = 'fa'): string {
  return numberFormatter(locale).format(num);
}

export function formatPercentage(value: number, locale = 'fa'): string {
  return percentFormatter(locale).format(value / 100);
}

export function formatMilliseconds(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

export function maskEmail(email: string): string {
  const [name, domain] = email.split('@');
  if (!domain) return email;
  const maskedName = name.length <= 2
    ? name[0] + '***'
    : name[0] + '***' + name[name.length - 1];
  return `${maskedName}@${domain}`;
}

export function maskPhone(phone: string): string {
  if (phone.length < 6) return phone;
  return phone.slice(0, 4) + '***' + phone.slice(-4);
}

export function shortId(id: string): string {
  if (id.length <= 8) return id;
  return id.slice(0, 8) + '...';
}

export function truncate(str: string, maxLength: number): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength) + '...';
}
