import { useRouter } from 'expo-router'
import React, { useMemo } from 'react'
import { StyleSheet, Text, View } from 'react-native'

import { OnboardingChoiceRow } from '../../src/components/OnboardingChoiceRow'
import { OnboardingDecisionSheet } from '../../src/components/OnboardingDecisionSheet'
import { OnboardingStep } from '../../src/components/OnboardingStep'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { useOnboardingStore } from '../../src/stores/onboarding'
import { useResponsiveScale } from '../../src/utils/responsive'

const options = [
  {
    value: 'Crypto',
    labelKey: 'onboarding.sc02.option.crypto',
    descriptionKey: 'onboarding.sc02.option.crypto.desc',
  },
  {
    value: 'Stocks',
    labelKey: 'onboarding.sc02.option.stocks',
    descriptionKey: 'onboarding.sc02.option.stocks.desc',
  },
  {
    value: 'Forex',
    labelKey: 'onboarding.sc02.option.forex',
    descriptionKey: 'onboarding.sc02.option.forex.desc',
  },
] as const

export default function SC02() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { markets, setMarkets } = useOnboardingStore()
  const { compact, font } = useResponsiveScale()
  const titleSize = font(32, 26)
  const subtitleSize = font(15, 13)
  const sectionGap = compact ? theme.spacing.md : theme.spacing.lg
  const nextButtonStyle = useMemo(
    () => ({
      borderRadius: 999,
      minHeight: compact ? 52 : 56,
      shadowOpacity: 0.16,
      shadowRadius: 18,
      shadowOffset: { width: 0, height: 10 },
      elevation: 2,
    }),
    [compact]
  )
  const styles = useMemo(() => createStyles(theme, { titleSize, subtitleSize }), [subtitleSize, theme, titleSize])

  const toggle = (value: string) => {
    if (markets.includes(value)) {
      setMarkets(markets.filter((item) => item !== value))
      return
    }
    setMarkets([...markets, value])
  }

  return (
    <OnboardingStep
      nextLabel={t('common.next')}
      onBack={() => router.dismissTo('/(auth)/sc01b')}
      onNext={() => router.push('/(auth)/sc03')}
      nextDisabled={markets.length === 0}
      swipeEnabled
      progressTotal={5}
      progressIndex={0}
      contentScroll
      nextButtonStyle={nextButtonStyle}
    >
      <View style={[styles.container, { gap: sectionGap }]}>
        <View style={styles.header}>
          <Text style={styles.title}>{t('onboarding.sc02.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc02.subtitle')}</Text>
        </View>

        <OnboardingDecisionSheet>
          {options.map((option, index) => (
            <OnboardingChoiceRow
              key={option.value}
              label={t(option.labelKey)}
              description={t(option.descriptionKey)}
              selected={markets.includes(option.value)}
              onPress={() => toggle(option.value)}
              showDivider={index < options.length - 1}
              selectionMode="multiple"
            />
          ))}
        </OnboardingDecisionSheet>
      </View>
    </OnboardingStep>
  )
}

function createStyles(
  theme: ReturnType<typeof useTheme>,
  sizes: {
    titleSize: number
    subtitleSize: number
  }
) {
  return StyleSheet.create({
    container: {
      flex: 1,
    },
    header: {
      gap: theme.spacing.sm,
      paddingRight: theme.spacing.sm,
    },
    title: {
      fontSize: sizes.titleSize,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      letterSpacing: -0.2,
    },
    subtitle: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: sizes.subtitleSize,
      lineHeight: Math.round(sizes.subtitleSize * 1.45),
    },
  })
}
