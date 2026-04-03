import { apiRequest } from './api'

export interface DeviceRegisterRequest {
  platform: 'ios' | 'android'
  push_provider: 'apns' | 'fcm'
  device_token: string
  client_device_id: string
  app_version: string
  os_version?: string
  locale: string
  timezone: string
  push_enabled?: boolean
  environment: 'production' | 'sandbox'
}

export async function registerDevice(token: string, payload: DeviceRegisterRequest) {
  return apiRequest<{ device_id: string; registered: boolean }>(`/v1/devices/register`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function deleteDevice(token: string, deviceId: string) {
  return apiRequest<{ revoked: boolean }>(`/v1/devices/${deviceId}`, {
    method: 'DELETE',
    token,
  })
}
