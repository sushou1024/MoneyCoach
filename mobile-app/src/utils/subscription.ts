import Constants from 'expo-constants'
import { Linking, Platform } from 'react-native'

const IOS_MANAGE_URL = 'https://apps.apple.com/account/subscriptions'
const IOS_MANAGE_DEEP = 'itms-apps://apps.apple.com/account/subscriptions'
const ANDROID_MANAGE_BASE = 'https://play.google.com/store/account/subscriptions'

function getAndroidPackageName() {
  const config: any = Constants.expoConfig ?? (Constants as any).manifest
  return config?.android?.package as string | undefined
}

export function getManageSubscriptionUrl(productId?: string) {
  if (Platform.OS === 'ios') {
    return IOS_MANAGE_URL
  }
  if (Platform.OS === 'android') {
    const params: string[] = []
    const pkg = getAndroidPackageName()
    if (pkg) params.push(`package=${encodeURIComponent(pkg)}`)
    if (productId) params.push(`sku=${encodeURIComponent(productId)}`)
    return params.length ? `${ANDROID_MANAGE_BASE}?${params.join('&')}` : ANDROID_MANAGE_BASE
  }
  return null
}

export async function openManageSubscription(productId?: string) {
  const url = getManageSubscriptionUrl(productId)
  if (!url) {
    throw new Error('Manage subscription not supported')
  }
  if (Platform.OS === 'ios') {
    try {
      await Linking.openURL(IOS_MANAGE_DEEP)
      return
    } catch {
      // fall back to https
    }
  }
  await Linking.openURL(url)
}
