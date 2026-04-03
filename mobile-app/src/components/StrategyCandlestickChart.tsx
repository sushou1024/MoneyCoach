import React, { useMemo, useState } from 'react'
import { LayoutChangeEvent, StyleSheet, Text, View } from 'react-native'
import Svg, { Line, Polyline, Rect, Text as SvgText } from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'

type ChartPadding = { top: number; right: number; bottom: number; left: number }

export type TraceLine = {
  value: number
  color: string
  label?: string
  dashed?: boolean
}

export type SeriesLine = {
  values: number[]
  color: string
  dashed?: boolean
  opacity?: number
  strokeWidth?: number
}

export type CandlestickPoint = {
  open: number
  high: number
  low: number
  close: number
  time: string | number
}

const labelHeight = 20
const labelGap = 16

export function valueToChartY({
  value,
  minValue,
  maxValue,
  chartTop,
  chartHeight,
}: {
  value: number
  minValue: number
  maxValue: number
  chartTop: number
  chartHeight: number
}) {
  const range = maxValue - minValue || 1
  return chartTop + chartHeight - ((value - minValue) / range) * chartHeight
}

export function StrategyCandlestickChart({
  candles,
  lines = [],
  series = [],
  height = 170,
  locale,
  formatPrice,
  axisSide = 'left',
  showXAxis = true,
  showYAxis = true,
  padding,
  backgroundColor,
}: {
  candles: CandlestickPoint[]
  lines?: TraceLine[]
  series?: SeriesLine[]
  height?: number
  locale?: string
  formatPrice?: (value: number) => string
  axisSide?: 'left' | 'right'
  showXAxis?: boolean
  showYAxis?: boolean
  padding?: ChartPadding
  backgroundColor?: string
}) {
  const theme = useTheme()
  const [width, setWidth] = useState(0)

  const { candleShapes, linePositions, minValue, maxValue, chartMetrics, seriesSegments } = useMemo(() => {
    if (!candles.length || width === 0) {
      return {
        candleShapes: [],
        linePositions: [],
        minValue: 0,
        maxValue: 1,
        chartMetrics: { left: 0, top: 0, width: 0, height: 0, step: 0 },
        seriesSegments: [] as (SeriesLine & { segments: string[] })[],
      }
    }

    const resolvedPadding = padding ?? {
      top: 8,
      right: showYAxis && axisSide === 'right' ? 48 : 10,
      bottom: showXAxis ? 22 : 8,
      left: showYAxis && axisSide === 'left' ? 48 : 10,
    }
    const chartWidth = Math.max(1, width - resolvedPadding.left - resolvedPadding.right)
    const chartHeight = Math.max(1, height - resolvedPadding.top - resolvedPadding.bottom)

    const lineValues = lines.map((line) => line.value).filter((value) => Number.isFinite(value))
    const seriesValues = series
      .flatMap((line) => line.values)
      .filter((value) => typeof value === 'number' && Number.isFinite(value))
    const highs = candles.map((candle) => candle.high).filter((value) => Number.isFinite(value))
    const lows = candles.map((candle) => candle.low).filter((value) => Number.isFinite(value))
    const allValues = [...highs, ...lows, ...lineValues, ...seriesValues]

    const min = Math.min(...allValues)
    const max = Math.max(...allValues)
    const range = max - min
    const valuePadding = range === 0 ? 1 : range * 0.06
    const paddedMin = min - valuePadding
    const paddedMax = max + valuePadding
    const safeRange = paddedMax - paddedMin || 1

    const toY = (value: number) => chartHeight - ((value - paddedMin) / safeRange) * chartHeight

    const step = chartWidth / candles.length
    const bodyWidth = Math.max(2, Math.min(10, step * 0.6))

    const shapes = candles.map((candle, index) => {
      const x = resolvedPadding.left + index * step + step / 2
      const openY = resolvedPadding.top + toY(candle.open)
      const closeY = resolvedPadding.top + toY(candle.close)
      const highY = resolvedPadding.top + toY(candle.high)
      const lowY = resolvedPadding.top + toY(candle.low)
      const bodyTop = Math.min(openY, closeY)
      const bodyHeight = Math.max(1, Math.abs(closeY - openY))
      return {
        x,
        bodyWidth,
        bodyTop,
        bodyHeight,
        highY,
        lowY,
        isUp: candle.close >= candle.open,
      }
    })

    const basePositions = lines.map((line) => ({
      ...line,
      y: resolvedPadding.top + toY(line.value),
    }))

    const sorted = [...basePositions].sort((a, b) => a.y - b.y)
    let lastY = -Infinity
    const adjusted = sorted.map((item) => {
      const minLabel = resolvedPadding.top + labelHeight / 2
      const maxLabel = resolvedPadding.top + chartHeight - labelHeight / 2
      const clamped = Math.min(Math.max(item.y, minLabel), maxLabel)
      const nextY = Math.max(clamped, lastY + labelGap)
      lastY = nextY
      return { ...item, labelY: nextY }
    })

    const overflow = lastY - (resolvedPadding.top + chartHeight - labelHeight / 2)
    const shifted =
      overflow > 0
        ? adjusted.map((item) => ({
            ...item,
            labelY: Math.max(resolvedPadding.top + labelHeight / 2, item.labelY - overflow),
          }))
        : adjusted

    const seriesSegmentData = series
      .map((line) => {
        const segments: string[] = []
        let current: string[] = []
        line.values.forEach((value, index) => {
          if (!Number.isFinite(value)) {
            if (current.length > 0) {
              segments.push(current.join(' '))
              current = []
            }
            return
          }
          const x = resolvedPadding.left + index * step + step / 2
          const y = resolvedPadding.top + toY(value)
          current.push(`${x},${y}`)
        })
        if (current.length > 0) {
          segments.push(current.join(' '))
        }
        return { ...line, segments }
      })
      .filter((line) => line.segments.length > 0)

    return {
      candleShapes: shapes,
      linePositions: shifted,
      minValue: paddedMin,
      maxValue: paddedMax,
      chartMetrics: {
        left: resolvedPadding.left,
        top: resolvedPadding.top,
        width: chartWidth,
        height: chartHeight,
        step,
      },
      seriesSegments: seriesSegmentData,
    }
  }, [axisSide, candles, height, lines, series, showXAxis, showYAxis, width, padding])

  const dateFormatter = useMemo(
    () => new Intl.DateTimeFormat(locale ?? undefined, { month: 'short', day: 'numeric' }),
    [locale]
  )

  const priceFormatter = useMemo(() => formatPrice ?? ((value: number) => value.toFixed(2)), [formatPrice])

  const styles = useMemo(() => createStyles(theme), [theme])

  const handleLayout = (event: LayoutChangeEvent) => {
    setWidth(event.nativeEvent.layout.width)
  }

  const showChart = width > 0 && candles.length > 1
  const valueToY = (value: number) =>
    valueToChartY({
      value,
      minValue,
      maxValue,
      chartTop: chartMetrics.top,
      chartHeight: chartMetrics.height,
    })

  const chartBackground = backgroundColor ?? theme.colors.surface

  return (
    <View style={[styles.container, { height }]} onLayout={handleLayout}>
      {showChart ? (
        <>
          <Svg width={width} height={height}>
            <Rect x={0} y={0} width={width} height={height} fill={chartBackground} />
            {showYAxis ? (
              <Line
                x1={axisSide === 'right' ? chartMetrics.left + chartMetrics.width : chartMetrics.left}
                x2={axisSide === 'right' ? chartMetrics.left + chartMetrics.width : chartMetrics.left}
                y1={chartMetrics.top}
                y2={chartMetrics.top + chartMetrics.height}
                stroke={theme.colors.border}
                strokeWidth={1}
              />
            ) : null}
            {showXAxis ? (
              <Line
                x1={chartMetrics.left}
                x2={chartMetrics.left + chartMetrics.width}
                y1={chartMetrics.top + chartMetrics.height}
                y2={chartMetrics.top + chartMetrics.height}
                stroke={theme.colors.border}
                strokeWidth={1}
              />
            ) : null}
            {candleShapes.map((shape, index) => {
              const color = shape.isUp ? theme.colors.success : theme.colors.danger
              return (
                <React.Fragment key={`candle-${index}`}>
                  <Line x1={shape.x} x2={shape.x} y1={shape.highY} y2={shape.lowY} stroke={color} strokeWidth={1} />
                  <Rect
                    x={shape.x - shape.bodyWidth / 2}
                    width={shape.bodyWidth}
                    y={shape.bodyTop}
                    height={shape.bodyHeight}
                    fill={color}
                    rx={2}
                  />
                </React.Fragment>
              )
            })}
            {lines.map((line, index) => {
              const lineValue = Number.isFinite(line.value) ? line.value : minValue
              const y = valueToY(lineValue)
              return (
                <Line
                  key={`line-${index}`}
                  x1={chartMetrics.left}
                  x2={chartMetrics.left + chartMetrics.width}
                  y1={y}
                  y2={y}
                  stroke={line.color}
                  strokeWidth={1}
                  strokeDasharray={line.dashed ? '4 4' : undefined}
                  opacity={line.dashed ? 0.5 : 0.7}
                />
              )
            })}
            {seriesSegments.flatMap((line, lineIndex) =>
              line.segments.map((segment, segmentIndex) => (
                <Polyline
                  key={`series-${lineIndex}-${segmentIndex}`}
                  points={segment}
                  fill="none"
                  stroke={line.color}
                  strokeWidth={line.strokeWidth ?? 1.5}
                  strokeDasharray={line.dashed ? '4 4' : undefined}
                  opacity={line.opacity ?? 0.8}
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              ))
            )}
            {showYAxis
              ? (() => {
                  const ticks = [maxValue, (maxValue + minValue) / 2, minValue]
                  const axisX = axisSide === 'right' ? chartMetrics.left + chartMetrics.width : chartMetrics.left
                  const tickEnd = axisSide === 'right' ? axisX - 4 : axisX + 4
                  const labelX = axisSide === 'right' ? axisX + 6 : 4
                  const labelAnchor = axisSide === 'right' ? 'start' : 'start'
                  return ticks.map((value, index) => {
                    const y = valueToY(value)
                    return (
                      <React.Fragment key={`price-${index}`}>
                        <Line x1={axisX} x2={tickEnd} y1={y} y2={y} stroke={theme.colors.border} strokeWidth={1} />
                        <SvgText
                          x={labelX}
                          y={y + 3}
                          fontSize={10}
                          fill={theme.colors.muted}
                          fontFamily={theme.fonts.bodyMedium}
                          textAnchor={labelAnchor}
                        >
                          {priceFormatter(value)}
                        </SvgText>
                      </React.Fragment>
                    )
                  })
                })()
              : null}
            {showXAxis
              ? (() => {
                  const lastIndex = candles.length - 1
                  const midIndex = Math.floor(candles.length / 2)
                  const indices = Array.from(new Set([0, midIndex, lastIndex])).filter((idx) => idx >= 0)
                  return indices.map((idx) => {
                    const candle = candles[idx]
                    const x = chartMetrics.left + idx * chartMetrics.step + chartMetrics.step / 2
                    const label = dateFormatter.format(new Date(candle.time))
                    return (
                      <SvgText
                        key={`time-${idx}`}
                        x={x}
                        y={chartMetrics.top + chartMetrics.height + 16}
                        fontSize={10}
                        fill={theme.colors.muted}
                        fontFamily={theme.fonts.bodyMedium}
                        textAnchor="middle"
                      >
                        {label}
                      </SvgText>
                    )
                  })
                })()
              : null}
          </Svg>
          {linePositions.map((line, index) =>
            line.label ? (
              <View
                key={`label-${index}`}
                style={[styles.label, { top: line.labelY - labelHeight / 2 }]}
                pointerEvents="none"
              >
                <View style={[styles.labelDot, { backgroundColor: line.color }]} />
                <Text numberOfLines={1} style={styles.labelText}>
                  {line.label}
                </Text>
              </View>
            ) : null
          )}
        </>
      ) : null}
    </View>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    container: {
      position: 'relative',
      borderRadius: theme.radius.md,
      backgroundColor: 'transparent',
      overflow: 'hidden',
    },
    label: {
      position: 'absolute',
      right: theme.spacing.sm,
      flexDirection: 'row',
      alignItems: 'center',
      paddingHorizontal: theme.spacing.sm,
      height: labelHeight,
      borderRadius: 999,
      backgroundColor: theme.colors.surface,
      borderWidth: 1,
      borderColor: theme.colors.border,
      gap: 6,
      maxWidth: 240,
    },
    labelDot: {
      width: 6,
      height: 6,
      borderRadius: 999,
    },
    labelText: {
      fontSize: 11,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyMedium,
    },
  })
