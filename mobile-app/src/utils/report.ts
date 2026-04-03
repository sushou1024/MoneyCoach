import { TranslationKey } from './i18n'
import { PaidReport, PreviewReport } from '../types/api'

type TranslateFn = (key: TranslationKey, params?: Record<string, string | number>) => string

export function isPaidReport(report: PaidReport | PreviewReport | null | undefined): report is PaidReport {
  if (!report) return false
  const paid = (report as PaidReport).report_header?.health_score?.value !== undefined
  const hasRadar = (report as PaidReport).charts?.radar_chart !== undefined
  const hasRisks = Array.isArray((report as PaidReport).risk_insights)
  return paid && (hasRadar || hasRisks)
}

export function reportTierLabelKey(tier?: string | null) {
  switch ((tier ?? '').toLowerCase()) {
    case 'paid':
      return 'me.reportTier.paid'
    case 'preview':
      return 'me.reportTier.preview'
    default:
      return 'me.reportTier.unknown'
  }
}

export function reportStatusLabelKey(status?: string | null) {
  switch ((status ?? '').toLowerCase()) {
    case 'ready':
      return 'me.reportStatus.ready'
    case 'processing':
    case 'queued':
      return 'me.reportStatus.processing'
    case 'failed':
      return 'me.reportStatus.failed'
    default:
      return 'me.reportStatus.unknown'
  }
}

export function resolveReportRiskTypeLabel(t: TranslateFn, type: string) {
  const normalized = type.toLowerCase().trim().replace(/\s+/g, '_')
  const key = `report.riskType.${normalized}` as TranslationKey
  const translated = t(key)
  return translated === key ? type : translated
}

export function resolveReportSeverityLabel(t: TranslateFn, severity?: string | null) {
  const normalized = (severity ?? '').toLowerCase().trim()
  const key = `insights.severity.${normalized || 'unknown'}` as TranslationKey
  const translated = t(key)
  return translated === key ? (severity ?? '') : translated
}
