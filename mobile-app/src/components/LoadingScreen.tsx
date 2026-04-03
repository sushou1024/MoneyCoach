import React from 'react'
import { ActivityIndicator, Text, View } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function LoadingScreen({ label = 'Loading...' }: { label?: string }) {
  const theme = useTheme()
  return (
    <View style={{ flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: theme.colors.background }}>
      <ActivityIndicator size="large" color={theme.colors.accent} />
      <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {label}
      </Text>
    </View>
  )
}
