import { Ionicons } from '@expo/vector-icons'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useEffect } from 'react'
import { ActivityIndicator, Animated, Modal, Platform, Pressable, Text, View, ViewStyle } from 'react-native'

import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { DonutChart } from '../../src/components/DonutChart'
import { HealthSpotlight } from '../../src/components/HealthSpotlight'
import { RadarChart } from '../../src/components/RadarChart'
import { ReportPlanActions } from '../../src/components/ReportPlanActions'
import { Screen } from '../../src/components/Screen'
import { useProfile } from '../../src/hooks/useProfile'
import { useReportNotifications } from '../../src/hooks/useReportNotifications'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchReport } from '../../src/services/reports'
import { PaidReport, PreviewReport } from '../../src/types/api'
import { allocationColor, allocationLabelKey } from '../../src/utils/allocations'
import {
  healthStatusFromScore,
  healthStatusLabelKey,
  toneForHealthStatus,
  toneForSeverity,
} from '../../src/utils/badges'
import { formatCompactCurrency, formatDateTime, formatPercent, formatPortfolioTotal } from '../../src/utils/format'
import { isPaidReport, resolveReportRiskTypeLabel, resolveReportSeverityLabel } from '../../src/utils/report'
import { getStrategyDisplayName } from '../../src/utils/strategies'

export default function ReportScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const queryClient = useQueryClient()
  const { accessToken, userId } = useAuth()
  const profile = useProfile()
  const entitlementStatus = profile.data?.entitlement?.status
  const isPaidUser = ['active', 'grace'].includes(entitlementStatus ?? '')
  const params = useLocalSearchParams<{ id?: string; return_to?: string }>()
  const calculationId = Array.isArray(params.id) ? (params.id[0] ?? '') : (params.id ?? '')
  const returnTo = Array.isArray(params.return_to) ? params.return_to[0] : params.return_to
  const {
    showNotifyModal,
    notifyPlanId,
    notifyInlineState,
    notifyEnabledAnim,
    notifyInlineAnim,
    notifyEnabled,
    isEnablingNotifications,
    handleNotify,
    enableNotifications,
    closeNotifyModal,
  } = useReportNotifications({ accessToken, userId: userId ?? undefined, queryClient })

  const query = useQuery<PaidReport | PreviewReport | null>({
    queryKey: ['report', calculationId],
    queryFn: async () => {
      if (!accessToken || !calculationId) return null
      const resp = await fetchReport(accessToken, calculationId)
      if (resp.error) {
        if (resp.status === 202 || resp.error.code === 'NOT_READY') {
          return null
        }
        throw new Error(resp.error.message)
      }
      if (resp.data && isPaidUser && !isPaidReport(resp.data)) {
        return null
      }
      return resp.data ?? null
    },
    enabled: !!accessToken && !!calculationId,
    refetchInterval: (currentQuery) => {
      const data = currentQuery.state.data
      if (!data) return 2000
      if (isPaidUser && !isPaidReport(data)) return 2000
      return false
    },
  })

  const report = query.data
  const paidReport = isPaidReport(report) ? report : null
  const healthScore = report?.report_header?.health_score?.value ?? report?.fixed_metrics?.health_score
  const volatilityScore = report?.report_header?.volatility_dashboard?.value ?? report?.fixed_metrics?.volatility_score
  // Fallback to score-derived status to keep the badge consistent with Assets.
  const healthStatus = report?.fixed_metrics?.health_status ?? healthStatusFromScore(healthScore)
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
  const notifyModalCardStyle: ViewStyle | undefined =
    Platform.OS === 'web'
      ? {
          width: '100%',
          maxWidth: 420,
          alignSelf: 'center',
        }
      : undefined

  const handleBack = () => {
    if (returnTo) {
      router.dismissTo(returnTo)
      return
    }
    router.back()
  }

  useEffect(() => {
    if (profile.isLoading || profile.isError) return
    if (!report || paidReport || isPaidUser) return
    router.replace({
      pathname: '/(modals)/preview',
      params: returnTo ? { id: calculationId, return_to: returnTo } : { id: calculationId },
    })
  }, [calculationId, isPaidUser, paidReport, profile.isError, profile.isLoading, report, returnTo, router])

  if (query.isError) {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted }}>{t('report.error')}</Text>
      </Screen>
    )
  }

  if (query.isLoading || !report) {
    return (
      <Screen>
        <ActivityIndicator size="large" color={theme.colors.accent} />
        <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted }}>{t('report.loading')}</Text>
      </Screen>
    )
  }

  if (!paidReport) {
    return (
      <Screen>
        <ActivityIndicator size="large" color={theme.colors.accent} />
        <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted }}>
          {isPaidUser ? t('processing.paid.title') : t('report.loading')}
        </Text>
      </Screen>
    )
  }

  const allocation = paidReport.asset_allocation ?? []
  const risks = paidReport.risk_insights?.length ? paidReport.risk_insights : (paidReport.exposure_analysis ?? [])
  const plan = paidReport.optimization_plan ?? []

  return (
    <Screen scroll>
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.sm }}>
        <Pressable onPress={handleBack} accessibilityRole="button" hitSlop={8}>
          <Ionicons name="arrow-back" size={24} color={theme.colors.ink} />
        </Pressable>
        <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
          {t('report.title')}
        </Text>
      </View>

      <Card style={{ marginTop: theme.spacing.md }}>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('report.netWorth')}</Text>
        <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
          {formatPortfolioTotal(paidReport.net_worth_display ?? 0, paidReport.base_currency ?? 'USD')}
        </Text>
        <Text style={{ marginTop: theme.spacing.xs, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {t('report.valuationAsOf', { time: formatDateTime(paidReport.valuation_as_of) })}
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

      {allocation.length > 0 ? (
        <View style={{ marginTop: theme.spacing.lg }}>
          <Text style={{ fontSize: 18, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
            {t('report.allocationTitle')}
          </Text>
          <Card style={{ marginTop: theme.spacing.sm }}>
            <View style={{ flexDirection: 'row', gap: theme.spacing.md, alignItems: 'center' }}>
              <DonutChart
                size={150}
                strokeWidth={18}
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
                        {formatPercent(item.weight_pct)}
                      </Text>
                      <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
                        {formatCompactCurrency(
                          item.value_display ?? item.value_usd,
                          item.display_currency ?? paidReport.base_currency ?? 'USD'
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
        {t('report.risksTitle')}
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
            {item.message}
          </Text>
        </Card>
      ))}

      <View style={{ marginTop: theme.spacing.lg }}>
        <Text style={{ fontSize: 18, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
          {t('report.radarTitle')}
        </Text>
        <Card style={{ marginTop: theme.spacing.sm, alignItems: 'center', paddingHorizontal: 0 }}>
          {paidReport.charts?.radar_chart ? (
            <>
              <RadarChart
                values={{
                  liquidity: paidReport.charts.radar_chart.liquidity ?? 0,
                  diversification: paidReport.charts.radar_chart.diversification ?? 0,
                  alpha: paidReport.charts.radar_chart.alpha ?? 0,
                  drawdown: paidReport.charts.radar_chart.drawdown ?? 0,
                }}
                labels={{
                  liquidity: t('report.radar.liquidity'),
                  diversification: t('report.radar.diversification'),
                  alpha: t('report.radar.alpha'),
                  drawdown: t('report.radar.drawdown'),
                }}
              />
            </>
          ) : (
            <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
              {t('report.radarUnavailable')}
            </Text>
          )}
        </Card>
      </View>

      <Text
        style={{ marginTop: theme.spacing.lg, fontSize: 18, color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}
      >
        {t('report.planTitle')}
      </Text>
      {plan.map((item) => {
        const linked = risks.find((risk) => risk.risk_id === item.linked_risk_id)
        const executionSummary = item.execution_summary?.trim()
        const badgeLabel = item.priority ?? linked?.severity
        return (
          <Card key={item.plan_id} style={{ marginTop: theme.spacing.sm }}>
            <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
              <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
                {getStrategyDisplayName(t, item.strategy_id)}
              </Text>
              {badgeLabel ? <Badge label={badgeLabel} tone={toneForSeverity(badgeLabel)} /> : null}
            </View>
            <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
              {item.rationale}
            </Text>
            {executionSummary ? (
              <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
                {executionSummary}
              </Text>
            ) : null}
            <View style={{ marginTop: theme.spacing.sm }}>
              <Text
                style={{
                  color: theme.colors.ink,
                  fontFamily: theme.fonts.bodyBold,
                  fontSize: 12,
                }}
              >
                {t('report.expectedOutcomeTitle')}
              </Text>
              <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
                {item.expected_outcome}
              </Text>
            </View>
            <ReportPlanActions
              onAutoExecute={() =>
                router.push({
                  pathname: '/(modals)/waitlist',
                  params: { strategy_id: item.strategy_id, calculation_id: calculationId },
                })
              }
              onViewStrategy={() =>
                router.push({
                  pathname: '/(modals)/strategy',
                  params: { plan_id: item.plan_id, calculation_id: calculationId },
                })
              }
              onNotify={() => handleNotify(item.plan_id)}
              notifyEnabled={notifyEnabled}
              notifyEnabledAnim={notifyEnabledAnim}
            />
            {item.plan_id === notifyPlanId && notifyInlineState ? (
              <Animated.View style={{ marginTop: theme.spacing.xs, opacity: notifyInlineAnim }}>
                <Text
                  style={{
                    color: notifyInlineState === 'success' ? theme.colors.accent : theme.colors.muted,
                    fontFamily: theme.fonts.bodyMedium,
                    fontSize: 13,
                  }}
                >
                  {notifyInlineState === 'success' ? t('report.notifyInlineSuccess') : t('report.notifyInlineDenied')}
                </Text>
              </Animated.View>
            ) : null}
          </Card>
        )
      })}

      <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {t('common.disclaimer')}
      </Text>

      <Button title={t('common.close')} variant="ghost" style={{ marginTop: theme.spacing.lg }} onPress={handleBack} />

      <Modal transparent visible={showNotifyModal} animationType="fade">
        <Pressable
          style={{
            flex: 1,
            backgroundColor: theme.colors.overlay,
            justifyContent: 'center',
            padding: theme.spacing.lg,
            alignItems: Platform.OS === 'web' ? 'center' : undefined,
          }}
          onPress={closeNotifyModal}
        >
          <Card style={notifyModalCardStyle}>
            <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
              {t('report.notifyModalTitle')}
            </Text>
            <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.sm, fontFamily: theme.fonts.body }}>
              {t('report.notifyModalBody')}
            </Text>
            <Button
              title={t('report.notifyEnable')}
              style={{ marginTop: theme.spacing.md }}
              onPress={enableNotifications}
              loading={isEnablingNotifications}
              disabled={isEnablingNotifications}
            />
            <Button
              title={t('common.cancel')}
              variant="ghost"
              style={{ marginTop: theme.spacing.sm }}
              onPress={closeNotifyModal}
              disabled={isEnablingNotifications}
            />
          </Card>
        </Pressable>
      </Modal>
    </Screen>
  )
}
