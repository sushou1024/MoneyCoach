import { apiRequest } from './api'
import { PaidReport, PreviewReport, ReportPlan } from '../types/api'

export async function fetchPreviewReport(token: string, calculationId: string) {
  return apiRequest<PreviewReport>(`/v1/reports/preview/${calculationId}`, { token })
}

export async function requestPaidReport(token: string, calculationId: string) {
  return apiRequest<{ calculation_id: string; status: string }>(`/v1/reports/${calculationId}/paid`, {
    method: 'POST',
    token,
  })
}

export async function requestActiveReport(token: string, tier: 'preview' | 'paid') {
  return apiRequest<{ calculation_id: string; status: string }>(`/v1/reports/active`, {
    method: 'POST',
    token,
    body: { tier },
  })
}

export async function fetchReport(token: string, calculationId: string) {
  return apiRequest<PaidReport | PreviewReport>(`/v1/reports/${calculationId}`, { token })
}

export async function fetchReportPlan(token: string, calculationId: string, planId: string) {
  return apiRequest<ReportPlan>(`/v1/reports/${calculationId}/plans/${planId}`, { token })
}
