import React, { useId, useMemo } from 'react'
import { View } from 'react-native'
import Svg, {
  Circle,
  Defs,
  G,
  Line,
  LinearGradient,
  Path,
  Polygon,
  RadialGradient,
  Stop,
  Text as SvgText,
  type AlignmentBaseline,
  type TextAnchor,
} from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'

export type RadarValues = {
  liquidity: number
  diversification: number
  alpha: number
  drawdown: number
}

type RadarLabels = Partial<Record<keyof RadarValues, string>>

const axes: (keyof RadarValues)[] = ['liquidity', 'diversification', 'alpha', 'drawdown']

function clamp(value: number) {
  if (!Number.isFinite(value)) return 0
  if (value < 0) return 0
  if (value > 100) return 100
  return value
}

function clampAngle(angle: number) {
  return ((angle % 360) + 360) % 360
}

function getDefaultTextAnchor(angle: number): TextAnchor {
  const adjusted = clampAngle(angle)
  if (adjusted <= 30 || adjusted >= 330) return 'middle'
  if (adjusted <= 210 && adjusted >= 150) return 'middle'
  if (adjusted <= 180) return 'end'
  return 'start'
}

function getDefaultBaseline(angle: number): AlignmentBaseline | 'auto' {
  const adjusted = clampAngle(angle)
  if (adjusted <= 30 || adjusted >= 330) return 'hanging'
  if (adjusted <= 210 && adjusted >= 150) return 'auto'
  return 'central'
}

function lerp(start: number, end: number, t: number) {
  return start * (1 - t) + end * t
}

function toTitle(label: string) {
  if (!label) return label
  return `${label[0].toUpperCase()}${label.slice(1)}`
}

function estimateLabelWidth(label: string, fontSize: number) {
  // Approximate text width so we can size the SVG without measuring text layout.
  // This keeps the chart layout deterministic and avoids per-platform text measurement.
  // Width factors are a heuristic: ASCII is narrower, non-ASCII tends to be wider.
  if (!label) return fontSize
  const isNonAscii = /[^\x00-\x7F]/.test(label)
  const widthFactor = isNonAscii ? 0.9 : 0.6
  return label.length * fontSize * widthFactor + fontSize * 0.2
}

type RadarChartProps = {
  size?: number
  values: RadarValues
  labels?: RadarLabels
  divisions?: number
  showLabels?: boolean
}

export function RadarChart({ size = 220, values, labels, divisions = 5, showLabels = true }: RadarChartProps) {
  const theme = useTheme()
  const gradientBaseId = useId().replace(/:/g, '-')
  const centerGradientId = `${gradientBaseId}-center`
  const axisColors = useMemo(
    () => [theme.colors.accent, theme.colors.success, theme.colors.warning, theme.colors.danger],
    [theme.colors.accent, theme.colors.danger, theme.colors.success, theme.colors.warning]
  )
  const labelFontSize = 11
  const baseCenter = size / 2
  const safeDivisions = Math.max(3, Math.round(divisions))
  const padding = Math.max(12, Math.round(size * 0.06))
  const labelGap = showLabels ? Math.max(10, Math.round(size * 0.07)) : 0
  // Keep left/right labels closer than top/bottom so long strings do not widen the chart too much.
  const labelGapHorizontal = showLabels ? Math.max(4, Math.round(labelGap * 0.6)) : 0
  const radius = Math.max(0, baseCenter - padding - labelGap)

  const angles = useMemo(() => {
    return axes.map((_, index) => (Math.PI * 2 * index) / axes.length)
  }, [])

  const baseLabelData = useMemo(() => {
    if (!showLabels) return []
    return axes.map((axis, index) => {
      const angle = angles[index]
      const angleDeg = (angle * 180) / Math.PI
      // Place labels outside the chart radius in the base (size x size) coordinate system.
      // We will later expand the canvas to fit these labels without shrinking the chart.
      // Left/right labels use a tighter gap to reduce horizontal span.
      const axisGap = axis === 'diversification' || axis === 'drawdown' ? labelGapHorizontal : labelGap
      const labelRadius = radius + axisGap
      const label = labels?.[axis] ?? toTitle(axis)
      const baseline = getDefaultBaseline(180 + angleDeg)
      return {
        axis,
        label,
        x: baseCenter + Math.sin(angle) * labelRadius,
        y: baseCenter - Math.cos(angle) * labelRadius,
        textAnchor: getDefaultTextAnchor(180 + angleDeg),
        alignmentBaseline: baseline === 'auto' ? undefined : baseline,
        rawBaseline: baseline,
      }
    })
  }, [angles, baseCenter, labelGap, labelGapHorizontal, labels, radius, showLabels])

  // Expand the canvas so outside labels stay visible without shrinking the chart.
  // We compute label bounds in base coordinates, then grow the SVG equally on both sides.
  const labelBounds = useMemo(() => {
    if (!showLabels || baseLabelData.length === 0) {
      return { minX: 0, maxX: size, minY: 0, maxY: size }
    }
    let minX = size
    let maxX = 0
    let minY = size
    let maxY = 0

    baseLabelData.forEach((label) => {
      const width = estimateLabelWidth(label.label, labelFontSize)
      const height = labelFontSize

      // Convert the anchor/baseline into a bounding box in SVG coordinates.
      // This matches how text is positioned by react-native-svg for each label.
      let labelMinX = label.x
      let labelMaxX = label.x
      if (label.textAnchor === 'start') {
        labelMaxX = label.x + width
      } else if (label.textAnchor === 'end') {
        labelMinX = label.x - width
      } else {
        labelMinX = label.x - width / 2
        labelMaxX = label.x + width / 2
      }

      let labelMinY = label.y
      let labelMaxY = label.y
      if (label.rawBaseline === 'hanging') {
        labelMaxY = label.y + height
      } else if (label.rawBaseline === 'central') {
        labelMinY = label.y - height / 2
        labelMaxY = label.y + height / 2
      } else {
        labelMinY = label.y - height
      }

      minX = Math.min(minX, labelMinX)
      maxX = Math.max(maxX, labelMaxX)
      minY = Math.min(minY, labelMinY)
      maxY = Math.max(maxY, labelMaxY)
    })

    return { minX, maxX, minY, maxY }
  }, [baseLabelData, labelFontSize, showLabels, size])

  // Allow a small vertical padding, but let horizontal labels touch edges if needed.
  // Horizontal padding is zero so long right/left labels can sit closer to the container edge.
  const layoutPaddingX = 0
  const layoutPaddingY = Math.max(2, Math.round(size * 0.02))
  const overflowLeft = showLabels ? Math.max(0, layoutPaddingX - labelBounds.minX) : 0
  const overflowRight = showLabels ? Math.max(0, labelBounds.maxX - (size - layoutPaddingX)) : 0
  const overflowTop = showLabels ? Math.max(0, layoutPaddingY - labelBounds.minY) : 0
  const overflowBottom = showLabels ? Math.max(0, labelBounds.maxY - (size - layoutPaddingY)) : 0
  // Expand symmetrically so the radar stays centered relative to its container.
  // We choose the larger overflow on each axis and apply it to both sides.
  const extraX = Math.max(overflowLeft, overflowRight)
  const extraY = Math.max(overflowTop, overflowBottom)
  const extraLeft = showLabels ? extraX : 0
  const extraRight = showLabels ? extraX : 0
  const extraTop = showLabels ? extraY : 0
  const extraBottom = showLabels ? extraY : 0
  // Keep expansion symmetric so the radar stays centered in its container.
  // The chart radius stays constant; only the canvas and center shift.
  const outerWidth = size + extraLeft + extraRight
  const outerHeight = size + extraTop + extraBottom
  const centerX = baseCenter + extraLeft
  const centerY = baseCenter + extraTop

  const corners = useMemo(() => {
    return angles.map((angle) => ({
      x: centerX + Math.sin(angle) * radius,
      y: centerY - Math.cos(angle) * radius,
      angle,
    }))
  }, [angles, centerX, centerY, radius])

  const valuePoints = useMemo(() => {
    return axes.map((axis, index) => {
      const value = clamp(values[axis] ?? 0)
      const scaled = (value / 100) * radius
      const angle = angles[index]
      return {
        x: centerX + Math.sin(angle) * scaled,
        y: centerY - Math.cos(angle) * scaled,
      }
    })
  }, [angles, centerX, centerY, radius, values])

  const polygonPoints = useMemo(() => {
    return valuePoints.map((point) => `${point.x},${point.y}`).join(' ')
  }, [valuePoints])

  const wedgeFills = useMemo(() => {
    if (valuePoints.length === 0) return []
    return valuePoints.map((point, index) => {
      const next = valuePoints[(index + 1) % valuePoints.length]
      return {
        id: `${gradientBaseId}-wedge-${index}`,
        d: `M ${centerX} ${centerY} L ${point.x} ${point.y} L ${next.x} ${next.y} Z`,
        start: point,
        end: next,
        startColor: axisColors[index % axisColors.length],
        endColor: axisColors[(index + 1) % axisColors.length],
      }
    })
  }, [axisColors, centerX, centerY, gradientBaseId, valuePoints])

  const edgeGradients = useMemo(() => {
    if (valuePoints.length === 0) return []
    return valuePoints.map((point, index) => {
      const next = valuePoints[(index + 1) % valuePoints.length]
      return {
        id: `${gradientBaseId}-edge-${index}`,
        start: point,
        end: next,
        startColor: axisColors[index % axisColors.length],
        endColor: axisColors[(index + 1) % axisColors.length],
      }
    })
  }, [axisColors, gradientBaseId, valuePoints])

  const ringRatios = useMemo(() => {
    return Array.from({ length: safeDivisions }, (_, index) => (index + 1) / safeDivisions)
  }, [safeDivisions])

  const ringPolygons = useMemo(() => {
    return ringRatios.map((ratio) =>
      corners.map((corner) => `${lerp(centerX, corner.x, ratio)},${lerp(centerY, corner.y, ratio)}`).join(' ')
    )
  }, [centerX, centerY, corners, ringRatios])

  const stripePaths = useMemo(() => {
    return ringRatios.map((ratio, index) => {
      const innerRatio = ringRatios[index - 1] ?? 0
      const outer = [...corners, corners[0]].map(
        (corner) => `${lerp(centerX, corner.x, ratio)} ${lerp(centerY, corner.y, ratio)}`
      )
      const inner = [...corners, corners[0]]
        .reverse()
        .map((corner) => `${lerp(centerX, corner.x, innerRatio)} ${lerp(centerY, corner.y, innerRatio)}`)
      return {
        ratio,
        d: `M ${outer.join(' L ')} L ${inner.join(' L ')} Z`,
      }
    })
  }, [centerX, centerY, corners, ringRatios])

  const labelData = useMemo(() => {
    if (!showLabels) return []
    // Shift labels into the expanded canvas so their absolute positions are preserved.
    // This keeps their polar placement while matching the new center offset.
    return baseLabelData.map((label) => ({
      ...label,
      x: label.x + extraLeft,
      y: label.y + extraTop,
    }))
  }, [baseLabelData, extraLeft, extraTop, showLabels])

  const gridStroke = theme.colors.border
  const gridOpacity = 0.85
  const axisOpacity = 0.32
  const stripeOpacity = 0.055
  const stripeAltOpacity = 0.032
  const wedgeOpacity = 0.22
  const markerRadius = Math.max(3, Math.round(size * 0.015))
  const markerStroke = Math.max(1.2, markerRadius * 0.6)
  const centerColor = theme.colors.surfaceElevated

  return (
    <View style={{ width: outerWidth, height: outerHeight }}>
      <Svg width={outerWidth} height={outerHeight} viewBox={`0 0 ${outerWidth} ${outerHeight}`}>
        <Defs>
          {wedgeFills.map((wedge) => (
            <LinearGradient
              key={wedge.id}
              id={wedge.id}
              x1={wedge.start.x}
              y1={wedge.start.y}
              x2={wedge.end.x}
              y2={wedge.end.y}
              gradientUnits="userSpaceOnUse"
            >
              <Stop offset="0%" stopColor={wedge.startColor} stopOpacity={wedgeOpacity} />
              <Stop offset="100%" stopColor={wedge.endColor} stopOpacity={wedgeOpacity} />
            </LinearGradient>
          ))}
          <RadialGradient
            id={centerGradientId}
            cx={centerX}
            cy={centerY}
            r={radius * 0.92}
            gradientUnits="userSpaceOnUse"
          >
            <Stop offset="0%" stopColor={centerColor} stopOpacity={0.55} />
            <Stop offset="65%" stopColor={centerColor} stopOpacity={0.22} />
            <Stop offset="100%" stopColor={centerColor} stopOpacity={0} />
          </RadialGradient>
          {edgeGradients.map((edge) => (
            <LinearGradient
              key={edge.id}
              id={edge.id}
              x1={edge.start.x}
              y1={edge.start.y}
              x2={edge.end.x}
              y2={edge.end.y}
              gradientUnits="userSpaceOnUse"
            >
              <Stop offset="0%" stopColor={edge.startColor} stopOpacity={0.96} />
              <Stop offset="100%" stopColor={edge.endColor} stopOpacity={0.96} />
            </LinearGradient>
          ))}
        </Defs>

        <G>
          {stripePaths.map((stripe, index) => (
            <Path
              key={`stripe-${stripe.ratio}`}
              d={stripe.d}
              fill={axisColors[index % axisColors.length]}
              fillOpacity={index % 2 === 0 ? stripeOpacity : stripeAltOpacity}
            />
          ))}

          {ringPolygons.map((ring, index) => (
            <Polygon
              key={`ring-${index}`}
              points={ring}
              fill="none"
              stroke={gridStroke}
              strokeOpacity={gridOpacity}
              strokeWidth={1}
            />
          ))}

          {corners.map((corner, index) => (
            <Line
              key={`axis-${index}`}
              x1={centerX}
              y1={centerY}
              x2={corner.x}
              y2={corner.y}
              stroke={axisColors[index % axisColors.length]}
              strokeOpacity={axisOpacity}
              strokeWidth={1}
            />
          ))}

          {wedgeFills.map((wedge) => (
            <Path key={`wedge-${wedge.id}`} d={wedge.d} fill={`url(#${wedge.id})`} />
          ))}

          <Polygon points={polygonPoints} fill={`url(#${centerGradientId})`} />

          {edgeGradients.map((edge) => (
            <Line
              key={`edge-${edge.id}`}
              x1={edge.start.x}
              y1={edge.start.y}
              x2={edge.end.x}
              y2={edge.end.y}
              stroke={`url(#${edge.id})`}
              strokeWidth={2.2}
              strokeLinecap="round"
            />
          ))}

          {valuePoints.map((point, index) => (
            <Circle
              key={`mark-${index}`}
              cx={point.x}
              cy={point.y}
              r={markerRadius}
              fill={axisColors[index % axisColors.length]}
              stroke={theme.colors.surface}
              strokeWidth={markerStroke}
            />
          ))}
        </G>

        {showLabels
          ? labelData.map((label, index) => (
              <SvgText
                key={`label-${label.axis}`}
                x={label.x}
                y={label.y}
                fill={axisColors[index % axisColors.length]}
                fillOpacity={0.82}
                fontSize={labelFontSize}
                fontFamily={theme.fonts.bodyMedium}
                textAnchor={label.textAnchor}
                alignmentBaseline={label.alignmentBaseline}
              >
                {label.label}
              </SvgText>
            ))
          : null}
      </Svg>
    </View>
  )
}
