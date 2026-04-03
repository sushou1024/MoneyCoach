import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useEffect, useMemo, useState } from 'react'
import { Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { ShimmerBlock } from '../../src/components/ShimmerBlock'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchUploadBatch } from '../../src/services/uploads'

export default function ProcessingOCR() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken } = useAuth()
  const params = useLocalSearchParams<{ id?: string }>()
  const batchId = params.id ?? ''
  const [message, setMessage] = useState(t('processing.ocr.title'))
  const [failed, setFailed] = useState(false)
  const [stepIndex, setStepIndex] = useState(0)

  const steps = useMemo(() => [t('processing.ocr.step1'), t('processing.ocr.step2'), t('processing.ocr.step3')], [t])

  useEffect(() => {
    const interval = setInterval(() => {
      setStepIndex((prev) => (prev + 1) % steps.length)
    }, 2000)
    return () => clearInterval(interval)
  }, [steps.length])

  useEffect(() => {
    let active = true
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    const poll = async () => {
      if (!active || !accessToken || !batchId) return
      const resp = await fetchUploadBatch(accessToken, batchId)
      if (!active) return
      if (resp.error) {
        setMessage(`${t('processing.ocr.fail')}: ${resp.error.message}`)
        setFailed(true)
        return
      }
      if (resp.data && 'status' in resp.data && resp.data.status === 'needs_review') {
        router.replace({ pathname: '/(modals)/review', params: { id: batchId } })
        return
      }
      if (resp.data && 'status' in resp.data && resp.data.status === 'failed') {
        setMessage(`${t('processing.ocr.fail')}: ${resp.data.error_code ?? t('processing.ocr.unknown')}`)
        setFailed(true)
        return
      }
      timeoutId = setTimeout(poll, 1500)
    }
    poll()
    return () => {
      active = false
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
    }
  }, [accessToken, batchId, router, t])

  return (
    <Screen>
      <Text style={{ fontSize: 22, color: theme.colors.ink, fontFamily: theme.fonts.display }}>{message}</Text>
      {!failed ? (
        <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {steps[stepIndex]}
        </Text>
      ) : (
        <Button
          title={t('review.retryCta')}
          style={{ marginTop: theme.spacing.md }}
          onPress={() => router.replace('/(modals)/upload')}
        />
      )}
      {!failed ? (
        <View style={{ marginTop: theme.spacing.lg }}>
          <View style={{ height: 170 }}>
            <ShimmerBlock height={120} radius={theme.radius.lg} style={{ position: 'absolute', top: 24, left: 0 }} />
            <ShimmerBlock height={120} radius={theme.radius.lg} style={{ position: 'absolute', top: 12, left: 0 }} />
            <ShimmerBlock height={120} radius={theme.radius.lg} style={{ position: 'absolute', top: 0, left: 0 }} />
          </View>
          <Card style={{ marginTop: theme.spacing.lg }}>
            <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.bodyMedium }}>
              {t('review.detectedTitle')}
            </Text>
            <View style={{ marginTop: theme.spacing.md, gap: theme.spacing.md }}>
              {Array.from({ length: 4 }).map((_, index) => (
                <View key={`row-${index}`} style={{ flexDirection: 'row', alignItems: 'center' }}>
                  <ShimmerBlock width={32} height={32} radius={16} />
                  <View style={{ flex: 1, marginLeft: theme.spacing.sm }}>
                    <ShimmerBlock width="46%" height={10} radius={6} />
                    <ShimmerBlock width="64%" height={8} radius={6} style={{ marginTop: theme.spacing.xs }} />
                  </View>
                  <ShimmerBlock width={72} height={12} radius={6} />
                </View>
              ))}
            </View>
          </Card>
        </View>
      ) : null}
    </Screen>
  )
}
