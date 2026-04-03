import { useQuery } from '@tanstack/react-query'
import Constants from 'expo-constants'
import { useFocusEffect, useRouter } from 'expo-router'
import React, { useCallback, useState } from 'react'
import { ActivityIndicator, FlatList, Pressable, Text, TouchableOpacity, View } from 'react-native'

import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { Tag } from '../../src/components/Tag'
import { useProfile } from '../../src/hooks/useProfile'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchPortfolioSnapshots } from '../../src/services/portfolio'
import { formatDateTime } from '../../src/utils/format'
import { reportStatusLabelKey, reportTierLabelKey } from '../../src/utils/report'

export default function MeScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken, clearSession, userId } = useAuth()
  const profile = useProfile()
  const [devTapCount, setDevTapCount] = useState(0)

  const reports = useQuery({
    queryKey: ['reports', userId],
    queryFn: async () => {
      if (!accessToken) return []
      const resp = await fetchPortfolioSnapshots(accessToken)
      if (resp.error) {
        if (resp.status === 401) {
          await clearSession()
          return []
        }
        throw new Error(resp.error.message)
      }
      return resp.data?.items ?? []
    },
    enabled: !!accessToken,
  })

  useFocusEffect(
    useCallback(() => {
      reports.refetch()
    }, [reports.refetch])
  )

  const handleVersionTap = () => {
    const next = devTapCount + 1
    setDevTapCount(next)
    if (next >= 7) {
      router.push('/(modals)/developer')
      setDevTapCount(0)
    }
  }

  const riskTags = [profile.data?.risk_level, profile.data?.experience, profile.data?.style]
    .filter(Boolean)
    .map((value) => String(value))
  const entitlementStatus = profile.data?.entitlement?.status
  const isPro = ['active', 'grace'].includes(entitlementStatus ?? '')
  const membershipLabel = isPro ? t('me.membershipPro') : t('me.membershipFree')
  const reportItems = reports.data ?? []

  const statusTone = (status: string) => {
    switch (status) {
      case 'ready':
        return 'low'
      case 'processing':
      case 'queued':
        return 'medium'
      case 'failed':
        return 'critical'
      default:
        return 'neutral'
    }
  }

  return (
    <Screen scroll style={{ paddingTop: theme.spacing.md }}>
      <Card style={{ marginTop: 0 }}>
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: theme.spacing.md }}>
          <View
            style={{
              width: 54,
              height: 54,
              borderRadius: 999,
              backgroundColor: theme.colors.accentSoft,
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Text style={{ color: theme.colors.accent, fontFamily: theme.fonts.bodyBold, fontSize: 18 }}>
              {(profile.data?.email ?? t('me.emailMissing')).slice(0, 1)}
            </Text>
          </View>
          <View style={{ flex: 1 }}>
            <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
              {profile.data?.email ?? t('me.emailMissing')}
            </Text>
            <View style={{ marginTop: theme.spacing.xs }}>
              <Badge label={membershipLabel} tone={isPro ? 'low' : 'neutral'} />
            </View>
          </View>
        </View>
        <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: theme.spacing.sm, marginTop: theme.spacing.sm }}>
          {riskTags.map((tag) => (
            <Tag key={tag} label={tag} />
          ))}
        </View>
      </Card>

      <Card style={{ marginTop: theme.spacing.md }}>
        <Text style={{ fontFamily: theme.fonts.bodyBold, color: theme.colors.ink, marginBottom: theme.spacing.sm }}>
          {t('me.reportsTitle')}
        </Text>
        {reports.isLoading ? (
          <View style={{ alignItems: 'center', paddingVertical: theme.spacing.sm }}>
            <ActivityIndicator size="small" color={theme.colors.accent} />
            <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
              {t('me.reportsLoading')}
            </Text>
          </View>
        ) : reports.isError ? (
          <View>
            <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('me.reportsError')}</Text>
            <Button
              title={t('common.retry')}
              variant="ghost"
              style={{ marginTop: theme.spacing.sm }}
              onPress={() => reports.refetch()}
            />
          </View>
        ) : reportItems.length === 0 ? (
          <View>
            <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('me.reportsEmpty')}</Text>
            <Button
              title={t('me.reportsEmptyCta')}
              variant="secondary"
              style={{ marginTop: theme.spacing.sm }}
              onPress={() => router.push('/(modals)/upload')}
            />
          </View>
        ) : (
          <FlatList
            data={reportItems}
            keyExtractor={(item) => item.calculation_id}
            scrollEnabled={false}
            renderItem={({ item, index }) => {
              const route =
                item.report_tier === 'preview'
                  ? '/(modals)/preview'
                  : item.is_active
                    ? '/(modals)/report'
                    : '/(modals)/report-archived'
              const title = t(reportTierLabelKey(item.report_tier) as any)
              const statusLabel = t(reportStatusLabelKey(item.status) as any)
              const divider = index < reportItems.length - 1
              return (
                <Pressable
                  style={{
                    paddingVertical: theme.spacing.sm,
                    borderBottomWidth: divider ? 1 : 0,
                    borderBottomColor: theme.colors.border,
                  }}
                  onPress={() =>
                    router.push({
                      pathname: route,
                      params:
                        item.is_active && item.report_tier === 'paid'
                          ? { id: item.calculation_id, return_to: '/(tabs)/me' }
                          : { id: item.calculation_id },
                    })
                  }
                >
                  <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                    <View style={{ flex: 1, paddingRight: theme.spacing.sm }}>
                      <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>{title}</Text>
                      <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
                        {formatDateTime(item.created_at)}
                      </Text>
                    </View>
                    <View style={{ alignItems: 'flex-end' }}>
                      <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.display, fontSize: 20 }}>
                        {item.health_score ?? '--'}
                      </Text>
                      <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
                        {t('report.healthLabel')}
                      </Text>
                    </View>
                  </View>
                  <View
                    style={{
                      flexDirection: 'row',
                      flexWrap: 'wrap',
                      gap: theme.spacing.sm,
                      marginTop: theme.spacing.sm,
                    }}
                  >
                    <Tag label={item.is_active ? t('me.reportActive') : t('me.reportArchived')} />
                    <Badge label={statusLabel} tone={statusTone(item.status)} />
                  </View>
                </Pressable>
              )
            }}
          />
        )}
      </Card>

      <Card style={{ marginTop: theme.spacing.md }}>
        <Button title={t('me.settingsCta')} variant="secondary" onPress={() => router.push('/(modals)/settings')} />
        <Button
          title={t('me.shareCta')}
          variant="ghost"
          style={{ marginTop: theme.spacing.sm }}
          onPress={() => router.push('/(modals)/share')}
        />
        <Button
          title={t('me.vaultsCta')}
          variant="ghost"
          style={{ marginTop: theme.spacing.sm }}
          onPress={() => router.push('/(modals)/vaults')}
        />
        <Button
          title={t('me.signOutCta')}
          variant="ghost"
          style={{ marginTop: theme.spacing.sm }}
          onPress={clearSession}
        />
      </Card>

      <View style={{ marginTop: theme.spacing.lg, alignItems: 'center' }}>
        <TouchableOpacity onPress={handleVersionTap}>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
            {t('me.version', { version: Constants.expoConfig?.version ?? '0.1.0' })}
          </Text>
        </TouchableOpacity>
      </View>
    </Screen>
  )
}
