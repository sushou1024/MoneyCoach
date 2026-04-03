import React, { useMemo } from 'react'
import { StyleProp, StyleSheet, Text, View, ViewStyle } from 'react-native'

import { Card } from './Card'
import { useLocalization } from '../providers/LocalizationProvider'
import { useTheme } from '../providers/ThemeProvider'

type WaitlistTicketProps = {
  strategyName?: string
  rank: number
  isPaid: boolean
  style?: StyleProp<ViewStyle>
}

export function WaitlistTicket({ strategyName, rank, isPaid, style }: WaitlistTicketProps) {
  const theme = useTheme()
  const { t } = useLocalization()
  const styles = useMemo(() => createStyles(theme), [theme])

  return (
    <Card style={[styles.card, style]}>
      <View style={styles.ticket}>
        <View style={styles.ticketRail} />
        <View style={styles.ticketBody}>
          {strategyName ? (
            <View style={styles.strategyBlock}>
              <Text style={styles.strategyLabel}>{t('waitlist.strategyLabel')}</Text>
              <Text style={styles.strategyName}>{strategyName}</Text>
            </View>
          ) : null}
          <View style={styles.rankBlock}>
            <Text style={styles.rankLabel}>{t('waitlist.rankLabel')}</Text>
            <Text style={styles.rankValue}>#{rank}</Text>
            <Text style={styles.rankHint}>{isPaid ? t('waitlist.rankHintPaid') : t('waitlist.rankHintFree')}</Text>
          </View>
        </View>
      </View>
    </Card>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    card: {
      marginTop: theme.spacing.md,
    },
    ticket: {
      flexDirection: 'row',
      gap: theme.spacing.md,
    },
    ticketRail: {
      width: 3,
      borderRadius: 999,
      backgroundColor: theme.colors.accent,
      opacity: 0.7,
    },
    ticketBody: {
      flex: 1,
    },
    strategyBlock: {
      gap: theme.spacing.xs,
    },
    strategyLabel: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.4,
      textTransform: 'uppercase',
    },
    strategyName: {
      fontSize: 18,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
    },
    rankBlock: {
      marginTop: theme.spacing.md,
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
      gap: theme.spacing.xs,
    },
    rankLabel: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.3,
      textTransform: 'uppercase',
    },
    rankValue: {
      fontSize: 32,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    rankHint: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
      lineHeight: 18,
    },
  })
