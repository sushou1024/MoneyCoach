import { useRouter } from 'expo-router'
import React from 'react'
import { Alert, Text } from 'react-native'

import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'

export default function VaultsScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()

  const handleJoin = () => {
    Alert.alert(t('vaults.joinTitle'), t('vaults.joinBody'))
  }

  return (
    <Screen>
      <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
        {t('vaults.title')}
      </Text>
      <Card style={{ marginTop: theme.spacing.md }}>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('vaults.body')}</Text>
      </Card>
      <Button title={t('vaults.cta')} style={{ marginTop: theme.spacing.lg }} onPress={handleJoin} />
      <Button
        title={t('common.close')}
        variant="ghost"
        style={{ marginTop: theme.spacing.sm }}
        onPress={() => router.back()}
      />
    </Screen>
  )
}
