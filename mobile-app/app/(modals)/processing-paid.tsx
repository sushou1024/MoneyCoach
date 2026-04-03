import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useEffect, useMemo, useState } from 'react'
import { Text, View } from 'react-native'

import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { ShimmerBlock } from '../../src/components/ShimmerBlock'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { requestPaidReport, fetchReport } from '../../src/services/reports'
import { fetchUploadBatch } from '../../src/services/uploads'
import { isPaidReport } from '../../src/utils/report'

export default function ProcessingPaid() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken } = useAuth()
  const params = useLocalSearchParams<{ id?: string; calculation_id?: string; return_to?: string }>()
  const batchId = Array.isArray(params.id) ? (params.id[0] ?? '') : (params.id ?? '')
  const initialCalculationID = Array.isArray(params.calculation_id)
    ? (params.calculation_id[0] ?? '')
    : (params.calculation_id ?? '')
  const returnTo = Array.isArray(params.return_to) ? params.return_to[0] : params.return_to
  const [calculationId, setCalculationId] = useState(initialCalculationID)
  const [message, setMessage] = useState(t('processing.paid.title'))
  const [requested, setRequested] = useState(false)
  const [stepIndex, setStepIndex] = useState(0)

  const steps = useMemo(() => [t('processing.paid.step1'), t('processing.paid.step2'), t('processing.paid.step3')], [t])

  useEffect(() => {
    const interval = setInterval(() => {
      setStepIndex((prev) => (prev + 1) % steps.length)
    }, 2000)
    return () => clearInterval(interval)
  }, [steps.length])

  useEffect(() => {
    let active = true
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    const pollBatch = async () => {
      if (!active || !accessToken || !batchId) return
      const resp = await fetchUploadBatch(accessToken, batchId)
      if (!active) return
      if (resp.error) {
        setMessage(`${t('processing.paid.fail')}: ${resp.error.message}`)
        return
      }
      if (resp.data && 'status' in resp.data && resp.data.status === 'completed' && 'calculation_id' in resp.data) {
        setCalculationId(resp.data.calculation_id ?? '')
        return
      }
      if (resp.data && 'status' in resp.data && resp.data.status === 'failed') {
        setMessage(`${t('processing.paid.fail')}: ${resp.data.error_code ?? t('processing.free.unknown')}`)
        return
      }
      timeoutId = setTimeout(pollBatch, 1500)
    }

    if (!calculationId) {
      pollBatch()
    }

    return () => {
      active = false
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
    }
  }, [accessToken, batchId, calculationId, t])

  useEffect(() => {
    let active = true
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    const pollReport = async () => {
      if (!active || !accessToken || !calculationId) return
      if (!requested) {
        const requestResp = await requestPaidReport(accessToken, calculationId)
        if (requestResp.error) {
          setMessage(`${t('processing.paid.fail')}: ${requestResp.error.message}`)
          return
        }
        setRequested(true)
      }
      const resp = await fetchReport(accessToken, calculationId)
      if (!active) return
      if (resp.error) {
        if (resp.status === 202 || resp.error.code === 'NOT_READY') {
          timeoutId = setTimeout(pollReport, 2000)
          return
        }
        setMessage(`${t('processing.paid.fail')}: ${resp.error.message}`)
        return
      }
      if (resp.data && isPaidReport(resp.data)) {
        router.replace({
          pathname: '/(modals)/report',
          params: returnTo ? { id: calculationId, return_to: returnTo } : { id: calculationId },
        })
        return
      }
      timeoutId = setTimeout(pollReport, 2000)
    }

    if (calculationId) {
      pollReport()
    }

    return () => {
      active = false
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
    }
  }, [accessToken, calculationId, requested, returnTo, router, t])

  return (
    <Screen>
      <Text style={{ fontSize: 22, color: theme.colors.ink, fontFamily: theme.fonts.display }}>{message}</Text>
      <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {steps[stepIndex]}
      </Text>
      <View style={{ marginTop: theme.spacing.lg }}>
        <Card>
          <ShimmerBlock width="52%" height={10} radius={6} />
          <ShimmerBlock width="72%" height={22} radius={8} style={{ marginTop: theme.spacing.sm }} />
          <View style={{ flexDirection: 'row', gap: theme.spacing.md, marginTop: theme.spacing.md }}>
            <ShimmerBlock width={80} height={80} radius={40} />
            <ShimmerBlock width={80} height={80} radius={40} />
          </View>
        </Card>
        <Card style={{ marginTop: theme.spacing.md, alignItems: 'center' }}>
          <ShimmerBlock width={140} height={140} radius={70} />
          <ShimmerBlock width="44%" height={10} radius={6} style={{ marginTop: theme.spacing.md }} />
        </Card>
        <View style={{ marginTop: theme.spacing.md, gap: theme.spacing.md }}>
          {Array.from({ length: 2 }).map((_, index) => (
            <Card key={`plan-${index}`}>
              <ShimmerBlock width="55%" height={10} radius={6} />
              <ShimmerBlock width="90%" height={8} radius={6} style={{ marginTop: theme.spacing.sm }} />
              <ShimmerBlock width="78%" height={8} radius={6} style={{ marginTop: theme.spacing.xs }} />
            </Card>
          ))}
        </View>
      </View>
    </Screen>
  )
}
