import React from 'react'
import { Pressable, StyleSheet, Text, View } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function ChoiceChip({
  label,
  description,
  selected,
  onPress,
}: {
  label: string
  description?: string
  selected?: boolean
  onPress?: () => void
}) {
  const theme = useTheme()
  const styles = makeStyles(theme, selected)

  return (
    <Pressable style={({ pressed }) => [styles.container, pressed && styles.pressed]} onPress={onPress}>
      <View>
        <Text style={styles.label}>{label}</Text>
        {description ? <Text style={styles.description}>{description}</Text> : null}
      </View>
    </Pressable>
  )
}

const makeStyles = (theme: ReturnType<typeof useTheme>, selected?: boolean) =>
  StyleSheet.create({
    container: {
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: selected ? theme.colors.accent : theme.colors.border,
      backgroundColor: selected ? theme.colors.accentSoft : theme.colors.surfaceElevated,
      paddingVertical: theme.spacing.md,
      paddingHorizontal: theme.spacing.md,
    },
    pressed: {
      transform: [{ scale: 0.99 }],
      opacity: 0.95,
    },
    label: {
      color: theme.colors.ink,
      fontSize: 16,
      fontFamily: theme.fonts.bodyBold,
      letterSpacing: 0.2,
    },
    description: {
      color: theme.colors.muted,
      marginTop: theme.spacing.xs,
      fontSize: 13,
      fontFamily: theme.fonts.body,
    },
  })
