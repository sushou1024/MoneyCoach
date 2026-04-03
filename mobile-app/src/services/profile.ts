import { apiRequest } from './api'
import { UserProfile } from '../types/api'

export async function fetchProfile(token: string) {
  return apiRequest<UserProfile>('/v1/users/me', { token })
}

export async function updateProfile(token: string, updates: Record<string, any>) {
  return apiRequest<{ updated: boolean }>('/v1/users/me', {
    method: 'PATCH',
    token,
    body: updates,
  })
}

export async function deleteAccount(token: string, confirmText: string) {
  return apiRequest<{ deleted: boolean }>('/v1/users/me', {
    method: 'DELETE',
    token,
    body: { confirm_text: confirmText },
  })
}
