import React, { useMemo, useState } from 'react'
import { LayoutChangeEvent, StyleSheet, Text, View } from 'react-native'
import Svg, { Line, Polyline } from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'

export type TraceLine = {
  value: number
  color: string
  label?: string
  dashed?: boolean
}

const labelHeight = 20
const labelGap = 16

export function StrategyTraceChart({
  data,
  lines = [],
  height = 170,
  strokeColor,
}: {
  data: number[]
  lines?: TraceLine[]
  height?: number
  strokeColor?: string
}) {
  const theme = useTheme()
  const [width, setWidth] = useState(0)

  const { points, linePositions, minValue, maxValue } = useMemo(() => {
    if (!data.length || width === 0) {
      return { points: '', linePositions: [], minValue: 0, maxValue: 1 }
    }

    const lineValues = lines.map((line) => line.value).filter((value) => Number.isFinite(value))
    const allValues = [...data, ...lineValues]
    const min = Math.min(...allValues)
    const max = Math.max(...allValues)
    const range = max - min
    const padding = range === 0 ? 1 : range * 0.08
    const paddedMin = min - padding
    const paddedMax = max + padding
    const safeRange = paddedMax - paddedMin || 1

    const toY = (value: number) => height - ((value - paddedMin) / safeRange) * height

    const chartPoints = data
      .map((value, index) => {
        const x = (index / Math.max(1, data.length - 1)) * width
        const y = toY(value)
        return `${x},${y}`
      })
      .join(' ')

    const basePositions = lines.map((line) => ({
      ...line,
      y: toY(line.value),
    }))

    const sorted = [...basePositions].sort((a, b) => a.y - b.y)
    let lastY = -Infinity
    const adjusted = sorted.map((item) => {
      const clamped = Math.min(Math.max(item.y, labelHeight / 2), height - labelHeight / 2)
      const nextY = Math.max(clamped, lastY + labelGap)
      lastY = nextY
      return { ...item, labelY: nextY }
    })

    const overflow = lastY - (height - labelHeight / 2)
    const shifted =
      overflow > 0
        ? adjusted.map((item) => ({ ...item, labelY: Math.max(labelHeight / 2, item.labelY - overflow) }))
        : adjusted

    return { points: chartPoints, linePositions: shifted, minValue: paddedMin, maxValue: paddedMax }
  }, [data, height, lines, width])

  const styles = useMemo(() => createStyles(theme), [theme])

  const handleLayout = (event: LayoutChangeEvent) => {
    setWidth(event.nativeEvent.layout.width)
  }

  const showChart = width > 0 && data.length > 1

  return (
    <View style={[styles.container, { height }]} onLayout={handleLayout}>
      {showChart ? (
        <>
          <Svg width={width} height={height}>
            <Polyline
              points={points}
              fill="none"
              stroke={strokeColor ?? theme.colors.accent}
              strokeWidth={2}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            {lines.map((line, index) => {
              const lineValue = Number.isFinite(line.value) ? line.value : minValue
              const y = height - ((lineValue - minValue) / Math.max(1, maxValue - minValue)) * height
              return (
                <Line
                  key={`line-${index}`}
                  x1={0}
                  x2={width}
                  y1={y}
                  y2={y}
                  stroke={line.color}
                  strokeWidth={1}
                  strokeDasharray={line.dashed ? '4 4' : undefined}
                  opacity={line.dashed ? 0.5 : 0.7}
                />
              )
            })}
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
      backgroundColor: theme.colors.surface,
      borderWidth: 1,
      borderColor: theme.colors.border,
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
