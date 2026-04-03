import React from 'react'
import { StyleProp, StyleSheet, View, ViewStyle } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function Card({ children, style }: { children: React.ReactNode; style?: StyleProp<ViewStyle> }) {
  const theme = useTheme()
  return <View style={[styles(theme).card, style]}>{children}</View>
}

const styles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    card: {
      backgroundColor: theme.colors.surface,
      borderRadius: theme.radius.lg,
      padding: theme.spacing.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      shadowColor: '#000',
      shadowOpacity: 0.04,
      shadowRadius: 12,
      shadowOffset: { width: 0, height: 6 },
      elevation: 1,
    },
  })
