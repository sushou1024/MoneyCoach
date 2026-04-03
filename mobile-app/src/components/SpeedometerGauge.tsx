import React, { useMemo } from 'react'
import { View } from 'react-native'
import Svg, { Circle, G, Line, Path, Text as SvgText } from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'

type SpeedometerGaugeProps = {
  value?: number | null
  label: string
  size?: number
  locked?: boolean
}

function clamp(value: number) {
  if (!Number.isFinite(value)) return 0
  if (value < 0) return 0
  if (value > 100) return 100
  return value
}

function scoreColor(score: number | null | undefined, theme: ReturnType<typeof useTheme>) {
  if (score === undefined || score === null) return theme.colors.muted
  if (score < 50) return theme.colors.danger
  if (score < 70) return theme.colors.warning
  if (score < 90) return theme.colors.success
  return theme.colors.accent
}

function polarToCartesian(cx: number, cy: number, radius: number, angleInDegrees: number) {
  const angleInRadians = (Math.PI / 180) * angleInDegrees
  return {
    x: cx + radius * Math.cos(angleInRadians),
    y: cy - radius * Math.sin(angleInRadians),
  }
}

function describeArc(cx: number, cy: number, radius: number, startAngle: number, endAngle: number) {
  const start = polarToCartesian(cx, cy, radius, startAngle)
  const end = polarToCartesian(cx, cy, radius, endAngle)
  const largeArcFlag = Math.abs(endAngle - startAngle) > 180 ? 1 : 0
  return `M ${start.x} ${start.y} A ${radius} ${radius} 0 ${largeArcFlag} 1 ${end.x} ${end.y}`
}

export function SpeedometerGauge({ value, label, size = 140, locked }: SpeedometerGaugeProps) {
  const theme = useTheme()
  const clampedValue = value === undefined || value === null ? null : clamp(value)
  const displayValue = locked ? '—' : clampedValue !== null ? Math.round(clampedValue).toString() : '--'
  const needleColor = locked ? theme.colors.border : scoreColor(clampedValue, theme)
  const needleOpacity = locked ? 0.5 : clampedValue === null ? 0.6 : 0.95
  const textColor = locked ? theme.colors.muted : theme.colors.ink

  const strokeWidth = Math.max(8, Math.round(size * 0.07))
  const segmentWidth = Math.max(6, Math.round(strokeWidth * 0.9))
  const tickLength = Math.max(6, Math.round(size * 0.05))
  const radius = Math.round(size * 0.42)
  const centerX = size / 2
  const centerY = radius + strokeWidth / 2 + tickLength
  const gaugeHeight = centerY + strokeWidth / 2
  const valueFontSize = Math.max(18, Math.round(size * 0.2))
  const labelFontSize = Math.max(12, Math.round(size * 0.085))
  const valueOffset = Math.max(6, Math.round(valueFontSize * 0.38))
  const valueY = gaugeHeight + valueOffset
  const labelGap = Math.round(labelFontSize * 1.35)
  const labelY = valueY + labelGap
  const svgHeight = Math.ceil(labelY + labelFontSize * 1.2)

  const trackPath = useMemo(() => describeArc(centerX, centerY, radius, 180, 0), [centerX, centerY, radius])
  const valueAngle = clampedValue === null ? 180 : 180 - (clampedValue / 100) * 180
  const needleLength = radius - strokeWidth * 1.35
  const needleEnd = polarToCartesian(centerX, centerY, needleLength, valueAngle)

  const ranges = useMemo(
    () => [
      { from: 0, to: 50, color: theme.colors.danger },
      { from: 50, to: 70, color: theme.colors.warning },
      { from: 70, to: 90, color: theme.colors.success },
      { from: 90, to: 100, color: theme.colors.accent },
    ],
    [theme.colors.accent, theme.colors.danger, theme.colors.success, theme.colors.warning]
  )

  const rangePaths = useMemo(
    () =>
      ranges.map((range, index) => {
        const startAngle = 180 - (range.from / 100) * 180
        const endAngle = 180 - (range.to / 100) * 180
        return {
          key: `range-${index}`,
          d: describeArc(centerX, centerY, radius, startAngle, endAngle),
          color: range.color,
        }
      }),
    [centerX, centerY, radius, ranges]
  )

  const ticks = useMemo(() => [0, 25, 50, 75, 100], [])
  const tickMarks = useMemo(
    () =>
      ticks.map((tick) => {
        const angle = 180 - (tick / 100) * 180
        const outer = polarToCartesian(centerX, centerY, radius + strokeWidth / 2 + 2, angle)
        const inner = polarToCartesian(centerX, centerY, radius - strokeWidth / 2 - tickLength, angle)
        return { tick, outer, inner }
      }),
    [centerX, centerY, radius, strokeWidth, tickLength, ticks]
  )

  const dialShadow = theme.colors.surfaceElevated
  const dialOpacity = 0.7

  return (
    <View style={{ width: size, height: svgHeight, alignItems: 'center' }}>
      <Svg width={size} height={svgHeight} viewBox={`0 0 ${size} ${svgHeight}`}>
        <G>
          <Path
            d={trackPath}
            fill="none"
            stroke={theme.colors.border}
            strokeWidth={strokeWidth}
            strokeLinecap="round"
          />
          {rangePaths.map((range) => (
            <Path
              key={range.key}
              d={range.d}
              fill="none"
              stroke={range.color}
              strokeWidth={segmentWidth}
              strokeLinecap="round"
              strokeOpacity={0.85}
            />
          ))}
          {tickMarks.map((tick) => (
            <Line
              key={`tick-${tick.tick}`}
              x1={tick.outer.x}
              y1={tick.outer.y}
              x2={tick.inner.x}
              y2={tick.inner.y}
              stroke={theme.colors.border}
              strokeOpacity={0.7}
              strokeWidth={1.2}
            />
          ))}
        </G>

        <G>
          <Path
            d={trackPath}
            fill="none"
            stroke={dialShadow}
            strokeWidth={strokeWidth * 0.45}
            strokeOpacity={dialOpacity}
          />
          <Line
            x1={centerX}
            y1={centerY}
            x2={needleEnd.x}
            y2={needleEnd.y}
            stroke={needleColor}
            strokeWidth={2.2}
            strokeLinecap="round"
            strokeOpacity={needleOpacity}
          />
          <Circle cx={centerX} cy={centerY} r={Math.max(4, strokeWidth * 0.45)} fill={needleColor} />
          <Circle
            cx={centerX}
            cy={centerY}
            r={Math.max(2, strokeWidth * 0.2)}
            fill={theme.colors.surface}
            opacity={0.9}
          />
        </G>

        <SvgText
          x={centerX}
          y={valueY}
          fill={textColor}
          fontSize={valueFontSize}
          fontFamily={theme.fonts.display}
          textAnchor="middle"
          alignmentBaseline="central"
        >
          {displayValue}
        </SvgText>
        <SvgText
          x={centerX}
          y={labelY}
          fill={theme.colors.muted}
          fontSize={labelFontSize}
          fontFamily={theme.fonts.bodyMedium}
          textAnchor="middle"
          alignmentBaseline="hanging"
        >
          {label}
        </SvgText>
      </Svg>
    </View>
  )
}
