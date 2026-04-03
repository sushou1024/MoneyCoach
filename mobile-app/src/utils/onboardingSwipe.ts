export const ONBOARDING_SWIPE_ACTIVATION_DX = 14
export const ONBOARDING_SWIPE_COMPLETION_DX = 72
export const ONBOARDING_SWIPE_VELOCITY = 0.35

const BLOCKED_DIRECTION_RESISTANCE = 0.18

export type OnboardingSwipeAction = 'back' | 'next'

interface ResolveOnboardingSwipeActionInput {
  dx: number
  vx: number
  canGoBack: boolean
  canGoForward: boolean
}

interface ApplyOnboardingSwipeResistanceInput {
  dx: number
  canGoBack: boolean
  canGoForward: boolean
}

export function resolveOnboardingSwipeAction({
  dx,
  vx,
  canGoBack,
  canGoForward,
}: ResolveOnboardingSwipeActionInput): OnboardingSwipeAction | null {
  if (canGoForward && dx <= -ONBOARDING_SWIPE_COMPLETION_DX) {
    return 'next'
  }
  if (canGoBack && dx >= ONBOARDING_SWIPE_COMPLETION_DX) {
    return 'back'
  }

  if (canGoForward && dx < 0 && vx <= -ONBOARDING_SWIPE_VELOCITY) {
    return 'next'
  }
  if (canGoBack && dx > 0 && vx >= ONBOARDING_SWIPE_VELOCITY) {
    return 'back'
  }

  return null
}

export function applyOnboardingSwipeResistance({
  dx,
  canGoBack,
  canGoForward,
}: ApplyOnboardingSwipeResistanceInput): number {
  if (dx > 0 && !canGoBack) {
    return dx * BLOCKED_DIRECTION_RESISTANCE
  }
  if (dx < 0 && !canGoForward) {
    return dx * BLOCKED_DIRECTION_RESISTANCE
  }
  return dx
}
