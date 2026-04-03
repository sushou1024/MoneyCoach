import { apiRequest } from './api'
import { PortfolioSnapshot } from '../types/api'

export async function fetchActivePortfolio(token: string) {
  return apiRequest<PortfolioSnapshot>('/v1/portfolio/active', { token })
}

export async function refreshActivePortfolio(token: string) {
  return apiRequest<PortfolioSnapshot>('/v1/portfolio/active/refresh', {
    method: 'POST',
    token,
  })
}

export async function fetchPortfolioSnapshots(token: string) {
  return apiRequest<{
    items: {
      calculation_id: string
      report_tier: string
      status: string
      health_score: number
      created_at: string
      is_active: boolean
    }[]
  }>('/v1/reports', { token })
}
