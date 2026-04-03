import React from 'react'
import { StyleProp, StyleSheet, View, ViewStyle } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function OnboardingDecisionSheet({
  children,
  style,
}: {
  children: React.ReactNode
  style?: StyleProp<ViewStyle>
}) {
  const theme = useTheme()
  return <View style={[styles(theme).container, style]}>{children}</View>
}

const styles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    container: {
      gap: theme.spacing.sm,
    },
  })
