import { applyOnboardingSwipeResistance, resolveOnboardingSwipeAction } from '../utils/onboardingSwipe'

describe('onboardingSwipe', () => {
  it('advances on a committed left swipe when forward navigation is allowed', () => {
    expect(
      resolveOnboardingSwipeAction({
        dx: -96,
        vx: -0.12,
        canGoBack: true,
        canGoForward: true,
      })
    ).toBe('next')
  })

  it('does not advance on a left swipe when the current step is still invalid', () => {
    expect(
      resolveOnboardingSwipeAction({
        dx: -120,
        vx: -0.44,
        canGoBack: true,
        canGoForward: false,
      })
    ).toBeNull()
  })

  it('returns to the previous step on a committed right swipe', () => {
    expect(
      resolveOnboardingSwipeAction({
        dx: 88,
        vx: 0.08,
        canGoBack: true,
        canGoForward: true,
      })
    ).toBe('back')
  })

  it('dampens drag distance when swiping into a blocked direction', () => {
    expect(
      applyOnboardingSwipeResistance({
        dx: 100,
        canGoBack: false,
        canGoForward: true,
      })
    ).toBeCloseTo(18)
    expect(
      applyOnboardingSwipeResistance({
        dx: -100,
        canGoBack: true,
        canGoForward: false,
      })
    ).toBeCloseTo(-18)
  })
})
