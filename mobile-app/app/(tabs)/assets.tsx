import { Ionicons } from '@expo/vector-icons'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useFocusEffect, useRouter } from 'expo-router'
import React, { useCallback, useMemo, useState } from 'react'
import {
  ActivityIndicator,
  Alert,
  FlatList,
  Pressable,
  StyleSheet,
  Text,
  View,
  ViewStyle,
  useWindowDimensions,
} from 'react-native'

import { AssetLogo } from '../../src/components/AssetLogo'
import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { LoadingScreen } from '../../src/components/LoadingScreen'
import { Screen } from '../../src/components/Screen'
import { SpeedometerGauge } from '../../src/components/SpeedometerGauge'
import { useEntitlement } from '../../src/hooks/useEntitlement'
import { useActivePortfolio } from '../../src/hooks/usePortfolio'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { getTodayBriefings } from '../../src/services/briefings'
import { fetchPortfolioSnapshots, refreshActivePortfolio } from '../../src/services/portfolio'
import { requestActiveReport, requestPaidReport } from '../../src/services/reports'
import { PortfolioHolding } from '../../src/types/api'
import {
  healthStatusFromScore,
  healthStatusLabelKey,
  toneForHealthStatus,
  toneForSeverity,
} from '../../src/utils/badges'
import {
  formatCurrency,
  formatDateTime,
  formatNumber,
  formatPercent,
  formatPortfolioTotal,
} from '../../src/utils/format'
import { resolveActionBiasLabel, toneForActionBias } from '../../src/utils/intelligence'
const hkSymbolPattern = /^0*(\d{1,5})\.?HK$/i

export default function AssetsScreen() {
  const theme = useTheme()
  const router = useRouter()
  const queryClient = useQueryClient()
  const { t } = useLocalization()
  const { width } = useWindowDimensions()
  const { accessToken, clearSession, userId } = useAuth()
  const portfolioQuery = useActivePortfolio()
  const { data: portfolio, isLoading } = portfolioQuery
  const entitlement = useEntitlement()
  const [diagnoseSubmitting, setDiagnoseSubmitting] = useState(false)
  const [refreshSubmitting, setRefreshSubmitting] = useState(false)
  const [pullRefreshing, setPullRefreshing] = useState(false)

  const reports = useQuery({
    queryKey: ['reports', 'active', userId],
    queryFn: async () => {
      if (!accessToken) return []
      const resp = await fetchPortfolioSnapshots(accessToken)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return []
        }
        throw new Error(resp.error.message)
      }
      return resp.data?.items ?? []
    },
    enabled: !!accessToken,
  })

  useFocusEffect(
    useCallback(() => {
      reports.refetch()
    }, [reports.refetch])
  )

  useFocusEffect(
    useCallback(() => {
      portfolioQuery.refetch()
    }, [portfolioQuery.refetch])
  )

  const briefingsQuery = useQuery({
    queryKey: ['briefings', 'today', userId],
    queryFn: async () => {
      if (!accessToken) return []
      const resp = await getTodayBriefings(accessToken)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return []
        }
        return []
      }
      const items = resp.data?.briefings ?? []
      return [...items].sort((a, b) => a.priority - b.priority)
    },
    enabled: !!accessToken,
  })

  useFocusEffect(
    useCallback(() => {
      briefingsQuery.refetch()
    }, [briefingsQuery.refetch])
  )

  const isPaid = ['active', 'grace'].includes(entitlement.data?.status ?? '')

  const sortedHoldings = useMemo(() => {
    if (!portfolio?.holdings) return []
    return [...portfolio.holdings].sort((a, b) => b.value_usd_priced - a.value_usd_priced)
  }, [portfolio?.holdings])

  const totalUsd = useMemo(() => sortedHoldings.reduce((sum, h) => sum + h.value_usd_priced, 0), [sortedHoldings])

  const reportsData = reports.data ?? []
  const activePreview = reportsData.find((item) => item.is_active && item.report_tier === 'preview')
  const activePaid = reportsData.find((item) => item.is_active && item.report_tier === 'paid')
  const desiredTier = isPaid ? 'paid' : 'preview'
  const hasAnyReadyReport = reportsData.some((item) => item.report_tier === desiredTier && item.status === 'ready')
  const hasReadyActiveReport = reportsData.some(
    (item) => item.is_active && item.report_tier === desiredTier && item.status === 'ready'
  )
  const showDiagnoseDot = hasAnyReadyReport && !hasReadyActiveReport
  const healthScore = portfolio?.dashboard_metrics?.health_score
  const previewHealthScore = !isPaid && activePreview?.status === 'ready' ? activePreview.health_score : undefined
  const resolvedHealthScore = healthScore ?? previewHealthScore
  const isHealthLocked = !isPaid && (resolvedHealthScore === undefined || resolvedHealthScore === null)
  const volatilityScore = portfolio?.dashboard_metrics?.volatility_score
  const valuationAsOf = portfolio?.dashboard_metrics?.valuation_as_of
  const metricsIncomplete = portfolio?.dashboard_metrics?.metrics_incomplete
  const healthStatus = portfolio?.dashboard_metrics?.health_status ?? healthStatusFromScore(resolvedHealthScore)
  const healthStatusLabel = healthStatus ? t(healthStatusLabelKey(healthStatus) as any) : null

  const netWorthDisplay = portfolio?.dashboard_metrics?.net_worth_display ?? portfolio?.net_worth_usd ?? 0
  const baseCurrency = portfolio?.dashboard_metrics?.base_currency ?? 'USD'
  const healthGaugeSize = Math.min(190, Math.max(150, Math.floor(width * 0.48)))
  const volatilityValue = volatilityScore ?? null
  const volatilityRatio = volatilityValue === null ? 0 : Math.max(0, Math.min(1, volatilityValue / 100))
  const volatilityLevel =
    volatilityValue === null ? null : volatilityValue < 35 ? 'low' : volatilityValue < 70 ? 'medium' : 'high'
  const volatilityLabel = volatilityLevel ? t(`insights.severity.${volatilityLevel}` as any) : null
  const volatilityTone = volatilityLevel ? toneForSeverity(volatilityLevel) : 'neutral'
  const volatilityColor =
    volatilityValue === null
      ? theme.colors.muted
      : volatilityValue < 40
        ? theme.colors.success
        : volatilityValue < 70
          ? theme.colors.warning
          : theme.colors.danger
  const styles = useMemo(() => createStyles(theme), [theme])

  const normalizeHKSymbol = useCallback((symbol: string) => {
    const trimmed = symbol.trim()
    const match = trimmed.match(hkSymbolPattern)
    if (!match) return trimmed
    return `${match[1]}.HK`
  }, [])

  const formatHoldingLabel = useCallback(
    (holding: PortfolioHolding) => {
      const symbol = normalizeHKSymbol(holding.symbol)
      const name = holding.name?.trim()
      if (!name) return symbol
      const mic = (holding.exchange_mic ?? '').toUpperCase()
      const isHongKong =
        holding.asset_type === 'stock' &&
        (mic === 'XHKG' || symbol.toUpperCase().endsWith('.HK') || holding.asset_key.toUpperCase().includes(':XHKG:'))
      if (!isHongKong) return symbol
      return `${name} (${symbol})`
    },
    [normalizeHKSymbol]
  )

  const holdingsActionStyle: ViewStyle = {
    minHeight: 32,
    flexDirection: 'row',
    alignItems: 'center',
    gap: theme.spacing.xs,
    borderRadius: theme.radius.lg,
    borderWidth: 1,
    borderColor: theme.colors.border,
    backgroundColor: theme.colors.surfaceElevated,
    paddingVertical: 6,
    paddingHorizontal: theme.spacing.sm,
  }

  const handlePullRefresh = useCallback(async () => {
    setPullRefreshing(true)
    try {
      await Promise.all([
        briefingsQuery.refetch(),
        portfolioQuery.refetch(),
        reports.refetch(),
      ])
    } finally {
      setPullRefreshing(false)
    }
  }, [briefingsQuery.refetch, portfolioQuery.refetch, reports.refetch])

  const handleDiagnose = async () => {
    if (!accessToken || diagnoseSubmitting) return
    const returnTo = '/(tabs)/assets'

    if (!isPaid) {
      if (
        activePreview?.calculation_id &&
        (activePreview.status === 'ready' || activePreview.status === 'processing')
      ) {
        router.push({ pathname: '/(modals)/preview', params: { id: activePreview.calculation_id } })
        return
      }
      setDiagnoseSubmitting(true)
      try {
        const resp = await requestActiveReport(accessToken, 'preview')
        if (resp.error || !resp.data?.calculation_id) {
          Alert.alert(t('assets.diagnoseFailTitle'), resp.error?.message ?? t('assets.diagnoseFailBody'))
          return
        }
        router.push({ pathname: '/(modals)/preview', params: { id: resp.data.calculation_id } })
      } finally {
        setDiagnoseSubmitting(false)
      }
      return
    }

    if (activePaid?.calculation_id) {
      if (activePaid.status === 'ready') {
        router.push({ pathname: '/(modals)/report', params: { id: activePaid.calculation_id, return_to: returnTo } })
        return
      }
      setDiagnoseSubmitting(true)
      try {
        if (activePaid.status !== 'processing') {
          const resp = await requestPaidReport(accessToken, activePaid.calculation_id)
          if (resp.error) {
            Alert.alert(t('assets.diagnoseFailTitle'), resp.error.message)
            return
          }
        }
        router.push({
          pathname: '/(modals)/processing-paid',
          params: { calculation_id: activePaid.calculation_id, return_to: returnTo },
        })
      } finally {
        setDiagnoseSubmitting(false)
      }
      return
    }

    if (activePreview?.calculation_id) {
      setDiagnoseSubmitting(true)
      try {
        const resp = await requestPaidReport(accessToken, activePreview.calculation_id)
        if (resp.error) {
          Alert.alert(t('assets.diagnoseFailTitle'), resp.error.message)
          return
        }
        router.push({
          pathname: '/(modals)/processing-paid',
          params: { calculation_id: activePreview.calculation_id, return_to: returnTo },
        })
      } finally {
        setDiagnoseSubmitting(false)
      }
      return
    }

    setDiagnoseSubmitting(true)
    try {
      const resp = await requestActiveReport(accessToken, 'paid')
      if (resp.error || !resp.data?.calculation_id) {
        Alert.alert(t('assets.diagnoseFailTitle'), resp.error?.message ?? t('assets.diagnoseFailBody'))
        return
      }
      router.push({
        pathname: '/(modals)/processing-paid',
        params: { calculation_id: resp.data.calculation_id, return_to: returnTo },
      })
    } finally {
      setDiagnoseSubmitting(false)
    }
  }

  const handleRefreshHoldings = async () => {
    if (!accessToken || !userId || refreshSubmitting) return

    setRefreshSubmitting(true)
    try {
      const resp = await refreshActivePortfolio(accessToken)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return
        }
        Alert.alert(t('assets.refreshFailTitle'), resp.error.message || t('assets.refreshFailBody'))
        return
      }
      if (!resp.data) {
        Alert.alert(t('assets.refreshFailTitle'), t('assets.refreshFailBody'))
        return
      }

      queryClient.setQueryData(['portfolio', 'active', userId], resp.data)
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['reports', 'active', userId] }),
        queryClient.invalidateQueries({ queryKey: ['market-regime', userId] }),
        queryClient.invalidateQueries({ queryKey: ['insights', userId] }),
        queryClient.invalidateQueries({ queryKey: ['asset-brief', userId] }),
      ])
    } finally {
      setRefreshSubmitting(false)
    }
  }

  const openAssetBrief = (assetKey: string) => {
    if (!isPaid) return
    router.push({ pathname: '/(modals)/asset-brief', params: { asset_key: assetKey } })
  }

  if (isLoading) {
    return <LoadingScreen label={t('common.loading')} />
  }

  if (!portfolio) {
    return (
      <Screen>
        <View style={{ flex: 1, justifyContent: 'center' }}>
          <Text style={{ fontSize: 28, color: theme.colors.ink, textAlign: 'center', fontFamily: theme.fonts.display }}>
            {t('assets.emptyTitle')}
          </Text>
          <Text
            style={{
              marginTop: theme.spacing.sm,
              color: theme.colors.muted,
              textAlign: 'center',
              fontFamily: theme.fonts.body,
            }}
          >
            {t('assets.emptySubtitle')}
          </Text>
          <Button
            title={t('assets.emptyCta')}
            style={{ marginTop: theme.spacing.lg }}
            onPress={() => router.push('/(modals)/upload')}
          />
        </View>
      </Screen>
    )
  }

  const briefingIconMap: Record<string, keyof typeof Ionicons.glyphMap> = {
    market_move: 'trending-up',
    portfolio_change: 'briefcase-outline',
    risk_alert: 'warning-outline',
    opportunity: 'flash-outline',
    education: 'school-outline',
  }

  const briefingTypeLabelMap: Record<string, string> = {
    market_move: 'market',
    portfolio_change: 'portfolio',
    risk_alert: 'risk',
    opportunity: 'action',
    education: 'default',
  }

  const briefingTypeLabel = (type: string) => {
    const mapped = briefingTypeLabelMap[type] ?? 'default'
    const key = `briefing.type.${mapped}` as any
    const resolved = t(key)
    return resolved !== key ? resolved : t('briefing.type.default' as any)
  }

  const briefings = briefingsQuery.data ?? []

  return (
    <Screen scroll refreshing={pullRefreshing} onRefresh={handlePullRefresh}>
      <Card style={{ marginBottom: theme.spacing.md }}>
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.xs }}>
          <Ionicons name="newspaper-outline" size={18} color={theme.colors.accent} />
          <Text style={{ fontSize: 16, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
            {t('briefing.title' as any)}
          </Text>
        </View>
        {briefingsQuery.isLoading ? (
          <View style={{ paddingVertical: theme.spacing.md, alignItems: 'center' }}>
            <ActivityIndicator size="small" color={theme.colors.muted} />
          </View>
        ) : briefings.length === 0 ? (
          <Text
            style={{
              marginTop: theme.spacing.sm,
              color: theme.colors.muted,
              fontFamily: theme.fonts.body,
              textAlign: 'center',
              paddingVertical: theme.spacing.md,
            }}
          >
            {t('briefing.empty' as any)}
          </Text>
        ) : (
          <View style={{ marginTop: theme.spacing.sm, gap: theme.spacing.sm }}>
            {briefings.map((item, index) => (
              <View
                key={`${item.type}-${index}`}
                style={{
                  flexDirection: 'row',
                  alignItems: 'flex-start',
                  gap: theme.spacing.sm,
                  paddingVertical: theme.spacing.xs,
                  borderTopWidth: index > 0 ? 1 : 0,
                  borderTopColor: theme.colors.border,
                }}
              >
                <View
                  style={{
                    width: 32,
                    height: 32,
                    borderRadius: theme.radius.md,
                    backgroundColor: theme.colors.accentSoft,
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Ionicons
                    name={briefingIconMap[item.type] ?? 'information-circle-outline'}
                    size={18}
                    color={theme.colors.accent}
                  />
                </View>
                <View style={{ flex: 1 }}>
                  <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.xs }}>
                    <Text style={{ fontSize: 14, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold, flex: 1 }}>
                      {item.title}
                    </Text>
                    <Badge label={briefingTypeLabel(item.type)} tone="neutral" />
                  </View>
                  <Text
                    style={{
                      marginTop: 2,
                      fontSize: 13,
                      color: theme.colors.muted,
                      fontFamily: theme.fonts.body,
                      lineHeight: 18,
                    }}
                  >
                    {item.body}
                  </Text>
                </View>
              </View>
            ))}
          </View>
        )}
      </Card>
      <Card>
        <View>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('assets.netWorth')}</Text>
          <Text style={{ fontSize: 28, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
            {formatPortfolioTotal(netWorthDisplay, baseCurrency)}
          </Text>
          {valuationAsOf ? (
            <Text style={{ marginTop: theme.spacing.xs, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
              {t('assets.valuationAsOf', { time: formatDateTime(valuationAsOf) })}
            </Text>
          ) : null}
        </View>
        <View style={{ marginTop: theme.spacing.md }}>
          <View style={styles.healthSpotlight}>
            <View style={styles.healthHeader}>
              <Text style={styles.healthEyebrow}>{t('assets.healthLabel')}</Text>
              {healthStatusLabel ? <Badge label={healthStatusLabel} tone={toneForHealthStatus(healthStatus)} /> : null}
            </View>
            <View style={styles.healthGaugeWrap}>
              <SpeedometerGauge
                size={healthGaugeSize}
                value={resolvedHealthScore}
                label={t('assets.healthShort')}
                locked={isHealthLocked}
              />
            </View>
            <View style={styles.volatilityStrip}>
              <View style={styles.volatilityHeader}>
                <Text style={styles.volatilityLabel}>{t('assets.volatilityShort')}</Text>
                {volatilityLabel ? <Badge label={volatilityLabel} tone={volatilityTone} /> : null}
              </View>
              <View style={styles.volatilityRow}>
                <Text style={styles.volatilityValue}>
                  {volatilityValue === null ? '--' : formatNumber(volatilityValue)}
                </Text>
                <View style={styles.volatilityTrack}>
                  <View
                    style={[
                      styles.volatilityFill,
                      { width: `${Math.round(volatilityRatio * 100)}%`, backgroundColor: volatilityColor },
                    ]}
                  />
                  <View
                    style={[
                      styles.volatilityMarker,
                      { left: `${Math.round(volatilityRatio * 100)}%`, borderColor: volatilityColor },
                    ]}
                  />
                </View>
              </View>
            </View>
          </View>
        </View>
        {metricsIncomplete ? (
          <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.warning, fontFamily: theme.fonts.body }}>
            {t('assets.metricsIncomplete')}
          </Text>
        ) : null}
        <View style={{ marginTop: theme.spacing.md }}>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.sm }}>
            <View style={{ flex: 1 }}>
              <Button
                title={t('assets.diagnoseCta')}
                icon={<Ionicons name="medkit-outline" size={18} color={theme.colors.surfaceElevated} />}
                style={{ minHeight: 44, paddingVertical: 10 }}
                onPress={handleDiagnose}
                disabled={diagnoseSubmitting}
                loading={diagnoseSubmitting}
              />
            </View>
            {showDiagnoseDot ? (
              <View
                style={{
                  width: 10,
                  height: 10,
                  borderRadius: 999,
                  backgroundColor: theme.colors.danger,
                  shadowColor: theme.colors.danger,
                  shadowOpacity: 0.4,
                  shadowRadius: 6,
                }}
              />
            ) : null}
          </View>
          {showDiagnoseDot ? (
            <Text style={{ marginTop: theme.spacing.xs, color: theme.colors.danger, fontFamily: theme.fonts.body }}>
              {t('assets.diagnoseWarning')}
            </Text>
          ) : null}
        </View>
      </Card>
      <View style={{ marginTop: theme.spacing.lg }}>
        <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.sm }}>
            <Text style={{ fontSize: 18, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
              {t('assets.holdingsTitle')}
            </Text>
            <Pressable
              onPress={handleRefreshHoldings}
              accessibilityRole="button"
              accessibilityLabel={t('assets.refreshAction')}
              disabled={refreshSubmitting}
              hitSlop={8}
              style={({ pressed }) => [
                { opacity: refreshSubmitting ? 0.55 : 1 },
                pressed && !refreshSubmitting ? { opacity: 0.75 } : null,
              ]}
            >
              {refreshSubmitting ? (
                <ActivityIndicator size="small" color={theme.colors.muted} />
              ) : (
                <Ionicons name="refresh" size={16} color={theme.colors.muted} />
              )}
            </Pressable>
          </View>
          <Pressable
            onPress={() => router.push('/(modals)/update-portfolio')}
            accessibilityRole="button"
            style={({ pressed }) => [holdingsActionStyle, pressed && { transform: [{ scale: 0.98 }], opacity: 0.9 }]}
          >
            <Ionicons name="pencil-outline" size={14} color={theme.colors.accent} />
            <Text style={{ color: theme.colors.accent, fontFamily: theme.fonts.bodyBold, fontSize: 13 }}>
              {t('assets.updateAction')}
            </Text>
          </Pressable>
        </View>
        {valuationAsOf ? (
          <Text
            style={{
              marginTop: theme.spacing.xs,
              color: theme.colors.muted,
              fontFamily: theme.fonts.body,
              fontSize: 12,
            }}
          >
            {t('assets.holdingsAsOf', { time: formatDateTime(valuationAsOf) })}
          </Text>
        ) : null}
        <FlatList
          data={sortedHoldings}
          keyExtractor={(item) => item.asset_key}
          scrollEnabled={false}
          renderItem={({ item }) => (
            <Pressable
              onPress={() => openAssetBrief(item.asset_key)}
              accessibilityRole={isPaid ? 'button' : undefined}
              disabled={!isPaid}
              style={({ pressed }) => [{ marginTop: theme.spacing.sm }, isPaid && pressed && { opacity: 0.92 }]}
            >
              <Card>
                <View style={{ flexDirection: 'row', alignItems: 'flex-start', gap: theme.spacing.sm }}>
                  <AssetLogo uri={item.logo_url} label={item.symbol} size={40} />
                  <View style={{ flex: 1 }}>
                    <Text
                      style={{
                        fontSize: 16,
                        color: theme.colors.ink,
                        fontFamily: theme.fonts.bodyBold,
                        flexWrap: 'wrap',
                      }}
                    >
                      {formatHoldingLabel(item)}
                    </Text>
                  </View>
                  {isPaid ? <Ionicons name="chevron-forward" size={18} color={theme.colors.muted} /> : null}
                </View>
                <View
                  style={{
                    marginTop: theme.spacing.sm,
                    flexDirection: 'row',
                    alignItems: 'flex-start',
                    justifyContent: 'space-between',
                    gap: theme.spacing.sm,
                  }}
                >
                  <View style={{ flex: 1 }}>
                    <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
                      {t('assets.amountLabel', { amount: formatNumber(item.amount) })}
                    </Text>
                    {item.avg_price && item.avg_price > 0 ? (
                      <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
                        {t('assets.avgCostLabel', {
                          value: formatCurrency(item.avg_price_quote ?? item.avg_price, item.quote_currency ?? 'USD'),
                        })}
                      </Text>
                    ) : null}
                    {totalUsd > 0 ? (
                      <Text style={{ color: theme.colors.accent, fontFamily: theme.fonts.bodyBold, fontSize: 13 }}>
                        {formatPercent(item.value_usd_priced / totalUsd, 1)}
                      </Text>
                    ) : null}
                  </View>
                  <View style={{ alignItems: 'flex-end' }}>
                    <Text style={{ fontSize: 16, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
                      {formatCurrency(item.value_quote ?? item.value_usd_priced, item.quote_currency ?? 'USD')}
                    </Text>
                    {isPaid && item.action_bias ? (
                      <View style={{ marginTop: theme.spacing.xs }}>
                        <Badge
                          label={resolveActionBiasLabel(t, item.action_bias)}
                          tone={toneForActionBias(item.action_bias)}
                        />
                      </View>
                    ) : null}
                  </View>
                </View>
              </Card>
            </Pressable>
          )}
        />
      </View>

      {!isPaid && (
        <Card style={{ marginTop: theme.spacing.lg }}>
          <Text style={{ fontSize: 16, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
            {t('assets.lockedTitle')}
          </Text>
          <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.sm, fontFamily: theme.fonts.body }}>
            {t('assets.lockedBody')}
          </Text>
          <Button
            title={t('assets.lockedCta')}
            style={{ marginTop: theme.spacing.md }}
            onPress={() => router.push('/(modals)/paywall')}
          />
        </Card>
      )}
    </Screen>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    healthSpotlight: {
      backgroundColor: theme.colors.surfaceElevated,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: theme.colors.border,
      paddingVertical: theme.spacing.md,
      paddingHorizontal: theme.spacing.md,
    },
    healthHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    healthEyebrow: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.4,
      textTransform: 'uppercase',
    },
    healthGaugeWrap: {
      alignItems: 'center',
      marginTop: theme.spacing.sm,
      marginBottom: theme.spacing.sm,
    },
    volatilityStrip: {
      marginTop: theme.spacing.sm,
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
    },
    volatilityHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    volatilityLabel: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.2,
    },
    volatilityRow: {
      marginTop: theme.spacing.sm,
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.sm,
    },
    volatilityValue: {
      fontSize: 18,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      minWidth: 42,
    },
    volatilityTrack: {
      flex: 1,
      height: 8,
      borderRadius: 999,
      backgroundColor: theme.colors.surface,
      borderWidth: 1,
      borderColor: theme.colors.border,
      overflow: 'visible',
      position: 'relative',
    },
    volatilityFill: {
      height: '100%',
      borderRadius: 999,
    },
    volatilityMarker: {
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
