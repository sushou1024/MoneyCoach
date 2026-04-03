import { BlurView } from 'expo-blur'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useEffect, useMemo, useState } from 'react'
import { Platform, StyleSheet, Text, View } from 'react-native'

import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { DonutChart } from '../../src/components/DonutChart'
import { HealthSpotlight } from '../../src/components/HealthSpotlight'
import { PreviewLoadingState } from '../../src/components/PreviewLoadingState'
import { RadarChart } from '../../src/components/RadarChart'
import { Screen } from '../../src/components/Screen'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchPreviewReport } from '../../src/services/reports'
import { fetchUploadBatch } from '../../src/services/uploads'
import { PreviewReport } from '../../src/types/api'
import { allocationColor, allocationLabelKey } from '../../src/utils/allocations'
import {
  healthStatusFromScore,
  healthStatusLabelKey,
  toneForHealthStatus,
  toneForSeverity,
} from '../../src/utils/badges'
import { formatCurrency, formatDateTime, formatPercent, formatPortfolioTotal } from '../../src/utils/format'
import { resolveReportRiskTypeLabel, resolveReportSeverityLabel } from '../../src/utils/report'

export default function PreviewScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken } = useAuth()
  const params = useLocalSearchParams<{ id?: string; calculation_id?: string; batch_id?: string }>()
  const initialCalculationId = params.calculation_id ?? params.id ?? ''
  const batchId = params.batch_id ?? ''
  const [calculationId, setCalculationId] = useState(initialCalculationId)
  const [preview, setPreview] = useState<PreviewReport | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const steps = useMemo(
    () => [
      t('processing.free.step1'),
      t('processing.free.step2'),
      t('processing.free.step3'),
      t('processing.free.step4'),
      t('processing.free.step5'),
    ],
    [t]
  )

  useEffect(() => {
    setPreview(null)
    setError(null)
    setLoading(true)
  }, [calculationId, batchId])

  useEffect(() => {
    if (!batchId || calculationId) return
    let active = true
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    const pollBatch = async () => {
      if (!active || !accessToken || !batchId || calculationId) return
      const resp = await fetchUploadBatch(accessToken, batchId)
      if (!active) return
      if (resp.error) {
        setError(`${t('processing.free.fail')}: ${resp.error.message}`)
        setLoading(false)
        return
      }
      if (resp.data && 'status' in resp.data) {
        if (resp.data.status === 'completed' && 'calculation_id' in resp.data && resp.data.calculation_id) {
          setCalculationId(resp.data.calculation_id)
          return
        }
        if (resp.data.status === 'failed') {
          setError(`${t('processing.free.fail')}: ${resp.data.error_code ?? t('processing.free.unknown')}`)
          setLoading(false)
          return
        }
      }
      timeoutId = setTimeout(pollBatch, 1500)
    }
    pollBatch()
    return () => {
      active = false
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
    }
  }, [accessToken, batchId, calculationId, t])

  useEffect(() => {
    if (!calculationId) return
    let active = true
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    const pollPreview = async () => {
      if (!active || !accessToken || !calculationId) return
      const resp = await fetchPreviewReport(accessToken, calculationId)
      if (!active) return
      if (resp.error) {
        if (resp.status === 202 || resp.error.code === 'NOT_READY') {
          setLoading(true)
          timeoutId = setTimeout(pollPreview, 2000)
          return
        }
        setError(resp.error.message)
        setLoading(false)
        return
      }
      if (resp.data) {
        setPreview(resp.data)
        setLoading(false)
        setError(null)
        return
      }
      timeoutId = setTimeout(pollPreview, 2000)
    }
    pollPreview()
    return () => {
      active = false
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
    }
  }, [accessToken, calculationId])

  if (loading) {
    return <PreviewLoadingState title={t('processing.free.title')} steps={steps} />
  }

  if (error) {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted }}>{t('preview.error', { message: error })}</Text>
        <Button
          title={t('common.close')}
          variant="ghost"
          style={{ marginTop: theme.spacing.md }}
          onPress={() => router.back()}
        />
      </Screen>
    )
  }

  const allocation = preview?.asset_allocation ?? []
  const risks = preview?.identified_risks ?? []
  const healthScore = preview?.fixed_metrics?.health_score ?? preview?.report_header?.health_score?.value
  const volatilityScore =
    preview?.fixed_metrics?.volatility_score ?? preview?.report_header?.volatility_dashboard?.value
  // Fallback to score-derived status to keep the badge consistent with Assets.
  const healthStatus = preview?.fixed_metrics?.health_status ?? healthStatusFromScore(healthScore)
  const healthStatusLabel = healthStatus ? t(healthStatusLabelKey(healthStatus) as any) : null
  const volatilityLevel =
    volatilityScore === undefined || volatilityScore === null
      ? null
      : volatilityScore < 35
        ? 'low'
        : volatilityScore < 70
          ? 'medium'
          : 'high'
  const volatilityLabel = volatilityLevel ? t(`insights.severity.${volatilityLevel}` as any) : null
  const volatilityTone = volatilityLevel ? toneForSeverity(volatilityLevel) : 'neutral'
  const volatilityColor =
    volatilityScore === undefined || volatilityScore === null
      ? theme.colors.muted
      : volatilityScore < 40
        ? theme.colors.success
        : volatilityScore < 70
          ? theme.colors.warning
          : theme.colors.danger
  const lockedRadarSample = {
    liquidity: 100,
    diversification: 54,
    alpha: 45,
    drawdown: 76,
  }
  const lockedPlanSample = {
    name: t('preview.lockedSample.name'),
    rationale: t('preview.lockedSample.rationale'),
    execution: t('preview.lockedSample.execution', {
      price: formatCurrency(80448.768, preview?.base_currency ?? 'USD'),
    }),
    expected: t('preview.lockedSample.expected'),
    stopLossPrice: 80448.768,
    stopLossPct: 0.096,
  }
  const lockedSeverityTone = 'medium'
  const lockedSeverityLabel = t('insights.severity.medium')
  const blurIntensity = Platform.select({ ios: 22, android: 35, default: 28 }) ?? 28
  const blurMethod = 'dimezisBlurView' as const
  const blurOverlayOpacity = Platform.select({ ios: 0.03, android: 0.06, default: 0.05 }) ?? 0.05
  const blurReductionFactor = Platform.OS === 'android' ? 4 : 1

  return (
    <Screen scroll>
      <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
        {t('preview.title')}
      </Text>
      {preview && (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('preview.netWorth')}</Text>
          <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
            {formatPortfolioTotal(preview.net_worth_display ?? 0, preview.base_currency ?? 'USD')}
          </Text>
          <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
            {t('preview.valuationAsOf', { time: formatDateTime(preview.valuation_as_of) })}
          </Text>
          <View style={{ marginTop: theme.spacing.md }}>
            <HealthSpotlight
              healthScore={healthScore}
              healthLabel={t('report.healthLabel')}
              healthGaugeLabel={t('assets.healthShort')}
              healthStatusLabel={healthStatusLabel}
              healthTone={toneForHealthStatus(healthStatus)}
              volatilityScore={volatilityScore}
              volatilityTitle={t('report.volatilityLabel')}
              volatilityLevelLabel={volatilityLabel}
              volatilityTone={volatilityTone}
              volatilityColor={volatilityColor}
            />
          </View>
        </Card>
      )}

      {allocation.length > 0 ? (
        <View style={{ marginTop: theme.spacing.lg }}>
          <Text style={{ fontSize: 18, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
            {t('preview.allocationTitle')}
          </Text>
          <Card style={{ marginTop: theme.spacing.sm }}>
            <View style={{ flexDirection: 'row', gap: theme.spacing.md, alignItems: 'center' }}>
              <DonutChart
                size={140}
                strokeWidth={16}
                segments={allocation.map((item) => ({
                  value: item.value_display ?? item.value_usd,
                  color: allocationColor(theme, item.label),
                }))}
              />
              <View style={{ flex: 1 }}>
                {allocation.map((item, index) => (
                  <View
                    key={`${item.label}-${index}`}
                    style={{ flexDirection: 'row', alignItems: 'center', marginBottom: theme.spacing.sm }}
                  >
                    <View
                      style={{
                        width: 10,
                        height: 10,
                        borderRadius: 999,
                        backgroundColor: allocationColor(theme, item.label),
                      }}
                    />
                    <View style={{ marginLeft: theme.spacing.sm, flex: 1 }}>
                      <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
                        {t(allocationLabelKey(item.label) as any)}
                      </Text>
                      <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
                        {formatPercent(item.weight_pct)} •{' '}
                        {formatCurrency(
                          item.value_display ?? item.value_usd,
                          item.display_currency ?? preview?.base_currency ?? 'USD'
                        )}
                      </Text>
                    </View>
                  </View>
                ))}
              </View>
            </View>
          </Card>
        </View>
      ) : null}

      <Text
        style={{ marginTop: theme.spacing.lg, fontSize: 18, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}
      >
        {t('preview.risksTitle')}
      </Text>
      {risks.map((item) => (
        <Card key={item.risk_id} style={{ marginTop: theme.spacing.sm }}>
          <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
              {resolveReportRiskTypeLabel(t, item.type)}
            </Text>
            <Badge label={resolveReportSeverityLabel(t, item.severity)} tone={toneForSeverity(item.severity)} />
          </View>
          <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
            {item.teaser_text}
          </Text>
          <View style={{ marginTop: theme.spacing.sm }}>
            <View style={{ height: 8, borderRadius: 999, backgroundColor: theme.colors.border, opacity: 0.4 }} />
            <View
              style={{
                height: 8,
                borderRadius: 999,
                backgroundColor: theme.colors.border,
                opacity: 0.4,
                marginTop: theme.spacing.xs,
                width: '70%',
              }}
            />
          </View>
        </Card>
      ))}

      <Card style={{ marginTop: theme.spacing.lg }}>
        <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>{t('preview.lockedRadar')}</Text>
        <View style={{ marginTop: theme.spacing.sm }}>
          <View
            style={[styles.blurFrame, { borderRadius: theme.radius.md, backgroundColor: theme.colors.surfaceElevated }]}
          >
            <View style={{ paddingVertical: theme.spacing.sm, alignItems: 'center' }}>
              <RadarChart
                size={190}
                values={lockedRadarSample}
                labels={{
                  liquidity: t('report.radar.liquidity'),
                  diversification: t('report.radar.diversification'),
                  alpha: t('report.radar.alpha'),
                  drawdown: t('report.radar.drawdown'),
                }}
              />
            </View>
            <BlurView
              intensity={blurIntensity}
              tint="light"
              experimentalBlurMethod={blurMethod}
              blurReductionFactor={blurReductionFactor}
              style={StyleSheet.absoluteFillObject}
            />
            <View
              pointerEvents="none"
              style={[
                StyleSheet.absoluteFillObject,
                { backgroundColor: theme.colors.surface, opacity: blurOverlayOpacity, borderRadius: theme.radius.md },
              ]}
            />
          </View>
        </View>
      </Card>

      {preview?.locked_projection ? (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>{t('preview.lockedTitle')}</Text>
          <View style={{ marginTop: theme.spacing.sm }}>
            <View
              style={[
                styles.blurFrame,
                { borderRadius: theme.radius.md, backgroundColor: theme.colors.surfaceElevated },
              ]}
            >
              <View style={{ padding: theme.spacing.md }}>
                <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
                    {lockedPlanSample.name}
                  </Text>
                  <Badge label={lockedSeverityLabel} tone={toneForSeverity(lockedSeverityTone)} />
                </View>
                <Text
                  style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}
                  numberOfLines={3}
                >
                  {lockedPlanSample.rationale}
                </Text>
                <Text
                  style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}
                  numberOfLines={3}
                >
                  {lockedPlanSample.execution}
                </Text>
                <Text
                  style={{
                    color: theme.colors.ink,
                    marginTop: theme.spacing.sm,
                    fontFamily: theme.fonts.bodyBold,
                    fontSize: 12,
                  }}
                >
                  {t('report.expectedOutcomeTitle')}
                </Text>
                <Text
                  style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}
                  numberOfLines={3}
                >
                  {lockedPlanSample.expected}
                </Text>
                <Text
                  style={{ color: theme.colors.muted, marginTop: theme.spacing.sm, fontFamily: theme.fonts.bodyMedium }}
                  numberOfLines={1}
                >
                  {t('strategy.visualization.stopLossLineWithPct', {
                    price: formatCurrency(lockedPlanSample.stopLossPrice, preview?.base_currency ?? 'USD'),
                    percent: formatPercent(lockedPlanSample.stopLossPct),
                  })}
                </Text>
              </View>
              <BlurView
                intensity={blurIntensity}
                tint="light"
                experimentalBlurMethod={blurMethod}
                blurReductionFactor={blurReductionFactor}
                style={StyleSheet.absoluteFillObject}
              />
              <View
                pointerEvents="none"
                style={[
                  StyleSheet.absoluteFillObject,
                  { backgroundColor: theme.colors.surface, opacity: blurOverlayOpacity, borderRadius: theme.radius.md },
                ]}
              />
            </View>
          </View>
        </Card>
      ) : null}

      <Button
        title={t('preview.unlockCta')}
        style={{ marginTop: theme.spacing.lg }}
        onPress={() => router.push({ pathname: '/(modals)/paywall', params: { id: calculationId } })}
      />

      <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {t('common.disclaimer')}
      </Text>

      <Button
        title={t('common.close')}
        variant="ghost"
        style={{ marginTop: theme.spacing.sm }}
        onPress={() => router.back()}
      />
    </Screen>
  )
}

const styles = StyleSheet.create({
  blurFrame: {
    position: 'relative',
    overflow: 'hidden',
  },
})
