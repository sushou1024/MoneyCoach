import React from 'react'
import { StyleSheet, Text, View } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function Tag({ label }: { label: string }) {
  const theme = useTheme()
  return (
    <View style={styles(theme).tag}>
      <Text style={styles(theme).text}>{label}</Text>
    </View>
  )
}

const styles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    tag: {
      backgroundColor: theme.colors.surfaceElevated,
      borderRadius: theme.radius.sm,
      paddingHorizontal: 10,
      paddingVertical: 5,
      borderWidth: 1,
      borderColor: theme.colors.border,
    },
    text: {
      fontSize: 12,
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyBold,
      letterSpacing: 0.2,
    },
  })
