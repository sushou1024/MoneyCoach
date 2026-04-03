import React, { useEffect, useRef } from 'react'
import { Animated, Easing, Pressable, StyleSheet, Text, View } from 'react-native'
import Svg, { Path } from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'
import { useResponsiveScale } from '../utils/responsive'

type SelectionMode = 'single' | 'multiple'

export function OnboardingChoiceRow({
  label,
  description,
  selected,
  onPress,
  showDivider = true,
  selectionMode = 'single',
}: {
  label: string
  description?: string
  selected?: boolean
  onPress?: () => void
  showDivider?: boolean
  selectionMode?: SelectionMode
}) {
  const theme = useTheme()
  const { font, scale } = useResponsiveScale()
  const labelSize = font(17, 15)
  const descriptionSize = font(13, 12)
  const markerSize = Math.max(16, Math.round(18 * scale))
  const progress = useRef(new Animated.Value(selected ? 1 : 0)).current
  const styles = makeStyles(theme, {
    labelSize,
    descriptionSize,
    markerSize,
    showDivider,
    selected: Boolean(selected),
    selectionMode,
  })

  useEffect(() => {
    Animated.timing(progress, {
      toValue: selected ? 1 : 0,
      duration: 220,
      easing: Easing.out(Easing.cubic),
      useNativeDriver: false,
    }).start()
  }, [progress, selected])

  const checkStrokeOffset = progress.interpolate({
    inputRange: [0, 1],
    outputRange: [32, 0],
  })
  const checkOpacity = progress.interpolate({
    inputRange: [0, 0.6, 1],
    outputRange: [0, 0, 1],
  })
  const indicatorScale = progress.interpolate({
    inputRange: [0, 1],
    outputRange: [0.96, 1],
  })

  return (
    <Pressable
      onPress={onPress}
      style={({ pressed }) => [styles.container, pressed && styles.pressed]}
      accessibilityRole={selectionMode === 'single' ? 'radio' : 'checkbox'}
      accessibilityState={{ selected: Boolean(selected) }}
    >
      <Animated.View
        style={[styles.indicator, selected && styles.indicatorSelected, { transform: [{ scale: indicatorScale }] }]}
      >
        <Svg width={markerSize} height={markerSize} viewBox="0 0 20 20">
          <AnimatedPath
            d="M4.2 10.6l3.2 3.4L15.8 6.6"
            stroke={theme.colors.surface}
            strokeWidth={2}
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeDasharray={32}
            strokeDashoffset={checkStrokeOffset}
            opacity={checkOpacity}
          />
        </Svg>
      </Animated.View>
      <View style={styles.textBlock}>
        <Text style={styles.label}>{label}</Text>
        {description ? <Text style={styles.description}>{description}</Text> : null}
      </View>
    </Pressable>
  )
}

const AnimatedPath = Animated.createAnimatedComponent(Path)

const makeStyles = (
  theme: ReturnType<typeof useTheme>,
  options: {
    labelSize: number
    descriptionSize: number
    markerSize: number
    showDivider: boolean
    selected: boolean
    selectionMode: SelectionMode
  }
) =>
  StyleSheet.create({
    container: {
      position: 'relative',
      flexDirection: 'row',
      alignItems: 'center',
      paddingVertical: theme.spacing.md,
      paddingHorizontal: theme.spacing.md,
      backgroundColor: options.selected ? theme.colors.surfaceElevated : theme.colors.surface,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: options.selected ? theme.colors.accent : theme.colors.border,
      shadowColor: '#0b0c10',
      shadowOpacity: options.selected ? 0.08 : 0.05,
      shadowRadius: options.selected ? 12 : 10,
      shadowOffset: { width: 0, height: options.selected ? 8 : 6 },
      elevation: options.selected ? 2 : 1,
    },
    pressed: {
      transform: [{ scale: 0.992 }],
      opacity: 0.96,
    },
    indicator: {
      width: options.markerSize,
      height: options.markerSize,
      borderRadius:
        options.selectionMode === 'multiple' ? Math.round(options.markerSize * 0.2) : options.markerSize / 2,
      borderWidth: 1.5,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surface,
      alignItems: 'center',
      justifyContent: 'center',
    },
    indicatorSelected: {
      borderColor: theme.colors.accent,
      backgroundColor: theme.colors.accent,
    },
    textBlock: {
      flex: 1,
      paddingLeft: theme.spacing.md,
      paddingRight: theme.spacing.xs,
    },
    label: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: options.labelSize,
      letterSpacing: 0.1,
    },
    description: {
      color: theme.colors.muted,
      marginTop: theme.spacing.xs,
      fontSize: options.descriptionSize,
      lineHeight: Math.round(options.descriptionSize * 1.4),
      fontFamily: theme.fonts.body,
    },
  })
