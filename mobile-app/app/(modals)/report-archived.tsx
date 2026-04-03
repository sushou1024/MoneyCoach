import { useQuery } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React from 'react'
import { ActivityIndicator, Text, View } from 'react-native'

import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { DonutChart } from '../../src/components/DonutChart'
import { HealthSpotlight } from '../../src/components/HealthSpotlight'
import { RadarChart } from '../../src/components/RadarChart'
import { Screen } from '../../src/components/Screen'
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
import { formatCurrency, formatDateTime, formatPercent, formatPortfolioTotal } from '../../src/utils/format'
import { isPaidReport, resolveReportRiskTypeLabel, resolveReportSeverityLabel } from '../../src/utils/report'
import { getStrategyDisplayName } from '../../src/utils/strategies'

export default function ReportArchivedScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken } = useAuth()
  const params = useLocalSearchParams<{ id?: string }>()
  const calculationId = params.id ?? ''

  const query = useQuery<PaidReport | PreviewReport | null>({
    queryKey: ['report', calculationId],
    queryFn: async () => {
      if (!accessToken || !calculationId) return null
      const resp = await fetchReport(accessToken, calculationId)
      if (resp.error) throw new Error(resp.error.message)
      return resp.data ?? null
    },
    enabled: !!accessToken && !!calculationId,
  })

  const report = query.data
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

  if (query.isLoading || !report) {
    return (
      <Screen>
        <ActivityIndicator size="large" color={theme.colors.accent} />
        <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {t('report.loading')}
        </Text>
      </Screen>
    )
  }

  if (!isPaidReport(report)) {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('report.error')}</Text>
      </Screen>
    )
  }

  const allocation = report.asset_allocation ?? []
  const risks = report.risk_insights?.length ? report.risk_insights : (report.exposure_analysis ?? [])
  const plan = report.optimization_plan ?? []

  return (
    <Screen scroll>
      <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
        {t('reportArchived.title')}
      </Text>

      <Card style={{ marginTop: theme.spacing.md }}>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('report.netWorth')}</Text>
        <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
          {formatPortfolioTotal(report.net_worth_display ?? 0, report.base_currency ?? 'USD')}
        </Text>
        <Text style={{ marginTop: theme.spacing.xs, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {t('report.valuationAsOf', { time: formatDateTime(report.valuation_as_of) })}
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
                        {formatPercent(item.weight_pct)} •{' '}
                        {formatCurrency(
                          item.value_display ?? item.value_usd,
                          item.display_currency ?? report.base_currency ?? 'USD'
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
          {report.charts?.radar_chart ? (
            <>
              <RadarChart
                values={{
                  liquidity: report.charts.radar_chart.liquidity ?? 0,
                  diversification: report.charts.radar_chart.diversification ?? 0,
                  alpha: report.charts.radar_chart.alpha ?? 0,
                  drawdown: report.charts.radar_chart.drawdown ?? 0,
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
            <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
              {item.expected_outcome}
            </Text>
            {executionSummary ? (
              <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
                {executionSummary}
              </Text>
            ) : null}
          </Card>
        )
      })}

      <Card style={{ marginTop: theme.spacing.lg }}>
        <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>{t('report.dailyAlpha')}</Text>
        {report.daily_alpha_signal ? (
          <>
            <View
              style={{
                flexDirection: 'row',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginTop: theme.spacing.sm,
              }}
            >
              <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
                {report.daily_alpha_signal.asset}
              </Text>
              <Badge
                label={report.daily_alpha_signal.severity}
                tone={toneForSeverity(report.daily_alpha_signal.severity)}
              />
            </View>
            <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
              {report.daily_alpha_signal.trigger_reason}
            </Text>
          </>
        ) : (
          <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.sm, fontFamily: theme.fonts.body }}>
            {t('report.dailyAlphaEmpty')}
          </Text>
        )}
      </Card>

      <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {t('report.archivedNote')}
      </Text>

      <Button
        title={t('common.close')}
        variant="ghost"
        style={{ marginTop: theme.spacing.lg }}
        onPress={() => router.back()}
      />
    </Screen>
  )
}
