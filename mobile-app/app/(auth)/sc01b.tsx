import { useRouter } from 'expo-router'
import React, { useMemo } from 'react'
import { DimensionValue, StyleSheet, Text, View } from 'react-native'

import { OnboardingStep } from '../../src/components/OnboardingStep'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { useResponsiveScale } from '../../src/utils/responsive'

export default function SC01b() {
  const router = useRouter()
  const theme = useTheme()
  const { t } = useLocalization()
  const { compact, font } = useResponsiveScale()
  const titleSize = font(32, 26)
  const subtitleSize = font(15, 13)
  const boardTitleSize = font(20, 17)
  const sectionGap = compact ? theme.spacing.md : theme.spacing.lg
  const boardRadius = compact ? theme.radius.sm : theme.radius.md
  const meterHeight = compact ? 10 : 12
  const styles = useMemo(
    () =>
      createStyles(theme, {
        boardRadius,
        meterHeight,
        boardTitleSize,
        titleSize,
        subtitleSize,
      }),
    [boardRadius, boardTitleSize, meterHeight, subtitleSize, theme, titleSize]
  )

  const percent = (value: number): DimensionValue => `${value}%`

  const planRows = [
    {
      id: 'risk',
      label: t('onboarding.sc01b.planRiskLabel'),
      value: t('onboarding.sc01b.planRiskValue'),
      fill: 68,
      color: theme.colors.danger,
    },
    {
      id: 'size',
      label: t('onboarding.sc01b.planSizeLabel'),
      value: t('onboarding.sc01b.planSizeValue'),
      fill: 46,
      color: theme.colors.accent,
    },
    {
      id: 'cadence',
      label: t('onboarding.sc01b.planCadenceLabel'),
      value: t('onboarding.sc01b.planCadenceValue'),
      fill: 80,
      color: theme.colors.success,
    },
  ]

  return (
    <OnboardingStep
      nextLabel={t('onboarding.sc01b.cta')}
      onBack={() => router.dismissTo('/(auth)/sc01')}
      onNext={() => router.push('/(auth)/sc02')}
      swipeEnabled
      progressTotal={3}
      progressIndex={2}
      contentScroll
      nextButtonStyle={{ borderRadius: theme.radius.md }}
    >
      <View style={{ gap: sectionGap }}>
        <View style={styles.copyBlock}>
          <Text style={styles.title}>{t('onboarding.sc01b.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc01b.subtitle')}</Text>
        </View>

        <View style={styles.board}>
          <View style={styles.boardContent}>
            <Text style={styles.boardTitle}>{t('onboarding.sc01b.planTitle')}</Text>

            <View style={styles.planTable}>
              {planRows.map((row, index) => (
                <View
                  key={row.id}
                  style={[styles.planRow, index < planRows.length - 1 ? styles.planDivider : undefined]}
                >
                  <View style={styles.planRowHeader}>
                    <Text style={styles.planLabel}>{row.label}</Text>
                    <Text style={styles.planValue}>{row.value}</Text>
                  </View>
                  <View style={styles.meterTrack}>
                    <View style={[styles.meterFill, { width: percent(row.fill), backgroundColor: row.color }]} />
                  </View>
                </View>
              ))}
              <View style={styles.principleBand}>
                <Text style={styles.principleText}>{t('onboarding.sc01b.principle')}</Text>
              </View>
            </View>
          </View>
        </View>
      </View>
    </OnboardingStep>
  )
}

function createStyles(
  theme: ReturnType<typeof useTheme>,
  sizes: {
    boardRadius: number
    meterHeight: number
    boardTitleSize: number
    titleSize: number
    subtitleSize: number
  }
) {
  const divider = {
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: theme.colors.border,
  } as const

  return StyleSheet.create({
    copyBlock: {
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
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: sizes.subtitleSize,
      lineHeight: Math.round(sizes.subtitleSize * 1.45),
    },
    board: {
      flexDirection: 'row',
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surface,
      borderRadius: sizes.boardRadius,
      paddingVertical: theme.spacing.md,
      paddingRight: theme.spacing.md,
      paddingLeft: theme.spacing.md,
      shadowColor: '#000',
      shadowOpacity: 0.04,
      shadowRadius: 14,
      shadowOffset: { width: 0, height: 8 },
      elevation: 1,
      overflow: 'hidden',
    },
    boardContent: {
      flex: 1,
      gap: theme.spacing.md,
      paddingTop: theme.spacing.xs,
    },
    boardTitle: {
      fontSize: sizes.boardTitleSize,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    planTable: {
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: Math.max(8, sizes.boardRadius - 4),
      backgroundColor: theme.colors.surfaceElevated,
      overflow: 'hidden',
    },
    planRow: {
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.md,
      gap: theme.spacing.sm,
      backgroundColor: theme.colors.surface,
    },
    planDivider: divider,
    planRowHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.md,
    },
    planLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 13,
      letterSpacing: 0.2,
    },
    planValue: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 15,
      letterSpacing: 0.2,
    },
    meterTrack: {
      height: sizes.meterHeight,
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: 6,
      overflow: 'hidden',
      backgroundColor: theme.colors.surface,
    },
    meterFill: {
      height: '100%',
      opacity: 0.3,
    },
    principleBand: {
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.md,
      borderTopWidth: StyleSheet.hairlineWidth,
      borderTopColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
    },
    principleText: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 13,
      lineHeight: 19,
    },
  })
}
