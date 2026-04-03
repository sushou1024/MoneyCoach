import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'expo-router'
import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Modal, Platform, Pressable, StyleSheet, Switch, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Input } from '../../src/components/Input'
import { Screen } from '../../src/components/Screen'
import { useProfile } from '../../src/hooks/useProfile'
import { useAuth } from '../../src/providers/AuthProvider'
import { useIAP } from '../../src/providers/IAPProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchBillingPlans } from '../../src/services/billing'
import { deleteAccount, updateProfile } from '../../src/services/profile'
import { useOnboardingStore } from '../../src/stores/onboarding'
import type { UserProfile } from '../../src/types/api'
import type { TranslationKey } from '../../src/utils/i18n'
import { openManageSubscription } from '../../src/utils/subscription'

const currencies = ['USD', 'CNY', 'EUR']
const languages = [
  { code: 'en', label: 'English' },
  { code: 'zh-CN', label: '简体中文' },
  { code: 'zh-TW', label: '繁體中文' },
  { code: 'ja', label: '日本語' },
  { code: 'ko', label: '한국어' },
]
const notificationKeys = ['portfolio_alerts', 'market_alpha', 'action_alerts'] as const

const marketLabelKeyByValue = {
  Crypto: 'onboarding.sc02.option.crypto',
  Stocks: 'onboarding.sc02.option.stocks',
  Forex: 'onboarding.sc02.option.forex',
} as const satisfies Record<string, TranslationKey>

const experienceLabelKeyByValue = {
  Beginner: 'onboarding.sc03.option.beginner',
  Intermediate: 'onboarding.sc03.option.intermediate',
  Expert: 'onboarding.sc03.option.expert',
} as const satisfies Record<string, TranslationKey>

const styleLabelKeyByValue = {
  Scalping: 'onboarding.sc04.option.scalping',
  'Day Trading': 'onboarding.sc04.option.daytrading',
  'Swing Trading': 'onboarding.sc04.option.swingtrading',
  'Long-Term': 'onboarding.sc04.option.long-term',
} as const satisfies Record<string, TranslationKey>

const painPointLabelKeyByValue = {
  Bagholder: 'onboarding.sc05.option.bagholder',
  FOMO: 'onboarding.sc05.option.fomo',
  'Messy Portfolio': 'onboarding.sc05.option.messyportfolio',
  'Seeking Stable Yield': 'onboarding.sc05.option.seekingstableyield',
} as const satisfies Record<string, TranslationKey>

const riskLabelKeyByValue = {
  'Yield Seeker': 'onboarding.sc06.option.yieldseeker',
  Speculator: 'onboarding.sc06.option.speculator',
} as const satisfies Record<string, TranslationKey>

type TranslateFn = (key: TranslationKey, params?: Record<string, string | number>) => string

const localizeValues = (values: string[], map: Record<string, TranslationKey>, t: TranslateFn) =>
  values.map((value) => {
    const key = map[value]
    return key ? t(key) : value
  })

const localizeSingleValue = (value: string | undefined, map: Record<string, TranslationKey>, t: TranslateFn) => {
  if (!value) return ''
  const key = map[value]
  return key ? t(key) : value
}

const summarizeValues = (values: string[], missingLabel: string, maxItems = 2) => {
  if (values.length === 0) return missingLabel
  const visible = values.slice(0, maxItems)
  const remaining = values.length - visible.length
  if (remaining > 0) {
    return `${visible.join(', ')} +${remaining}`
  }
  return visible.join(', ')
}

function withAlpha(hex: string, alpha: number) {
  const normalized = hex.replace('#', '')
  if (normalized.length !== 6) {
    return `rgba(0,0,0,${alpha})`
  }
  const r = parseInt(normalized.slice(0, 2), 16)
  const g = parseInt(normalized.slice(2, 4), 16)
  const b = parseInt(normalized.slice(4, 6), 16)
  return `rgba(${r},${g},${b},${alpha})`
}

export default function SettingsScreen() {
  const theme = useTheme()
  const router = useRouter()
  const queryClient = useQueryClient()
  const { t, setLocale } = useLocalization()
  const { accessToken, userId, clearSession } = useAuth()
  const { supported: iapSupported, ready: iapReady, restorePurchases } = useIAP()
  const profile = useProfile()
  const startRetake = useOnboardingStore((state) => state.startRetake)
  const styles = makeStyles(theme)

  const [baseCurrency, setBaseCurrency] = useState('USD')
  const [language, setLanguage] = useState('en')
  const [timezone, setTimezone] = useState('')
  const [prefs, setPrefs] = useState<Record<string, boolean>>({})
  const [restoring, setRestoring] = useState(false)
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [deleteConfirmText, setDeleteConfirmText] = useState('')
  const [deletingAccount, setDeletingAccount] = useState(false)

  const plans = useQuery({
    queryKey: ['billing-plans', userId],
    queryFn: async () => {
      if (!accessToken) return []
      const resp = await fetchBillingPlans(accessToken)
      if (resp.error) throw new Error(resp.error.message)
      return resp.data?.plans ?? []
    },
    enabled: !!accessToken,
  })

  useEffect(() => {
    if (!profile.data) return
    setBaseCurrency(profile.data.base_currency || 'USD')
    setLanguage(profile.data.language || 'en')
    setTimezone(profile.data.timezone || '')
    setPrefs(profile.data.notification_prefs || {})
  }, [profile.data])

  const togglePref = (key: string) => {
    setPrefs((prev) => ({ ...prev, [key]: !prev[key] }))
  }

  const handleSave = async () => {
    if (!accessToken) return
    const updates = {
      base_currency: baseCurrency,
      language,
      timezone,
      notification_prefs: prefs,
    }
    const resp = await updateProfile(accessToken, updates)
    if (resp.error) {
      Alert.alert(t('settings.saveFailTitle'), resp.error.message)
      return
    }
    queryClient.setQueryData<UserProfile | null>(['profile', userId], (prev) => (prev ? { ...prev, ...updates } : prev))
    queryClient.invalidateQueries({ queryKey: ['profile', userId] })
    await setLocale(language as any)
    Alert.alert(t('settings.saveSuccessTitle'), t('settings.saveSuccessBody'))
    router.back()
  }

  const entitlement = profile.data?.entitlement
  const entitlementStatus = entitlement?.status ?? 'expired'
  const entitlementPlan = plans.data?.find((plan) => plan.plan_id === entitlement?.plan_id)
  const providerLabel = entitlement?.provider
    ? entitlement.provider.toUpperCase()
    : t('settings.membershipProviderUnknown')
  const planLabel =
    entitlementPlan?.name ?? (entitlement?.plan_id ? entitlement.plan_id : t('settings.membershipFreePlan'))
  const coverageTone = entitlementStatus === 'active' ? 'active' : entitlementStatus === 'grace' ? 'grace' : 'expired'
  const manageProductId =
    Platform.OS === 'ios'
      ? entitlementPlan?.product_ids?.apple
      : Platform.OS === 'android'
        ? entitlementPlan?.product_ids?.google
        : undefined
  const profileRows = useMemo(() => {
    const snapshot = profile.data
    const missingLabel = t('onboarding.sc07.selectionMissing')
    const marketsValues = localizeValues(snapshot?.markets ?? [], marketLabelKeyByValue, t)
    const focusValues = localizeValues(snapshot?.pain_points ?? [], painPointLabelKeyByValue, t)
    const experienceValue = localizeSingleValue(snapshot?.experience, experienceLabelKeyByValue, t)
    const styleValue = localizeSingleValue(snapshot?.style, styleLabelKeyByValue, t)
    const riskValue = localizeSingleValue(snapshot?.risk_preference, riskLabelKeyByValue, t)

    return [
      {
        key: 'markets',
        label: t('onboarding.sc07.summary.markets'),
        value: summarizeValues(marketsValues, missingLabel, 2),
      },
      {
        key: 'experience',
        label: t('onboarding.sc07.summary.experience'),
        value: experienceValue || missingLabel,
      },
      {
        key: 'style',
        label: t('onboarding.sc07.summary.style'),
        value: styleValue || missingLabel,
      },
      {
        key: 'focus',
        label: t('onboarding.sc07.summary.focus'),
        value: summarizeValues(focusValues, missingLabel, 2),
      },
      {
        key: 'risk',
        label: t('onboarding.sc07.summary.risk'),
        value: riskValue || missingLabel,
      },
    ]
  }, [profile.data, t])

  const handleManageSubscription = async () => {
    try {
      await openManageSubscription(manageProductId)
    } catch (err) {
      Alert.alert(t('common.manageSubscription'), err instanceof Error ? err.message : t('common.restoreFailBody'))
    }
  }

  const handleRestorePurchases = async () => {
    if (!iapSupported || !iapReady) {
      Alert.alert(t('paywall.storeUnavailableTitle'), t('paywall.storeUnavailableBody'))
      return
    }
    setRestoring(true)
    try {
      const restored = await restorePurchases()
      const refreshed = await profile.refetch()
      const status = refreshed.data?.entitlement?.status
      if (restored || status === 'active' || status === 'grace') {
        Alert.alert(t('common.restoreSuccessTitle'), t('common.restoreSuccessBody'))
        return
      }
      Alert.alert(t('common.restoreNoneTitle'), t('common.restoreNoneBody'))
    } catch (err) {
      Alert.alert(t('common.restoreFailTitle'), err instanceof Error ? err.message : t('common.restoreFailBody'))
    } finally {
      setRestoring(false)
    }
  }

  const handleRetakeQuiz = () => {
    const snapshot = profile.data
    startRetake({
      markets: snapshot?.markets ?? [],
      experience: snapshot?.experience ?? '',
      style: snapshot?.style ?? '',
      painPoints: snapshot?.pain_points ?? [],
      riskPreference: snapshot?.risk_preference ?? '',
    })
    router.push('/(auth)/sc02')
  }

  const openDeleteModal = () => {
    setDeleteConfirmText('')
    setDeleteModalVisible(true)
  }

  const closeDeleteModal = () => {
    if (deletingAccount) return
    setDeleteModalVisible(false)
    setDeleteConfirmText('')
  }

  const handleDeleteAccount = async () => {
    if (!accessToken) return
    const confirmText = deleteConfirmText.trim()
    if (confirmText !== 'DELETE') {
      Alert.alert(t('settings.deleteAccountFailTitle'), t('settings.deleteAccountConfirmMismatch'))
      return
    }
    setDeletingAccount(true)
    try {
      const resp = await deleteAccount(accessToken, confirmText)
      if (resp.error) {
        Alert.alert(t('settings.deleteAccountFailTitle'), resp.error.message)
        return
      }
      setDeleteModalVisible(false)
      setDeleteConfirmText('')
      Alert.alert(t('settings.deleteAccountSuccessTitle'), t('settings.deleteAccountSuccessBody'))
      await clearSession()
      router.replace('/(auth)/sc07')
    } finally {
      setDeletingAccount(false)
    }
  }

  const deleteMatches = deleteConfirmText.trim() === 'DELETE'

  const SectionHeader = ({ title, description }: { title: string; description?: string }) => (
    <View style={styles.sectionHeader}>
      <Text style={styles.sectionTitle}>{title}</Text>
      {description ? <Text style={styles.sectionDescription}>{description}</Text> : null}
    </View>
  )

  const SettingLabel = ({ label, hint }: { label: string; hint?: string }) => (
    <View style={styles.settingLabel}>
      <Text style={styles.settingLabelText}>{label}</Text>
      {hint ? <Text style={styles.settingHint}>{hint}</Text> : null}
    </View>
  )

  const SelectionIndicator = ({ selected }: { selected: boolean }) => (
    <View
      style={[
        styles.selectionOuter,
        {
          borderColor: selected ? theme.colors.accent : theme.colors.border,
          backgroundColor: selected ? withAlpha(theme.colors.accent, 0.1) : theme.colors.surface,
        },
      ]}
    >
      {selected ? <View style={styles.selectionInner} /> : null}
    </View>
  )

  const SegmentedControl = ({
    options,
    value,
    onChange,
  }: {
    options: { value: string; label: string }[]
    value: string
    onChange: (next: string) => void
  }) => (
    <View style={styles.segmented}>
      {options.map((option) => {
        const selected = option.value === value
        return (
          <Pressable
            key={option.value}
            onPress={() => onChange(option.value)}
            style={({ pressed }) => [
              styles.segmentedOption,
              selected && styles.segmentedOptionSelected,
              pressed && styles.segmentedOptionPressed,
            ]}
          >
            <Text style={[styles.segmentedText, selected && styles.segmentedTextSelected]}>{option.label}</Text>
          </Pressable>
        )
      })}
    </View>
  )

  return (
    <Screen scroll decorativeBackground={false}>
      <View style={styles.header}>
        <Text style={styles.title}>{t('settings.title')}</Text>
      </View>

      <Card style={[styles.sectionCard, styles.firstSectionCard]}>
        <SectionHeader title={t('settings.contextTitle')} />

        <View style={styles.block}>
          <SettingLabel label={t('settings.baseCurrency')} />
          <SegmentedControl
            options={currencies.map((code) => ({ value: code, label: code }))}
            value={baseCurrency}
            onChange={setBaseCurrency}
          />
        </View>

        <View style={styles.block}>
          <SettingLabel label={t('settings.language')} />
          <View style={styles.list}>
            {languages.map((lang, index) => {
              const selected = language === lang.code
              const isLast = index === languages.length - 1
              return (
                <Pressable
                  key={lang.code}
                  onPress={() => setLanguage(lang.code)}
                  style={({ pressed }) => [
                    styles.listRow,
                    !isLast && styles.listRowDivider,
                    pressed && styles.listRowPressed,
                  ]}
                >
                  <Text style={[styles.listRowLabel, selected && styles.listRowLabelSelected]}>{lang.label}</Text>
                  <SelectionIndicator selected={selected} />
                </Pressable>
              )
            })}
          </View>
        </View>
      </Card>

      <Card style={styles.sectionCard}>
        <SectionHeader title={t('settings.signalHygieneTitle')} />
        <View style={styles.list}>
          {notificationKeys.map((key, index) => {
            const isLast = index === notificationKeys.length - 1
            return (
              <View key={key} style={[styles.toggleRow, !isLast && styles.listRowDivider]}>
                <Text style={styles.toggleLabel}>{t(`settings.notif.${key}` as any)}</Text>
                <Switch value={Boolean(prefs[key])} onValueChange={() => togglePref(key)} />
              </View>
            )
          })}
        </View>
      </Card>

      <Card style={styles.sectionCard}>
        <SectionHeader title={t('settings.membershipTitle')} />
        <View style={styles.membershipGrid}>
          <View style={styles.membershipCell}>
            <Text style={styles.membershipLabel}>{t('settings.membershipStatus')}</Text>
            <Text style={styles.membershipValue}>{t(`settings.coverage.${coverageTone}` as any)}</Text>
          </View>
          <View style={styles.membershipCell}>
            <Text style={styles.membershipLabel}>{t('settings.membershipPlan')}</Text>
            <Text style={styles.membershipValue}>{planLabel}</Text>
          </View>
          <View style={styles.membershipCell}>
            <Text style={styles.membershipLabel}>{t('settings.membershipProvider')}</Text>
            <Text style={styles.membershipValue}>{providerLabel}</Text>
          </View>
        </View>

        {(Platform.OS === 'ios' || Platform.OS === 'android') && (
          <View style={styles.membershipActions}>
            <Button title={t('common.manageSubscription')} variant="secondary" onPress={handleManageSubscription} />
            <Button
              title={t('common.restorePurchases')}
              variant="ghost"
              style={{ marginTop: theme.spacing.sm }}
              onPress={handleRestorePurchases}
              loading={restoring}
              disabled={restoring}
            />
          </View>
        )}
      </Card>

      <Card style={styles.sectionCard}>
        <SectionHeader title={t('settings.riskProfileTitle')} description={t('settings.riskProfileHint')} />
        <View style={styles.profileStrip}>
          {profileRows.map((row, index) => {
            const isLast = index === profileRows.length - 1
            return (
              <View key={row.key} style={[styles.profileRow, !isLast && styles.profileRowDivider]}>
                <Text style={styles.profileLabel}>{row.label}</Text>
                <Text style={styles.profileValue} numberOfLines={2}>
                  {row.value}
                </Text>
              </View>
            )
          })}
        </View>
        <Button
          title={t('settings.retakeQuiz')}
          variant="secondary"
          style={styles.riskProfileAction}
          onPress={handleRetakeQuiz}
        />
      </Card>

      <Card style={styles.sectionCard}>
        <SectionHeader title={t('settings.accountDataTitle')} description={t('settings.accountDataHint')} />
        <View style={styles.accountDeletePanel}>
          <Text style={styles.accountDeleteHint}>{t('settings.deleteAccountDescription')}</Text>
          <Button
            title={t('settings.deleteAccountCta')}
            variant="secondary"
            onPress={openDeleteModal}
            style={styles.deleteTriggerButton}
            textStyle={styles.deleteTriggerText}
          />
        </View>
      </Card>

      <View style={styles.footer}>
        <Button title={t('common.save')} onPress={handleSave} />
        <Button
          title={t('common.close')}
          variant="ghost"
          style={{ marginTop: theme.spacing.sm }}
          onPress={() => router.back()}
        />
      </View>

      <Modal visible={deleteModalVisible} transparent animationType="fade" onRequestClose={closeDeleteModal}>
        <View style={styles.modalOverlay}>
          <View style={styles.modalCard}>
            <Text style={styles.modalTitle}>{t('settings.deleteAccountModalTitle')}</Text>
            <Text style={styles.modalBody}>{t('settings.deleteAccountModalBody')}</Text>
            <Text style={styles.deleteInputHint}>{t('settings.deleteAccountConfirmLabel')}</Text>
            <Input
              value={deleteConfirmText}
              onChangeText={setDeleteConfirmText}
              autoCapitalize="characters"
              autoCorrect={false}
              autoComplete="off"
              placeholder={t('settings.deleteAccountConfirmPlaceholder')}
              editable={!deletingAccount}
            />
            <View style={styles.deleteActions}>
              <Button
                title={t('common.cancel')}
                variant="ghost"
                onPress={closeDeleteModal}
                disabled={deletingAccount}
                style={styles.deleteActionButton}
              />
              <Button
                title={t('settings.deleteAccountAction')}
                variant="secondary"
                onPress={handleDeleteAccount}
                disabled={!deleteMatches || deletingAccount}
                loading={deletingAccount}
                style={styles.deleteActionButton}
                textStyle={styles.deleteActionText}
              />
            </View>
          </View>
        </View>
      </Modal>
    </Screen>
  )
}

const makeStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    header: {
      marginBottom: theme.spacing.md,
    },
    title: {
      fontSize: 26,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      letterSpacing: 0.2,
    },
    sectionCard: {
      marginTop: theme.spacing.lg,
    },
    firstSectionCard: {
      marginTop: theme.spacing.sm,
    },
    sectionHeader: {
      marginBottom: theme.spacing.sm,
      gap: 4,
    },
    sectionTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 16,
      letterSpacing: 0.2,
    },
    sectionDescription: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
      lineHeight: 18,
    },
    block: {
      marginTop: theme.spacing.sm,
      gap: theme.spacing.sm,
    },
    settingLabel: {
      gap: 2,
    },
    settingLabelText: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
    },
    settingHint: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
      lineHeight: 16,
    },
    segmented: {
      flexDirection: 'row',
      gap: theme.spacing.xs,
      padding: 6,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
    },
    segmentedOption: {
      flex: 1,
      minHeight: 40,
      borderRadius: theme.radius.md,
      alignItems: 'center',
      justifyContent: 'center',
      borderWidth: 1,
      borderColor: 'transparent',
      paddingHorizontal: 10,
    },
    segmentedOptionSelected: {
      borderColor: withAlpha(theme.colors.accent, 0.4),
      backgroundColor: theme.colors.accentSoft,
    },
    segmentedOptionPressed: {
      opacity: 0.92,
    },
    segmentedText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 13,
      letterSpacing: 0.2,
    },
    segmentedTextSelected: {
      color: theme.colors.accent,
    },
    list: {
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
      overflow: 'hidden',
    },
    listRow: {
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.sm + 2,
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    listRowDivider: {
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.border,
    },
    listRowPressed: {
      backgroundColor: withAlpha(theme.colors.accent, 0.06),
    },
    listRowLabel: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.body,
      fontSize: 14,
    },
    listRowLabelSelected: {
      color: theme.colors.accent,
      fontFamily: theme.fonts.bodyBold,
    },
    selectionOuter: {
      width: 22,
      height: 22,
      borderRadius: 999,
      borderWidth: 1.5,
      alignItems: 'center',
      justifyContent: 'center',
    },
    selectionInner: {
      width: 10,
      height: 10,
      borderRadius: 999,
      backgroundColor: theme.colors.accent,
    },
    toggleRow: {
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.sm + 2,
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.md,
    },
    toggleLabel: {
      flex: 1,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
    },
    membershipGrid: {
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: theme.spacing.sm,
      marginTop: theme.spacing.sm,
    },
    membershipCell: {
      minWidth: 120,
      flexGrow: 1,
      padding: theme.spacing.sm,
      borderRadius: theme.radius.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
      gap: 4,
    },
    membershipLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    membershipValue: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 15,
    },
    membershipActions: {
      marginTop: theme.spacing.md,
    },
    footer: {
      marginTop: theme.spacing.lg,
    },
    profileStrip: {
      marginTop: theme.spacing.xs,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
      overflow: 'hidden',
    },
    profileRow: {
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.sm,
      flexDirection: 'row',
      alignItems: 'flex-start',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    profileRowDivider: {
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.border,
    },
    profileLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
    },
    profileValue: {
      flexShrink: 1,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 13,
      lineHeight: 18,
      textAlign: 'right',
    },
    riskProfileAction: {
      marginTop: theme.spacing.md,
    },
    accountDeletePanel: {
      marginTop: theme.spacing.xs,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: withAlpha(theme.colors.danger, 0.3),
      backgroundColor: withAlpha(theme.colors.danger, 0.06),
      padding: theme.spacing.md,
      gap: theme.spacing.sm,
    },
    accountDeleteHint: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
      lineHeight: 18,
    },
    deleteTriggerButton: {
      borderColor: withAlpha(theme.colors.danger, 0.35),
      backgroundColor: withAlpha(theme.colors.danger, 0.1),
    },
    deleteTriggerText: {
      color: theme.colors.danger,
    },
    modalOverlay: {
      flex: 1,
      backgroundColor: withAlpha(theme.colors.ink, 0.45),
      justifyContent: 'center',
      paddingHorizontal: theme.spacing.lg,
    },
    modalCard: {
      borderRadius: theme.radius.xl,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surface,
      padding: theme.spacing.md,
    },
    modalTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      fontSize: 20,
    },
    modalBody: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 13,
      lineHeight: 18,
    },
    deleteInputHint: {
      marginTop: theme.spacing.md,
      marginBottom: theme.spacing.xs,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 13,
    },
    deleteActions: {
      marginTop: theme.spacing.md,
      flexDirection: 'row',
      gap: theme.spacing.sm,
    },
    deleteActionButton: {
      flex: 1,
    },
    deleteActionText: {
      color: theme.colors.danger,
    },
  })
