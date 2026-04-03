import React, { useMemo, useState } from 'react'
import { LayoutChangeEvent, View } from 'react-native'
import Svg, { Line, Polyline, Text as SvgText } from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'

type ChartPadding = { top: number; right: number; bottom: number; left: number }
type HorizontalLine = {
  value: number
  color?: string
  dashed?: boolean
  opacity?: number
  strokeWidth?: number
}

export function LineChart({
  data,
  height = 120,
  strokeColor,
  minValue,
  maxValue,
  strokeWidth = 2,
  opacity = 1,
  axisSide = 'left',
  showYAxis = false,
  showXAxis = false,
  padding,
  yTicks,
  formatY,
  alignToCandles = false,
  horizontalLines,
}: {
  data: number[]
  height?: number
  strokeColor?: string
  minValue?: number
  maxValue?: number
  strokeWidth?: number
  opacity?: number
  axisSide?: 'left' | 'right'
  showYAxis?: boolean
  showXAxis?: boolean
  padding?: ChartPadding
  yTicks?: number[]
  formatY?: (value: number) => string
  alignToCandles?: boolean
  horizontalLines?: HorizontalLine[]
}) {
  const theme = useTheme()
  const [width, setWidth] = useState(0)

  const resolvedPadding = useMemo(
    () =>
      padding ?? {
        top: 6,
        right: showYAxis && axisSide === 'right' ? 36 : 10,
        bottom: showXAxis ? 18 : 6,
        left: showYAxis && axisSide === 'left' ? 36 : 10,
      },
    [axisSide, padding, showXAxis, showYAxis]
  )

  const segments = useMemo(() => {
    if (!data.length || width === 0) return [] as string[]
    const finiteValues = data.filter((value) => Number.isFinite(value))
    if (!finiteValues.length) return [] as string[]
    const min = typeof minValue === 'number' ? minValue : Math.min(...finiteValues)
    const max = typeof maxValue === 'number' ? maxValue : Math.max(...finiteValues)
    const range = max - min
    const safeRange = range === 0 ? 1 : range
    const chartWidth = Math.max(1, width - resolvedPadding.left - resolvedPadding.right)
    const chartHeight = Math.max(1, height - resolvedPadding.top - resolvedPadding.bottom)
    const step = data.length > 0 ? chartWidth / data.length : chartWidth
    const output: string[] = []
    let current: string[] = []
    data.forEach((value, index) => {
      if (!Number.isFinite(value)) {
        if (current.length > 0) {
          output.push(current.join(' '))
          current = []
        }
        return
      }
      const x = alignToCandles
        ? resolvedPadding.left + index * step + step / 2
        : resolvedPadding.left + (index / Math.max(1, data.length - 1)) * chartWidth
      const y = resolvedPadding.top + chartHeight - ((value - min) / safeRange) * chartHeight
      current.push(`${x},${y}`)
    })
    if (current.length > 0) {
      output.push(current.join(' '))
    }
    return output
  }, [alignToCandles, data, height, maxValue, minValue, resolvedPadding, width])

  const handleLayout = (event: LayoutChangeEvent) => {
    setWidth(event.nativeEvent.layout.width)
  }

  const chartMetrics = useMemo(() => {
    const chartWidth = Math.max(1, width - resolvedPadding.left - resolvedPadding.right)
    const chartHeight = Math.max(1, height - resolvedPadding.top - resolvedPadding.bottom)
    return { chartWidth, chartHeight }
  }, [height, resolvedPadding.left, resolvedPadding.right, resolvedPadding.top, resolvedPadding.bottom, width])

  const tickValues = useMemo(() => {
    if (!showYAxis) return []
    const finiteValues = data.filter((value) => Number.isFinite(value))
    if (!finiteValues.length) return []
    const min = typeof minValue === 'number' ? minValue : Math.min(...finiteValues)
    const max = typeof maxValue === 'number' ? maxValue : Math.max(...finiteValues)
    if (yTicks && yTicks.length > 0) return yTicks
    return [max, (max + min) / 2, min]
  }, [data, maxValue, minValue, showYAxis, yTicks])

  const tickRange = useMemo(() => {
    const finiteValues = data.filter((value) => Number.isFinite(value))
    if (!finiteValues.length) {
      return { min: 0, max: 1 }
    }
    const min = typeof minValue === 'number' ? minValue : Math.min(...finiteValues)
    const max = typeof maxValue === 'number' ? maxValue : Math.max(...finiteValues)
    return { min, max }
  }, [data, maxValue, minValue])

  const valueToY = (value: number) => {
    const range = tickRange.max - tickRange.min
    const safeRange = range === 0 ? 1 : range
    return (
      resolvedPadding.top + chartMetrics.chartHeight - ((value - tickRange.min) / safeRange) * chartMetrics.chartHeight
    )
  }

  return (
    <View style={{ height }} onLayout={handleLayout}>
      {width > 0 && data.length > 1 ? (
        <Svg width={width} height={height}>
          {showYAxis ? (
            <Line
              x1={axisSide === 'right' ? resolvedPadding.left + chartMetrics.chartWidth : resolvedPadding.left}
              x2={axisSide === 'right' ? resolvedPadding.left + chartMetrics.chartWidth : resolvedPadding.left}
              y1={resolvedPadding.top}
              y2={resolvedPadding.top + chartMetrics.chartHeight}
              stroke={theme.colors.border}
              strokeWidth={1}
            />
          ) : null}
          {showXAxis ? (
            <Line
              x1={resolvedPadding.left}
              x2={resolvedPadding.left + chartMetrics.chartWidth}
              y1={resolvedPadding.top + chartMetrics.chartHeight}
              y2={resolvedPadding.top + chartMetrics.chartHeight}
              stroke={theme.colors.border}
              strokeWidth={1}
            />
          ) : null}
          {showYAxis
            ? tickValues.map((value, index) => {
                const y = valueToY(value)
                const axisX =
                  axisSide === 'right' ? resolvedPadding.left + chartMetrics.chartWidth : resolvedPadding.left
                const tickEnd = axisSide === 'right' ? axisX - 4 : axisX + 4
                const labelX = axisSide === 'right' ? axisX + 6 : 4
                const labelAnchor = axisSide === 'right' ? 'start' : 'start'
                return (
                  <React.Fragment key={`tick-${index}`}>
                    <Line x1={axisX} x2={tickEnd} y1={y} y2={y} stroke={theme.colors.border} strokeWidth={1} />
                    <SvgText
                      x={labelX}
                      y={y + 3}
                      fontSize={10}
                      fill={theme.colors.muted}
                      fontFamily={theme.fonts.bodyMedium}
                      textAnchor={labelAnchor}
                    >
                      {formatY ? formatY(value) : value.toString()}
                    </SvgText>
                  </React.Fragment>
                )
              })
            : null}
          {horizontalLines
            ?.filter((line) => Number.isFinite(line.value))
            .map((line, index) => {
              const y = valueToY(line.value)
              return (
                <Line
                  key={`h-${index}`}
                  x1={resolvedPadding.left}
                  x2={resolvedPadding.left + chartMetrics.chartWidth}
                  y1={y}
                  y2={y}
                  stroke={line.color ?? theme.colors.border}
                  strokeWidth={line.strokeWidth ?? 1}
                  strokeDasharray={line.dashed ? '4 4' : undefined}
                  opacity={line.opacity ?? 0.7}
                />
              )
            })}
          {segments.map((segment, index) => (
            <Polyline
              key={`segment-${index}`}
              points={segment}
              fill="none"
              stroke={strokeColor ?? theme.colors.accent}
              strokeWidth={strokeWidth}
              strokeLinecap="round"
              strokeLinejoin="round"
              opacity={opacity}
            />
          ))}
        </Svg>
      ) : null}
    </View>
  )
}
