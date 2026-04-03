import { useOnboardingStore } from '../stores/onboarding'
import { allocationLabelKey } from '../utils/allocations'
import { healthStatusFromScore, healthStatusLabelKey, toneForHealthStatus, toneForSeverity } from '../utils/badges'
import { formatCurrency, formatNumber, formatPercent, formatPortfolioTotal } from '../utils/format'
import {
  isPaidReport,
  reportStatusLabelKey,
  reportTierLabelKey,
  resolveReportRiskTypeLabel,
  resolveReportSeverityLabel,
} from '../utils/report'

describe('format helpers', () => {
  it('formats currency with symbol', () => {
    expect(formatCurrency(1234.5, 'USD')).toContain('$')
  })

  it('formats non-ISO quote units without crashing', () => {
    expect(formatCurrency(1, 'USDT')).toBe('1 USDT')
    expect(formatCurrency(1234.5, 'USDT')).toBe('1,234.5 USDT')
  })

  it('formats portfolio totals with K/M/B suffixes and two decimals', () => {
    expect(formatPortfolioTotal(999.5, 'USD')).toBe('$999.50')
    expect(formatPortfolioTotal(12_345.67, 'USD')).toBe('$12.35K')
    expect(formatPortfolioTotal(12_345_678, 'USD')).toBe('$12.35M')
    expect(formatPortfolioTotal(1_234_567_890, 'USD')).toBe('$1.23B')
    expect(formatPortfolioTotal(5_225.26, 'USDT')).toBe('5.23K USDT')
  })

  it('formats number with max decimals', () => {
    expect(formatNumber(1234.567, 2)).toBe('1,234.57')
  })

  it('formats percent values', () => {
    expect(formatPercent(0.123)).toContain('%')
  })
})

describe('onboarding store', () => {
  it('resets state', () => {
    const store = useOnboardingStore.getState()
    store.startRetake({ markets: ['Crypto'], experience: 'Beginner' })
    store.reset()
    const next = useOnboardingStore.getState()
    expect(next.mode).toBe('onboarding')
    expect(next.markets).toEqual([])
    expect(next.experience).toBe('')
  })
})

describe('allocation helpers', () => {
  it('maps allocation labels to translation keys', () => {
    expect(allocationLabelKey('crypto')).toBe('allocation.crypto')
    expect(allocationLabelKey('stock')).toBe('allocation.stock')
    expect(allocationLabelKey('cash')).toBe('allocation.cash')
    expect(allocationLabelKey('manual')).toBe('allocation.manual')
    expect(allocationLabelKey('other')).toBe('allocation.other')
  })
})

describe('badge helpers', () => {
  it('maps severity to badge tone', () => {
    expect(toneForSeverity('critical')).toBe('critical')
    expect(toneForSeverity('high')).toBe('high')
    expect(toneForSeverity('medium')).toBe('medium')
    expect(toneForSeverity('low')).toBe('low')
    expect(toneForSeverity('unknown')).toBe('neutral')
  })

  it('maps health status to badge tone + label', () => {
    expect(toneForHealthStatus('critical')).toBe('critical')
    expect(toneForHealthStatus('warning')).toBe('high')
    expect(toneForHealthStatus('stable')).toBe('medium')
    expect(toneForHealthStatus('excellent')).toBe('low')
    expect(toneForHealthStatus('other')).toBe('neutral')
    expect(healthStatusFromScore(20)).toBe('critical')
    expect(healthStatusFromScore(65)).toBe('warning')
    expect(healthStatusFromScore(85)).toBe('stable')
    expect(healthStatusFromScore(95)).toBe('excellent')
    expect(healthStatusFromScore(undefined)).toBeNull()
    expect(healthStatusLabelKey('critical')).toBe('healthStatus.critical')
    expect(healthStatusLabelKey('warning')).toBe('healthStatus.warning')
    expect(healthStatusLabelKey('stable')).toBe('healthStatus.stable')
    expect(healthStatusLabelKey('excellent')).toBe('healthStatus.excellent')
    expect(healthStatusLabelKey('other')).toBe('healthStatus.unknown')
  })
})

describe('report helpers', () => {
  it('detects paid report payloads', () => {
    const preview = {
      meta_data: { calculation_id: 'calc_1' },
      valuation_as_of: '2026-01-01T00:00:00Z',
      market_data_snapshot_id: 'snap_1',
      fixed_metrics: { net_worth_usd: 1000, health_score: 50, health_status: 'Stable', volatility_score: 20 },
      identified_risks: [],
      locked_projection: { potential_upside: 'x', cta: 'y' },
    } as any
    const paid = {
      meta_data: { calculation_id: 'calc_2' },
      valuation_as_of: '2026-01-01T00:00:00Z',
      market_data_snapshot_id: 'snap_2',
      report_header: {
        health_score: { value: 80, status: 'Green' },
        volatility_dashboard: { value: 20, status: 'Green' },
      },
      charts: { radar_chart: { liquidity: 10, diversification: 20, alpha: 30, drawdown: 40 } },
      risk_insights: [],
    } as any

    expect(isPaidReport(preview)).toBe(false)
    expect(isPaidReport(paid)).toBe(true)
  })

  it('maps report tier and status labels', () => {
    expect(reportTierLabelKey('paid')).toBe('me.reportTier.paid')
    expect(reportTierLabelKey('preview')).toBe('me.reportTier.preview')
    expect(reportTierLabelKey('other')).toBe('me.reportTier.unknown')
    expect(reportStatusLabelKey('ready')).toBe('me.reportStatus.ready')
    expect(reportStatusLabelKey('processing')).toBe('me.reportStatus.processing')
    expect(reportStatusLabelKey('queued')).toBe('me.reportStatus.processing')
    expect(reportStatusLabelKey('failed')).toBe('me.reportStatus.failed')
    expect(reportStatusLabelKey('unknown')).toBe('me.reportStatus.unknown')
  })

  it('resolves localized risk type labels from stable backend keys', () => {
    const t = (key: any) => {
      if (key === 'report.riskType.concentration_risk') return '集中度风险'
      return key
    }

    expect(resolveReportRiskTypeLabel(t as any, 'concentration_risk')).toBe('集中度风险')
    expect(resolveReportRiskTypeLabel(t as any, 'custom_risk')).toBe('custom_risk')
  })

  it('resolves localized severity labels for report risk badges', () => {
    const t = (key: any) => {
      if (key === 'insights.severity.high') return '高'
      return key
    }

    expect(resolveReportSeverityLabel(t as any, 'High')).toBe('高')
    expect(resolveReportSeverityLabel(t as any, 'Unknown')).toBe('Unknown')
  })
})
