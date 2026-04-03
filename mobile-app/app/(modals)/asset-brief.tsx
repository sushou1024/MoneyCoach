import { useQuery } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useMemo } from 'react'
import { Pressable, StyleSheet, Text, View } from 'react-native'

import { AssetLogo } from '../../src/components/AssetLogo'
import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { ShimmerBlock } from '../../src/components/ShimmerBlock'
import { StrategyCandlestickChart, TraceLine } from '../../src/components/StrategyCandlestickChart'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchAssetBrief } from '../../src/services/intelligence'
import { fetchOHLCV } from '../../src/services/marketData'
import { toneForSeverity } from '../../src/utils/badges'
import { formatCurrency, formatDateTime, formatNumber, formatPercent } from '../../src/utils/format'
import {
  buildChartDateRange,
  chartIntervalForAsset,
  formatSignedPercent,
  resolveActionBiasLabel,
  resolveConcentrationImpactLabel,
  resolveEntryBasisLabel,
  resolveInvalidationReasonLabel,
  resolvePortfolioRoleLabel,
  resolveRiskFlagLabel,
  resolveSummarySignalLabel,
  resolveTrendStateLabel,
  resolveTrendStrengthLabel,
  shouldShowAssetBriefLoading,
  toneForActionBias,
  toneForRiskFlag,
} from '../../src/utils/intelligence'
import { getStrategyDisplayName } from '../../src/utils/strategies'
import { buildChartCandles } from '../../src/utils/strategyDetail'

export default function AssetBriefScreen() {
  const theme = useTheme()
  const styles = useMemo(() => createStyles(theme), [theme])
  const router = useRouter()
  const { t, locale } = useLocalization()
  const { accessToken, clearSession, userId, isLoading: authLoading } = useAuth()
  const params = useLocalSearchParams<{ asset_key?: string }>()
  const assetKey = Array.isArray(params.asset_key) ? (params.asset_key[0] ?? '') : (params.asset_key ?? '')

  const briefQuery = useQuery({
    queryKey: ['asset-brief', userId, assetKey],
    queryFn: async () => {
      if (!accessToken || !assetKey) return null
      const resp = await fetchAssetBrief(accessToken, assetKey)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return null
        }
        if (resp.status === 404) return null
        throw new Error(resp.error.message)
      }
      return resp.data ?? null
    },
    enabled: !authLoading && !!accessToken && !!assetKey,
  })

  const showLoadingState = shouldShowAssetBriefLoading({
    authLoading,
    hasAccessToken: !!accessToken,
    hasAssetKey: !!assetKey,
    isFetched: briefQuery.isFetched,
    isFetching: briefQuery.isFetching,
    isPending: briefQuery.isPending,
  })

  const brief = briefQuery.data
  const chartInterval = brief ? chartIntervalForAsset(brief.asset_type) : '1d'
  const chartDateRange = useMemo(() => buildChartDateRange(chartInterval), [chartInterval])
  const chartAssetType = brief?.asset_type === 'crypto' || brief?.asset_type === 'stock' ? brief.asset_type : undefined

  const chartQuery = useQuery({
    queryKey: ['asset-brief-chart', userId, brief?.asset_key, chartInterval, chartDateRange.start, chartDateRange.end],
    queryFn: async () => {
      if (!accessToken || !brief) return null
      const resp = await fetchOHLCV(accessToken, {
        asset_key: brief.asset_key,
        asset_type: chartAssetType,
        symbol: brief.symbol,
        interval: chartInterval,
        start: chartDateRange.start,
        end: chartDateRange.end,
      })
      if (resp.error) throw new Error(resp.error.message)
      return resp.data ?? null
    },
    enabled: !!accessToken && !!brief,
  })

  const candles = useMemo(() => buildChartCandles(chartQuery.data?.series ?? []), [chartQuery.data?.series])
  const chartLines = useMemo<TraceLine[]>(() => {
    if (!brief) return []
    return [
      { value: brief.current_price, color: theme.colors.accent },
      { value: brief.entry_zone.low, color: theme.colors.success, dashed: true },
      { value: brief.entry_zone.high, color: theme.colors.success, dashed: true },
      { value: brief.invalidation.price, color: theme.colors.danger, dashed: true },
    ]
  }, [brief, theme.colors.accent, theme.colors.danger, theme.colors.success])

  const currentPriceCard = brief
    ? {
        key: 'current',
        label: t('intelligence.assetBrief.currentPrice'),
        value: formatCurrency(brief.current_price, brief.quote_currency),
      }
    : null

  const changeCards = brief
    ? [
        {
          key: '24h',
          label: t('intelligence.assetBrief.change24h'),
          value: formatSignedPercent(brief.price_change_24h),
        },
        { key: '7d', label: t('intelligence.assetBrief.change7d'), value: formatSignedPercent(brief.price_change_7d) },
        {
          key: '30d',
          label: t('intelligence.assetBrief.change30d'),
          value: formatSignedPercent(brief.price_change_30d),
        },
      ]
    : []

  const technicalItems = brief
    ? [
        { key: 'rsi', label: t('intelligence.technical.rsi'), value: formatNumber(brief.technicals.rsi_14, 1) },
        {
          key: 'trend-state',
          label: t('intelligence.technical.trendState'),
          value: resolveTrendStateLabel(t, brief.technicals.trend_state),
        },
        {
          key: 'trend-strength',
          label: t('intelligence.technical.trendStrength'),
          value: resolveTrendStrengthLabel(t, brief.technicals.trend_strength),
        },
        {
          key: 'ma20',
          label: t('intelligence.technical.ma20'),
          value: formatCurrency(brief.technicals.ma_20, brief.quote_currency),
        },
        {
          key: 'ma50',
          label: t('intelligence.technical.ma50'),
          value: formatCurrency(brief.technicals.ma_50, brief.quote_currency),
        },
        {
          key: 'ma200',
          label: t('intelligence.technical.ma200'),
          value: formatCurrency(brief.technicals.ma_200, brief.quote_currency),
        },
        {
          key: 'bb-upper',
          label: t('intelligence.technical.bollingerUpper'),
          value: formatCurrency(brief.technicals.bollinger_upper, brief.quote_currency),
        },
        {
          key: 'bb-lower',
          label: t('intelligence.technical.bollingerLower'),
          value: formatCurrency(brief.technicals.bollinger_lower, brief.quote_currency),
        },
      ]
    : []

  const openStrategy = (planId: string, calculationId: string) => {
    router.push({ pathname: '/(modals)/strategy', params: { plan_id: planId, calculation_id: calculationId } })
  }

  const openInsight = (insightId: string, type: string) => {
    if (type === 'market_alpha') {
      router.push({ pathname: '/(modals)/insight-detail', params: { id: insightId } })
      return
    }
    router.push({ pathname: '/(modals)/quick-update', params: { id: insightId } })
  }

  const resolveSeverityLabel = (severity?: string) => {
    const normalized = (severity ?? '').toLowerCase()
    const key = `insights.severity.${normalized || 'unknown'}` as const
    const value = t(key as any)
    return value === key ? severity || t('insights.severity.unknown') : value
  }

  return (
    <Screen scroll>
      <View style={styles.header}>
        <View style={{ flex: 1 }}>
          <Text style={styles.title}>{t('intelligence.assetBrief.title')}</Text>
        </View>
        <Pressable onPress={() => router.back()} accessibilityRole="button">
          <Text style={styles.close}>{t('common.close')}</Text>
        </Pressable>
      </View>

      {showLoadingState ? (
        <View style={styles.stack}>
          <Card>
            <ShimmerBlock width="38%" height={18} radius={8} />
            <ShimmerBlock width="62%" height={12} radius={6} style={{ marginTop: theme.spacing.sm }} />
            <View style={styles.grid}>
              {Array.from({ length: 4 }).map((_, index) => (
                <View key={`price-${index}`} style={styles.metricCard}>
                  <ShimmerBlock width="60%" height={10} radius={6} />
                  <ShimmerBlock width="52%" height={16} radius={8} style={{ marginTop: theme.spacing.sm }} />
                </View>
              ))}
            </View>
          </Card>
        </View>
      ) : !brief ? (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={styles.emptyText}>{t('insights.featuredEmpty')}</Text>
        </Card>
      ) : (
        <View style={styles.stack}>
          <Card>
            <View style={styles.assetHeader}>
              <View style={styles.assetMeta}>
                <AssetLogo uri={brief.logo_url} label={brief.symbol} size={48} />
                <View style={{ flex: 1 }}>
                  <Text style={styles.assetSymbol}>{brief.symbol}</Text>
                  <Text style={styles.assetName}>
                    {brief.name ?? brief.asset_type}
                    {brief.exchange_mic ? ` · ${brief.exchange_mic}` : ''}
                  </Text>
                  <Text style={styles.timestamp}>{formatDateTime(brief.as_of)}</Text>
                </View>
              </View>
              <Badge label={resolveActionBiasLabel(t, brief.action_bias)} tone={toneForActionBias(brief.action_bias)} />
            </View>
            <Text style={styles.summarySignal}>{resolveSummarySignalLabel(t, brief.summary_signal)}</Text>
            {currentPriceCard ? (
              <View style={styles.grid}>
                <View style={[styles.metricCard, { width: '100%' }]}>
                  <Text style={styles.metricLabel}>{currentPriceCard.label}</Text>
                  <Text style={styles.metricValue}>{currentPriceCard.value}</Text>
                </View>
              </View>
            ) : null}
            {changeCards.length > 0 ? (
              <>
                <Text style={[styles.sectionTitle, { marginTop: theme.spacing.sm, fontSize: 14 }]}>
                  {t('intelligence.assetBrief.priceChanges')}
                </Text>
                <View style={{ marginTop: theme.spacing.xs, flexDirection: 'row', gap: theme.spacing.sm }}>
                  {changeCards.map((item) => (
                    <View key={item.key} style={[styles.metricCard, { flex: 1, minWidth: 0 }]}>
                      <Text style={styles.metricLabel}>{item.label}</Text>
                      <Text style={styles.metricValue}>{item.value}</Text>
                    </View>
                  ))}
                </View>
              </>
            ) : null}
          </Card>

          {candles.length > 1 ? (
            <Card>
              <Text style={styles.sectionTitle}>{t('intelligence.assetBrief.chartTitle')}</Text>
              <View style={{ marginTop: theme.spacing.sm }}>
                <StrategyCandlestickChart
                  candles={candles}
                  lines={chartLines}
                  height={204}
                  locale={locale}
                  formatPrice={(value) => formatCurrency(value, brief.quote_currency)}
                  axisSide="right"
                  backgroundColor={theme.colors.surfaceElevated}
                />
              </View>
            </Card>
          ) : null}

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.assetBrief.entryZone')}</Text>
            <View style={styles.grid}>
              <MetricTile
                label={resolveEntryBasisLabel(t, brief.entry_zone.basis)}
                value={`${formatCurrency(brief.entry_zone.low, brief.quote_currency)} - ${formatCurrency(brief.entry_zone.high, brief.quote_currency)}`}
              />
              <MetricTile
                label={t('intelligence.assetBrief.invalidation')}
                value={formatCurrency(brief.invalidation.price, brief.quote_currency)}
                detail={resolveInvalidationReasonLabel(t, brief.invalidation.reason)}
              />
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.assetBrief.technicals')}</Text>
            <View style={styles.grid}>
              {technicalItems.map((item) => (
                <MetricTile key={item.key} label={item.label} value={item.value} />
              ))}
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.assetBrief.portfolioFit')}</Text>
            <View style={styles.grid}>
              <MetricTile
                label={t('intelligence.assetBrief.positionWeight')}
                value={brief.portfolio_fit.is_held ? formatPercent(brief.portfolio_fit.weight_pct, 0) : '0%'}
              />
              <MetricTile
                label={t('intelligence.featured.beta')}
                value={formatNumber(brief.portfolio_fit.beta_to_portfolio, 2)}
              />
              <MetricTile
                label={t('intelligence.assetBrief.portfolioRole')}
                value={resolvePortfolioRoleLabel(t, brief.portfolio_fit.role)}
              />
              <MetricTile
                label={t('intelligence.assetBrief.concentrationImpact')}
                value={resolveConcentrationImpactLabel(t, brief.portfolio_fit.concentration_impact)}
              />
            </View>
            <View style={styles.fitFooter}>
              <Text style={styles.fitLabel}>{t('intelligence.assetBrief.riskFlag')}</Text>
              <Badge
                label={resolveRiskFlagLabel(t, brief.portfolio_fit.risk_flag)}
                tone={toneForRiskFlag(brief.portfolio_fit.risk_flag)}
              />
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.assetBrief.relatedPlans')}</Text>
            <View style={styles.listWrap}>
              {brief.related_plans.length === 0 ? (
                <Text style={styles.emptyText}>{t('intelligence.assetBrief.noLinks')}</Text>
              ) : (
                brief.related_plans.map((plan) => (
                  <View key={plan.plan_id} style={styles.linkCard}>
                    <View style={styles.linkHeader}>
                      <Text style={styles.linkTitle}>{getStrategyDisplayName(t, plan.strategy_id)}</Text>
                      <Badge label={resolveSeverityLabel(plan.priority)} tone={toneForSeverity(plan.priority)} />
                    </View>
                    <Text style={styles.linkBody}>{plan.rationale}</Text>
                    <Button
                      title={t('intelligence.assetBrief.openStrategy')}
                      variant="secondary"
                      style={{ marginTop: theme.spacing.sm }}
                      onPress={() => openStrategy(plan.plan_id, plan.calculation_id)}
                    />
                  </View>
                ))
              )}
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.assetBrief.relatedInsights')}</Text>
            <View style={styles.listWrap}>
              {brief.related_insights.length === 0 ? (
                <Text style={styles.emptyText}>{t('intelligence.assetBrief.noLinks')}</Text>
              ) : (
                brief.related_insights.map((insight) => {
                  const typeKey = `insight.type.${insight.type}` as any
                  const typeLabel = t(typeKey)
                  return (
                    <View key={insight.id} style={styles.linkCard}>
                      <View style={styles.linkHeader}>
                        <Text style={styles.linkTitle}>{typeLabel === typeKey ? insight.type : typeLabel}</Text>
                        <Badge
                          label={resolveSeverityLabel(insight.severity)}
                          tone={toneForSeverity(insight.severity)}
                        />
                      </View>
                      <Text style={styles.linkBody}>{insight.trigger_reason}</Text>
                      <Button
                        title={t('intelligence.assetBrief.openSignal')}
                        variant="secondary"
                        style={{ marginTop: theme.spacing.sm }}
                        onPress={() => openInsight(insight.id, insight.type)}
                      />
                    </View>
                  )
                })
              )}
            </View>
          </Card>
        </View>
      )}
    </Screen>
  )
}

function MetricTile({ label, value, detail }: { label: string; value: string; detail?: string }) {
  const theme = useTheme()
  const styles = useMemo(() => createStyles(theme), [theme])
  return (
    <View style={styles.metricCard}>
      <Text style={styles.metricLabel}>{label}</Text>
      <Text style={styles.metricValue}>{value}</Text>
      {detail ? <Text style={styles.metricDetail}>{detail}</Text> : null}
    </View>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    header: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      gap: theme.spacing.md,
    },
    title: {
      fontSize: 28,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    subtitle: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    close: {
      color: theme.colors.accent,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
      marginTop: 4,
    },
    stack: {
      marginTop: theme.spacing.lg,
      gap: theme.spacing.sm,
    },
    assetHeader: {
      flexDirection: 'row',
      justifyContent: 'space-between',
      alignItems: 'flex-start',
      gap: theme.spacing.sm,
    },
    assetMeta: {
      flexDirection: 'row',
      flex: 1,
      gap: theme.spacing.sm,
      alignItems: 'center',
    },
    assetSymbol: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 24,
    },
    assetName: {
      marginTop: 2,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    timestamp: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    summarySignal: {
      marginTop: theme.spacing.sm,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 16,
      lineHeight: 24,
    },
    sectionTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 18,
    },
    grid: {
      marginTop: theme.spacing.sm,
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: theme.spacing.sm,
    },
    metricCard: {
      width: '47%',
      minWidth: 110,
      borderRadius: theme.radius.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
      padding: theme.spacing.sm,
    },
    metricLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
    },
    metricValue: {
      marginTop: theme.spacing.xs,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 18,
    },
    metricDetail: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
      lineHeight: 18,
    },
    listWrap: {
      marginTop: theme.spacing.sm,
      gap: theme.spacing.sm,
    },
    bulletRow: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      gap: theme.spacing.sm,
    },
    bullet: {
      width: 8,
      height: 8,
      borderRadius: 999,
      marginTop: 7,
      backgroundColor: theme.colors.accent,
    },
    linkCard: {
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
      borderRadius: theme.radius.md,
      padding: theme.spacing.sm,
    },
    linkHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    linkTitle: {
      flex: 1,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 15,
    },
    linkBody: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
      flex: 1,
      marginTop: 0,
    },
    fitFooter: {
      marginTop: theme.spacing.md,
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    fitLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
    },
    emptyText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
    },
  })
