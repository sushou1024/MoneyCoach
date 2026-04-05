import React from 'react'
import { RefreshControl, ScrollView, StyleProp, StyleSheet, View, ViewStyle, useWindowDimensions } from 'react-native'
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context'

import { useTheme } from '../providers/ThemeProvider'

interface ScreenProps {
  children: React.ReactNode
  style?: StyleProp<ViewStyle>
  scroll?: boolean
  decorativeBackground?: boolean
  refreshing?: boolean
  onRefresh?: () => void
}

export function Screen({ children, style, scroll, decorativeBackground = true, refreshing, onRefresh }: ScreenProps) {
  const theme = useTheme()
  const insets = useSafeAreaInsets()
  const { width } = useWindowDimensions()
  const bottomOffset = Math.max(insets.bottom, theme.spacing.md)
  const haloSize = Math.min(420, Math.max(240, Math.round(width * 0.9)))
  const haloRadius = Math.round(haloSize / 2)
  const haloTop = -Math.round(haloSize * 0.45)
  const haloSide = -Math.round(haloSize * 0.4)
  const haloBottom = -Math.round(haloSize * 0.55)
  const content = scroll ? (
    <ScrollView
      contentContainerStyle={[
        {
          paddingHorizontal: theme.spacing.lg,
          paddingTop: theme.spacing.lg,
          paddingBottom: theme.spacing.xxl + bottomOffset,
          flexGrow: 1,
        },
        style,
      ]}
      showsVerticalScrollIndicator={false}
      refreshControl={
        onRefresh ? (
          <RefreshControl refreshing={refreshing ?? false} onRefresh={onRefresh} tintColor={theme.colors.muted} />
        ) : undefined
      }
    >
      {children}
    </ScrollView>
  ) : (
    <View
      style={[
        {
          paddingHorizontal: theme.spacing.lg,
          paddingTop: theme.spacing.lg,
          paddingBottom: theme.spacing.lg + bottomOffset,
          flex: 1,
        },
        style,
      ]}
    >
      {children}
    </View>
  )

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: theme.colors.background }} edges={['top', 'left', 'right']}>
      {decorativeBackground ? (
        <View pointerEvents="none" style={StyleSheet.absoluteFill}>
          <View
            style={[
              styles.halo,
              {
                width: haloSize,
                height: haloSize,
                borderRadius: haloRadius,
                top: haloTop,
                right: haloSide,
                backgroundColor: theme.colors.accentSoft,
              },
            ]}
          />
          <View
            style={[
              styles.halo,
              {
                width: haloSize,
                height: haloSize,
                borderRadius: haloRadius,
                bottom: haloBottom,
                left: haloSide,
                backgroundColor: theme.colors.glow,
              },
            ]}
          />
        </View>
      ) : null}
      {content}
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  halo: {
    position: 'absolute',
    opacity: 0.3,
  },
})
