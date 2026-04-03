import { getCurrentLocale } from './i18n'
import { fetchEntitlement } from '../services/billing'
import { fetchProfile, updateProfile } from '../services/profile'

interface OnboardingSnapshot {
  markets: string[]
  experience: string
  style: string
  painPoints: string[]
  riskPreference: string
  reset: () => void
}

export async function finalizeAuthFlow(options: {
  accessToken: string
  refreshToken: string
  userId: string
  setSession: (accessToken: string, refreshToken: string, userId: string) => Promise<void>
  onboarding: OnboardingSnapshot
}) {
  const { accessToken, refreshToken, userId, setSession, onboarding } = options

  await setSession(accessToken, refreshToken, userId)

  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const locale = getCurrentLocale()
  const profileResp = await fetchProfile(accessToken)

  const updates: Record<string, any> = {}
  if (onboarding.markets.length > 0) updates.markets = onboarding.markets
  if (onboarding.experience) updates.experience = onboarding.experience
  if (onboarding.style) updates.style = onboarding.style
  if (onboarding.painPoints.length > 0) updates.pain_points = onboarding.painPoints
  if (onboarding.riskPreference) updates.risk_preference = onboarding.riskPreference
  if (!profileResp.data?.timezone) updates.timezone = timezone
  if (!profileResp.data?.language) updates.language = locale

  if (Object.keys(updates).length > 0) {
    await updateProfile(accessToken, updates)
  }

  onboarding.reset()

  const entitlementResp = await fetchEntitlement(accessToken)
  if (entitlementResp.data && (entitlementResp.data.status === 'active' || entitlementResp.data.status === 'grace')) {
    return '/(tabs)/insights'
  }
  return '/(tabs)/assets'
}
