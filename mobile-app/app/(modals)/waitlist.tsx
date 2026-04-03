import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { StyleSheet, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Screen } from '../../src/components/Screen'
import { WaitlistTicket } from '../../src/components/WaitlistTicket'
import { useEntitlement } from '../../src/hooks/useEntitlement'
import { useWaitlistEntry } from '../../src/hooks/useWaitlistEntry'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { getStrategyDisplayName } from '../../src/utils/strategies'

export default function WaitlistScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const entitlement = useEntitlement()
  const params = useLocalSearchParams<{ strategy_id?: string; calculation_id?: string }>()
  const strategyId = params.strategy_id ?? ''
  const calculationId = params.calculation_id ?? ''
  const [didTapNotify, setDidTapNotify] = useState(false)
  const { rank, error, isReady } = useWaitlistEntry(strategyId, calculationId)

  const isPaid = ['active', 'grace'].includes(entitlement.data?.status ?? '')
  const strategyName = useMemo(() => (strategyId ? getStrategyDisplayName(t, strategyId) : ''), [strategyId, t])
  const styles = useMemo(() => createStyles(theme), [theme])

  return (
    <Screen>
      <View style={styles.header}>
        <Text style={styles.title}>{t('waitlist.title')}</Text>
        <Text style={styles.subtitle}>{t('waitlist.subtitle')}</Text>
      </View>

      {error ? (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      ) : null}

      {!error && rank === null ? <Text style={styles.loadingText}>{t('waitlist.loading')}</Text> : null}

      {rank !== null ? (
        <WaitlistTicket strategyName={strategyName} rank={rank} isPaid={isPaid} style={styles.ticketCard} />
      ) : null}

      {didTapNotify && isReady ? <Text style={styles.successText}>{t('waitlist.notifySuccess')}</Text> : null}

      <Button
        title={didTapNotify ? t('waitlist.notifyDone') : t('waitlist.cta')}
        style={styles.primaryCta}
        disabled={!isReady || didTapNotify}
        onPress={() => setDidTapNotify(true)}
      />
      <Button
        title={t('waitlist.backCta')}
        variant="ghost"
        style={styles.backCta}
        textStyle={styles.backCtaText}
        onPress={() => router.back()}
      />
    </Screen>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    header: {
      gap: theme.spacing.xs,
    },
    title: {
      fontSize: 24,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    subtitle: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 14,
      lineHeight: 20,
    },
    loadingText: {
      marginTop: theme.spacing.md,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 14,
    },
    errorBanner: {
      marginTop: theme.spacing.md,
      padding: theme.spacing.sm,
      borderRadius: theme.radius.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
    },
    errorText: {
      color: theme.colors.danger,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
    },
    ticketCard: {
      marginTop: theme.spacing.md,
    },
    successText: {
      marginTop: theme.spacing.sm,
      color: theme.colors.success,
      fontFamily: theme.fonts.bodyMedium,
    },
    primaryCta: {
      marginTop: theme.spacing.lg,
    },
    backCta: {
      marginTop: theme.spacing.xs,
    },
    backCtaText: {
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
      letterSpacing: 0,
    },
  })
