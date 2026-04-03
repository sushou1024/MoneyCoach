import { useQuery } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { Alert, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Input } from '../../src/components/Input'
import { Screen } from '../../src/components/Screen'
import { useProfile } from '../../src/hooks/useProfile'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { executeInsight, fetchInsights } from '../../src/services/insights'
import { formatCurrency, formatNumber } from '../../src/utils/format'

export default function QuickUpdateScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken, userId } = useAuth()
  const profile = useProfile()
  const params = useLocalSearchParams<{ id?: string }>()
  const insightId = params.id ?? ''
  const [amount, setAmount] = useState('')
  const [unit, setUnit] = useState<'base' | 'asset'>('base')

  const insightQuery = useQuery({
    queryKey: ['insight', userId, insightId],
    queryFn: async () => {
      if (!accessToken || !insightId) return null
      const resp = await fetchInsights(accessToken, 'all')
      if (resp.error) throw new Error(resp.error.message)
      return resp.data?.items.find((item) => item.id === insightId) ?? null
    },
    enabled: !!accessToken && !!insightId,
  })

  const insight = insightQuery.data
  const suggestedQuantity = insight?.suggested_quantity
  const isRebalance = suggestedQuantity?.mode === 'rebalance'
  const baseCurrency = suggestedQuantity?.display_currency ?? profile.data?.base_currency ?? 'USD'

  const suggestedSummary = useMemo(() => {
    if (!suggestedQuantity) return null
    if (suggestedQuantity.mode === 'usd') {
      return t('quickUpdate.suggestedUsd', {
        amount: formatCurrency(
          suggestedQuantity.amount_display ?? suggestedQuantity.amount_usd ?? 0,
          suggestedQuantity.amount_display !== undefined && suggestedQuantity.amount_display !== null
            ? baseCurrency
            : 'USD'
        ),
      })
    }
    if (suggestedQuantity.mode === 'asset') {
      return t('quickUpdate.suggestedAsset', {
        amount: formatNumber(suggestedQuantity.amount_asset ?? 0),
        symbol: suggestedQuantity.symbol ?? insight?.asset ?? '',
      })
    }
    if (suggestedQuantity.mode === 'rebalance') {
      const trades = Array.isArray(suggestedQuantity.trades) ? suggestedQuantity.trades.length : 0
      return t('quickUpdate.suggestedRebalance', { count: trades })
    }
    return null
  }, [baseCurrency, insight?.asset, suggestedQuantity, t])

  const handleExecuteSuggested = async () => {
    if (!accessToken || !insight) return
    const resp = await executeInsight(accessToken, insight.id, { method: 'suggested' })
    if (resp.error) {
      Alert.alert(t('quickUpdate.failTitle'), resp.error.message)
      return
    }
    Alert.alert(t('quickUpdate.successTitle'), t('quickUpdate.successBody'))
    router.replace('/(tabs)/insights')
  }

  const handleSubmit = async () => {
    if (!accessToken || !insightId) return
    const quantity = Number(amount)
    if (!Number.isFinite(quantity) || quantity <= 0) {
      Alert.alert(t('quickUpdate.invalidTitle'), t('quickUpdate.invalidBody'))
      return
    }
    const resp = await executeInsight(accessToken, insightId, {
      method: 'manual',
      quantity,
      quantity_unit: unit,
    })
    if (resp.error) {
      Alert.alert(t('quickUpdate.failTitle'), resp.error.message)
      return
    }
    Alert.alert(t('quickUpdate.successTitle'), t('quickUpdate.successBody'))
    router.replace('/(tabs)/insights')
  }

  const openTradeSlip = () => {
    router.push({ pathname: '/(modals)/trade-slip', params: { insight_id: insightId } })
  }

  if (!insight) {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('quickUpdate.loading')}</Text>
      </Screen>
    )
  }

  return (
    <Screen>
      <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
        {t('quickUpdate.title')}
      </Text>
      <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {t('quickUpdate.subtitle')}
      </Text>
      <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
        {insight.asset}
      </Text>
      <Text style={{ marginTop: theme.spacing.xs, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {insight.trigger_reason}
      </Text>
      {suggestedSummary ? (
        <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {suggestedSummary}
        </Text>
      ) : null}

      {suggestedQuantity ? (
        <Button
          title={t('quickUpdate.applySuggested')}
          style={{ marginTop: theme.spacing.md }}
          onPress={handleExecuteSuggested}
        />
      ) : null}

      {!isRebalance && (
        <View style={{ marginTop: theme.spacing.md }}>
          <Input
            value={amount}
            onChangeText={setAmount}
            placeholder={t('quickUpdate.amountPlaceholder')}
            keyboardType="decimal-pad"
          />
          <View style={{ flexDirection: 'row', gap: theme.spacing.sm, marginTop: theme.spacing.sm }}>
            <Button
              title={t('quickUpdate.unitBase', { currency: baseCurrency })}
              variant={unit === 'base' ? 'primary' : 'secondary'}
              onPress={() => setUnit('base')}
            />
            <Button
              title={t('quickUpdate.unitAsset', { symbol: insight.asset })}
              variant={unit === 'asset' ? 'primary' : 'secondary'}
              onPress={() => setUnit('asset')}
            />
          </View>
          <Button title={t('quickUpdate.submit')} style={{ marginTop: theme.spacing.md }} onPress={handleSubmit} />
        </View>
      )}

      <Button
        title={t('quickUpdate.tradeSlip')}
        variant="ghost"
        style={{ marginTop: theme.spacing.md }}
        onPress={openTradeSlip}
      />
    </Screen>
  )
}
