import { translate } from '../utils/i18n'
import {
  describeRegimeDriver,
  formatSignedPercent,
  resolveActionBiasLabel,
  resolveRegimeLabel,
  resolveSummarySignalLabel,
  shouldShowAssetBriefLoading,
  toneForActionBias,
  toneForRegime,
} from '../utils/intelligence'

describe('intelligence helpers', () => {
  const t = (key: Parameters<typeof translate>[1], params?: Record<string, string | number>) =>
    translate('en', key, params)

  it('formats signed percent values', () => {
    expect(formatSignedPercent(0.123)).toBe('+12.3%')
    expect(formatSignedPercent(-0.045)).toBe('-4.5%')
  })

  it('maps regime and action labels', () => {
    expect(resolveRegimeLabel(t, 'risk_on')).toBe('Risk-On')
    expect(resolveActionBiasLabel(t, 'accumulate')).toBe('Accumulate')
    expect(resolveSummarySignalLabel(t, 'trend_up_pullback')).toContain('workable entry zone')
  })

  it('maps badge tones for regime and action bias', () => {
    expect(toneForRegime('risk_on')).toBe('low')
    expect(toneForRegime('risk_off')).toBe('high')
    expect(toneForActionBias('accumulate')).toBe('low')
    expect(toneForActionBias('reduce')).toBe('high')
  })

  it('renders driver copy from semantic fields', () => {
    expect(
      describeRegimeDriver(t, {
        id: 'trend_breadth',
        kind: 'trend_breadth',
        tone: 'positive',
        value_text: '',
        up_count: 3,
        down_count: 1,
        total_count: 4,
      })
    ).toBe('3 of 4 held assets still trend up')

    expect(
      describeRegimeDriver(t, {
        id: 'alpha_30d',
        kind: 'alpha_30d',
        tone: 'positive',
        value_text: '',
        value: 0.087,
      })
    ).toContain('+8.7%')
  })

  it('keeps asset brief in loading state until auth and first fetch settle', () => {
    expect(
      shouldShowAssetBriefLoading({
        authLoading: true,
        hasAccessToken: false,
        hasAssetKey: true,
        isFetched: false,
        isFetching: false,
        isPending: true,
      })
    ).toBe(true)

    expect(
      shouldShowAssetBriefLoading({
        authLoading: false,
        hasAccessToken: true,
        hasAssetKey: true,
        isFetched: false,
        isFetching: false,
        isPending: true,
      })
    ).toBe(true)

    expect(
      shouldShowAssetBriefLoading({
        authLoading: false,
        hasAccessToken: true,
        hasAssetKey: true,
        isFetched: true,
        isFetching: false,
        isPending: false,
      })
    ).toBe(false)
  })
})
