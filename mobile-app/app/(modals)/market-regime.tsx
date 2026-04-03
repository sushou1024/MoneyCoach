import { useQuery } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import React, { useMemo } from 'react'
import { Pressable, StyleSheet, Text, View } from 'react-native'

import { AssetLogo } from '../../src/components/AssetLogo'
import { Badge } from '../../src/components/Badge'
import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { ShimmerBlock } from '../../src/components/ShimmerBlock'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchMarketRegime } from '../../src/services/intelligence'
import { toneForSeverity } from '../../src/utils/badges'
import { formatNumber, formatPercent } from '../../src/utils/format'
import {
  describeRegimeDriver,
  formatSignedPercent,
  resolveDriverLabel,
  resolveRegimeLabel,
  resolveRegimeSummary,
  resolveSemanticLabel,
  resolveTrendStrengthLabel,
  toneForRegime,
} from '../../src/utils/intelligence'

export default function MarketRegimeScreen() {
  const theme = useTheme()
  const styles = useMemo(() => createStyles(theme), [theme])
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken, clearSession, userId } = useAuth()

  const query = useQuery({
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
    enabled: !!accessToken,
  })

  const regime = query.data

  const handleOpenAsset = (assetKey: string) => {
    router.push({ pathname: '/(modals)/asset-brief', params: { asset_key: assetKey } })
  }

  return (
    <Screen scroll>
      <View style={styles.header}>
        <View style={{ flex: 1 }}>
          <Text style={styles.title}>{t('intelligence.marketRegime.title')}</Text>
          <Text style={styles.subtitle}>{t('intelligence.marketRegime.subtitle')}</Text>
        </View>
        <Pressable onPress={() => router.back()} accessibilityRole="button">
          <Text style={styles.close}>{t('common.close')}</Text>
        </Pressable>
      </View>

      {query.isLoading ? (
        <View style={styles.stack}>
          <Card>
            <ShimmerBlock width="42%" height={18} radius={8} />
            <ShimmerBlock width="78%" height={12} radius={6} style={{ marginTop: theme.spacing.sm }} />
            <View style={styles.metricsGrid}>
              {Array.from({ length: 4 }).map((_, index) => (
                <View key={`metric-${index}`} style={styles.metricCard}>
                  <ShimmerBlock width="70%" height={10} radius={6} />
                  <ShimmerBlock width="50%" height={18} radius={8} style={{ marginTop: theme.spacing.sm }} />
                </View>
              ))}
            </View>
          </Card>
          <Card>
            {Array.from({ length: 4 }).map((_, index) => (
              <View key={`driver-${index}`} style={index > 0 ? styles.listItem : undefined}>
                <ShimmerBlock width="36%" height={12} radius={6} />
                <ShimmerBlock width="84%" height={10} radius={6} style={{ marginTop: theme.spacing.xs }} />
              </View>
            ))}
          </Card>
        </View>
      ) : !regime ? (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={styles.emptyText}>{t('insights.regimeUnavailable')}</Text>
        </Card>
      ) : (
        <View style={styles.stack}>
          <Card>
            <View style={styles.heroHeader}>
              <View style={{ flex: 1 }}>
                <Text style={styles.heroTitle}>{resolveRegimeLabel(t, regime.regime)}</Text>
                <Text style={styles.heroSummary}>{resolveRegimeSummary(t, regime.regime)}</Text>
              </View>
              <Badge label={resolveTrendStrengthLabel(t, regime.trend_strength)} tone={toneForRegime(regime.regime)} />
            </View>
            <View style={styles.metricsGrid}>
              <MetricCard
                label={t('intelligence.metric.alpha30d')}
                value={formatSignedPercent(regime.metrics.alpha_30d)}
              />
              <MetricCard
                label={t('intelligence.metric.volatility30d')}
                value={formatPercent(regime.metrics.volatility_30d_annualized, 0)}
              />
              <MetricCard
                label={t('intelligence.metric.topPosition')}
                value={formatPercent(regime.metrics.top_asset_pct, 0)}
              />
              <MetricCard
                label={t('intelligence.metric.cashBuffer')}
                value={formatPercent(regime.metrics.cash_pct, 0)}
              />
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.marketRegime.breadthTitle')}</Text>
            <View style={styles.metricsGrid}>
              <MetricCard label={t('intelligence.breadth.up')} value={formatNumber(regime.trend_breadth.up_count, 0)} />
              <MetricCard
                label={t('intelligence.breadth.neutral')}
                value={formatNumber(regime.trend_breadth.neutral_count, 0)}
              />
              <MetricCard
                label={t('intelligence.breadth.down')}
                value={formatNumber(regime.trend_breadth.down_count, 0)}
              />
              <MetricCard
                label={t('intelligence.metric.weightedBreadth')}
                value={formatNumber(regime.trend_breadth.weighted_score, 2)}
              />
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.marketRegime.driversTitle')}</Text>
            <View style={styles.listWrap}>
              {regime.drivers.map((driver) => (
                <View key={driver.id} style={styles.listItem}>
                  <Text style={styles.listTitle}>{resolveDriverLabel(t, driver.kind)}</Text>
                  <Text style={styles.listBody}>{describeRegimeDriver(t, driver)}</Text>
                </View>
              ))}
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.marketRegime.portfolioImpactTitle')}</Text>
            <View style={styles.listWrap}>
              {regime.portfolio_impact.map((item) => (
                <BulletRow key={item.id} label={resolveSemanticLabel(t, item.kind)} />
              ))}
            </View>
          </Card>

          <Card>
            <Text style={styles.sectionTitle}>{t('intelligence.marketRegime.actionsTitle')}</Text>
            <View style={styles.listWrap}>
              {regime.actions.map((item) => (
                <BulletRow key={item.id} label={resolveSemanticLabel(t, item.kind)} />
              ))}
            </View>
          </Card>

          {regime.leaders.length > 0 ? (
            <Card>
              <Text style={styles.sectionTitle}>{t('intelligence.marketRegime.leadersTitle')}</Text>
              <View style={styles.stackTight}>
                {regime.leaders.map((asset) => (
                  <Pressable
                    key={asset.asset_key}
                    onPress={() => handleOpenAsset(asset.asset_key)}
                    accessibilityRole="button"
                    style={({ pressed }) => [styles.assetRow, pressed && styles.pressed]}
                  >
                    <View style={styles.assetMeta}>
                      <AssetLogo uri={asset.logo_url} label={asset.symbol} size={38} />
                      <View style={{ flex: 1 }}>
                        <Text style={styles.assetSymbol}>{asset.symbol}</Text>
                        {asset.name ? <Text style={styles.assetName}>{asset.name}</Text> : null}
                      </View>
                    </View>
                    <Badge label={formatSignedPercent(asset.change_30d)} tone={toneForSeverity('low')} />
                  </Pressable>
                ))}
              </View>
            </Card>
          ) : null}

          {regime.laggards.length > 0 ? (
            <Card>
              <Text style={styles.sectionTitle}>{t('intelligence.marketRegime.laggardsTitle')}</Text>
              <View style={styles.stackTight}>
                {regime.laggards.map((asset) => (
                  <Pressable
                    key={asset.asset_key}
                    onPress={() => handleOpenAsset(asset.asset_key)}
                    accessibilityRole="button"
                    style={({ pressed }) => [styles.assetRow, pressed && styles.pressed]}
                  >
                    <View style={styles.assetMeta}>
                      <AssetLogo uri={asset.logo_url} label={asset.symbol} size={38} />
                      <View style={{ flex: 1 }}>
                        <Text style={styles.assetSymbol}>{asset.symbol}</Text>
                        {asset.name ? <Text style={styles.assetName}>{asset.name}</Text> : null}
                      </View>
                    </View>
                    <Badge label={formatSignedPercent(asset.change_30d)} tone={toneForSeverity('high')} />
                  </Pressable>
                ))}
              </View>
            </Card>
          ) : null}
        </View>
      )}
    </Screen>
  )
}

function MetricCard({ label, value }: { label: string; value: string }) {
  const theme = useTheme()
  const styles = useMemo(() => createStyles(theme), [theme])
  return (
    <View style={styles.metricCard}>
      <Text style={styles.metricLabel}>{label}</Text>
      <Text style={styles.metricValue}>{value}</Text>
    </View>
  )
}

function BulletRow({ label }: { label: string }) {
  const theme = useTheme()
  const styles = useMemo(() => createStyles(theme), [theme])
  return (
    <View style={styles.bulletRow}>
      <View style={styles.bullet} />
      <Text style={styles.listBody}>{label}</Text>
    </View>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    header: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      gap: theme.spacing.md,
    },
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
    close: {
      color: theme.colors.accent,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
      marginTop: 4,
    },
    stack: {
      marginTop: theme.spacing.lg,
      gap: theme.spacing.sm,
    },
    stackTight: {
      gap: theme.spacing.sm,
      marginTop: theme.spacing.sm,
    },
    heroHeader: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      gap: theme.spacing.sm,
    },
    heroTitle: {
      fontSize: 22,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    heroSummary: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
    },
    metricsGrid: {
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: theme.spacing.sm,
      marginTop: theme.spacing.md,
      alignItems: 'stretch',
    },
    metricCard: {
      width: '47%',
      minWidth: 132,
      minHeight: 96,
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
      fontSize: 20,
    },
    sectionTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 18,
    },
    listWrap: {
      marginTop: theme.spacing.sm,
      gap: theme.spacing.sm,
    },
    listItem: {
      paddingTop: theme.spacing.sm,
      borderTopWidth: 1,
      borderTopColor: theme.colors.border,
    },
    listTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
    },
    listBody: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
      flex: 1,
    },
    bulletRow: {
      flexDirection: 'row',
      alignItems: 'flex-start',
      gap: theme.spacing.sm,
    },
    bullet: {
      width: 8,
      height: 8,
      borderRadius: 999,
      marginTop: 7,
      backgroundColor: theme.colors.accent,
    },
    assetRow: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
      borderRadius: theme.radius.md,
      padding: theme.spacing.sm,
    },
    assetMeta: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.sm,
      flex: 1,
    },
    assetSymbol: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 16,
    },
    assetName: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      marginTop: 2,
    },
    pressed: {
      opacity: 0.9,
      transform: [{ scale: 0.99 }],
    },
    emptyText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      lineHeight: 22,
    },
  })
