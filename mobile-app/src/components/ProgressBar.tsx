import React from 'react'
import { StyleSheet, View } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function ProgressBar({ progress }: { progress: number }) {
  const theme = useTheme()
  const clamped = Math.max(0, Math.min(1, progress))
  return (
    <View style={[styles.track, { backgroundColor: theme.colors.surfaceElevated, borderColor: theme.colors.border }]}>
      <View style={[styles.fill, { backgroundColor: theme.colors.accent, width: `${clamped * 100}%` }]} />
    </View>
  )
}

const styles = StyleSheet.create({
  track: {
    height: 6,
    borderRadius: 999,
    overflow: 'hidden',
    borderWidth: 1,
  },
  fill: {
    height: '100%',
    borderRadius: 999,
  },
})
