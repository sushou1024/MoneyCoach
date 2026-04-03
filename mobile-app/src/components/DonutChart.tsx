import React, { useMemo } from 'react'
import { View } from 'react-native'
import Svg, { Circle, G } from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'

type DonutSegment = {
  value: number
  color?: string
}

export function DonutChart({
  size = 180,
  strokeWidth = 18,
  segments,
}: {
  size?: number
  strokeWidth?: number
  segments: DonutSegment[]
}) {
  const theme = useTheme()
  const { normalized, total } = useMemo(() => {
    const filtered = segments.filter((segment) => Number.isFinite(segment.value) && segment.value > 0)
    const sum = filtered.reduce((acc, segment) => acc + segment.value, 0)
    return { normalized: filtered, total: sum }
  }, [segments])

  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius

  let offset = 0

  return (
    <View style={{ width: size, height: size }}>
      <Svg width={size} height={size}>
        <Circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke={theme.colors.border}
          strokeWidth={strokeWidth}
          fill="none"
        />
        <G rotation={-90} originX={size / 2} originY={size / 2}>
          {normalized.map((segment, index) => {
            const length = total > 0 ? (segment.value / total) * circumference : 0
            const dashArray = `${length} ${circumference - length}`
            const dashOffset = -offset
            offset += length
            return (
              <Circle
                key={`segment-${index}`}
                cx={size / 2}
                cy={size / 2}
                r={radius}
                stroke={segment.color ?? theme.colors.accent}
                strokeWidth={strokeWidth}
                strokeDasharray={dashArray}
                strokeDashoffset={dashOffset}
                strokeLinecap="round"
                fill="none"
              />
            )
          })}
        </G>
      </Svg>
    </View>
  )
}
