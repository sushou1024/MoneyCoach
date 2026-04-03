import { Ionicons } from '@expo/vector-icons'
import { useQuery } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { Alert, FlatList, Pressable, StyleSheet, Text, View } from 'react-native'

import { AssetLogo } from '../../src/components/AssetLogo'
import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Input } from '../../src/components/Input'
import { Screen } from '../../src/components/Screen'
import { ShimmerBlock } from '../../src/components/ShimmerBlock'
import { useEntitlement } from '../../src/hooks/useEntitlement'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { dismissInsight, fetchInsights } from '../../src/services/insights'
import { fetchMarketRegime } from '../../src/services/intelligence'
import { toneForSeverity } from '../../src/utils/badges'
import { formatDateTime, formatNumber, formatPercent } from '../../src/utils/format'
import {
  describeRegimeDriver,
  formatSignedPercent,
  resolveActionBiasLabel,
  resolveRegimeLabel,
  resolveRegimeSummary,
  resolveSemanticLabel,
  resolveSummarySignalLabel,
  resolveTrendStrengthLabel,
  toneForActionBias,
  toneForRegime,
} from '../../src/utils/intelligence'

const filters = [
  { key: 'all', labelKey: 'insights.filter.all' },
  { key: 'portfolio_watch', labelKey: 'insights.filter.portfolio' },
  { key: 'market_alpha', labelKey: 'insights.filter.market' },
  { key: 'action_alert', labelKey: 'insights.filter.action' },
]

export default function InsightsScreen() {
  const theme = useTheme()
  const styles = useMemo(() => createStyles(theme), [theme])
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken, clearSession, userId } = useAuth()
  const entitlement = useEntitlement()
  const [filter, setFilter] = useState('all')
  const [search, setSearch] = useState('')

  const isPaid = ['active', 'grace'].includes(entitlement.data?.status ?? '')
  const subtitle = t('insights.subtitle')

  const insightsQuery = useQuery({
    queryKey: ['insights', userId, filter],
    queryFn: async () => {
      if (!accessToken) return []
      const resp = await fetchInsights(accessToken, filter)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return []
        }
        throw new Error(resp.error.message)
      }
      return resp.data?.items ?? []
    },
    enabled: !!accessToken && isPaid,
  })

  const regimeQuery = useQuery({
    queryKey: ['market-regime', userId],
    queryFn: async () => {
      if (!accessToken) return null
      const resp = await fetchMarketRegime(accessToken)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return null
        }
        if (resp.status === 404) return null
        throw new Error(resp.error.message)
      }
      return resp.data ?? null
    },
    enabled: !!accessToken && isPaid,
  })

  const insights = insightsQuery.data ?? []
  const filteredInsights = insights.filter((item) => {
    if (!search.trim()) return true
    const keyword = search.trim().toLowerCase()
    return (
      item.asset?.toLowerCase().includes(keyword) ||
      item.trigger_reason?.toLowerCase().includes(keyword) ||
      item.suggested_action?.toLowerCase().includes(keyword)
    )
  })

  const featuredAssets = (regimeQuery.data?.featured_assets ?? []).filter((item) => {
    if (!search.trim()) return true
    const keyword = search.trim().toLowerCase()
    return (
      item.symbol.toLowerCase().includes(keyword) ||
      item.name?.toLowerCase().includes(keyword) ||
      item.summary_signal.toLowerCase().includes(keyword)
    )
  })

  const handleDismiss = async (id: string) => {
    if (!accessToken) return
    const resp = await dismissInsight(accessToken, id, 'not_relevant')
    if (resp.error) {
      Alert.alert(t('insights.dismissFailTitle'), resp.error.message)
      return
    }
    insightsQuery.refetch()
    regimeQuery.refetch()
  }

  const resolveTranslation = (key: string, fallback: string) => {
    const value = t(key as any)
    return value === key ? fallback : value
  }

  const resolveSeverityLabel = (severity?: string) => {
    const normalized = (severity ?? '').toLowerCase()
    const key = `insights.severity.${normalized || 'unknown'}`
    const fallback = severity || t('insights.severity.unknown')
    return resolveTranslation(key, fallback)
  }

  const severityRailColor = (severity?: string) => {
    const tone = toneForSeverity(severity)
    if (tone === 'critical') return theme.colors.danger
    if (tone === 'high') return theme.colors.warning
    if (tone === 'medium') return theme.colors.accent
    if (tone === 'low') return theme.colors.success
    return theme.colors.border
  }

  const openAssetBrief = (assetKey: string) => {
    router.push({ pathname: '/(modals)/asset-brief', params: { asset_key: assetKey } })
  }

  const openQueueItem = (id: string, type: string, assetKey?: string) => {
    if (type === 'market_alpha' && assetKey) {
      openAssetBrief(assetKey)
      return
    }
    router.push({
      pathname: type === 'market_alpha' ? '/(modals)/insight-detail' : '/(modals)/quick-update',
      params: { id },
    })
  }

  if (!isPaid) {
    return (
      <Screen scroll>
        <Text style={styles.title}>{t('insights.title')}</Text>
        <Text style={styles.subtitle}>{subtitle}</Text>
        <Card style={{ marginTop: theme.spacing.lg }}>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.sm }}>
            <View style={styles.lockIconWrap}>
              <Ionicons name="lock-closed-outline" size={18} color={theme.colors.accent} />
            </View>
            <View style={{ flex: 1 }}>
              <Text style={styles.lockTitle}>{t('insights.lockedTitle')}</Text>
              <Text style={styles.lockBody}>{t('insights.lockedBody')}</Text>
            </View>
          </View>
          <Button
            title={t('insights.lockedCta')}
            style={{ marginTop: theme.spacing.md }}
            onPress={() => router.push('/(modals)/paywall')}
          />
        </Card>
        <View style={{ marginTop: theme.spacing.lg }}>
          <Text style={styles.previewLabel}>{t('insights.previewTitle')}</Text>
          <View style={{ marginTop: theme.spacing.sm, gap: theme.spacing.sm }}>
            {Array.from({ length: 3 }).map((_, index) => (
              <Card key={`preview-${index}`} style={{ opacity: 0.6 }}>
                <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                  <View style={styles.previewBarWide} />
                  <View style={styles.previewBarShort} />
                </View>
                <View style={styles.previewBarHeadline} />
                <View style={styles.previewBarBody} />
                <View style={styles.previewBarBodyShort} />
              </Card>
            ))}
          </View>
        </View>
      </Screen>
    )
  }

  return (
    <Screen scroll>
      <Text style={styles.title}>{t('insights.title')}</Text>
      <Text style={styles.subtitle}>{subtitle}</Text>

      {regimeQuery.isLoading ? (
        <Card style={{ marginTop: theme.spacing.lg }}>
          <ShimmerBlock width="36%" height={18} radius={8} />
          <ShimmerBlock width="84%" height={10} radius={6} style={{ marginTop: theme.spacing.sm }} />
          <View style={styles.metricRow}>
            {Array.from({ length: 3 }).map((_, index) => (
              <View key={`regime-loading-${index}`} style={styles.metricShell}>
                <ShimmerBlock width="72%" height={10} radius={6} />
                <ShimmerBlock width="54%" height={16} radius={8} style={{ marginTop: theme.spacing.sm }} />
              </View>
            ))}
          </View>
        </Card>
      ) : regimeQuery.data ? (
        <Card style={{ marginTop: theme.spacing.lg }}>
          <View style={styles.sectionHeader}>
            <View style={{ flex: 1 }}>
              <Text style={styles.sectionEyebrow}>{t('insights.regimeTitle')}</Text>
              <Text style={styles.sectionTitle}>{resolveRegimeLabel(t, regimeQuery.data.regime)}</Text>
            </View>
            <Badge
              label={resolveTrendStrengthLabel(t, regimeQuery.data.trend_strength)}
              tone={toneForRegime(regimeQuery.data.regime)}
            />
          </View>
          <Text style={styles.sectionBody}>{resolveRegimeSummary(t, regimeQuery.data.regime)}</Text>
          {regimeQuery.data.drivers[0] ? (
            <Text style={styles.actionHint}>{describeRegimeDriver(t, regimeQuery.data.drivers[0])}</Text>
          ) : null}
          <View style={styles.metricRow}>
            <View style={styles.metricShell}>
              <Text style={styles.metricLabel}>{t('intelligence.metric.alpha30d')}</Text>
              <Text style={styles.metricValue}>{formatSignedPercent(regimeQuery.data.metrics.alpha_30d, 1)}</Text>
            </View>
            <View style={styles.metricShell}>
              <Text style={styles.metricLabel}>{t('intelligence.metric.topPosition')}</Text>
              <Text style={styles.metricValue}>{formatPercent(regimeQuery.data.metrics.top_asset_pct, 0)}</Text>
            </View>
            <View style={styles.metricShell}>
              <Text style={styles.metricLabel}>{t('intelligence.metric.cashBuffer')}</Text>
              <Text style={styles.metricValue}>{formatPercent(regimeQuery.data.metrics.cash_pct, 0)}</Text>
            </View>
          </View>
          <View style={{ marginTop: theme.spacing.md, gap: theme.spacing.xs }}>
            {regimeQuery.data.actions.slice(0, 2).map((item) => (
              <Text key={item.id} style={styles.actionHint}>
                • {resolveSemanticLabel(t, item.kind)}
              </Text>
            ))}
          </View>
          <Button
            title={t('insights.regimeCta')}
            variant="secondary"
            style={{ marginTop: theme.spacing.md }}
            onPress={() => router.push('/(modals)/market-regime')}
          />
        </Card>
      ) : (
        <Card style={{ marginTop: theme.spacing.lg }}>
          <Text style={styles.emptyText}>{t('insights.regimeUnavailable')}</Text>
        </Card>
      )}

      <View style={{ marginTop: theme.spacing.lg }}>
        <View style={styles.sectionHeaderSimple}>
          <Text style={styles.sectionTitle}>{t('insights.featuredTitle')}</Text>
        </View>
        <View style={{ marginTop: theme.spacing.sm, gap: theme.spacing.sm }}>
          {featuredAssets.length === 0 ? (
            <Card>
              <Text style={styles.emptyText}>{t('insights.featuredEmpty')}</Text>
            </Card>
          ) : (
            featuredAssets.map((item) => (
              <Pressable
                key={item.asset_key}
                onPress={() => openAssetBrief(item.asset_key)}
                accessibilityRole="button"
                style={({ pressed }) => [pressed && styles.pressed]}
              >
                <Card>
                  <View style={styles.featuredHeader}>
                    <View style={styles.featuredMeta}>
                      <AssetLogo uri={item.logo_url} label={item.symbol} size={42} />
                      <View style={{ flex: 1 }}>
                        <Text style={styles.featuredSymbol}>{item.symbol}</Text>
                        <Text style={styles.featuredName}>{item.name ?? item.asset_type}</Text>
                      </View>
                    </View>
                    <Badge
                      label={resolveActionBiasLabel(t, item.action_bias)}
                      tone={toneForActionBias(item.action_bias)}
                    />
                  </View>
                  <Text style={styles.featuredSummary}>{resolveSummarySignalLabel(t, item.summary_signal)}</Text>
                  <View style={styles.featuredStats}>
                    <Text style={styles.featuredStat}>
                      {t('intelligence.featured.weight')}: {formatPercent(item.weight_pct, 0)}
                    </Text>
                    <Text style={styles.featuredStat}>
                      {t('intelligence.featured.signals')}: {formatNumber(item.signal_count, 0)}
                    </Text>
                    <Text style={styles.featuredStat}>
                      {t('intelligence.featured.beta')}: {formatNumber(item.beta_to_portfolio, 2)}
                    </Text>
                  </View>
                </Card>
              </Pressable>
            ))
          )}
        </View>
      </View>

      <Input
        value={search}
        onChangeText={setSearch}
        placeholder={t('insights.searchPlaceholder')}
        style={{ marginTop: theme.spacing.lg, paddingVertical: 10 }}
      />
      <View style={styles.filterWrap}>
        {filters.map((item) => {
          const isActive = filter === item.key
          return (
            <Pressable
              key={item.key}
              onPress={() => setFilter(item.key)}
              accessibilityRole="button"
              style={({ pressed }) => [
                styles.filterChip,
                {
                  borderColor: isActive ? theme.colors.accent : theme.colors.border,
                  backgroundColor: isActive ? theme.colors.accentSoft : theme.colors.surfaceElevated,
                },
                pressed && { opacity: 0.85 },
              ]}
            >
              <Text
                style={{
                  color: isActive ? theme.colors.accent : theme.colors.muted,
                  fontFamily: theme.fonts.bodyBold,
                  fontSize: 12,
                  letterSpacing: 0.2,
                }}
              >
                {t(item.labelKey as any)}
              </Text>
            </Pressable>
          )
        })}
      </View>

      <View style={{ marginTop: theme.spacing.lg }}>
        <Text style={styles.sectionTitle}>{t('insights.actionQueueTitle')}</Text>
        {insightsQuery.isLoading ? (
          <View style={{ gap: theme.spacing.sm, marginTop: theme.spacing.sm }}>
            {Array.from({ length: 3 }).map((_, index) => (
              <Card key={`loading-${index}`}>
                <ShimmerBlock width="38%" height={10} radius={6} />
                <ShimmerBlock width="26%" height={10} radius={6} style={{ marginTop: theme.spacing.xs }} />
                <ShimmerBlock width="70%" height={14} radius={6} style={{ marginTop: theme.spacing.md }} />
                <ShimmerBlock width="92%" height={10} radius={6} style={{ marginTop: theme.spacing.xs }} />
                <ShimmerBlock width="64%" height={10} radius={6} style={{ marginTop: theme.spacing.xs }} />
                <View style={{ flexDirection: 'row', gap: theme.spacing.sm, marginTop: theme.spacing.md }}>
                  <ShimmerBlock width={74} height={26} radius={13} />
                  <ShimmerBlock width={62} height={26} radius={13} />
                </View>
              </Card>
            ))}
          </View>
        ) : filteredInsights.length === 0 ? (
          <Card style={{ marginTop: theme.spacing.sm }}>
            <Text style={styles.emptyText}>{t('insights.empty')}</Text>
          </Card>
        ) : (
          <FlatList
            data={filteredInsights}
            keyExtractor={(item) => item.id}
            scrollEnabled={false}
            contentContainerStyle={{ marginTop: theme.spacing.sm }}
            renderItem={({ item }) => {
              const typeKey = `insight.type.${item.type}`
              const isMarketAlpha = item.type === 'market_alpha'
              const assetValue = item.asset?.trim() ?? ''
              const isPortfolioAsset = assetValue.toUpperCase() === 'PORTFOLIO'
              const typeLabel = resolveTranslation(typeKey, item.type)
              const headline =
                isPortfolioAsset && item.type === 'action_alert'
                  ? t('insights.portfolioRebalance')
                  : assetValue || typeLabel
              const severityLabel = resolveSeverityLabel(item.severity)
              const timestamp = item.created_at ? formatDateTime(item.created_at) : ''
              return (
                <Card
                  style={{
                    marginBottom: theme.spacing.sm,
                    borderLeftWidth: 3,
                    borderLeftColor: severityRailColor(item.severity),
                  }}
                >
                  <View style={styles.queueHeader}>
                    <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.xs, flex: 1 }}>
                      <Text style={styles.queueType}>{typeLabel}</Text>
                      {timestamp ? <Text style={styles.queueType}>{timestamp}</Text> : null}
                    </View>
                    <Badge label={severityLabel} tone={toneForSeverity(item.severity)} />
                  </View>
                  <Text style={styles.queueHeadline}>{headline}</Text>
                  <Text style={styles.queueBody}>{item.trigger_reason}</Text>
                  {item.suggested_action ? (
                    <Text style={styles.queueBody}>
                      {t('insights.suggestedLabel')}: {item.suggested_action}
                    </Text>
                  ) : null}
                  <View style={styles.queueActions}>
                    <Pressable
                      onPress={() => openQueueItem(item.id, item.type, item.asset_key)}
                      accessibilityRole="button"
                      style={({ pressed }) => [styles.primaryAction, pressed && { opacity: 0.9 }]}
                    >
                      <Text style={styles.primaryActionLabel}>
                        {isMarketAlpha ? t('insights.viewCta') : t('insights.executedCta')}
                      </Text>
                    </Pressable>
                    <Pressable onPress={() => handleDismiss(item.id)} accessibilityRole="button">
                      <Text style={styles.dismissAction}>{t('insights.dismissCta')}</Text>
                    </Pressable>
                  </View>
                </Card>
              )
            }}
          />
        )}
      </View>
    </Screen>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    title: {
      fontSize: 28,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    subtitle: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    lockIconWrap: {
      width: 44,
      height: 44,
      borderRadius: 999,
      backgroundColor: theme.colors.accentSoft,
      alignItems: 'center',
      justifyContent: 'center',
    },
    lockTitle: {
      fontSize: 18,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
    },
    lockBody: {
      color: theme.colors.muted,
      marginTop: theme.spacing.xs,
      fontFamily: theme.fonts.body,
    },
    previewLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
    },
    previewBarWide: {
      height: 10,
      borderRadius: 999,
      backgroundColor: theme.colors.border,
      width: '40%',
    },
    previewBarShort: {
      height: 10,
      borderRadius: 999,
      backgroundColor: theme.colors.border,
      width: 64,
    },
    previewBarHeadline: {
      height: 12,
      borderRadius: 999,
      backgroundColor: theme.colors.border,
      width: '68%',
      marginTop: theme.spacing.sm,
    },
    previewBarBody: {
      height: 10,
      borderRadius: 999,
      backgroundColor: theme.colors.border,
      width: '92%',
      marginTop: theme.spacing.xs,
    },
    previewBarBodyShort: {
      height: 10,
      borderRadius: 999,
      backgroundColor: theme.colors.border,
      width: '62%',
      marginTop: theme.spacing.xs,
    },
    sectionHeader: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    sectionHeaderSimple: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    sectionEyebrow: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
      textTransform: 'uppercase',
      letterSpacing: 0.3,
    },
    sectionTitle: {
      marginTop: theme.spacing.xs,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 18,
    },
    sectionBody: {
      marginTop: theme.spacing.sm,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
    },
    metricRow: {
      flexDirection: 'row',
      gap: theme.spacing.sm,
      marginTop: theme.spacing.md,
      alignItems: 'stretch',
    },
    metricShell: {
      flex: 1,
      minHeight: 88,
      borderRadius: theme.radius.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
      padding: theme.spacing.sm,
      justifyContent: 'space-between',
    },
    metricLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
      lineHeight: 16,
      minHeight: 32,
    },
    metricValue: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 18,
    },
    actionHint: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
    },
    featuredHeader: {
      flexDirection: 'row',
      justifyContent: 'space-between',
      alignItems: 'flex-start',
      gap: theme.spacing.sm,
    },
    featuredMeta: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.sm,
      flex: 1,
    },
    featuredSymbol: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 16,
    },
    featuredName: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      marginTop: 2,
    },
    featuredSummary: {
      marginTop: theme.spacing.sm,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 15,
      lineHeight: 22,
    },
    featuredStats: {
      marginTop: theme.spacing.sm,
      gap: 4,
    },
    featuredStat: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    filterWrap: {
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: theme.spacing.sm,
      marginTop: theme.spacing.md,
    },
    filterChip: {
      paddingVertical: 6,
      paddingHorizontal: 12,
      borderRadius: 999,
      borderWidth: 1,
    },
    emptyText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
    },
    queueHeader: {
      flexDirection: 'row',
      justifyContent: 'space-between',
      alignItems: 'center',
    },
    queueType: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
    },
    queueHeadline: {
      marginTop: theme.spacing.sm,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 16,
    },
    queueBody: {
      color: theme.colors.muted,
      marginTop: theme.spacing.xs,
      fontFamily: theme.fonts.body,
    },
    queueActions: {
      flexDirection: 'row',
      alignItems: 'center',
      marginTop: theme.spacing.md,
    },
    primaryAction: {
      paddingVertical: 6,
      paddingHorizontal: 14,
      borderRadius: 999,
      backgroundColor: theme.colors.accent,
      marginRight: theme.spacing.sm,
    },
    primaryActionLabel: {
      color: theme.colors.surfaceElevated,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 13,
    },
    dismissAction: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 13,
    },
    pressed: {
      opacity: 0.92,
      transform: [{ scale: 0.99 }],
    },
  })
