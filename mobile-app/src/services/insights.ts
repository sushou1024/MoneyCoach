import { apiRequest } from './api'
import { InsightItem } from '../types/api'

export async function fetchInsights(token: string, filter: string) {
  const query = filter && filter !== 'all' ? `?filter=${encodeURIComponent(filter)}` : ''
  return apiRequest<{ items: InsightItem[]; next_cursor?: string }>(`/v1/insights${query}`, { token })
}

export async function executeInsight(
  token: string,
  insightId: string,
  payload: {
    method: 'suggested' | 'manual' | 'trade_slip'
    quantity?: number
    quantity_unit?: string
    transaction_ids?: string[]
  }
) {
  return apiRequest<{
    executed: boolean
    transaction_ids?: string[]
    portfolio_snapshot_id?: string
    warnings?: string[]
  }>(`/v1/insights/${insightId}/execute`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function dismissInsight(token: string, insightId: string, reason: string) {
  return apiRequest<{ dismissed: boolean }>(`/v1/insights/${insightId}/dismiss`, {
    method: 'POST',
    token,
    body: { reason },
  })
}
