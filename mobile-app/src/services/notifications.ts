import Constants from 'expo-constants'
import * as Localization from 'expo-localization'
import * as Notifications from 'expo-notifications'
import { Platform } from 'react-native'

import { registerDevice } from './devices'

export type PushPermissionStatus = 'granted' | 'denied' | 'undetermined' | 'unknown'

export interface PushPermissionState {
  status: PushPermissionStatus
  canAskAgain?: boolean
}

export async function getPushPermission(): Promise<PushPermissionState> {
  try {
    const permissions = await Notifications.getPermissionsAsync()
    const status = (permissions.status ?? 'unknown') as PushPermissionStatus
    return { status, canAskAgain: permissions.canAskAgain }
  } catch {
    return { status: 'unknown' }
  }
}

export async function registerForPush(token: string) {
  try {
    const permissions = await Notifications.requestPermissionsAsync()
    const status = (permissions.status ?? 'unknown') as PushPermissionStatus
    if (status !== 'granted') {
      return { registered: false, permission: { status, canAskAgain: permissions.canAskAgain } as PushPermissionState }
    }

    if (Platform.OS === 'web') {
      // Web pushes are handled by the browser; skip native device registration.
      return {
        registered: true,
        permission: { status: 'granted', canAskAgain: permissions.canAskAgain } as PushPermissionState,
      }
    }

    const deviceToken = await Notifications.getDevicePushTokenAsync()
    const pushProvider = Platform.OS === 'ios' ? 'apns' : 'fcm'

    const locale = Localization.getLocales?.()[0]?.languageTag ?? 'en'
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
    const appVersion = Constants.expoConfig?.version ?? '0.1.0'
    const osVersion = typeof Platform.Version === 'string' ? Platform.Version : String(Platform.Version)

    const resp = await registerDevice(token, {
      platform: Platform.OS === 'ios' ? 'ios' : 'android',
      push_provider: pushProvider,
      device_token: deviceToken.data,
      client_device_id: `${Platform.OS}-${deviceToken.data.slice(0, 10)}`,
      app_version: appVersion,
      os_version: osVersion,
      locale,
      timezone,
      push_enabled: true,
      environment: __DEV__ ? 'sandbox' : 'production',
    })

    if (resp.error) {
      return {
        registered: false,
        permission: { status: 'granted', canAskAgain: permissions.canAskAgain } as PushPermissionState,
        error: resp.error.message,
      }
    }

    return {
      registered: true,
      permission: { status: 'granted', canAskAgain: permissions.canAskAgain } as PushPermissionState,
    }
  } catch (err) {
    return {
      registered: false,
      permission: { status: 'unknown' } as PushPermissionState,
      error: err instanceof Error ? err.message : 'Push registration failed',
    }
  }
}
