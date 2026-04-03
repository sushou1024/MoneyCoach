export type BadgeTone = 'critical' | 'high' | 'medium' | 'low' | 'neutral'

export function toneForSeverity(severity?: string | null): BadgeTone {
  switch ((severity ?? '').toLowerCase()) {
    case 'critical':
      return 'critical'
    case 'high':
      return 'high'
    case 'medium':
      return 'medium'
    case 'low':
      return 'low'
    default:
      return 'neutral'
  }
}

export function toneForHealthStatus(status?: string | null): BadgeTone {
  switch ((status ?? '').toLowerCase()) {
    case 'critical':
      return 'critical'
    case 'warning':
      return 'high'
    case 'stable':
      return 'medium'
    case 'excellent':
      return 'low'
    default:
      return 'neutral'
  }
}

export function healthStatusFromScore(score?: number | null) {
  if (score === undefined || score === null) return null
  if (score < 50) return 'critical'
  if (score < 70) return 'warning'
  if (score < 90) return 'stable'
  return 'excellent'
}

export function healthStatusLabelKey(status?: string | null) {
  switch ((status ?? '').toLowerCase()) {
    case 'critical':
      return 'healthStatus.critical'
    case 'warning':
      return 'healthStatus.warning'
    case 'stable':
      return 'healthStatus.stable'
    case 'excellent':
      return 'healthStatus.excellent'
    default:
      return 'healthStatus.unknown'
  }
}
