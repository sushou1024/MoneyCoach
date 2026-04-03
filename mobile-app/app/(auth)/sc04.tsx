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
    value: 'Scalping',
    labelKey: 'onboarding.sc04.option.scalping',
    descriptionKey: 'onboarding.sc04.option.scalping.desc',
  },
  {
    value: 'Day Trading',
    labelKey: 'onboarding.sc04.option.daytrading',
    descriptionKey: 'onboarding.sc04.option.daytrading.desc',
  },
  {
    value: 'Swing Trading',
    labelKey: 'onboarding.sc04.option.swingtrading',
    descriptionKey: 'onboarding.sc04.option.swingtrading.desc',
  },
  {
    value: 'Long-Term',
    labelKey: 'onboarding.sc04.option.long-term',
    descriptionKey: 'onboarding.sc04.option.long-term.desc',
  },
] as const

export default function SC04() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { style, setStyle } = useOnboardingStore()
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

  return (
    <OnboardingStep
      nextLabel={t('common.next')}
      onBack={() => router.dismissTo('/(auth)/sc03')}
      onNext={() => router.push('/(auth)/sc05')}
      nextDisabled={!style}
      swipeEnabled
      progressTotal={5}
      progressIndex={2}
      contentScroll
      nextButtonStyle={nextButtonStyle}
    >
      <View style={[styles.container, { gap: sectionGap }]}>
        <View style={styles.header}>
          <Text style={styles.title}>{t('onboarding.sc04.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc04.subtitle')}</Text>
        </View>

        <OnboardingDecisionSheet>
          {options.map((option, index) => (
            <OnboardingChoiceRow
              key={option.value}
              label={t(option.labelKey)}
              description={t(option.descriptionKey)}
              selected={style === option.value}
              onPress={() => setStyle(option.value)}
              showDivider={index < options.length - 1}
              selectionMode="single"
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
