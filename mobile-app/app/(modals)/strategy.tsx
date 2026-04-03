import { Ionicons } from '@expo/vector-icons'
import { useQuery } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useCallback, useMemo, useState } from 'react'
import { Alert, Pressable, StyleSheet, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { DonutChart } from '../../src/components/DonutChart'
import { Screen } from '../../src/components/Screen'
import { ShimmerBlock } from '../../src/components/ShimmerBlock'
import { StrategyCandlestickChart, TraceLine } from '../../src/components/StrategyCandlestickChart'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchInsights } from '../../src/services/insights'
import { fetchOHLCV } from '../../src/services/marketData'
import { fetchReport, fetchReportPlan } from '../../src/services/reports'
import { allocationColor, allocationLabelKey } from '../../src/utils/allocations'
import { formatCurrency, formatPercent } from '../../src/utils/format'
import { getStrategyDisplayName } from '../../src/utils/strategies'
import {
  buildChartCandles,
  buildDcaSchedule,
  buildStrategyIntroduction,
  buildStrategyRules,
  formatInline,
  formatPercentValue,
  humanizeKey,
  normalizePercent,
  parseNumber,
} from '../../src/utils/strategyDetail'

export default function StrategyScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t, locale } = useLocalization()
  const { accessToken, userId } = useAuth()
  const params = useLocalSearchParams<{ plan_id?: string; calculation_id?: string }>()
  const planId = params.plan_id ?? ''
  const calculationId = params.calculation_id ?? ''
  const [showAdvancedParameters, setShowAdvancedParameters] = useState(false)

  const query = useQuery({
    queryKey: ['plan', userId, calculationId, planId],
    queryFn: async () => {
      if (!accessToken || !planId || !calculationId) return null
      const resp = await fetchReportPlan(accessToken, calculationId, planId)
      if (resp.error) throw new Error(resp.error.message)
      return resp.data ?? null
    },
    enabled: !!accessToken && !!planId && !!calculationId,
  })

  const plan = query.data ?? null
  const assetType = plan?.asset_type
  const chartInterval: '1d' = '1d'
  const chartDateRange = useMemo(() => {
    const end = new Date()
    const start = new Date()
    start.setUTCDate(start.getUTCDate() - 30)
    const format = (value: Date) => value.toISOString().slice(0, 10)
    return { start: format(start), end: format(end) }
  }, [])

  const chartQuery = useQuery({
    queryKey: ['strategy-ohlcv', userId, plan?.asset_key, chartInterval, chartDateRange.start, chartDateRange.end],
    queryFn: async () => {
      if (!accessToken || !plan || (assetType !== 'crypto' && assetType !== 'stock')) return null
      const resp = await fetchOHLCV(accessToken, {
        asset_key: plan.asset_key,
        asset_type: assetType,
        symbol: plan.symbol,
        interval: chartInterval,
        start: chartDateRange.start,
        end: chartDateRange.end,
      })
      if (resp.error) throw new Error(resp.error.message)
      return resp.data ?? null
    },
    enabled: !!accessToken && !!plan && (assetType === 'crypto' || assetType === 'stock'),
  })

  const reportQuery = useQuery({
    queryKey: ['report', userId, calculationId],
    queryFn: async () => {
      if (!accessToken || !calculationId) return null
      const resp = await fetchReport(accessToken, calculationId)
      if (resp.error) throw new Error(resp.error.message)
      return resp.data ?? null
    },
    enabled: !!accessToken && !!calculationId && assetType === 'portfolio',
  })

  const chartCandles = useMemo(() => buildChartCandles(chartQuery.data?.series ?? []), [chartQuery.data?.series])

  const chartPoints = useMemo(() => chartCandles.map((candle) => candle.close), [chartCandles])

  const parameters = plan?.parameters ?? {}
  const portfolioAllocation = reportQuery.data?.asset_allocation ?? []
  const portfolioCurrency = reportQuery.data?.base_currency ?? 'USD'
  const quoteCurrency =
    plan?.quote_currency ?? chartQuery.data?.quote_currency ?? (assetType === 'portfolio' ? portfolioCurrency : 'USD')

  const strategyDisplayName = useMemo(() => getStrategyDisplayName(t, plan?.strategy_id), [plan?.strategy_id, t])

  const assetTypeLabel = useMemo(() => {
    if (!assetType) return ''
    switch (assetType) {
      case 'crypto':
        return t('strategy.assetType.crypto')
      case 'stock':
        return t('strategy.assetType.stock')
      case 'forex':
        return t('strategy.assetType.forex')
      case 'portfolio':
        return t('strategy.assetType.portfolio')
      default:
        return ''
    }
  }, [assetType, t])

  const planMetaLine = plan
    ? assetTypeLabel
      ? t('strategy.planMeta', { symbol: plan.symbol, assetType: assetTypeLabel })
      : plan.symbol
    : ''

  const handleMarkExecuted = async () => {
    if (!accessToken || !plan) return
    const resp = await fetchInsights(accessToken, 'all')
    if (resp.error) {
      Alert.alert(t('strategy.executeFailTitle'), resp.error.message)
      return
    }
    const match = resp.data?.items.find(
      (item) => item.plan_id === plan.plan_id || item.strategy_id === plan.strategy_id
    )
    if (!match) {
      Alert.alert(t('strategy.noInsightTitle'), t('strategy.noInsightBody'), [
        { text: t('strategy.tradeSlipCta'), onPress: () => router.push('/(modals)/trade-slip') },
        { text: t('common.close'), style: 'cancel' },
      ])
      return
    }
    router.push({ pathname: '/(modals)/quick-update', params: { id: match.id } })
  }

  const styles = useMemo(() => createStyles(theme), [theme])

  const stopLossData = useMemo(() => {
    const stopLossPrice = parseNumber(parameters.stop_loss_price)
    const supportLevel = parseNumber(parameters.support_level)
    const stopLossPct = parseNumber(parameters.stop_loss_pct)
    if (stopLossPrice === null && supportLevel === null) return null
    return { stopLossPrice, supportLevel, stopLossPct }
  }, [parameters])

  const takeProfitLayers = useMemo(() => {
    const rawLayers = Array.isArray(parameters.layers) ? parameters.layers : []
    return rawLayers
      .map((entry: any, index: number) => {
        const label =
          typeof entry?.layer_name === 'string' && entry.layer_name.trim().length > 0
            ? entry.layer_name
            : t('strategy.visualization.layerLabel', { index: index + 1 })
        const targetPrice = parseNumber(entry?.target_price)
        const sellPct = parseNumber(entry?.sell_percentage) ?? parseNumber(entry?.sell_pct)
        const expectedProfit = parseNumber(entry?.expected_profit_usd)
        return { label, targetPrice, sellPct, expectedProfit }
      })
      .filter((entry: any) => entry.targetPrice !== null || entry.sellPct !== null)
  }, [parameters, t])

  const pyramidingAdditions = useMemo(() => {
    const rawAdditions = Array.isArray(parameters.additions) ? parameters.additions : []
    const normalized = rawAdditions
      .map((entry: any, index: number) => {
        const additionNumber = parseNumber(entry?.addition_number) ?? index + 1
        const triggerProfitPct = parseNumber(entry?.trigger_profit_pct)
        const additionAmount = parseNumber(entry?.addition_amount_usd)
        return { additionNumber, triggerProfitPct, additionAmount }
      })
      .filter((entry: any) => entry.triggerProfitPct !== null || entry.additionAmount !== null)
      .sort((a: any, b: any) => a.additionNumber - b.additionNumber)

    if (normalized.length > 0) return normalized

    const profitStepPct = parseNumber(parameters.profit_step_pct)
    const maxAdditions = parseNumber(parameters.max_additions)
    const baseAmount = parseNumber(parameters.base_addition_usd)
    if (!profitStepPct || !maxAdditions || !baseAmount) return []

    const count = Math.min(Math.max(1, Math.round(maxAdditions)), 4)
    return Array.from({ length: count }).map((_, index) => {
      const additionNumber = index + 1
      const triggerProfitPct = profitStepPct * additionNumber
      const additionAmount = baseAmount * Math.pow(1.2, additionNumber - 1)
      return { additionNumber, triggerProfitPct, additionAmount }
    })
  }, [parameters])

  const ladderOrders = useMemo(() => {
    const rawOrders = Array.isArray(parameters.safety_orders) ? parameters.safety_orders : []
    const normalized = rawOrders
      .map((entry: any, index: number) => {
        const orderNumber = parseNumber(entry?.order_number) ?? index + 1
        const triggerPrice = parseNumber(entry?.trigger_price)
        const orderAmount = parseNumber(entry?.order_amount_usd)
        return { orderNumber, triggerPrice, orderAmount }
      })
      .filter((entry: any) => entry.triggerPrice || entry.orderAmount)
      .sort((a: any, b: any) => a.orderNumber - b.orderNumber)

    if (normalized.length > 0) return normalized

    const priceStepPct = parseNumber(parameters.price_step_pct)
    const maxOrders = parseNumber(parameters.max_safety_orders)
    const baseAmount = parseNumber(parameters.safety_order_base_usd)
    const multiplier = parseNumber(parameters.order_multiplier) ?? 1
    const currentPrice = chartPoints.length > 0 ? chartPoints[chartPoints.length - 1] : null

    if (!priceStepPct || !maxOrders || !baseAmount || !currentPrice) return []

    const count = Math.min(Math.max(1, Math.round(maxOrders)), 4)
    return Array.from({ length: count }).map((_, index) => {
      const orderNumber = index + 1
      const triggerPrice = currentPrice * (1 - priceStepPct * orderNumber)
      const orderAmount = baseAmount * Math.pow(multiplier, orderNumber - 1)
      return { orderNumber, triggerPrice, orderAmount }
    })
  }, [parameters, chartPoints])

  const dcaSchedule = useMemo(() => {
    const nextExecution = typeof parameters.next_execution_at === 'string' ? parameters.next_execution_at : ''
    const frequency = typeof parameters.frequency === 'string' ? parameters.frequency : ''
    return buildDcaSchedule(nextExecution, frequency)
  }, [parameters])

  const trailingStop = useMemo(() => {
    const activationPrice = parseNumber(parameters.activation_price)
    const trailingStopPct = parseNumber(parameters.trailing_stop_pct) ?? parseNumber(parameters.callback_rate)
    const hasTrailingStopPct = trailingStopPct !== null && trailingStopPct !== undefined
    const initialStop =
      parseNumber(parameters.initial_trailing_stop_price) ??
      (activationPrice !== null && hasTrailingStopPct ? activationPrice * (1 - (trailingStopPct as number)) : null)
    if (activationPrice === null || initialStop === null) return null
    return { activationPrice, trailingStopPrice: initialStop, trailingStopPct }
  }, [parameters])

  const frequencyLabel = useMemo(() => {
    const frequency = typeof parameters.frequency === 'string' ? parameters.frequency : ''
    if (!frequency) return ''
    const map: Record<string, string> = {
      daily: t('strategy.visualization.frequency.daily'),
      weekly: t('strategy.visualization.frequency.weekly'),
      biweekly: t('strategy.visualization.frequency.biweekly'),
      monthly: t('strategy.visualization.frequency.monthly'),
    }
    return map[frequency] ?? frequency
  }, [parameters.frequency, t])

  const dateFormatter = useMemo(() => new Intl.DateTimeFormat(locale, { month: 'short', day: 'numeric' }), [locale])
  const weekdayFormatter = useMemo(() => new Intl.DateTimeFormat(locale, { weekday: 'short' }), [locale])
  const hasPriceChart = (assetType === 'crypto' || assetType === 'stock') && chartCandles.length > 1
  const currentPrice = chartCandles.length > 0 ? chartCandles[chartCandles.length - 1].close : null

  const traceLines = useMemo<TraceLine[]>(() => {
    if (!plan || !hasPriceChart) return []

    switch (plan.strategy_id) {
      case 'S01': {
        if (!stopLossData) return []
        const lines: TraceLine[] = []
        if (stopLossData.stopLossPrice !== null) {
          const label =
            stopLossData.stopLossPct !== null
              ? t('strategy.visualization.stopLossLineWithPct', {
                  price: formatCurrency(stopLossData.stopLossPrice, quoteCurrency),
                  percent: formatPercentValue(stopLossData.stopLossPct),
                })
              : t('strategy.visualization.stopLossLine', {
                  price: formatCurrency(stopLossData.stopLossPrice, quoteCurrency),
                })
          lines.push({
            value: stopLossData.stopLossPrice,
            color: theme.colors.danger,
            label,
          })
        }
        if (stopLossData.supportLevel !== null) {
          lines.push({
            value: stopLossData.supportLevel,
            color: theme.colors.warning,
            label: t('strategy.visualization.supportLine', {
              price: formatCurrency(stopLossData.supportLevel, quoteCurrency),
            }),
            dashed: true,
          })
        }
        return lines
      }
      case 'S02':
        return ladderOrders.slice(0, 4).reduce<TraceLine[]>((acc, order) => {
          if (!order.triggerPrice) return acc
          const tier = t('strategy.visualization.tierShort', { index: order.orderNumber })
          const price = formatCurrency(order.triggerPrice, quoteCurrency)
          const amount = order.orderAmount ? formatCurrency(order.orderAmount, quoteCurrency) : ''
          const label = amount ? `${tier} ${price} · ${amount}` : `${tier} ${price}`
          acc.push({ value: order.triggerPrice, color: theme.colors.accent, label })
          return acc
        }, [])
      case 'S03': {
        if (!trailingStop) return []
        const activationLabel = t('strategy.visualization.activationLine', {
          price: formatCurrency(trailingStop.activationPrice, quoteCurrency),
        })
        const trailingLabel = t('strategy.visualization.trailingLine', {
          price: formatCurrency(trailingStop.trailingStopPrice, quoteCurrency),
          percent: typeof trailingStop.trailingStopPct === 'number' ? formatPercent(trailingStop.trailingStopPct) : '-',
        })
        return [
          { value: trailingStop.activationPrice, color: theme.colors.accent, label: activationLabel },
          {
            value: trailingStop.trailingStopPrice,
            color: theme.colors.warning,
            label: trailingLabel,
            dashed: true,
          },
        ]
      }
      case 'S04':
        return takeProfitLayers.slice(0, 4).reduce<TraceLine[]>((acc, layer, index) => {
          if (!layer.targetPrice) return acc
          const tier = t('strategy.visualization.tierShort', { index: index + 1 })
          const price = formatCurrency(layer.targetPrice, quoteCurrency)
          const percent = formatPercentValue(layer.sellPct)
          const label = `${tier} ${price} · ${percent}`
          acc.push({ value: layer.targetPrice, color: theme.colors.warning, label })
          return acc
        }, [])
      case 'S09':
        return pyramidingAdditions.slice(0, 4).reduce<TraceLine[]>((acc, addition) => {
          if (!currentPrice) return acc
          const triggerPct = normalizePercent(addition.triggerProfitPct)
          if (triggerPct === null) return acc
          const targetPrice = currentPrice * (1 + triggerPct)
          const amount = addition.additionAmount ? formatCurrency(addition.additionAmount, quoteCurrency) : ''
          const pctLabel = triggerPct ? `+${formatPercent(triggerPct)}` : ''
          const label = amount ? `${pctLabel} · ${amount}` : pctLabel
          acc.push({ value: targetPrice, color: theme.colors.success, label })
          return acc
        }, [])
      default:
        return []
    }
  }, [
    plan,
    hasPriceChart,
    stopLossData,
    ladderOrders,
    trailingStop,
    takeProfitLayers,
    pyramidingAdditions,
    currentPrice,
    t,
    quoteCurrency,
    theme.colors,
  ])

  const ruleResult = useMemo(
    () =>
      buildStrategyRules({
        plan,
        parameters,
        stopLossData,
        ladderOrders,
        trailingStop,
        takeProfitLayers,
        pyramidingAdditions,
        currentPrice,
        frequencyLabel,
        dateFormatter,
        weekdayFormatter,
        t,
        currencyCode: quoteCurrency,
      }),
    [
      plan,
      parameters,
      stopLossData,
      ladderOrders,
      trailingStop,
      takeProfitLayers,
      pyramidingAdditions,
      currentPrice,
      frequencyLabel,
      dateFormatter,
      weekdayFormatter,
      t,
      quoteCurrency,
    ]
  )

  const advancedEntries = useMemo(() => {
    return Object.entries(parameters).filter(([key]) => !ruleResult.usedKeys.has(key))
  }, [parameters, ruleResult.usedKeys])

  const strategyIntroduction = useMemo(
    () =>
      buildStrategyIntroduction({
        plan,
        parameters,
        stopLossData,
        ladderOrders,
        trailingStop,
        takeProfitLayers,
        pyramidingAdditions,
        frequencyLabel,
        dateFormatter,
        weekdayFormatter,
        t,
        currencyCode: quoteCurrency,
      }),
    [
      plan,
      parameters,
      stopLossData,
      ladderOrders,
      trailingStop,
      takeProfitLayers,
      pyramidingAdditions,
      frequencyLabel,
      dateFormatter,
      weekdayFormatter,
      t,
      quoteCurrency,
    ]
  )

  const rules = ruleResult.rules
  const hasRules = rules.length > 0
  const hasAdvanced = advancedEntries.length > 0
  const advancedOpen = !hasRules && hasAdvanced ? true : showAdvancedParameters
  const isRulesLoading = query.isLoading || (query.isFetching && !plan)

  const toggleAdvancedParameters = useCallback(() => {
    setShowAdvancedParameters((prev) => !prev)
  }, [])

  return (
    <Screen scroll>
      <Text style={styles.pageTitle}>{t('strategy.title')}</Text>
      {plan && (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={styles.planTitle}>{strategyDisplayName}</Text>
          <Text style={styles.planMeta}>{planMetaLine}</Text>

          {hasPriceChart ? (
            <View style={styles.tracePanel}>
              <StrategyCandlestickChart
                candles={chartCandles}
                lines={traceLines}
                locale={locale}
                formatPrice={(value) => formatCurrency(value, quoteCurrency)}
              />
              {plan.strategy_id === 'S05' ? (
                <View style={styles.cadenceWrap}>
                  <View style={styles.cadenceHeader}>
                    <Text style={styles.cadenceTitle}>{t('strategy.visualization.calendarTitle')}</Text>
                    <Text style={styles.cadenceMeta}>
                      {t('strategy.visualization.calendarAmount', {
                        amount: (() => {
                          const amt = parseNumber(parameters.amount)
                          return amt !== null ? formatCurrency(amt, quoteCurrency) : '-'
                        })(),
                        frequency: frequencyLabel || '-',
                      })}
                    </Text>
                  </View>
                  {dcaSchedule.length > 0 ? (
                    <View style={styles.cadenceTrack}>
                      <View style={styles.cadenceLine} />
                      {dcaSchedule.map((date, index) => (
                        <View key={`dca-${date.toISOString()}-${index}`} style={styles.cadenceItem}>
                          <View style={[styles.cadenceDot, index === 0 && styles.cadenceDotActive]} />
                          <Text style={styles.cadenceDow}>{weekdayFormatter.format(date)}</Text>
                          <Text style={styles.cadenceDate}>{dateFormatter.format(date)}</Text>
                        </View>
                      ))}
                    </View>
                  ) : (
                    <Text style={styles.unavailableText}>{t('strategy.visualization.unavailable')}</Text>
                  )}
                </View>
              ) : null}
            </View>
          ) : null}

          {strategyIntroduction.length > 0 ? (
            <View style={styles.introSection}>
              <Text style={styles.introTitle}>{t('strategy.introTitle')}</Text>
              {strategyIntroduction.map((paragraph, index) => (
                <Text key={`intro-${index}`} style={styles.introParagraph}>
                  {paragraph}
                </Text>
              ))}
            </View>
          ) : null}

          {assetType === 'portfolio' && portfolioAllocation.length > 0 ? (
            <View style={{ marginTop: theme.spacing.md }}>
              <DonutChart
                size={150}
                strokeWidth={16}
                segments={portfolioAllocation.map((item) => ({
                  value: item.value_display ?? item.value_usd,
                  color: allocationColor(theme, item.label),
                }))}
              />
              <View style={{ marginTop: theme.spacing.sm }}>
                {portfolioAllocation.map((item, index) => (
                  <View
                    key={`${item.label}-${index}`}
                    style={{ flexDirection: 'row', alignItems: 'center', marginBottom: theme.spacing.xs }}
                  >
                    <View
                      style={{
                        width: 10,
                        height: 10,
                        borderRadius: 999,
                        backgroundColor: allocationColor(theme, item.label),
                      }}
                    />
                    <View style={{ marginLeft: theme.spacing.sm }}>
                      <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
                        {t(allocationLabelKey(item.label) as any)}
                      </Text>
                      <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
                        {formatPercent(item.weight_pct)} •{' '}
                        {formatCurrency(
                          item.value_display ?? item.value_usd,
                          item.display_currency ?? portfolioCurrency
                        )}
                      </Text>
                    </View>
                  </View>
                ))}
              </View>
            </View>
          ) : assetType === 'forex' ? (
            <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
              {t('strategy.forexNote')}
            </Text>
          ) : null}
        </Card>
      )}

      <Text style={styles.sectionTitle}>{t('strategy.rulesTitle')}</Text>

      {isRulesLoading ? (
        <Card style={{ marginTop: theme.spacing.sm }}>
          <View style={styles.rulesLoading}>
            <View style={styles.rulesLoadingSpine}>
              <View style={styles.rulesLoadingLine} />
              <View style={styles.rulesLoadingNodes}>
                {Array.from({ length: 3 }).map((_, index) => (
                  <ShimmerBlock
                    key={`rules-node-${index}`}
                    width={10}
                    height={10}
                    radius={5}
                    style={styles.rulesLoadingNode}
                  />
                ))}
              </View>
            </View>
            <View style={styles.rulesLoadingRows}>
              {Array.from({ length: 3 }).map((_, index) => (
                <View key={`rules-row-${index}`} style={styles.rulesLoadingRow}>
                  <ShimmerBlock height={10} width="55%" style={styles.rulesLoadingTitle} />
                  <ShimmerBlock height={16} width="82%" style={styles.rulesLoadingValue} />
                </View>
              ))}
            </View>
          </View>
          <Text style={styles.rulesLoadingCaption}>{t('common.loading')}</Text>
        </Card>
      ) : hasRules ? (
        <Card style={{ marginTop: theme.spacing.sm }}>
          {rules.map((rule, index) => {
            const toneColor =
              rule.tone === 'danger'
                ? theme.colors.danger
                : rule.tone === 'warning'
                  ? theme.colors.warning
                  : rule.tone === 'success'
                    ? theme.colors.success
                    : theme.colors.ink

            return (
              <View key={rule.id}>
                <View style={styles.ruleRow}>
                  <View style={styles.ruleTextWrap}>
                    <Text style={styles.ruleTitle}>{rule.title}</Text>
                    <Text style={[styles.ruleValue, { color: toneColor }]}>{rule.value}</Text>
                  </View>
                </View>
                {index < rules.length - 1 ? <View style={styles.ruleDivider} /> : null}
              </View>
            )
          })}
        </Card>
      ) : (
        <Card style={{ marginTop: theme.spacing.sm }}>
          <Text style={styles.emptyRulesText}>{t('strategy.rulesEmpty')}</Text>
        </Card>
      )}

      {hasAdvanced ? (
        <Card style={{ marginTop: theme.spacing.sm }}>
          <Pressable
            onPress={toggleAdvancedParameters}
            style={styles.advancedHeader}
            disabled={!hasRules}
            accessibilityRole="button"
            accessibilityState={{ expanded: advancedOpen, disabled: !hasRules }}
            accessibilityLabel={t('strategy.advancedTitle')}
          >
            <View>
              <Text style={styles.advancedTitle}>{t('strategy.advancedTitle')}</Text>
              {hasRules ? (
                <Text style={styles.advancedHint}>
                  {advancedOpen ? t('strategy.advancedHide') : t('strategy.advancedShow')}
                </Text>
              ) : null}
            </View>
            {hasRules ? (
              <Ionicons name={advancedOpen ? 'chevron-up' : 'chevron-down'} size={18} color={theme.colors.muted} />
            ) : null}
          </Pressable>

          {advancedOpen ? (
            <View style={styles.advancedBody}>
              {advancedEntries.map(([key, value], index) => {
                const isArray = Array.isArray(value)
                const isObject = value && typeof value === 'object' && !isArray
                const objectEntries = isObject ? Object.entries(value as Record<string, unknown>) : []
                const arrayItems = isArray ? (value as unknown[]) : []
                const arrayOfObjects =
                  isArray &&
                  arrayItems.length > 0 &&
                  arrayItems.every((entry) => entry && typeof entry === 'object' && !Array.isArray(entry))

                const label = humanizeKey(key)
                const showRawKey = label.toLowerCase() !== key.toLowerCase()

                return (
                  <View key={key}>
                    <View style={styles.advancedItem}>
                      <Text style={styles.advancedKeyLabel}>{label || key}</Text>
                      {showRawKey ? <Text style={styles.advancedRawKey}>{key}</Text> : null}

                      {arrayOfObjects ? (
                        <View style={styles.advancedNestedBlock}>
                          {arrayItems.map((entry, entryIndex) => (
                            <View key={`${key}-${entryIndex}`} style={styles.advancedNestedItem}>
                              <Text style={styles.advancedSubLabel}>{`Item ${entryIndex + 1}`}</Text>
                              {Object.entries(entry as Record<string, unknown>).map(([entryKey, entryValue]) => (
                                <Text key={entryKey} style={styles.advancedLine}>
                                  {entryKey}: {formatInline(entryValue)}
                                </Text>
                              ))}
                            </View>
                          ))}
                        </View>
                      ) : isObject ? (
                        <View style={styles.advancedNestedBlock}>
                          {objectEntries.map(([entryKey, entryValue]) => (
                            <Text key={entryKey} style={styles.advancedLine}>
                              {entryKey}: {formatInline(entryValue)}
                            </Text>
                          ))}
                        </View>
                      ) : (
                        <Text style={styles.advancedValue}>{formatInline(value)}</Text>
                      )}
                    </View>
                    {index < advancedEntries.length - 1 ? <View style={styles.ruleDivider} /> : null}
                  </View>
                )
              })}
            </View>
          ) : null}
        </Card>
      ) : null}

      <Button title={t('strategy.executeCta')} style={{ marginTop: theme.spacing.lg }} onPress={handleMarkExecuted} />
      <Button
        title={t('common.close')}
        variant="ghost"
        style={{ marginTop: theme.spacing.md }}
        onPress={() => router.back()}
      />
    </Screen>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    pageTitle: {
      fontSize: 24,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    planTitle: {
      fontSize: 20,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    planMeta: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    introSection: {
      marginTop: theme.spacing.md,
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
      gap: theme.spacing.xs,
    },
    introTitle: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.3,
      textTransform: 'uppercase',
    },
    introParagraph: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.body,
      fontSize: 14,
      lineHeight: 20,
    },
    sectionTitle: {
      marginTop: theme.spacing.lg,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
    },
    ruleRow: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      paddingVertical: theme.spacing.xs,
    },
    ruleTextWrap: {
      flex: 1,
    },
    ruleTitle: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.2,
    },
    ruleValue: {
      marginTop: theme.spacing.xs,
      fontSize: 16,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      lineHeight: 22,
    },
    ruleDivider: {
      height: 1,
      backgroundColor: theme.colors.border,
      marginVertical: theme.spacing.xs,
      opacity: 0.9,
    },
    emptyRulesText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 14,
      lineHeight: 20,
    },
    rulesLoading: {
      flexDirection: 'row',
      gap: theme.spacing.md,
      paddingVertical: theme.spacing.xs,
    },
    rulesLoadingSpine: {
      width: 18,
      position: 'relative',
      alignItems: 'center',
    },
    rulesLoadingLine: {
      position: 'absolute',
      top: theme.spacing.xs,
      bottom: theme.spacing.xs,
      width: 2,
      borderRadius: 999,
      backgroundColor: theme.colors.border,
      opacity: 0.7,
    },
    rulesLoadingNodes: {
      flex: 1,
      justifyContent: 'space-between',
      paddingVertical: theme.spacing.xs,
    },
    rulesLoadingNode: {
      alignSelf: 'center',
    },
    rulesLoadingRows: {
      flex: 1,
      gap: theme.spacing.sm,
      paddingRight: theme.spacing.xs,
    },
    rulesLoadingRow: {
      gap: theme.spacing.xs,
    },
    rulesLoadingTitle: {
      height: 10,
    },
    rulesLoadingValue: {
      height: 16,
    },
    rulesLoadingCaption: {
      marginTop: theme.spacing.sm,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    advancedHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    advancedTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 15,
    },
    advancedHint: {
      marginTop: 2,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    advancedBody: {
      marginTop: theme.spacing.sm,
    },
    advancedItem: {
      paddingVertical: theme.spacing.xs,
    },
    advancedKeyLabel: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
    },
    advancedRawKey: {
      marginTop: 2,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    advancedNestedBlock: {
      marginTop: theme.spacing.xs,
    },
    advancedNestedItem: {
      marginTop: theme.spacing.xs,
    },
    advancedSubLabel: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 13,
    },
    advancedLine: {
      marginTop: 2,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
    },
    advancedValue: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 14,
      lineHeight: 20,
    },
    tracePanel: {
      marginTop: theme.spacing.md,
      backgroundColor: theme.colors.surfaceElevated,
      borderRadius: theme.radius.lg,
      padding: theme.spacing.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
    },
    cadenceWrap: {
      marginTop: theme.spacing.md,
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
    },
    cadenceHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    cadenceTitle: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      letterSpacing: 0.3,
    },
    cadenceMeta: {
      fontSize: 12,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
    },
    cadenceTrack: {
      marginTop: theme.spacing.sm,
      color: theme.colors.ink,
      flexDirection: 'row',
      justifyContent: 'space-between',
      alignItems: 'flex-start',
      position: 'relative',
    },
    cadenceLine: {
      position: 'absolute',
      left: theme.spacing.sm,
      right: theme.spacing.sm,
      top: 6,
      height: 1,
      backgroundColor: theme.colors.border,
    },
    cadenceItem: {
      alignItems: 'center',
      flex: 1,
      paddingHorizontal: theme.spacing.xs,
    },
    cadenceDot: {
      width: 8,
      height: 8,
      borderRadius: 999,
      backgroundColor: theme.colors.accentSoft,
      borderWidth: 1,
      borderColor: theme.colors.accent,
    },
    cadenceDotActive: {
      backgroundColor: theme.colors.accent,
      borderColor: theme.colors.accent,
    },
    cadenceDow: {
      marginTop: theme.spacing.xs,
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    cadenceDate: {
      marginTop: 2,
      fontSize: 14,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
    },
    unavailableText: {
      marginTop: theme.spacing.sm,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
  })
