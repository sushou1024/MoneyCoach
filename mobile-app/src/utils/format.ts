import { getCurrentLocale } from './i18n'

const currencyFormatterSupportCache = new Map<string, boolean>()

function normalizeMoneyUnit(currency: string) {
  const unit = currency?.trim().toUpperCase()
  return unit || 'USD'
}

function supportsIntlCurrency(unit: string) {
  if (currencyFormatterSupportCache.has(unit)) {
    return currencyFormatterSupportCache.get(unit) ?? false
  }

  let supported = false
  try {
    const resolved = new Intl.NumberFormat(getCurrentLocale(), {
      style: 'currency',
      currency: unit,
      maximumFractionDigits: 2,
    }).resolvedOptions()
    supported = resolved.currency === unit
  } catch {
    supported = false
  }
  currencyFormatterSupportCache.set(unit, supported)
  return supported
}

function formatUnitAmount(value: number, unit: string, options?: Intl.NumberFormatOptions) {
  const amount = Number.isFinite(value) ? value : 0
  const formatted = new Intl.NumberFormat(getCurrentLocale(), options).format(amount)
  return `${formatted} ${unit}`
}

function formatMoneyAmount(amount: number, unit: string, options?: Intl.NumberFormatOptions) {
  if (!supportsIntlCurrency(unit)) {
    return formatUnitAmount(amount, unit, options)
  }
  return new Intl.NumberFormat(getCurrentLocale(), {
    style: 'currency',
    currency: unit,
    currencyDisplay: 'narrowSymbol',
    ...options,
  }).format(amount)
}

export function formatCurrency(value: number, currency: string) {
  const amount = Number.isFinite(value) ? value : 0
  const unit = normalizeMoneyUnit(currency)
  return formatMoneyAmount(amount, unit, { maximumFractionDigits: 2 })
}

export function formatNumber(value: number, maxDecimals = 2) {
  const amount = Number.isFinite(value) ? value : 0
  return new Intl.NumberFormat(getCurrentLocale(), { maximumFractionDigits: maxDecimals }).format(amount)
}

export function formatPercent(value: number, maxDecimals = 1) {
  const amount = Number.isFinite(value) ? value : 0
  return new Intl.NumberFormat(getCurrentLocale(), {
    style: 'percent',
    maximumFractionDigits: maxDecimals,
  }).format(amount)
}

export function formatCompactCurrency(value: number, currency: string) {
  const amount = Number.isFinite(value) ? value : 0
  const unit = normalizeMoneyUnit(currency)
  const abs = Math.abs(amount)
  const useCompact = abs >= 100_000
  return formatMoneyAmount(amount, unit, {
    notation: useCompact ? 'compact' : 'standard',
    maximumFractionDigits: useCompact ? 1 : 2,
  })
}

export function formatPortfolioTotal(value: number, currency: string) {
  const amount = Number.isFinite(value) ? value : 0
  const unit = normalizeMoneyUnit(currency)
  const abs = Math.abs(amount)

  if (abs < 1_000) {
    return formatMoneyAmount(amount, unit, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })
  }

  let divisor = 1_000
  let suffix = 'K'
  if (abs >= 1_000_000_000) {
    divisor = 1_000_000_000
    suffix = 'B'
  } else if (abs >= 1_000_000) {
    divisor = 1_000_000
    suffix = 'M'
  }

  const scaled = amount / divisor
  const formatted = formatMoneyAmount(scaled, unit, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
  if (!supportsIntlCurrency(unit)) {
    const unitSuffix = ` ${unit}`
    if (formatted.endsWith(unitSuffix)) {
      return `${formatted.slice(0, -unitSuffix.length)}${suffix}${unitSuffix}`
    }
  }
  return `${formatted}${suffix}`
}

export function formatDateTime(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(getCurrentLocale())
}
