import { BadgeTone } from './badges'
import { formatNumber, formatPercent } from './format'
import { TranslationKey } from './i18n'
import { AssetBrief, MarketRegime } from '../types/api'

type TranslateFn = (key: TranslationKey, params?: Record<string, string | number>) => string

function resolveText(t: TranslateFn, key: TranslationKey, fallback: string, params?: Record<string, string | number>) {
  const value = t(key, params)
  return value === key ? fallback : value
}

export function formatSignedPercent(value: number, maxDecimals = 1) {
  const prefix = value > 0 ? '+' : ''
  return `${prefix}${formatPercent(value, maxDecimals)}`
}

export function toneForRegime(regime: MarketRegime['regime']): BadgeTone {
  switch (regime) {
    case 'risk_on':
      return 'low'
    case 'risk_off':
      return 'high'
    default:
      return 'medium'
  }
}

export function toneForActionBias(actionBias: AssetBrief['action_bias']): BadgeTone {
  switch (actionBias) {
    case 'accumulate':
      return 'low'
    case 'hold':
      return 'medium'
    case 'reduce':
      return 'high'
    default:
      return 'neutral'
  }
}

export function toneForRiskFlag(riskFlag: AssetBrief['portfolio_fit']['risk_flag']): BadgeTone {
  switch (riskFlag) {
    case 'high_concentration':
      return 'critical'
    case 'high_beta':
      return 'high'
    case 'balanced':
      return 'low'
    default:
      return 'neutral'
  }
}

export function resolveRegimeLabel(t: TranslateFn, regime: MarketRegime['regime']) {
  return resolveText(t, `intelligence.regime.${regime}` as TranslationKey, regime)
}

export function resolveRegimeSummary(t: TranslateFn, regime: MarketRegime['regime']) {
  return resolveText(t, `intelligence.regimeSummary.${regime}` as TranslationKey, regime)
}

export function resolveTrendStrengthLabel(t: TranslateFn, strength: MarketRegime['trend_strength'] | string) {
  return resolveText(t, `intelligence.trendStrength.${strength}` as TranslationKey, strength)
}

export function resolveTrendStateLabel(t: TranslateFn, state: AssetBrief['technicals']['trend_state']) {
  return resolveText(t, `intelligence.trendState.${state}` as TranslationKey, state)
}

export function resolveActionBiasLabel(t: TranslateFn, actionBias: AssetBrief['action_bias']) {
  return resolveText(t, `intelligence.actionBias.${actionBias}` as TranslationKey, actionBias)
}

export function resolveSummarySignalLabel(t: TranslateFn, summarySignal: AssetBrief['summary_signal']) {
  return resolveText(t, `intelligence.summarySignal.${summarySignal}` as TranslationKey, summarySignal)
}

export function resolveSemanticLabel(t: TranslateFn, kind: string) {
  return resolveText(t, `intelligence.semantic.${kind}` as TranslationKey, kind)
}

export function resolveDriverLabel(t: TranslateFn, kind: string) {
  return resolveText(t, `intelligence.driver.${kind}` as TranslationKey, kind)
}

export function resolvePortfolioRoleLabel(t: TranslateFn, role: AssetBrief['portfolio_fit']['role']) {
  return resolveText(t, `intelligence.role.${role}` as TranslationKey, role)
}

export function resolveConcentrationImpactLabel(
  t: TranslateFn,
  impact: AssetBrief['portfolio_fit']['concentration_impact']
) {
  return resolveText(t, `intelligence.concentration.${impact}` as TranslationKey, impact)
}

export function resolveRiskFlagLabel(t: TranslateFn, riskFlag: AssetBrief['portfolio_fit']['risk_flag']) {
  return resolveText(t, `intelligence.riskFlag.${riskFlag}` as TranslationKey, riskFlag)
}

export function resolveEntryBasisLabel(t: TranslateFn, basis: AssetBrief['entry_zone']['basis']) {
  return resolveText(t, `intelligence.entryBasis.${basis}` as TranslationKey, basis)
}

export function resolveInvalidationReasonLabel(t: TranslateFn, reason: AssetBrief['invalidation']['reason']) {
  return resolveText(t, `intelligence.invalidation.${reason}` as TranslationKey, reason)
}

export function describeRegimeDriver(t: TranslateFn, driver: MarketRegime['drivers'][number]) {
  switch (driver.kind) {
    case 'trend_breadth':
      return t('intelligence.driverValue.trend_breadth', {
        up: driver.up_count ?? 0,
        total: driver.total_count ?? 0,
      })
    case 'alpha_30d':
      return t('intelligence.driverValue.alpha_30d', {
        value: formatSignedPercent(driver.value ?? 0),
      })
    case 'volatility':
      return t('intelligence.driverValue.volatility', {
        value: formatPercent(driver.value ?? 0, 0),
      })
    case 'concentration':
      return t('intelligence.driverValue.concentration', {
        value: formatPercent(driver.value ?? 0, 0),
      })
    case 'cash_buffer':
      return t('intelligence.driverValue.cash_buffer', {
        value: formatPercent(driver.value ?? 0, 0),
      })
    case 'correlation':
      return t('intelligence.driverValue.correlation', {
        value: formatNumber(driver.value ?? 0, 2),
      })
    default:
      return driver.value_text
  }
}

export function chartIntervalForAsset(assetType: string): '4h' | '1d' {
  return assetType === 'crypto' ? '4h' : '1d'
}

export function buildChartDateRange(interval: '4h' | '1d') {
  const end = new Date()
  const start = new Date()
  start.setUTCDate(start.getUTCDate() - (interval === '4h' ? 21 : 120))
  const formatDate = (value: Date) => value.toISOString().slice(0, 10)
  return {
    start: formatDate(start),
    end: formatDate(end),
  }
}

export function shouldShowAssetBriefLoading(params: {
  authLoading: boolean
  hasAccessToken: boolean
  hasAssetKey: boolean
  isFetched: boolean
  isFetching: boolean
  isPending: boolean
}) {
  if (!params.hasAssetKey) return false
  if (params.authLoading) return true
  if (!params.hasAccessToken) return true
  if (params.isFetching) return true
  if (params.isPending) return true
  return !params.isFetched
}
