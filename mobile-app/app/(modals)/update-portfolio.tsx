import { Ionicons } from '@expo/vector-icons'
import { useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import React, { useState } from 'react'
import { ActivityIndicator, Alert, Pressable, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Input } from '../../src/components/Input'
import { Screen } from '../../src/components/Screen'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { apiRequest } from '../../src/services/api'

export default function UpdatePortfolioModal() {
  const theme = useTheme()
  const router = useRouter()
  const queryClient = useQueryClient()
  const { t } = useLocalization()
  const { accessToken, userId } = useAuth()
  const [command, setCommand] = useState('')
  const [commandSubmitting, setCommandSubmitting] = useState(false)

  const tradeSlipRoute = { pathname: '/(modals)/trade-slip', params: { return_to: '/(tabs)/assets' } }

  const handleCommand = async () => {
    if (!command.trim() || !accessToken || commandSubmitting) return
    setCommandSubmitting(true)
    try {
      const resp = await apiRequest<{ status: string; toast?: string }>('/v1/assets/commands', {
        method: 'POST',
        token: accessToken,
        body: { text: command.trim() },
      })
      if (resp.error) {
        Alert.alert(t('assets.commandFailTitle'), resp.error.message)
        return
      }
      if (resp.data?.status === 'ignored') {
        Alert.alert(t('assets.commandIgnoredTitle'), resp.data?.toast ?? t('assets.commandIgnoredBody'))
        return
      }
      Alert.alert(t('assets.commandSuccessTitle'), resp.data?.toast ?? t('assets.commandSuccessBody'))
      setCommand('')
      await queryClient.invalidateQueries({ queryKey: ['portfolio', 'active', userId] })
      await queryClient.invalidateQueries({ queryKey: ['reports', 'active', userId] })
    } finally {
      setCommandSubmitting(false)
    }
  }

  return (
    <Screen scroll>
      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
        <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
          {t('strategy.updateCta')}
        </Text>
        <Button title={t('common.close')} variant="ghost" onPress={() => router.back()} />
      </View>

      <Card style={{ marginTop: theme.spacing.lg }}>
        <Pressable
          onPress={() => router.replace('/(modals)/upload')}
          accessibilityRole="button"
          style={({ pressed }) => [
            { flexDirection: 'row', alignItems: 'center', gap: theme.spacing.md },
            pressed && { opacity: 0.9 },
          ]}
        >
          <View
            style={{
              width: 40,
              height: 40,
              borderRadius: 999,
              backgroundColor: theme.colors.accentSoft,
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Ionicons name="camera-outline" size={18} color={theme.colors.accent} />
          </View>
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: 16, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
              {t('upload.title')}
            </Text>
            <Text style={{ color: theme.colors.muted, marginTop: 2, fontFamily: theme.fonts.body }}>
              {t('upload.subtitle')}
            </Text>
          </View>
          <Ionicons name="chevron-forward" size={18} color={theme.colors.muted} />
        </Pressable>
      </Card>

      <Card style={{ marginTop: theme.spacing.md }}>
        <Pressable
          onPress={() => router.replace(tradeSlipRoute)}
          accessibilityRole="button"
          style={({ pressed }) => [
            { flexDirection: 'row', alignItems: 'center', gap: theme.spacing.md },
            pressed && { opacity: 0.9 },
          ]}
        >
          <View
            style={{
              width: 40,
              height: 40,
              borderRadius: 999,
              backgroundColor: theme.colors.glow,
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Ionicons name="document-text-outline" size={18} color={theme.colors.warning} />
          </View>
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: 16, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
              {t('tradeSlip.uploadCta')}
            </Text>
            <Text style={{ color: theme.colors.muted, marginTop: 2, fontFamily: theme.fonts.body }}>
              {t('tradeSlip.subtitle')}
            </Text>
          </View>
          <Ionicons name="chevron-forward" size={18} color={theme.colors.muted} />
        </Pressable>
      </Card>

      <Card style={{ marginTop: theme.spacing.lg }}>
        <Text style={{ fontSize: 16, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
          {t('assets.commandTitle')}
        </Text>
        <View
          style={{
            marginTop: theme.spacing.sm,
            flexDirection: 'row',
            alignItems: 'center',
            borderRadius: theme.radius.lg,
            borderWidth: 1,
            borderColor: theme.colors.border,
            backgroundColor: theme.colors.surfaceElevated,
            paddingHorizontal: theme.spacing.sm,
          }}
        >
          <Input
            value={command}
            onChangeText={setCommand}
            placeholder={t('assets.commandPlaceholder')}
            editable={!commandSubmitting}
            returnKeyType="send"
            onSubmitEditing={handleCommand}
            style={{ flex: 1, borderWidth: 0, backgroundColor: 'transparent' }}
          />
          <Pressable
            onPress={handleCommand}
            disabled={commandSubmitting}
            style={{ padding: theme.spacing.sm, opacity: commandSubmitting ? 0.6 : 1 }}
          >
            {commandSubmitting ? (
              <ActivityIndicator size="small" color={theme.colors.accent} />
            ) : (
              <Ionicons name="arrow-up-circle" size={22} color={theme.colors.accent} />
            )}
          </Pressable>
        </View>
        <View style={{ marginTop: theme.spacing.xs }}>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body, fontSize: 12 }}>
            {t('assets.commandHintTitle')}
          </Text>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body, fontSize: 12 }}>
            {`- ${t('assets.commandExample1')}`}
          </Text>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body, fontSize: 12 }}>
            {`- ${t('assets.commandExample2')}`}
          </Text>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body, fontSize: 12 }}>
            {`- ${t('assets.commandExample3')}`}
          </Text>
        </View>
      </Card>
    </Screen>
  )
}
