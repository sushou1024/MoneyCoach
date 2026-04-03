import { apiRequest } from './api'

export async function joinWaitlist(token: string, payload: { strategy_id: string; calculation_id: string }) {
  return apiRequest<{ rank: number }>(`/v1/waitlist`, {
    method: 'POST',
    token,
    body: payload,
  })
}
