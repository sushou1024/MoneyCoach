import { useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import React, { useEffect, useState } from 'react'
import { Alert, Switch, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Screen } from '../../src/components/Screen'
import { useEntitlement } from '../../src/hooks/useEntitlement'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { devActivateEntitlement, devClearEntitlement } from '../../src/services/billing'

export default function DeveloperToolsScreen() {
  const theme = useTheme()
  const router = useRouter()
  const queryClient = useQueryClient()
  const { t } = useLocalization()
  const { accessToken, clearSession } = useAuth()
  const entitlement = useEntitlement()
  const [enabled, setEnabled] = useState(false)

  useEffect(() => {
    if (!entitlement.data) return
    setEnabled(entitlement.data.provider === 'dev')
  }, [entitlement.data])

  const toggleDev = async (value: boolean) => {
    if (!accessToken) return
    setEnabled(value)
    const resp = value ? await devActivateEntitlement(accessToken) : await devClearEntitlement(accessToken)
    if (resp.error) {
      Alert.alert(t('developer.failTitle'), resp.error.message)
      setEnabled(!value)
      return
    }
    entitlement.refetch()
  }

  const handleReset = async () => {
    queryClient.clear()
    await clearSession()
    router.replace('/(auth)/sc07')
  }

  if (!__DEV__) {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('developer.hidden')}</Text>
        <Button
          title={t('common.close')}
          variant="ghost"
          style={{ marginTop: theme.spacing.md }}
          onPress={() => router.back()}
        />
      </Screen>
    )
  }

  return (
    <Screen>
      <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
        {t('developer.title')}
      </Text>
      <View style={{ flexDirection: 'row', alignItems: 'center', marginTop: theme.spacing.lg }}>
        <Text style={{ flex: 1, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {t('developer.toggle')}
        </Text>
        <Switch value={enabled} onValueChange={toggleDev} />
      </View>
      <Button
        title={t('developer.reset')}
        variant="secondary"
        style={{ marginTop: theme.spacing.lg }}
        onPress={handleReset}
      />
      <Button
        title={t('common.close')}
        variant="ghost"
        style={{ marginTop: theme.spacing.sm }}
        onPress={() => router.back()}
      />
    </Screen>
  )
}
