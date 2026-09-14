const currencyFormatters: Record<string, Intl.NumberFormat> = {};

function getFormatter(currency: string): Intl.NumberFormat {
  if (!currencyFormatters[currency]) {
    currencyFormatters[currency] = new Intl.NumberFormat(undefined, {
      style: "currency",
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  }
  return currencyFormatters[currency];
}

export function formatMoney(amountMinor: number, currency: string): string {
  const amount = amountMinor / 100;
  return getFormatter(currency).format(amount);
}

export function formatMoneyCompact(amountMinor: number, currency: string): string {
  const amount = amountMinor / 100;
  const absAmount = Math.abs(amount);

  if (absAmount >= 1_000_000) {
    return `${amount < 0 ? "-" : ""}${(absAmount / 1_000_000).toFixed(1)}M ${currency}`;
  }
  if (absAmount >= 1_000) {
    return `${amount < 0 ? "-" : ""}${(absAmount / 1_000).toFixed(1)}K ${currency}`;
  }
  return getFormatter(currency).format(amount);
}

export function parseMoneyInput(value: string): number | null {
  const cleaned = value.replace(/[^0-9.-]/g, "");
  const num = parseFloat(cleaned);
  if (isNaN(num)) return null;
  return Math.round(num * 100);
}
