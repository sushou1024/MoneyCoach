import { useRouter } from 'expo-router'
import React from 'react'
import { Text } from 'react-native'

import { Button } from '../src/components/Button'
import { Screen } from '../src/components/Screen'
import { useTheme } from '../src/providers/ThemeProvider'

export default function NotFound() {
  const theme = useTheme()
  const router = useRouter()
  return (
    <Screen>
      <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>Screen not found</Text>
      <Button title="Go Home" style={{ marginTop: theme.spacing.lg }} onPress={() => router.replace('/')} />
    </Screen>
  )
}
