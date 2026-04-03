import { useRouter } from 'expo-router'
import React, { useMemo } from 'react'
import { StyleSheet, Text, View } from 'react-native'

import { OnboardingStep } from '../../src/components/OnboardingStep'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { useResponsiveScale } from '../../src/utils/responsive'

export default function SC01() {
  const router = useRouter()
  const theme = useTheme()
  const { t } = useLocalization()
  const { compact, font } = useResponsiveScale()
  const titleSize = font(32, 26)
  const subtitleSize = font(15, 13)
  const boardTitleSize = font(20, 17)
  const sectionGap = compact ? theme.spacing.md : theme.spacing.lg
  const boardRadius = compact ? theme.radius.sm : theme.radius.md
  const styles = useMemo(
    () =>
      createStyles(theme, {
        boardRadius,
        boardTitleSize,
        titleSize,
        subtitleSize,
      }),
    [boardRadius, boardTitleSize, subtitleSize, theme, titleSize]
  )

  const signals = [
    {
      id: 'risk',
      time: '10:42',
      type: t('onboarding.sc01.signalRiskLabel'),
      text: t('onboarding.sc01.alert1'),
      hint: t('onboarding.sc01.signalRiskHint'),
      color: theme.colors.danger,
    },
    {
      id: 'entry',
      time: '11:18',
      type: t('onboarding.sc01.signalEntryLabel'),
      text: t('onboarding.sc01.alert2'),
      hint: t('onboarding.sc01.signalEntryHint'),
      color: theme.colors.accent,
    },
  ]

  return (
    <OnboardingStep
      nextLabel={t('common.next')}
      onBack={() => router.dismissTo('/(auth)/sc00')}
      onNext={() => router.push('/(auth)/sc01b')}
      swipeEnabled
      progressTotal={3}
      progressIndex={1}
      contentScroll
      nextButtonStyle={{ borderRadius: theme.radius.md }}
    >
      <View style={{ gap: sectionGap }}>
        <View style={styles.copyBlock}>
          <Text style={styles.title}>{t('onboarding.sc01.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc01.subtitle')}</Text>
        </View>

        <View style={styles.board}>
          <View style={styles.boardContent}>
            <Text style={styles.boardTitle}>{t('onboarding.sc01.boardTitle')}</Text>

            <View style={styles.timeline}>
              {signals.map((signal, index) => (
                <View
                  key={signal.id}
                  style={[styles.timelineRow, index < signals.length - 1 ? styles.timelineDivider : undefined]}
                >
                  <View style={styles.metaColumn}>
                    <Text style={styles.metaTime}>{signal.time}</Text>
                    <View style={styles.metaTypeRow}>
                      <View style={[styles.metaSwatch, { backgroundColor: signal.color }]} />
                      <Text style={styles.metaType}>{signal.type}</Text>
                    </View>
                  </View>
                  <View style={styles.signalColumn}>
                    <Text style={styles.signalText}>{signal.text}</Text>
                    <Text style={styles.signalHint}>{signal.hint}</Text>
                  </View>
                </View>
              ))}
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
    timeline: {
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: Math.max(8, sizes.boardRadius - 4),
      backgroundColor: theme.colors.surfaceElevated,
      overflow: 'hidden',
    },
    timelineRow: {
      flexDirection: 'row',
      gap: theme.spacing.md,
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.md,
      backgroundColor: theme.colors.surface,
    },
    timelineDivider: divider,
    metaColumn: {
      width: 84,
      gap: theme.spacing.xs,
    },
    metaTime: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 15,
      letterSpacing: 0.2,
    },
    metaTypeRow: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.xs,
    },
    metaSwatch: {
      width: 8,
      height: 8,
      borderRadius: 2,
    },
    metaType: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 11,
      letterSpacing: 0.6,
      textTransform: 'uppercase',
    },
    signalColumn: {
      flex: 1,
      gap: theme.spacing.xs,
      paddingRight: theme.spacing.xs,
    },
    signalText: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
      lineHeight: 20,
    },
    signalHint: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
  })
}
