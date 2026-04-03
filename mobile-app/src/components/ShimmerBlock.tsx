import React, { useEffect, useRef } from 'react'
import { Animated, Easing, StyleProp, View, ViewStyle, useWindowDimensions, type DimensionValue } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

type ShimmerBlockProps = {
  width?: DimensionValue
  height: number
  radius?: number
  style?: StyleProp<ViewStyle>
}

export function ShimmerBlock({ width = '100%', height, radius, style }: ShimmerBlockProps) {
  const theme = useTheme()
  const { width: windowWidth } = useWindowDimensions()
  const safeWidth = Math.max(windowWidth, 1)
  const shimmer = useRef(new Animated.Value(0)).current

  useEffect(() => {
    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(shimmer, {
          toValue: 1,
          duration: 1100,
          easing: Easing.inOut(Easing.quad),
          useNativeDriver: true,
        }),
        Animated.timing(shimmer, {
          toValue: 0,
          duration: 1100,
          easing: Easing.inOut(Easing.quad),
          useNativeDriver: true,
        }),
      ])
    )
    animation.start()
    return () => animation.stop()
  }, [shimmer])

  const shimmerWidth = Math.max(120, Math.round(safeWidth * 0.6))
  const translateX = shimmer.interpolate({
    inputRange: [0, 1],
    outputRange: [-shimmerWidth, safeWidth],
  })
  const shimmerOpacity = shimmer.interpolate({
    inputRange: [0, 1],
    outputRange: [0.12, 0.32],
  })

  return (
    <View
      style={[
        {
          width,
          height,
          borderRadius: radius ?? theme.radius.md,
          backgroundColor: theme.colors.surfaceElevated,
          overflow: 'hidden',
        },
        style,
      ]}
    >
      <Animated.View
        style={{
          position: 'absolute',
          top: 0,
          bottom: 0,
          width: shimmerWidth,
          backgroundColor: theme.colors.glow,
          opacity: shimmerOpacity,
          transform: [{ translateX }],
        }}
      />
    </View>
  )
}
