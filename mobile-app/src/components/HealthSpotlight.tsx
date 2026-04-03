import React, { useMemo } from 'react'
import { StyleSheet, Text, View, useWindowDimensions } from 'react-native'

import { Badge } from './Badge'
import { SpeedometerGauge } from './SpeedometerGauge'
import { useTheme } from '../providers/ThemeProvider'
import { BadgeTone } from '../utils/badges'
import { formatNumber } from '../utils/format'

type HealthSpotlightProps = {
  healthScore?: number | null
  healthLabel: string
  healthGaugeLabel?: string
  healthStatusLabel?: string | null
  healthTone?: BadgeTone
  volatilityScore?: number | null
  volatilityTitle: string
  volatilityLevelLabel?: string | null
  volatilityTone?: BadgeTone
  volatilityColor?: string
  locked?: boolean
}

function toneColor(theme: ReturnType<typeof useTheme>, tone?: BadgeTone) {
  switch (tone) {
    case 'critical':
      return theme.colors.danger
    case 'high':
      return theme.colors.warning
    case 'medium':
      return theme.colors.accent
    case 'low':
      return theme.colors.success
    default:
      return theme.colors.muted
  }
}

export function HealthSpotlight({
  healthScore,
  healthLabel,
  healthGaugeLabel,
  healthStatusLabel,
  healthTone,
  volatilityScore,
  volatilityTitle,
  volatilityLevelLabel,
  volatilityTone,
  volatilityColor,
  locked = false,
}: HealthSpotlightProps) {
  const theme = useTheme()
  const { width } = useWindowDimensions()
  const gaugeSize = Math.min(180, Math.max(140, Math.floor(width * 0.46)))
  const volatilityValue = volatilityScore ?? null
  const ratio = volatilityValue === null ? 0 : Math.max(0, Math.min(1, volatilityValue / 100))
  const percent = Math.round(ratio * 100)
  const barColor = volatilityColor ?? toneColor(theme, volatilityTone)
  const styles = useMemo(() => createStyles(theme), [theme])

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.healthTitle}>{healthLabel}</Text>
        {healthStatusLabel ? <Badge label={healthStatusLabel} tone={healthTone} /> : null}
      </View>
      <View style={styles.gaugeWrap}>
        <SpeedometerGauge
          size={gaugeSize}
          value={healthScore}
          label={healthGaugeLabel ?? healthLabel}
          locked={locked}
        />
      </View>
      <View style={styles.strip}>
        <View style={styles.stripHeader}>
          <Text style={styles.stripLabel}>{volatilityTitle}</Text>
          {volatilityLevelLabel ? <Badge label={volatilityLevelLabel} tone={volatilityTone} /> : null}
        </View>
        <View style={styles.stripRow}>
          <Text style={styles.stripValue}>{volatilityValue === null ? '--' : formatNumber(volatilityValue)}</Text>
          <View style={styles.track}>
            <View style={[styles.fill, { width: `${percent}%`, backgroundColor: barColor }]} />
            <View
              style={[
                styles.marker,
                { left: `${percent}%`, borderColor: barColor, opacity: volatilityValue === null ? 0 : 1 },
              ]}
            />
          </View>
        </View>
      </View>
    </View>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    container: {
      backgroundColor: theme.colors.surfaceElevated,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: theme.colors.border,
      paddingVertical: theme.spacing.md,
      paddingHorizontal: theme.spacing.md,
    },
    header: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    healthTitle: {
      fontSize: 14,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      letterSpacing: 0.2,
    },
    gaugeWrap: {
      alignItems: 'center',
      marginTop: theme.spacing.sm,
      marginBottom: theme.spacing.sm,
    },
    strip: {
      marginTop: theme.spacing.sm,
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
    },
    stripHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    stripLabel: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.2,
    },
    stripRow: {
      marginTop: theme.spacing.sm,
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.sm,
    },
    stripValue: {
      fontSize: 18,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      minWidth: 42,
    },
    track: {
      flex: 1,
      height: 8,
      borderRadius: 999,
      backgroundColor: theme.colors.surface,
      borderWidth: 1,
      borderColor: theme.colors.border,
      overflow: 'visible',
      position: 'relative',
    },
    fill: {
      height: '100%',
      borderRadius: 999,
    },
    marker: {
      position: 'absolute',
      top: -4,
      width: 10,
      height: 16,
      borderRadius: 6,
      borderWidth: 2,
      backgroundColor: theme.colors.surface,
      transform: [{ translateX: -5 }],
    },
  })
