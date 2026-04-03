import React from 'react'
import { StyleSheet, Text, View } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

function scoreColor(score?: number, theme?: ReturnType<typeof useTheme>) {
  if (!theme) return '#fff'
  if (score === undefined || score === null) return theme.colors.muted
  if (score < 50) return theme.colors.danger
  if (score < 70) return theme.colors.warning
  if (score < 90) return theme.colors.success
  return theme.colors.accent
}

export function ScoreRing({ score, label, locked }: { score?: number | null; label?: string; locked?: boolean }) {
  const theme = useTheme()
  const ringColor = locked ? theme.colors.border : scoreColor(score ?? undefined, theme)

  return (
    <View style={[styles.container, { borderColor: ringColor, backgroundColor: theme.colors.surfaceElevated }]}>
      <Text
        style={[
          styles.score,
          { color: locked ? theme.colors.muted : theme.colors.ink, fontFamily: theme.fonts.display },
        ]}
      >
        {locked ? '—' : (score?.toFixed(0) ?? '--')}
      </Text>
      <Text style={[styles.label, { color: theme.colors.muted, fontFamily: theme.fonts.bodyMedium }]}>{label}</Text>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    width: 104,
    height: 104,
    borderRadius: 999,
    borderWidth: 4,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 4,
  },
  score: {
    fontSize: 22,
    letterSpacing: 0.4,
  },
  label: {
    fontSize: 11,
    textTransform: 'uppercase',
    letterSpacing: 1.4,
  },
})
