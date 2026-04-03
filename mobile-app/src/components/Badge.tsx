import React from 'react'
import { StyleSheet, Text, View } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'
import { BadgeTone } from '../utils/badges'

function hexToRgba(hex: string, alpha: number) {
  const normalized = hex.replace('#', '')
  if (normalized.length !== 6) {
    return `rgba(0,0,0,${alpha})`
  }
  const r = parseInt(normalized.slice(0, 2), 16)
  const g = parseInt(normalized.slice(2, 4), 16)
  const b = parseInt(normalized.slice(4, 6), 16)
  return `rgba(${r},${g},${b},${alpha})`
}

function toneColor(theme: ReturnType<typeof useTheme>, tone: BadgeTone) {
  switch (tone) {
    case 'critical':
      return theme.colors.danger
    case 'high':
      return theme.colors.warning
    case 'medium':
      return theme.colors.accent
    case 'low':
      return theme.colors.success
    default:
      return theme.colors.muted
  }
}

export function Badge({ label, tone = 'neutral' }: { label: string; tone?: BadgeTone }) {
  const theme = useTheme()
  const color = toneColor(theme, tone)
  return (
    <View style={[styles.container, { backgroundColor: hexToRgba(color, 0.12), borderColor: hexToRgba(color, 0.22) }]}>
      <Text style={[styles.text, { color, fontFamily: theme.fonts.bodyBold }]}>{label}</Text>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    alignSelf: 'flex-start',
    paddingHorizontal: 10,
    paddingVertical: 5,
    borderRadius: 999,
    borderWidth: 1,
  },
  text: {
    fontSize: 12,
    letterSpacing: 0.2,
  },
})
