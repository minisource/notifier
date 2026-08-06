// Shared helpers for displaying provider account balance values.
// Some providers report a monetary credit balance (e.g. Kavenegar's
// remaincredit is in Rial), others report a message count or quota — the
// unit must be shown truthfully, never assumed to be a count.

export function formatBalanceValue(value?: number | null): string {
  if (value === null || value === undefined) return '';
  return new Intl.NumberFormat('en-US', { maximumFractionDigits: 2 }).format(value);
}

// balanceUnitLabel returns the translated label for a provider-reported
// balance unit (rial, count, usd, ...), falling back to the raw unit string
// when no translation exists.
export function balanceUnitLabel(t: (key: string) => string, unit?: string | null): string {
  if (!unit) return '';
  const key = `providers.balance.units.${unit}`;
  const label = t(key);
  return label && label !== key ? label : unit;
}
