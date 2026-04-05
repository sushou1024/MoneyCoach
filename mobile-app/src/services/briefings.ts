import { apiRequest } from './api'

export interface BriefingItem {
  id: string
  type: string
  priority: number
  title: string
  body: string
  push_text: string
  created_at: string
}

export async function getTodayBriefings(token: string) {
  return apiRequest<{ briefings: BriefingItem[]; date: string }>('/v1/briefings/today', { token })
}
