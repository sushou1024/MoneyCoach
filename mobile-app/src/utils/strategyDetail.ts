import { formatCurrency, formatNumber, formatPercent } from './format'
import { TranslationKey } from './i18n'
import { ReportPlan } from '../types/api'

export type StrategyRuleTone = 'neutral' | 'danger' | 'warning' | 'success'

export type StrategyRule = {
  id: string
  title: string
  value: string
  explanation: string
  tone?: StrategyRuleTone
}

export type StrategyRuleResult = {
  rules: StrategyRule[]
  usedKeys: Set<string>
}

export type StrategyCandle = {
  open: number
  high: number
  low: number
  close: number
  time: string
}

export type StopLossData = {
  stopLossPrice: number | null
  supportLevel: number | null
  stopLossPct: number | null
}

export type TakeProfitLayer = {
  label: string
  targetPrice: number | null
  sellPct: number | null
  expectedProfit: number | null
}

export type PyramidingAddition = {
  additionNumber: number
  triggerProfitPct: number | null
  additionAmount: number | null
}

export type LadderOrder = {
  orderNumber: number
  triggerPrice: number | null
  orderAmount: number | null
}

export type TrailingStop = {
  activationPrice: number
  trailingStopPrice: number
  trailingStopPct: number | null
}

type Translator = (key: TranslationKey, params?: Record<string, string | number>) => string

type StrategyRulesInput = {
  plan: ReportPlan | null
  parameters: Record<string, unknown>
  stopLossData: StopLossData | null
  ladderOrders: LadderOrder[]
  trailingStop: TrailingStop | null
  takeProfitLayers: TakeProfitLayer[]
  pyramidingAdditions: PyramidingAddition[]
  currentPrice: number | null
  frequencyLabel: string
  dateFormatter: Intl.DateTimeFormat
  weekdayFormatter: Intl.DateTimeFormat
  t: Translator
  currencyCode: string
}

type StrategyIntroInput = {
  plan: ReportPlan | null
  parameters: Record<string, unknown>
  stopLossData: StopLossData | null
  ladderOrders: LadderOrder[]
  trailingStop: TrailingStop | null
  takeProfitLayers: TakeProfitLayer[]
  pyramidingAdditions: PyramidingAddition[]
  frequencyLabel: string
  dateFormatter: Intl.DateTimeFormat
  weekdayFormatter: Intl.DateTimeFormat
  t: Translator
  currencyCode: string
}

export function parseNumber(value: unknown) {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : null
  }
  return null
}

export function normalizePercent(value: number | null | undefined) {
  if (value === null || value === undefined) return null
  const normalized = value > 1 ? value / 100 : value
  return Math.min(Math.max(normalized, 0), 1)
}

export function formatPercentValue(value: number | null) {
  const normalized = normalizePercent(value)
  if (normalized === null) return '-'
  return formatPercent(normalized)
}

export function formatPercentMaybe(value: number | null) {
  const normalized = normalizePercent(value)
  if (normalized === null) return null
  return formatPercent(normalized)
}

function formatPrimitive(value: unknown) {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  return String(value)
}

export function formatInline(value: unknown) {
  if (Array.isArray(value)) {
    const primitives = value.every(
      (entry) => entry === null || entry === undefined || ['string', 'number', 'boolean'].includes(typeof entry)
    )
    if (primitives) {
      return value.map((entry) => formatPrimitive(entry)).join(', ')
    }
    return `${value.length} items`
  }
  if (value && typeof value === 'object') {
    try {
      return JSON.stringify(value)
    } catch {
      return String(value)
    }
  }
  return formatPrimitive(value)
}

export function humanizeText(value: string) {
  return value.replace(/[_-]+/g, ' ').replace(/\s+/g, ' ').trim()
}

export function titleCase(value: string) {
  const normalized = humanizeText(value)
  if (!normalized) return ''
  return normalized.replace(/\b\w/g, (char) => char.toUpperCase())
}

export function humanizeKey(key: string) {
  const normalized = key.replace(/\.(?=\w)/g, ' ')
  return titleCase(normalized)
}

export function buildChartCandles(series: unknown): StrategyCandle[] {
  if (!Array.isArray(series) || series.length === 0) return []

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

  return series
    .map((point: any) => {
      const raw = Array.isArray(point) ? point : []
      const time = toTime(raw[0] ?? point?.time ?? point?.timestamp ?? point?.date)
      const open = toNumber(raw[1] ?? point?.open)
      const high = toNumber(raw[2] ?? point?.high)
      const low = toNumber(raw[3] ?? point?.low)
      const close = toNumber(raw[4] ?? point?.close)
      if (!time || open === null || high === null || low === null || close === null) return null
      return { open, high, low, close, time }
    })
    .filter(isCandle)
}

export function buildDcaSchedule(nextExecution: string, frequency: string) {
  if (!nextExecution || !frequency) return []
  const baseDate = new Date(nextExecution)
  if (Number.isNaN(baseDate.getTime())) return []
  const dates: Date[] = []
  const recognized = ['daily', 'weekly', 'biweekly', 'monthly']
  if (!recognized.includes(frequency)) {
    return [baseDate]
  }
  const addDays = (date: Date, days: number) => {
    const next = new Date(date)
    next.setUTCDate(next.getUTCDate() + days)
    return next
  }
  const addMonths = (date: Date, months: number) => {
    const next = new Date(date)
    next.setUTCMonth(next.getUTCMonth() + months)
    return next
  }
  let cursor = baseDate
  for (let i = 0; i < 4; i += 1) {
    dates.push(cursor)
    switch (frequency) {
      case 'daily':
        cursor = addDays(cursor, 1)
        break
      case 'weekly':
        cursor = addDays(cursor, 7)
        break
      case 'biweekly':
        cursor = addDays(cursor, 14)
        break
      case 'monthly':
        cursor = addMonths(cursor, 1)
        break
      default:
        cursor = addMonths(cursor, 1)
        break
    }
  }
  return dates
}

export function buildStrategyRules({
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
  currencyCode,
}: StrategyRulesInput): StrategyRuleResult {
  const usedKeys = new Set<string>()
  const rules: StrategyRule[] = []
  const markKeys = (...keys: string[]) => {
    keys.forEach((key) => usedKeys.add(key))
  }

  if (!plan) {
    return { rules, usedKeys }
  }

  const frequencyRaw = typeof parameters.frequency === 'string' ? parameters.frequency : ''
  const cadenceLabel = frequencyLabel || humanizeText(frequencyRaw)

  const formatMoney = (value: number | null) => {
    if (value === null || value === undefined) return '-'
    return formatCurrency(value, currencyCode)
  }

  switch (plan.strategy_id) {
    case 'S01': {
      markKeys('stop_loss_price', 'stop_loss_pct', 'support_level')
      if (stopLossData?.stopLossPrice !== null && stopLossData?.stopLossPrice !== undefined) {
        rules.push({
          id: 's01-stop-loss-price',
          title: t('strategy.rules.s01.stop.title'),
          value: t('strategy.rules.s01.stop.value', { price: formatMoney(stopLossData.stopLossPrice) }),
          explanation: t('strategy.rules.s01.stop.explanation'),
          tone: 'danger',
        })
      }
      if (stopLossData?.stopLossPct !== null && stopLossData?.stopLossPct !== undefined) {
        rules.push({
          id: 's01-stop-loss-distance',
          title: t('strategy.rules.s01.distance.title'),
          value: t('strategy.rules.s01.distance.value', { percent: formatPercentValue(stopLossData.stopLossPct) }),
          explanation: t('strategy.rules.s01.distance.explanation'),
          tone: 'danger',
        })
      }
      if (stopLossData?.supportLevel !== null && stopLossData?.supportLevel !== undefined) {
        rules.push({
          id: 's01-support-reference',
          title: t('strategy.rules.s01.support.title'),
          value: t('strategy.rules.s01.support.value', { price: formatMoney(stopLossData.supportLevel) }),
          explanation: t('strategy.rules.s01.support.explanation'),
          tone: 'warning',
        })
      }
      break
    }
    case 'S02': {
      markKeys('safety_orders', 'price_step_pct', 'max_safety_orders', 'safety_order_base_usd', 'order_multiplier')

      ladderOrders.slice(0, 4).forEach((order) => {
        if (!order.triggerPrice && !order.orderAmount) return
        const priceLabel = order.triggerPrice ? formatMoney(order.triggerPrice) : t('strategy.value.nextTierPrice')
        const amountLabel = order.orderAmount ? formatMoney(order.orderAmount) : t('strategy.value.nextTierAmount')
        rules.push({
          id: `s02-tier-${order.orderNumber}`,
          title: t('strategy.rules.s02.order.title', { index: order.orderNumber }),
          value: t('strategy.rules.s02.order.value', { price: priceLabel, amount: amountLabel }),
          explanation: t('strategy.rules.s02.order.explanation'),
          tone: 'warning',
        })
      })

      const stepPct = formatPercentMaybe(parseNumber(parameters.price_step_pct))
      if (stepPct) {
        rules.push({
          id: 's02-tier-spacing',
          title: t('strategy.rules.s02.spacing.title'),
          value: t('strategy.rules.s02.spacing.value', { percent: stepPct }),
          explanation: t('strategy.rules.s02.spacing.explanation'),
          tone: 'neutral',
        })
      }

      const baseAmount = parseNumber(parameters.safety_order_base_usd)
      const multiplier = parseNumber(parameters.order_multiplier)
      if (baseAmount !== null || multiplier !== null) {
        const baseLabel = baseAmount !== null ? formatMoney(baseAmount) : t('strategy.value.baseOrderSize')
        const multiplierLabel =
          multiplier !== null
            ? t('strategy.value.multiplierValue', { value: formatNumber(multiplier, 2) })
            : t('strategy.value.fixedSize')
        rules.push({
          id: 's02-sizing',
          title: t('strategy.rules.s02.sizing.title'),
          value: t('strategy.rules.s02.sizing.value', { base: baseLabel, multiplier: multiplierLabel }),
          explanation: t('strategy.rules.s02.sizing.explanation'),
          tone: 'neutral',
        })
      }
      break
    }
    case 'S03': {
      markKeys('activation_price', 'trailing_stop_pct', 'callback_rate', 'initial_trailing_stop_price')
      if (!trailingStop) break

      rules.push({
        id: 's03-activation',
        title: t('strategy.rules.s03.activation.title'),
        value: t('strategy.rules.s03.activation.value', { price: formatMoney(trailingStop.activationPrice) }),
        explanation: t('strategy.rules.s03.activation.explanation'),
        tone: 'neutral',
      })

      const trailingPctLabel =
        typeof trailingStop.trailingStopPct === 'number' ? formatPercentMaybe(trailingStop.trailingStopPct) : null
      if (trailingPctLabel) {
        rules.push({
          id: 's03-trail-distance',
          title: t('strategy.rules.s03.trail.title'),
          value: t('strategy.rules.s03.trail.value', { percent: trailingPctLabel }),
          explanation: t('strategy.rules.s03.trail.explanation'),
          tone: 'warning',
        })
      }

      rules.push({
        id: 's03-initial-stop',
        title: t('strategy.rules.s03.initial.title'),
        value: t('strategy.rules.s03.initial.value', { price: formatMoney(trailingStop.trailingStopPrice) }),
        explanation: t('strategy.rules.s03.initial.explanation'),
        tone: 'warning',
      })
      break
    }
    case 'S04': {
      markKeys('layers')
      takeProfitLayers.slice(0, 4).forEach((layer, index) => {
        if (!layer.targetPrice && layer.sellPct === null) return
        const sellPct = formatPercentMaybe(layer.sellPct)
        const sellLabel = sellPct ?? t('strategy.value.portionOfPosition')
        const priceLabel = layer.targetPrice ? formatMoney(layer.targetPrice) : t('strategy.value.targetPrice')
        const profitSnippet =
          layer.expectedProfit !== null && layer.expectedProfit !== undefined
            ? t('strategy.value.estimatedProfitSnippet', { amount: formatMoney(layer.expectedProfit) })
            : ''
        rules.push({
          id: `s04-layer-${index + 1}`,
          title: layer.label,
          value: t('strategy.rules.s04.layer.value', {
            price: priceLabel,
            percent: sellLabel,
            profitSnippet,
          }),
          explanation: t('strategy.rules.s04.layer.explanation'),
          tone: 'success',
        })
      })
      break
    }
    case 'S05': {
      markKeys('amount', 'frequency', 'next_execution_at')
      const amount = parseNumber(parameters.amount)
      if (amount !== null) {
        const cadence = cadenceLabel ? cadenceLabel.toLowerCase() : t('strategy.value.onSchedule')
        rules.push({
          id: 's05-cadence',
          title: t('strategy.rules.s05.cadence.title'),
          value: t('strategy.rules.s05.cadence.value', { amount: formatMoney(amount), cadence }),
          explanation: t('strategy.rules.s05.cadence.explanation'),
          tone: 'neutral',
        })
      }

      const nextExecution = typeof parameters.next_execution_at === 'string' ? parameters.next_execution_at : ''
      if (nextExecution) {
        const nextDate = new Date(nextExecution)
        if (!Number.isNaN(nextDate.getTime())) {
          const dateLabel = `${weekdayFormatter.format(nextDate)} ${dateFormatter.format(nextDate)}`
          rules.push({
            id: 's05-next-execution',
            title: t('strategy.rules.s05.next.title'),
            value: t('strategy.rules.s05.next.value', { date: dateLabel }),
            explanation: t('strategy.rules.s05.next.explanation'),
            tone: 'neutral',
          })
        }
      }
      break
    }
    case 'S09': {
      markKeys('additions', 'profit_step_pct', 'max_additions', 'base_addition_usd')

      pyramidingAdditions.slice(0, 4).forEach((addition) => {
        const triggerPct = formatPercentMaybe(addition.triggerProfitPct)
        if (!triggerPct && addition.additionAmount === null) return
        const amountLabel = addition.additionAmount
          ? formatMoney(addition.additionAmount)
          : t('strategy.value.nextAddSize')
        const triggerLabel = triggerPct
          ? t('strategy.value.plusPercent', { percent: triggerPct })
          : t('strategy.value.nextProfitStep')
        rules.push({
          id: `s09-addition-${addition.additionNumber}`,
          title: t('strategy.rules.s09.addition.title', { index: addition.additionNumber }),
          value: t('strategy.rules.s09.addition.value', { trigger: triggerLabel, amount: amountLabel }),
          explanation: t('strategy.rules.s09.addition.explanation'),
          tone: 'success',
        })
      })

      const stepPct = formatPercentMaybe(parseNumber(parameters.profit_step_pct))
      const baseAddition = parseNumber(parameters.base_addition_usd)
      const maxAdditions = parseNumber(parameters.max_additions)
      if (stepPct || baseAddition !== null || maxAdditions !== null) {
        const capLabel =
          maxAdditions !== null
            ? t('strategy.value.tierCount', { count: Math.max(1, Math.round(maxAdditions)) })
            : t('strategy.value.severalTiers')
        const baseLabel = baseAddition !== null ? formatMoney(baseAddition) : t('strategy.value.baseAddSize')
        const stepLabel = stepPct
          ? t('strategy.value.plusPercent', { percent: stepPct })
          : t('strategy.value.nextProfitStep')
        rules.push({
          id: 's09-sizing',
          title: t('strategy.rules.s09.step.title'),
          value: t('strategy.rules.s09.step.value', { base: baseLabel, step: stepLabel, cap: capLabel }),
          explanation: t('strategy.rules.s09.step.explanation'),
          tone: 'neutral',
        })
      }
      break
    }
    case 'S16': {
      markKeys(
        'funding_rate_8h',
        'spot_price',
        'mark_price',
        'basis_pct',
        'holding_period_hours',
        'fee_pct',
        'hedge_notional_usd',
        'spot_amount',
        'futures_symbol',
        'trigger_funding_rate',
        'trigger_basis_pct_max'
      )

      const triggerFunding = formatPercentMaybe(parseNumber(parameters.trigger_funding_rate))
      const triggerBasis = formatPercentMaybe(parseNumber(parameters.trigger_basis_pct_max))
      if (triggerFunding || triggerBasis) {
        rules.push({
          id: 's16-entry-criteria',
          title: t('strategy.rules.s16.entry.title'),
          value: t('strategy.rules.s16.entry.value', { funding: triggerFunding ?? '-', basis: triggerBasis ?? '-' }),
          explanation: t('strategy.rules.s16.entry.explanation'),
          tone: 'warning',
        })
      }

      const hedgeNotional = parseNumber(parameters.hedge_notional_usd)
      if (hedgeNotional !== null) {
        rules.push({
          id: 's16-hedge-size',
          title: t('strategy.rules.s16.size.title'),
          value: t('strategy.rules.s16.size.value', { amount: formatMoney(hedgeNotional) }),
          explanation: t('strategy.rules.s16.size.explanation'),
          tone: 'neutral',
        })
      }

      const futuresSymbol = typeof parameters.futures_symbol === 'string' ? parameters.futures_symbol : ''
      if (futuresSymbol) {
        rules.push({
          id: 's16-instrument',
          title: t('strategy.rules.s16.instrument.title'),
          value: t('strategy.rules.s16.instrument.value', { symbol: futuresSymbol }),
          explanation: t('strategy.rules.s16.instrument.explanation'),
          tone: 'neutral',
        })
      }

      const currentFunding = formatPercentMaybe(parseNumber(parameters.funding_rate_8h))
      const currentBasis = formatPercentMaybe(parseNumber(parameters.basis_pct))
      if (currentFunding || currentBasis) {
        rules.push({
          id: 's16-current-conditions',
          title: t('strategy.rules.s16.current.title'),
          value: t('strategy.rules.s16.current.value', {
            funding: currentFunding ?? '-',
            basis: currentBasis ?? '-',
          }),
          explanation: t('strategy.rules.s16.current.explanation'),
          tone: 'neutral',
        })
      }
      break
    }
    case 'S18': {
      markKeys(
        'trend_state',
        'trend_strength',
        'trend_action',
        'ma_short',
        'ma_medium',
        'ma_long',
        'current_price',
        'ma_20',
        'ma_50',
        'ma_200'
      )

      const trendState = typeof parameters.trend_state === 'string' ? parameters.trend_state : ''
      const trendStrength = typeof parameters.trend_strength === 'string' ? parameters.trend_strength : ''
      if (trendState || trendStrength) {
        const stateLabel = trendState ? titleCase(trendState) : t('strategy.value.unknownTrend')
        const strengthLabel = trendStrength ? titleCase(trendStrength) : t('strategy.value.unknownStrength')
        rules.push({
          id: 's18-trend-state',
          title: t('strategy.rules.s18.regime.title'),
          value: t('strategy.rules.s18.regime.value', { state: stateLabel, strength: strengthLabel }),
          explanation: t('strategy.rules.s18.regime.explanation'),
          tone: 'neutral',
        })
      }

      const trendAction = typeof parameters.trend_action === 'string' ? parameters.trend_action : ''
      if (trendAction) {
        rules.push({
          id: 's18-trend-action',
          title: t('strategy.rules.s18.action.title'),
          value: t('strategy.rules.s18.action.value', { action: humanizeText(trendAction) }),
          explanation: t('strategy.rules.s18.action.explanation'),
          tone: 'success',
        })
      }

      const ma20 = parseNumber(parameters.ma_20)
      const ma50 = parseNumber(parameters.ma_50)
      const ma200 = parseNumber(parameters.ma_200)
      const price = parseNumber(parameters.current_price) ?? currentPrice
      if (price !== null || ma20 !== null || ma50 !== null || ma200 !== null) {
        const parts = [
          price !== null ? t('strategy.value.priceAt', { price: formatMoney(price) }) : null,
          ma20 !== null ? t('strategy.value.ma20At', { price: formatMoney(ma20) }) : null,
          ma50 !== null ? t('strategy.value.ma50At', { price: formatMoney(ma50) }) : null,
          ma200 !== null ? t('strategy.value.ma200At', { price: formatMoney(ma200) }) : null,
        ].filter((part): part is string => Boolean(part))
        if (parts.length > 0) {
          rules.push({
            id: 's18-moving-averages',
            title: t('strategy.rules.s18.ma.title'),
            value: t('strategy.rules.s18.ma.value', { parts: parts.join(' · ') }),
            explanation: t('strategy.rules.s18.ma.explanation'),
            tone: 'neutral',
          })
        }
      }
      break
    }
    case 'S22': {
      markKeys('target_weights', 'vol_floor', 'rebalance_threshold_pct', 'rebalance_frequency')
      const targetWeights = Array.isArray(parameters.target_weights) ? parameters.target_weights : []
      const weightSummary = targetWeights
        .slice(0, 4)
        .map((entry: any, index: number) => {
          const symbol =
            typeof entry?.symbol === 'string' && entry.symbol
              ? entry.symbol
              : t('strategy.value.assetFallback', { index: index + 1 })
          const weight = formatPercentMaybe(parseNumber(entry?.weight_pct))
          return weight ? `${symbol} ${weight}` : symbol
        })
        .join(' · ')
      if (weightSummary) {
        rules.push({
          id: 's22-target-weights',
          title: t('strategy.rules.s22.targets.title'),
          value: t('strategy.rules.s22.targets.value', { summary: weightSummary }),
          explanation: t('strategy.rules.s22.targets.explanation'),
          tone: 'neutral',
        })
      }

      const rebalanceThreshold = formatPercentMaybe(parseNumber(parameters.rebalance_threshold_pct))
      const rebalanceFrequencyRaw =
        typeof parameters.rebalance_frequency === 'string' ? parameters.rebalance_frequency : ''
      const rebalanceFrequencyLabel = rebalanceFrequencyRaw
        ? (() => {
            const frequencyMap: Record<string, string> = {
              daily: t('strategy.visualization.frequency.daily'),
              weekly: t('strategy.visualization.frequency.weekly'),
              biweekly: t('strategy.visualization.frequency.biweekly'),
              monthly: t('strategy.visualization.frequency.monthly'),
            }
            return frequencyMap[rebalanceFrequencyRaw] ?? humanizeText(rebalanceFrequencyRaw)
          })()
        : ''
      if (rebalanceThreshold || rebalanceFrequencyLabel) {
        const frequencyPart = rebalanceFrequencyLabel
          ? t('strategy.value.frequencyPart', { frequency: rebalanceFrequencyLabel })
          : ''
        rules.push({
          id: 's22-rebalance-rule',
          title: t('strategy.rules.s22.rebalance.title'),
          value: t('strategy.rules.s22.rebalance.value', {
            threshold: rebalanceThreshold ?? '-',
            frequencyPart,
          }),
          explanation: t('strategy.rules.s22.rebalance.explanation'),
          tone: 'warning',
        })
      }

      const volFloor = formatPercentMaybe(parseNumber(parameters.vol_floor))
      if (volFloor) {
        rules.push({
          id: 's22-vol-floor',
          title: t('strategy.rules.s22.volFloor.title'),
          value: t('strategy.rules.s22.volFloor.value', { percent: volFloor }),
          explanation: t('strategy.rules.s22.volFloor.explanation'),
          tone: 'neutral',
        })
      }
      break
    }
    default:
      break
  }

  return { rules, usedKeys }
}

export function buildStrategyIntroduction({
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
  currencyCode,
}: StrategyIntroInput) {
  if (!plan) return []

  const paragraphs: string[] = []
  const nextExecution = typeof parameters.next_execution_at === 'string' ? parameters.next_execution_at : ''
  const nextExecutionDate = nextExecution ? new Date(nextExecution) : null
  const nextExecutionLabel =
    nextExecutionDate && !Number.isNaN(nextExecutionDate.getTime())
      ? `${weekdayFormatter.format(nextExecutionDate)} ${dateFormatter.format(nextExecutionDate)}`
      : ''

  const formatMoney = (value: number | null) => {
    if (value === null || value === undefined) return '-'
    return formatCurrency(value, currencyCode)
  }

  switch (plan.strategy_id) {
    case 'S01': {
      const stopPrice = stopLossData?.stopLossPrice ?? null
      const stopPct = stopLossData?.stopLossPct ?? null
      const support = stopLossData?.supportLevel ?? null
      if (stopPrice !== null) {
        paragraphs.push(t('strategy.intro.s01.p1', { price: formatMoney(stopPrice) }))
      }
      const stopPctLabel = stopPct !== null ? formatPercentValue(stopPct) : null
      if (stopPctLabel) {
        paragraphs.push(t('strategy.intro.s01.p2', { percent: stopPctLabel }))
      }
      if (support !== null) {
        paragraphs.push(t('strategy.intro.s01.p3', { price: formatMoney(support) }))
      }
      break
    }
    case 'S02': {
      paragraphs.push(t('strategy.intro.s02.p1'))
      const firstOrder = ladderOrders[0]
      if (firstOrder?.triggerPrice || firstOrder?.orderAmount) {
        const priceLabel = firstOrder.triggerPrice
          ? formatMoney(firstOrder.triggerPrice)
          : t('strategy.value.nextTierPrice')
        const amountLabel = firstOrder.orderAmount
          ? formatMoney(firstOrder.orderAmount)
          : t('strategy.value.tierAmount')
        paragraphs.push(t('strategy.intro.s02.p2', { amount: amountLabel, price: priceLabel }))
      }
      const stepPct = formatPercentMaybe(parseNumber(parameters.price_step_pct))
      const multiplier = parseNumber(parameters.order_multiplier)
      const multiplierLabel =
        multiplier !== null
          ? t('strategy.value.multiplierValue', { value: formatNumber(multiplier, 2) })
          : t('strategy.value.fixedSize')
      if (stepPct || multiplier !== null) {
        const stepLabel = stepPct
          ? t('strategy.value.stepPerTier', { percent: stepPct })
          : t('strategy.value.eachTierDrop')
        paragraphs.push(t('strategy.intro.s02.p3', { stepLabel, multiplierLabel }))
      }
      break
    }
    case 'S03': {
      if (!trailingStop) break
      paragraphs.push(t('strategy.intro.s03.p1'))
      const trailPctLabel =
        typeof trailingStop.trailingStopPct === 'number'
          ? formatPercentMaybe(trailingStop.trailingStopPct)
          : t('strategy.value.setDistance')
      paragraphs.push(
        t('strategy.intro.s03.p2', {
          activationPrice: formatMoney(trailingStop.activationPrice),
          trailPercent: trailPctLabel ?? t('strategy.value.setDistance'),
        })
      )
      paragraphs.push(t('strategy.intro.s03.p3', { price: formatMoney(trailingStop.trailingStopPrice) }))
      break
    }
    case 'S04': {
      paragraphs.push(t('strategy.intro.s04.p1'))
      paragraphs.push(t('strategy.intro.s04.p2'))
      const firstLayer = takeProfitLayers[0]
      if (firstLayer?.targetPrice || firstLayer?.sellPct !== null) {
        const priceLabel = firstLayer.targetPrice
          ? formatMoney(firstLayer.targetPrice)
          : t('strategy.value.firstTarget')
        const sellLabel = formatPercentMaybe(firstLayer.sellPct) ?? t('strategy.value.portion')
        paragraphs.push(t('strategy.intro.s04.p3', { percent: sellLabel, price: priceLabel }))
      }
      break
    }
    case 'S05': {
      paragraphs.push(t('strategy.intro.s05.p1'))
      const amount = parseNumber(parameters.amount)
      const frequencyRaw = typeof parameters.frequency === 'string' ? parameters.frequency : ''
      const cadence = frequencyLabel
        ? frequencyLabel.toLowerCase()
        : humanizeText(frequencyRaw || t('strategy.value.onSchedule'))
      if (amount !== null) {
        paragraphs.push(t('strategy.intro.s05.p2', { amount: formatMoney(amount), cadence }))
      }
      if (nextExecutionLabel) {
        paragraphs.push(t('strategy.intro.s05.p3', { date: nextExecutionLabel }))
      }
      break
    }
    case 'S09': {
      paragraphs.push(t('strategy.intro.s09.p1'))
      paragraphs.push(t('strategy.intro.s09.p2'))
      const firstAddition = pyramidingAdditions[0]
      if (firstAddition) {
        const triggerLabel = formatPercentMaybe(firstAddition.triggerProfitPct)
        const amountLabel = firstAddition.additionAmount
          ? formatMoney(firstAddition.additionAmount)
          : t('strategy.value.nextAddSize')
        const triggerPart = triggerLabel
          ? t('strategy.value.plusPercent', { percent: triggerLabel })
          : t('strategy.value.nextProfitStep')
        if (triggerLabel || firstAddition.additionAmount) {
          paragraphs.push(t('strategy.intro.s09.p3', { amount: amountLabel, trigger: triggerPart }))
        }
      }
      break
    }
    case 'S16': {
      const triggerFunding = formatPercentMaybe(parseNumber(parameters.trigger_funding_rate))
      const triggerBasis = formatPercentMaybe(parseNumber(parameters.trigger_basis_pct_max))
      const hedgeNotional = parseNumber(parameters.hedge_notional_usd)
      paragraphs.push(t('strategy.intro.s16.p1'))
      paragraphs.push(t('strategy.intro.s16.p2'))
      if (triggerFunding || triggerBasis || hedgeNotional !== null) {
        const hedgeLabel = hedgeNotional !== null ? formatMoney(hedgeNotional) : t('strategy.value.plannedNotional')
        paragraphs.push(
          t('strategy.intro.s16.p3', {
            funding: triggerFunding ?? '-',
            basis: triggerBasis ?? '-',
            hedge: hedgeLabel,
          })
        )
      }
      break
    }
    case 'S18': {
      const trendAction = typeof parameters.trend_action === 'string' ? parameters.trend_action : ''
      paragraphs.push(t('strategy.intro.s18.p1'))
      paragraphs.push(t('strategy.intro.s18.p2'))
      if (trendAction) {
        paragraphs.push(t('strategy.intro.s18.p3', { action: humanizeText(trendAction) }))
      }
      break
    }
    case 'S22': {
      const rebalanceThreshold = formatPercentMaybe(parseNumber(parameters.rebalance_threshold_pct))
      const rebalanceFrequencyRaw =
        typeof parameters.rebalance_frequency === 'string' ? parameters.rebalance_frequency : ''
      const rebalanceFrequencyLabel = rebalanceFrequencyRaw
        ? (() => {
            const frequencyMap: Record<string, string> = {
              daily: t('strategy.visualization.frequency.daily'),
              weekly: t('strategy.visualization.frequency.weekly'),
              biweekly: t('strategy.visualization.frequency.biweekly'),
              monthly: t('strategy.visualization.frequency.monthly'),
            }
            return frequencyMap[rebalanceFrequencyRaw] ?? humanizeText(rebalanceFrequencyRaw)
          })()
        : ''
      const frequencyPart = rebalanceFrequencyLabel
        ? t('strategy.value.frequencyPart', { frequency: rebalanceFrequencyLabel })
        : ''
      paragraphs.push(t('strategy.intro.s22.p1'))
      paragraphs.push(t('strategy.intro.s22.p2'))
      if (rebalanceThreshold) {
        paragraphs.push(t('strategy.intro.s22.p3', { threshold: rebalanceThreshold, frequencyPart }))
      }
      break
    }
    default:
      break
  }

  const cleaned = paragraphs.map((text) => text.trim()).filter((text) => text.length > 0)
  if (cleaned.length > 0) {
    return cleaned.slice(0, 3)
  }
  const fallback = [plan.rationale, plan.expected_outcome]
    .map((text) => (typeof text === 'string' ? text.trim() : ''))
    .filter((text) => text.length > 0)
  return fallback.slice(0, 3)
}
