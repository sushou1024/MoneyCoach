import { apiRequest } from './api'

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user_id: string
}

export async function oauthLogin(provider: 'google' | 'apple', idToken: string) {
  return apiRequest<AuthResponse>('/v1/auth/oauth', {
    method: 'POST',
    body: { provider, id_token: idToken },
  })
}

export async function startEmailRegistration(email: string) {
  return apiRequest<{ sent: boolean; code?: string }>('/v1/auth/email/register/start', {
    method: 'POST',
    body: { email },
  })
}

export async function registerWithEmail(email: string, password: string, code: string) {
  return apiRequest<AuthResponse>('/v1/auth/email/register', {
    method: 'POST',
    body: { email, password, code },
  })
}

export async function loginWithEmail(email: string, password: string) {
  return apiRequest<AuthResponse>('/v1/auth/email/login', {
    method: 'POST',
    body: { email, password },
  })
}

export async function refreshSession(refreshToken: string) {
  return apiRequest<{ access_token: string }>('/v1/auth/refresh', {
    method: 'POST',
    body: { refresh_token: refreshToken },
  })
}

export async function logout(refreshToken: string) {
  return apiRequest<{ revoked: boolean }>('/v1/auth/logout', {
    method: 'POST',
    body: { refresh_token: refreshToken },
  })
}
