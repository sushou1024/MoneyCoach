import React from 'react'
import { StyleSheet, TextInput, TextInputProps } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function Input(props: TextInputProps) {
  const theme = useTheme()
  return <TextInput placeholderTextColor={theme.colors.muted} {...props} style={[styles(theme).input, props.style]} />
}

const styles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    input: {
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: theme.radius.md,
      paddingHorizontal: 14,
      paddingVertical: 12,
      fontSize: 15,
      color: theme.colors.ink,
      backgroundColor: theme.colors.surfaceElevated,
      fontFamily: theme.fonts.body,
    },
  })
