import { useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { Alert, StyleSheet, Text, View } from 'react-native'

import { OnboardingChoiceRow } from '../../src/components/OnboardingChoiceRow'
import { OnboardingDecisionSheet } from '../../src/components/OnboardingDecisionSheet'
import { OnboardingStep } from '../../src/components/OnboardingStep'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { updateProfile } from '../../src/services/profile'
import { useOnboardingStore } from '../../src/stores/onboarding'
import type { UserProfile } from '../../src/types/api'
import { useResponsiveScale } from '../../src/utils/responsive'

const options = [
  {
    value: 'Yield Seeker',
    labelKey: 'onboarding.sc06.option.yieldseeker',
    descriptionKey: 'onboarding.sc06.option.yieldseeker.desc',
  },
  {
    value: 'Speculator',
    labelKey: 'onboarding.sc06.option.speculator',
    descriptionKey: 'onboarding.sc06.option.speculator.desc',
  },
] as const

export default function SC06() {
  const theme = useTheme()
  const router = useRouter()
  const queryClient = useQueryClient()
  const { t } = useLocalization()
  const { accessToken, userId } = useAuth()
  const { mode, markets, experience, style, painPoints, riskPreference, setRiskPreference, reset } =
    useOnboardingStore()
  const isRetake = mode === 'retake'
  const [saving, setSaving] = useState(false)
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

  const handleNext = async () => {
    if (!isRetake) {
      router.push('/(auth)/sc07')
      return
    }
    if (!accessToken || !userId || saving) {
      return
    }

    setSaving(true)
    const updates = {
      markets,
      experience,
      style,
      pain_points: painPoints,
      risk_preference: riskPreference,
    }

    const resp = await updateProfile(accessToken, updates)
    if (resp.error) {
      Alert.alert(t('settings.saveFailTitle'), resp.error.message)
      setSaving(false)
      return
    }

    queryClient.setQueryData<UserProfile | null>(['profile', userId], (prev) => (prev ? { ...prev, ...updates } : prev))
    queryClient.invalidateQueries({ queryKey: ['profile', userId] })

    Alert.alert(t('settings.saveSuccessTitle'), t('settings.saveSuccessBody'))
    router.replace('/(modals)/settings')
    requestAnimationFrame(() => {
      reset()
    })
    setSaving(false)
  }

  return (
    <OnboardingStep
      nextLabel={isRetake ? t('common.save') : t('common.next')}
      onBack={() => router.dismissTo('/(auth)/sc05')}
      onNext={handleNext}
      nextDisabled={!riskPreference || saving}
      swipeEnabled
      progressTotal={5}
      progressIndex={4}
      contentScroll
      nextButtonStyle={nextButtonStyle}
    >
      <View style={[styles.container, { gap: sectionGap }]}>
        <View style={styles.header}>
          <Text style={styles.title}>{t('onboarding.sc06.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc06.subtitle')}</Text>
        </View>

        <OnboardingDecisionSheet>
          {options.map((option, index) => (
            <OnboardingChoiceRow
              key={option.value}
              label={t(option.labelKey)}
              description={t(option.descriptionKey)}
              selected={riskPreference === option.value}
              onPress={() => setRiskPreference(option.value)}
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
