import { useQuery } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { Alert, LayoutChangeEvent, StyleSheet, Text, View } from 'react-native'
import Svg, { Line as SvgLine, Text as SvgText } from 'react-native-svg'

import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { LineChart } from '../../src/components/LineChart'
import { Screen } from '../../src/components/Screen'
import { SeriesLine, StrategyCandlestickChart, TraceLine } from '../../src/components/StrategyCandlestickChart'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { dismissInsight, fetchInsights } from '../../src/services/insights'
import { fetchOHLCV } from '../../src/services/marketData'
import { fetchPortfolioSnapshots } from '../../src/services/portfolio'
import { fetchReportPlan } from '../../src/services/reports'
import { toneForSeverity } from '../../src/utils/badges'
import { formatCurrency, formatDateTime, formatNumber, formatPercent } from '../../src/utils/format'
import type { TranslationKey } from '../../src/utils/i18n'
import { getStrategyDisplayName } from '../../src/utils/strategies'

const rsiPeriod = 14
const bollingerPeriod = 20
const bollingerStdDev = 2

type IndicatorItem = {
  id: string
  label: string
  value?: string
  color?: string
}

type TranslateFn = (key: TranslationKey, params?: Record<string, string | number>) => string

export default function InsightDetailScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t, locale } = useLocalization()
  const { accessToken, userId } = useAuth()
  const params = useLocalSearchParams<{ id?: string }>()
  const insightId = params.id ?? ''

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
  const isMarketAlpha = insight?.type === 'market_alpha'

  const assetType = useMemo(() => {
    if (!insight?.asset_key) return undefined
    if (insight.asset_key.startsWith('crypto:')) return 'crypto'
    if (insight.asset_key.startsWith('stock:')) return 'stock'
    return undefined
  }, [insight?.asset_key])

  const interval = insight?.timeframe ? (insight.timeframe as '4h' | '1d') : assetType === 'crypto' ? '4h' : '1d'

  const displayDateRange = useMemo(() => {
    const endDate = new Date()
    const startDate = new Date()
    const daySpan = interval === '4h' ? 5 : 30
    startDate.setUTCDate(startDate.getUTCDate() - daySpan)
    const format = (value: Date) => value.toISOString().slice(0, 10)
    return {
      start: format(startDate),
      end: format(endDate),
      startDate,
      endDate,
    }
  }, [interval])

  const indicatorLookbackDays = useMemo(() => {
    const maxLookbackCandles = Math.max(rsiPeriod + 1, bollingerPeriod)
    const hoursPerCandle = interval === '4h' ? 4 : 24
    const baseDays = Math.ceil((maxLookbackCandles * hoursPerCandle) / 24)
    const tradingBufferDays = assetType === 'stock' && interval === '1d' ? 10 : 0
    return baseDays + tradingBufferDays
  }, [assetType, interval])

  const queryDateRange = useMemo(() => {
    const startDate = new Date(displayDateRange.startDate)
    startDate.setUTCDate(startDate.getUTCDate() - indicatorLookbackDays)
    const format = (value: Date) => value.toISOString().slice(0, 10)
    return {
      start: format(startDate),
      end: displayDateRange.end,
      startDate,
    }
  }, [displayDateRange.end, displayDateRange.startDate, indicatorLookbackDays])

  const chartQuery = useQuery({
    queryKey: ['ohlcv', userId, insight?.id, interval, queryDateRange.start, queryDateRange.end],
    queryFn: async () => {
      if (!accessToken || !insight) return null
      const resp = await fetchOHLCV(accessToken, {
        asset_key: insight.asset_key,
        asset_type: assetType,
        symbol: insight.asset,
        interval,
        start: queryDateRange.start,
        end: queryDateRange.end,
      })
      if (resp.error) throw new Error(resp.error.message)
      return resp.data ?? null
    },
    enabled: !!accessToken && !!insight,
  })

  const fullCandles = useMemo(() => {
    if (!chartQuery.data?.series) return []
    const toNumber = (value: unknown) => {
      if (typeof value === 'number' && Number.isFinite(value)) return value
      if (typeof value === 'string' && value.trim() !== '') {
        const parsed = Number(value)
        return Number.isFinite(parsed) ? parsed : null
      }
      return null
    }
    const toTime = (value: unknown) => {
      if (typeof value === 'string' && value.trim() !== '') return value
      if (typeof value === 'number' && Number.isFinite(value)) {
        return new Date(value).toISOString()
      }
      return null
    }
    const isCandle = (
      value: { open: number; high: number; low: number; close: number; time: string } | null
    ): value is { open: number; high: number; low: number; close: number; time: string } => value !== null
    return chartQuery.data.series
      .map((point: any) => {
        const raw = Array.isArray(point) ? point : []
        const time = toTime(raw[0] ?? point?.time ?? point?.timestamp ?? point?.date)
        const open = toNumber(raw[1] ?? point?.open)
        const high = toNumber(raw[2] ?? point?.high)
        const low = toNumber(raw[3] ?? point?.low)
        const close = toNumber(raw[4] ?? point?.close)
        if (!time || open === null || high === null || low === null || close === null) return null
        return { time, open, high, low, close }
      })
      .filter(isCandle)
  }, [chartQuery.data])

  const displaySlice = useMemo(() => {
    if (!fullCandles.length) {
      return { candles: [] as typeof fullCandles, startIndex: 0, endIndex: -1 }
    }
    const startMs = displayDateRange.startDate.getTime()
    const endMs = displayDateRange.endDate.getTime()
    let startIndex = 0
    while (startIndex < fullCandles.length && new Date(fullCandles[startIndex].time).getTime() < startMs) {
      startIndex += 1
    }
    let endIndex = fullCandles.length - 1
    while (endIndex >= startIndex && new Date(fullCandles[endIndex].time).getTime() > endMs) {
      endIndex -= 1
    }
    if (endIndex < startIndex) {
      return { candles: [] as typeof fullCandles, startIndex, endIndex }
    }
    return {
      candles: fullCandles.slice(startIndex, endIndex + 1),
      startIndex,
      endIndex,
    }
  }, [displayDateRange.endDate, displayDateRange.startDate, fullCandles])

  const displayCandles = displaySlice.candles
  const fullCloses = useMemo(() => fullCandles.map((candle) => candle.close), [fullCandles])
  const displayCloses = useMemo(() => displayCandles.map((candle) => candle.close), [displayCandles])
  const rsiSeries = useMemo(() => computeRSISeries(fullCloses, rsiPeriod), [fullCloses])
  const bollingerSeries = useMemo(
    () => computeBollingerSeries(fullCloses, bollingerPeriod, bollingerStdDev),
    [fullCloses]
  )
  const displayRsiSeries = useMemo(
    () => rsiSeries.slice(displaySlice.startIndex, displaySlice.endIndex + 1),
    [displaySlice.endIndex, displaySlice.startIndex, rsiSeries]
  )
  const displayBollingerSeries = useMemo(
    () => ({
      lower: bollingerSeries.lower.slice(displaySlice.startIndex, displaySlice.endIndex + 1),
      upper: bollingerSeries.upper.slice(displaySlice.startIndex, displaySlice.endIndex + 1),
    }),
    [bollingerSeries.lower, bollingerSeries.upper, displaySlice.endIndex, displaySlice.startIndex]
  )

  const planQuery = useQuery({
    queryKey: ['insight-plan', userId, insight?.plan_id],
    queryFn: async () => {
      if (!accessToken || !insight?.plan_id) return null
      const reportsResp = await fetchPortfolioSnapshots(accessToken)
      if (reportsResp.error) throw new Error(reportsResp.error.message)
      const activeReport = reportsResp.data?.items.find((item) => item.is_active)
      if (!activeReport?.calculation_id) return null
      const planResp = await fetchReportPlan(accessToken, activeReport.calculation_id, insight.plan_id)
      if (planResp.error) throw new Error(planResp.error.message)
      return planResp.data ?? null
    },
    enabled: !!accessToken && !!insight?.plan_id,
  })

  const linkedPlan = planQuery.data
  const quoteCurrency = linkedPlan?.quote_currency ?? chartQuery.data?.quote_currency ?? 'USD'
  const chartPadding = useMemo(
    () => ({
      top: 8,
      right: 56,
      bottom: 8,
      left: 12,
    }),
    []
  )

  const indicatorSeries = useMemo<SeriesLine[]>(() => {
    const hasBollinger =
      displayBollingerSeries.lower.some((value) => Number.isFinite(value)) ||
      displayBollingerSeries.upper.some((value) => Number.isFinite(value))
    if (!isMarketAlpha || !hasBollinger) return []
    return [
      {
        values: displayBollingerSeries.upper,
        color: theme.colors.muted,
        dashed: true,
        opacity: 0.45,
        strokeWidth: 1.2,
      },
      {
        values: displayBollingerSeries.lower,
        color: theme.colors.warning,
        dashed: true,
        opacity: 0.7,
        strokeWidth: 1.4,
      },
    ]
  }, [
    displayBollingerSeries.lower,
    displayBollingerSeries.upper,
    isMarketAlpha,
    theme.colors.muted,
    theme.colors.warning,
  ])

  const indicatorLines = useMemo<TraceLine[]>(() => {
    if (!insight || isMarketAlpha) return []
    return buildTriggerLines(insight.trigger_key, linkedPlan?.parameters ?? {}, theme.colors)
  }, [insight, isMarketAlpha, linkedPlan?.parameters, theme.colors])

  const indicatorItems = useMemo(() => {
    if (!insight) return [] as IndicatorItem[]

    const items: IndicatorItem[] = []

    if (insight.type === 'market_alpha') {
      const rsi = computeRSI(fullCloses, rsiPeriod)
      if (rsi !== null) {
        items.push({
          id: 'rsi',
          label: t('insight.indicator.rsi'),
          value: formatNumber(rsi, 1),
          color: theme.colors.accent,
        })
      }
      const bollinger = computeBollinger(fullCloses, bollingerPeriod, bollingerStdDev)
      if (bollinger && Number.isFinite(bollinger.upper)) {
        items.push({
          id: 'bollinger-upper',
          label: t('insight.indicator.bollingerUpper'),
          value: formatCurrency(bollinger.upper, quoteCurrency),
          color: theme.colors.muted,
        })
      }
      if (bollinger && Number.isFinite(bollinger.lower)) {
        items.push({
          id: 'bollinger-lower',
          label: t('insight.indicator.bollingerLower'),
          value: formatCurrency(bollinger.lower, quoteCurrency),
          color: theme.colors.warning,
        })
      }
      const lastClose = displayCloses.length ? displayCloses[displayCloses.length - 1] : null
      if (lastClose !== null) {
        items.push({
          id: 'last-close',
          label: t('insight.indicator.close'),
          value: formatCurrency(lastClose, quoteCurrency),
        })
      }
      if (items.length === 0 && insight.trigger_reason) {
        items.push({ id: 'trigger', label: t('insight.indicator.condition'), value: insight.trigger_reason })
      }
      return items
    }

    const triggerItems = buildTriggerIndicators(insight.trigger_key, linkedPlan?.parameters ?? {}, quoteCurrency, t)
    if (triggerItems.length > 0) {
      return triggerItems.map((item) => ({
        ...item,
        color: indicatorColorForId(item.id, theme.colors),
      }))
    }
    if (insight.trigger_reason) {
      return [{ id: 'trigger', label: t('insight.indicator.condition'), value: insight.trigger_reason }]
    }
    return items
  }, [displayCloses, fullCloses, insight, linkedPlan?.parameters, quoteCurrency, t, theme.colors])

  const handleDismiss = async () => {
    if (!accessToken || !insight) return
    const resp = await dismissInsight(accessToken, insight.id, 'later')
    if (resp.error) {
      Alert.alert(t('insight.dismissFailTitle'), resp.error.message)
      return
    }
    router.back()
  }

  if (!insight) {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('insight.loading')}</Text>
      </Screen>
    )
  }

  const styles = makeStyles(theme)
  const indicatorCount = indicatorItems.length
  const showRsiChart = Boolean(isMarketAlpha && displayRsiSeries.some((value) => Number.isFinite(value)))

  return (
    <Screen scroll>
      <Text style={styles.pageTitle}>{t('insight.title')}</Text>

      <Card style={styles.summaryCard}>
        <View style={styles.summaryHeader}>
          <Text style={styles.assetLabel}>{insight.asset}</Text>
          <Badge label={insight.severity} tone={toneForSeverity(insight.severity)} />
        </View>
        {insight.timeframe ? (
          <Text style={styles.timeframeText}>{t('insight.timeframe', { timeframe: insight.timeframe })}</Text>
        ) : null}
        <Text style={styles.triggerText}>{insight.trigger_reason}</Text>
        {insight.suggested_action ? (
          <Text style={styles.suggestedText}>
            {t('insights.suggestedLabel')}: {insight.suggested_action}
          </Text>
        ) : null}
        <View style={styles.metaRow}>
          <Text style={styles.metaText}>{t('insight.createdAt', { time: formatDateTime(insight.created_at) })}</Text>
          <Text style={styles.metaText}>{t('insight.expiresAt', { time: formatDateTime(insight.expires_at) })}</Text>
        </View>
      </Card>

      {displayCandles.length > 1 ? (
        <View style={styles.chartPanel}>
          <StrategyCandlestickChart
            candles={displayCandles}
            lines={indicatorLines}
            series={indicatorSeries}
            locale={locale}
            formatPrice={(value) => formatCurrency(value, quoteCurrency)}
            axisSide="right"
            showXAxis={false}
            showYAxis
            padding={chartPadding}
            backgroundColor="transparent"
          />
          {showRsiChart ? (
            <View style={styles.rsiPanel}>
              <View style={styles.chartDivider} />
              <LineChart
                data={displayRsiSeries}
                height={64}
                minValue={0}
                maxValue={100}
                strokeColor={theme.colors.accent}
                opacity={0.9}
                strokeWidth={1.6}
                axisSide="right"
                showYAxis
                showXAxis={false}
                padding={chartPadding}
                yTicks={[100, 70, 30, 0]}
                formatY={(value) => formatNumber(value, 0)}
                alignToCandles
                horizontalLines={[
                  { value: 70, color: theme.colors.border, dashed: true, opacity: 0.6 },
                  { value: 30, color: theme.colors.border, dashed: true, opacity: 0.6 },
                ]}
              />
            </View>
          ) : null}
          <ChartXAxis candles={displayCandles} locale={locale} padding={chartPadding} />
          {indicatorCount > 0 ? (
            <View style={styles.indicatorShelf}>
              <View style={styles.indicatorList}>
                {indicatorItems.map((item, index) => {
                  const isLast = index === indicatorCount - 1
                  return (
                    <View key={item.id} style={[styles.indicatorRow, !isLast && styles.indicatorRowDivider]}>
                      <View style={styles.indicatorLabelWrap}>
                        {item.color ? <View style={[styles.indicatorDot, { backgroundColor: item.color }]} /> : null}
                        <Text style={styles.indicatorLabel}>{item.label}</Text>
                      </View>
                      {item.value ? <Text style={styles.indicatorValue}>{item.value}</Text> : null}
                    </View>
                  )
                })}
              </View>
            </View>
          ) : null}
        </View>
      ) : null}

      {linkedPlan ? (
        <Card style={styles.planCard}>
          <Text style={styles.planTitle}>{t('insight.planTitle')}</Text>
          <Text style={styles.planName}>{getStrategyDisplayName(t, linkedPlan.strategy_id)}</Text>
          <Text style={styles.planBody}>{linkedPlan.rationale}</Text>
          <Text style={styles.planBody}>{linkedPlan.expected_outcome}</Text>
        </Card>
      ) : null}

      {isMarketAlpha ? (
        <Button title={t('insight.dismissCta')} style={styles.primaryCta} onPress={handleDismiss} />
      ) : (
        <Button
          title={t('insight.executedCta')}
          style={styles.primaryCta}
          onPress={() => router.push({ pathname: '/(modals)/quick-update', params: { id: insightId } })}
        />
      )}

      {!isMarketAlpha ? (
        <Button title={t('insight.dismissCta')} variant="ghost" style={styles.secondaryCta} onPress={handleDismiss} />
      ) : null}
    </Screen>
  )
}

function makeStyles(theme: ReturnType<typeof useTheme>) {
  return StyleSheet.create({
    pageTitle: {
      fontSize: 24,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    summaryCard: {
      marginTop: theme.spacing.md,
      gap: theme.spacing.xs,
    },
    summaryHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    assetLabel: {
      fontSize: 18,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
    },
    timeframeText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
    },
    triggerText: {
      marginTop: theme.spacing.xs,
      color: theme.colors.ink,
      fontFamily: theme.fonts.body,
      fontSize: 14,
      lineHeight: 20,
    },
    suggestedText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
      lineHeight: 18,
    },
    metaRow: {
      marginTop: theme.spacing.xs,
      gap: 2,
    },
    metaText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    chartPanel: {
      marginTop: theme.spacing.lg,
      backgroundColor: theme.colors.surfaceElevated,
      borderRadius: theme.radius.lg,
      padding: theme.spacing.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      gap: theme.spacing.xs,
    },
    rsiPanel: {
      gap: theme.spacing.xs,
    },
    chartDivider: {
      height: 1,
      backgroundColor: theme.colors.border,
    },
    indicatorShelf: {
      gap: theme.spacing.sm,
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
    },
    indicatorList: {
      borderRadius: theme.radius.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surface,
      overflow: 'hidden',
    },
    indicatorRow: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      justifyContent: 'space-between',
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.sm,
      gap: theme.spacing.sm,
    },
    indicatorLabelWrap: {
      flex: 1,
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.xs,
    },
    indicatorDot: {
      width: 6,
      height: 6,
      borderRadius: 999,
    },
    indicatorRowDivider: {
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.border,
    },
    indicatorLabel: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 13,
    },
    indicatorValue: {
      flexShrink: 1,
      textAlign: 'right',
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
      lineHeight: 18,
    },
    planCard: {
      marginTop: theme.spacing.lg,
      gap: theme.spacing.xs,
    },
    planTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
    },
    planName: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 13,
    },
    planBody: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
      lineHeight: 19,
    },
    primaryCta: {
      marginTop: theme.spacing.lg,
    },
    secondaryCta: {
      marginTop: theme.spacing.sm,
    },
  })
}

function computeRSI(closes: number[], period: number) {
  if (closes.length < period + 1) return null
  let gains = 0
  let losses = 0
  for (let i = 1; i <= period; i += 1) {
    const delta = closes[i] - closes[i - 1]
    if (delta >= 0) {
      gains += delta
    } else {
      losses -= delta
    }
  }
  let avgGain = gains / period
  let avgLoss = losses / period
  for (let i = period + 1; i < closes.length; i += 1) {
    const delta = closes[i] - closes[i - 1]
    const gain = delta > 0 ? delta : 0
    const loss = delta < 0 ? -delta : 0
    avgGain = (avgGain * (period - 1) + gain) / period
    avgLoss = (avgLoss * (period - 1) + loss) / period
  }
  if (avgLoss === 0) return 100
  const rs = avgGain / avgLoss
  return 100 - 100 / (1 + rs)
}

function computeBollinger(closes: number[], period: number, stdDev: number) {
  if (closes.length < period) return null
  const slice = closes.slice(-period)
  const mean = slice.reduce((sum, value) => sum + value, 0) / period
  const variance = slice.reduce((sum, value) => sum + (value - mean) ** 2, 0) / period
  const deviation = Math.sqrt(variance)
  return {
    middle: mean,
    upper: mean + stdDev * deviation,
    lower: mean - stdDev * deviation,
  }
}

function computeRSISeries(closes: number[], period: number) {
  if (closes.length < period + 1) return []
  const series = Array(closes.length).fill(Number.NaN)
  let gains = 0
  let losses = 0
  for (let i = 1; i <= period; i += 1) {
    const delta = closes[i] - closes[i - 1]
    if (delta >= 0) {
      gains += delta
    } else {
      losses -= delta
    }
  }
  let avgGain = gains / period
  let avgLoss = losses / period
  const rsInitial = avgLoss === 0 ? 0 : avgGain / avgLoss
  series[period] = avgLoss === 0 ? 100 : 100 - 100 / (1 + rsInitial)
  for (let i = period + 1; i < closes.length; i += 1) {
    const delta = closes[i] - closes[i - 1]
    const gain = delta > 0 ? delta : 0
    const loss = delta < 0 ? -delta : 0
    avgGain = (avgGain * (period - 1) + gain) / period
    avgLoss = (avgLoss * (period - 1) + loss) / period
    if (avgLoss === 0) {
      series[i] = 100
    } else {
      const rs = avgGain / avgLoss
      series[i] = 100 - 100 / (1 + rs)
    }
  }
  return series
}

function computeBollingerSeries(closes: number[], period: number, stdDev: number) {
  if (closes.length < period) return { lower: [], upper: [] }
  const lower = Array(closes.length).fill(Number.NaN)
  const upper = Array(closes.length).fill(Number.NaN)
  for (let i = period - 1; i < closes.length; i += 1) {
    const slice = closes.slice(i - period + 1, i + 1)
    const mean = slice.reduce((sum, value) => sum + value, 0) / period
    const variance = slice.reduce((sum, value) => sum + (value - mean) ** 2, 0) / period
    const deviation = Math.sqrt(variance)
    lower[i] = mean - stdDev * deviation
    upper[i] = mean + stdDev * deviation
  }
  return { lower, upper }
}

function buildTriggerIndicators(
  triggerKey: string | undefined,
  parameters: Record<string, any>,
  currency: string,
  t: TranslateFn
) {
  if (!triggerKey) return [] as IndicatorItem[]
  const parts = triggerKey.split(':')
  if (parts.length < 4) return [] as IndicatorItem[]
  const thresholdId = parts.slice(3).join(':')
  const items: IndicatorItem[] = []

  const add = (id: string, label: string, value?: string) => {
    items.push({ id, label, value })
  }

  const getNumber = (key: string) => {
    const raw = parameters?.[key]
    return typeof raw === 'number' && Number.isFinite(raw) ? raw : null
  }

  const formatMaybePercent = (value: number | null) => (value === null ? undefined : formatPercent(value, 1))

  if (thresholdId === 'stop_loss') {
    const stopPrice = getNumber('stop_loss_price') ?? getNumber('support_level')
    const stopPct = getNumber('stop_loss_pct')
    add(
      'stop-loss',
      t('insight.indicator.stopLoss'),
      stopPrice !== null
        ? formatCurrency(stopPrice, currency)
        : stopPct !== null
          ? formatPercent(stopPct, 1)
          : undefined
    )
    return items
  }

  if (thresholdId.startsWith('layer_')) {
    const idx = Number(thresholdId.replace('layer_', ''))
    const layers = Array.isArray(parameters?.layers) ? parameters.layers : []
    const layer = Number.isFinite(idx) ? layers[idx - 1] : null
    const targetPrice = typeof layer?.target_price === 'number' ? layer.target_price : null
    add(
      `take-profit-${idx}`,
      t('insight.indicator.takeProfit', { index: idx }),
      targetPrice !== null ? formatCurrency(targetPrice, currency) : undefined
    )
    return items
  }

  if (thresholdId.startsWith('safety_')) {
    const idx = Number(thresholdId.replace('safety_', ''))
    add(`safety-${idx}`, t('insight.indicator.safetyOrder', { index: idx }))
    return items
  }

  if (thresholdId === 'trailing_stop') {
    const trailingPct = getNumber('trailing_stop_pct') ?? getNumber('callback_rate')
    add('trailing-stop', t('insight.indicator.trailingStop'), formatMaybePercent(trailingPct))
    return items
  }

  if (thresholdId.startsWith('execution_')) {
    const amount = getNumber('amount')
    add('dca', t('insight.indicator.dcaExecution'), amount !== null ? formatCurrency(amount, currency) : undefined)
    return items
  }

  if (thresholdId.startsWith('addition_')) {
    const idx = Number(thresholdId.replace('addition_', ''))
    add(`addition-${idx}`, t('insight.indicator.addOn', { index: idx }))
    return items
  }

  if (thresholdId.startsWith('funding_')) {
    const triggerRate = getNumber('trigger_funding_rate')
    const basisMax = getNumber('trigger_basis_pct_max')
    const value = triggerRate !== null ? formatMaybePercent(triggerRate) : formatMaybePercent(basisMax)
    add('funding', t('insight.indicator.funding'), value)
    return items
  }

  if (thresholdId.startsWith('trend_')) {
    const segments = thresholdId.split('_')
    const trend = segments.length >= 2 ? segments[1] : ''
    const trendLabel = trend ? trend.charAt(0).toUpperCase() + trend.slice(1) : undefined
    add('trend', t('insight.indicator.trendShift'), trendLabel)
    return items
  }

  if (thresholdId.startsWith('rebalance_')) {
    const threshold = getNumber('rebalance_threshold_pct')
    add('rebalance', t('insight.indicator.rebalance'), formatMaybePercent(threshold))
    return items
  }

  return items
}

function buildTriggerLines(
  triggerKey: string | undefined,
  parameters: Record<string, any>,
  colors: { danger: string; warning: string; accent: string }
) {
  if (!triggerKey) return [] as TraceLine[]
  const parts = triggerKey.split(':')
  if (parts.length < 4) return [] as TraceLine[]
  const thresholdId = parts.slice(3).join(':')
  const lines: TraceLine[] = []

  const getNumber = (key: string) => {
    const raw = parameters?.[key]
    if (typeof raw === 'number' && Number.isFinite(raw)) return raw
    if (typeof raw === 'string' && raw.trim() !== '') {
      const parsed = Number(raw)
      return Number.isFinite(parsed) ? parsed : null
    }
    return null
  }

  const addLine = (value: number | null, color: string, dashed = false) => {
    if (value === null) return
    lines.push({ value, color, dashed })
  }

  if (thresholdId === 'stop_loss') {
    addLine(getNumber('stop_loss_price'), colors.danger)
    addLine(getNumber('support_level'), colors.warning, true)
    return lines
  }

  if (thresholdId.startsWith('layer_')) {
    const idx = Number(thresholdId.replace('layer_', ''))
    const layers = Array.isArray(parameters?.layers) ? parameters.layers : []
    const layer = Number.isFinite(idx) ? layers[idx - 1] : null
    const targetRaw = layer?.target_price
    const targetPrice =
      typeof targetRaw === 'number'
        ? targetRaw
        : typeof targetRaw === 'string' && targetRaw.trim() !== ''
          ? Number(targetRaw)
          : null
    const safeTargetPrice = targetPrice !== null && Number.isFinite(targetPrice) ? targetPrice : null
    addLine(safeTargetPrice, colors.accent)
    return lines
  }

  if (thresholdId === 'trailing_stop') {
    const activationPrice = getNumber('activation_price')
    const trailingPct = getNumber('trailing_stop_pct') ?? getNumber('callback_rate')
    const initialStop =
      getNumber('initial_trailing_stop_price') ??
      (activationPrice !== null && trailingPct !== null ? activationPrice * (1 - trailingPct) : null)
    addLine(activationPrice, colors.accent)
    addLine(initialStop, colors.warning, true)
    return lines
  }

  return lines
}

function indicatorColorForId(
  id: string,
  colors: { accent: string; warning: string; danger: string }
): string | undefined {
  if (id.startsWith('stop-loss')) return colors.danger
  if (id.startsWith('take-profit')) return colors.accent
  if (id.startsWith('trailing-stop')) return colors.warning
  if (id.startsWith('dca')) return colors.accent
  if (id.startsWith('funding')) return colors.warning
  if (id.startsWith('rebalance')) return colors.accent
  if (id.startsWith('trend')) return colors.accent
  return undefined
}

function ChartXAxis({
  candles,
  locale,
  padding,
  height = 18,
}: {
  candles: { time: string | number }[]
  locale: string
  padding: { left: number; right: number }
  height?: number
}) {
  const theme = useTheme()
  const [width, setWidth] = useState(0)

  const dateFormatter = useMemo(() => new Intl.DateTimeFormat(locale, { month: 'short', day: 'numeric' }), [locale])

  const labelIndices = useMemo(() => {
    if (!candles.length) return []
    const lastIndex = candles.length - 1
    const midIndex = Math.floor(candles.length / 2)
    return Array.from(new Set([0, midIndex, lastIndex])).filter((idx) => idx >= 0)
  }, [candles.length])

  const handleLayout = (event: LayoutChangeEvent) => {
    setWidth(event.nativeEvent.layout.width)
  }

  if (candles.length < 2) return null

  const chartWidth = Math.max(1, width - padding.left - padding.right)
  const step = candles.length > 0 ? chartWidth / candles.length : chartWidth
  const axisY = 2

  return (
    <View style={{ height }} onLayout={handleLayout}>
      {width > 0 ? (
        <Svg width={width} height={height}>
          <SvgLine
            x1={padding.left}
            x2={padding.left + chartWidth}
            y1={axisY}
            y2={axisY}
            stroke={theme.colors.border}
            strokeWidth={1}
          />
          {labelIndices.map((idx) => {
            const candle = candles[idx]
            const x = padding.left + idx * step + step / 2
            const label = dateFormatter.format(new Date(candle.time))
            return (
              <SvgText
                key={`x-${idx}`}
                x={x}
                y={height - 4}
                fontSize={10}
                fill={theme.colors.muted}
                fontFamily={theme.fonts.bodyMedium}
                textAnchor="middle"
              >
                {label}
              </SvgText>
            )
          })}
        </Svg>
      ) : null}
    </View>
  )
}
