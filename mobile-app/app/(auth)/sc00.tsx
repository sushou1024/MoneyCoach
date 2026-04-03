import { useRouter } from 'expo-router'
import React, { useMemo } from 'react'
import { StyleSheet, Text, View } from 'react-native'

import { OnboardingStep } from '../../src/components/OnboardingStep'
import { SpeedometerGauge } from '../../src/components/SpeedometerGauge'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { useResponsiveScale } from '../../src/utils/responsive'

export default function SC00() {
  const router = useRouter()
  const theme = useTheme()
  const { t } = useLocalization()
  const { compact, font } = useResponsiveScale()
  const titleSize = font(32, 26)
  const subtitleSize = font(15, 13)
  const boardTitleSize = font(20, 17)
  const portfolioValueSize = font(32, 26)
  const gaugeSize = compact ? 96 : 106
  const sectionGap = compact ? theme.spacing.md : theme.spacing.lg
  const boardRadius = compact ? theme.radius.sm : theme.radius.md
  const styles = useMemo(
    () =>
      createStyles(theme, {
        boardRadius,
        boardTitleSize,
        portfolioValueSize,
        gaugeSize,
        titleSize,
        subtitleSize,
      }),
    [boardRadius, boardTitleSize, gaugeSize, portfolioValueSize, subtitleSize, theme, titleSize]
  )

  const ledgerRows = [
    {
      id: 'binance-btc',
      platform: 'Binance',
      asset: 'BTC · 1.24',
      value: '$109,864',
      color: theme.colors.accent,
    },
    {
      id: 'fidelity-nvda',
      platform: 'Fidelity',
      asset: 'NVDA · 400 sh',
      value: '$74,588',
      color: theme.colors.warning,
    },
    {
      id: 'chase-usd',
      platform: 'Chase',
      asset: 'USD cash',
      value: '$39,200',
      color: theme.colors.muted,
    },
  ]

  return (
    <OnboardingStep
      nextLabel={t('common.next')}
      onNext={() => router.push('/(auth)/sc01')}
      swipeEnabled
      progressTotal={3}
      progressIndex={0}
      animateIn={false}
      contentScroll
      nextButtonStyle={{ borderRadius: theme.radius.md }}
    >
      <View style={{ gap: sectionGap }}>
        <View style={styles.board}>
          <View style={styles.boardContent}>
            <View style={styles.boardHeader}>
              <View style={styles.headerCopy}>
                <Text style={styles.boardTitle}>{t('onboarding.sc00.boardTitle')}</Text>
                <Text style={styles.portfolioValue}>$223,652</Text>
              </View>
              <View style={styles.gaugeWrap}>
                <SpeedometerGauge value={72} label={t('onboarding.sc00.healthCornerLabel')} size={gaugeSize} />
              </View>
            </View>

            <View style={styles.ledger}>
              {ledgerRows.map((row, index) => (
                <View
                  key={row.id}
                  style={[styles.ledgerRow, index < ledgerRows.length - 1 ? styles.ledgerRowDivider : undefined]}
                >
                  <View style={styles.ledgerLabelWrap}>
                    <View style={[styles.ledgerMarker, { backgroundColor: row.color }]} />
                    <View style={styles.ledgerTextBlock}>
                      <Text style={styles.ledgerPlatform}>{row.platform}</Text>
                      <Text style={styles.ledgerAsset}>{row.asset}</Text>
                    </View>
                  </View>
                  <Text style={styles.ledgerValue}>{row.value}</Text>
                </View>
              ))}
            </View>
          </View>
        </View>

        <View style={styles.copyBlock}>
          <Text style={styles.title}>{t('onboarding.sc00.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc00.subtitle')}</Text>
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
    portfolioValueSize: number
    gaugeSize: number
    titleSize: number
    subtitleSize: number
  }
) {
  const divider = {
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: theme.colors.border,
  } as const

  return StyleSheet.create({
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
    },
    boardHeader: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      justifyContent: 'space-between',
      gap: theme.spacing.md,
    },
    headerCopy: {
      flex: 1,
      gap: theme.spacing.xs,
    },
    gaugeWrap: {
      alignItems: 'center',
      justifyContent: 'flex-start',
      minWidth: sizes.gaugeSize,
    },
    boardTitle: {
      fontSize: sizes.boardTitleSize,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      letterSpacing: -0.1,
    },
    portfolioValue: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: sizes.portfolioValueSize,
      letterSpacing: 0.2,
      marginTop: -2,
    },
    ledger: {
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: Math.max(8, sizes.boardRadius - 4),
      backgroundColor: theme.colors.surfaceElevated,
      overflow: 'hidden',
      position: 'relative',
    },
    ledgerRow: {
      minHeight: 60,
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.sm,
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      backgroundColor: theme.colors.surface,
    },
    ledgerRowDivider: divider,
    ledgerLabelWrap: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.sm,
      flex: 1,
    },
    ledgerMarker: {
      width: 8,
      height: 8,
      borderRadius: 2,
    },
    ledgerTextBlock: {
      gap: 2,
      flex: 1,
      paddingRight: theme.spacing.sm,
    },
    ledgerPlatform: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 15,
      letterSpacing: 0.2,
    },
    ledgerAsset: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
      letterSpacing: 0.2,
    },
    ledgerValue: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 15,
      letterSpacing: 0.2,
      marginLeft: theme.spacing.sm,
    },
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
  })
}
