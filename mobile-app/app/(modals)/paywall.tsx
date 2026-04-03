import { Ionicons } from '@expo/vector-icons'
import { useQuery } from '@tanstack/react-query'
import * as Linking from 'expo-linking'
import { useLocalSearchParams, useRouter } from 'expo-router'
import * as WebBrowser from 'expo-web-browser'
import React, { useEffect, useMemo, useState } from 'react'
import { Alert, Platform, Pressable, Text, View } from 'react-native'

import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { useEntitlement } from '../../src/hooks/useEntitlement'
import { useAuth } from '../../src/providers/AuthProvider'
import { useIAP } from '../../src/providers/IAPProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { createStripeCheckoutSession, fetchBillingPlans } from '../../src/services/billing'
import { formatCurrency } from '../../src/utils/format'

const DEFAULT_TERMS_URL = 'https://moneycoach.cc/terms'
const DEFAULT_PRIVACY_URL = 'https://moneycoach.cc/privacy'

export default function PaywallScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken, userId } = useAuth()
  const {
    supported: iapSupported,
    ready: iapReady,
    products: iapProducts,
    loadProducts: loadIapProducts,
    purchaseProduct,
    restorePurchases,
  } = useIAP()
  const entitlement = useEntitlement()
  const params = useLocalSearchParams<{ id?: string; status?: string }>()
  const calculationId = useMemo(() => {
    const value = params.id
    return Array.isArray(value) ? (value[0] ?? '') : (value ?? '')
  }, [params.id])
  const statusParam = useMemo(() => {
    const value = params.status
    return Array.isArray(value) ? (value[0] ?? '') : (value ?? '')
  }, [params.status])
  const pwaBaseUrl = process.env.EXPO_PUBLIC_PWA_URL ?? ''
  const termsUrl = process.env.EXPO_PUBLIC_TERMS_URL ?? DEFAULT_TERMS_URL
  const privacyUrl = process.env.EXPO_PUBLIC_PRIVACY_URL ?? DEFAULT_PRIVACY_URL
  const [purchasingPlan, setPurchasingPlan] = useState<string | null>(null)
  const [restoring, setRestoring] = useState(false)
  const [polling, setPolling] = useState(false)
  const entitlementRefetch = entitlement.refetch
  const isWeb = Platform.OS === 'web'
  const isIOS = Platform.OS === 'ios'
  const isAndroid = Platform.OS === 'android'
  const hasActiveEntitlement = ['active', 'grace'].includes(entitlement.data?.status ?? '')

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

  const buildReturnUrl = (status: 'success' | 'cancel') => {
    const query = calculationId ? `status=${status}&id=${calculationId}` : `status=${status}`
    if (pwaBaseUrl.startsWith('http')) {
      return `${pwaBaseUrl.replace(/\/$/, '')}/paywall?${query}`
    }
    return Linking.createURL('paywall', { queryParams: { status, id: calculationId || undefined } })
  }

  const handleStripePurchase = async (planId: string) => {
    if (!accessToken) return
    setPurchasingPlan(planId)
    try {
      const successUrl = buildReturnUrl('success')
      const cancelUrl = buildReturnUrl('cancel')
      const resp = await createStripeCheckoutSession(accessToken, { planId, successUrl, cancelUrl })
      if (resp.error ?? !resp.data?.checkout_url) {
        Alert.alert(t('paywall.paymentFailTitle'), resp.error?.message ?? t('paywall.paymentFailBody'))
        return
      }
      const checkoutUrl = resp.data.checkout_url
      if (Platform.OS === 'web') {
        await Linking.openURL(checkoutUrl)
      } else {
        await WebBrowser.openAuthSessionAsync(checkoutUrl, successUrl)
      }
    } finally {
      setPurchasingPlan(null)
    }
  }

  const handleIapPurchase = async (planId: string, productId?: string) => {
    if (!accessToken) return
    if (!productId) {
      Alert.alert(t('paywall.comingSoonTitle'), t('paywall.comingSoonBody'))
      return
    }
    if (!iapReady) {
      Alert.alert(t('paywall.storeUnavailableTitle'), t('paywall.storeUnavailableBody'))
      return
    }
    setPurchasingPlan(planId)
    try {
      const result = await purchaseProduct(productId)
      if (!result) return
      await entitlement.refetch()
      if (calculationId) {
        router.replace({ pathname: '/(modals)/processing-paid', params: { calculation_id: calculationId } })
      } else {
        router.back()
      }
    } catch (err) {
      Alert.alert(t('paywall.paymentFailTitle'), err instanceof Error ? err.message : t('paywall.paymentFailBody'))
    } finally {
      setPurchasingPlan(null)
    }
  }

  const handleRestorePurchases = async () => {
    if (!iapSupported) {
      Alert.alert(t('paywall.storeUnavailableTitle'), t('paywall.storeUnavailableBody'))
      return
    }
    if (!iapReady) {
      Alert.alert(t('paywall.storeUnavailableTitle'), t('paywall.storeUnavailableBody'))
      return
    }
    setRestoring(true)
    try {
      const restored = await restorePurchases()
      const refreshed = await entitlement.refetch()
      const status = refreshed.data?.status
      if (restored || status === 'active' || status === 'grace') {
        Alert.alert(t('common.restoreSuccessTitle'), t('common.restoreSuccessBody'))
        if (calculationId) {
          router.replace({ pathname: '/(modals)/processing-paid', params: { calculation_id: calculationId } })
        } else {
          router.back()
        }
        return
      }
      Alert.alert(t('common.restoreNoneTitle'), t('common.restoreNoneBody'))
    } catch (err) {
      Alert.alert(t('common.restoreFailTitle'), err instanceof Error ? err.message : t('common.restoreFailBody'))
    } finally {
      setRestoring(false)
    }
  }

  useEffect(() => {
    if (!iapSupported) return
    const ids =
      plans.data
        ?.map((plan) => (isIOS ? plan.product_ids?.apple : plan.product_ids?.google))
        .filter((id): id is string => !!id) ?? []
    if (ids.length === 0) return
    loadIapProducts(ids).catch(() => null)
  }, [iapSupported, isIOS, loadIapProducts, plans.data])

  useEffect(() => {
    if (!accessToken || statusParam !== 'success') return
    let active = true
    let timeoutId: ReturnType<typeof setTimeout> | null = null
    let attempts = 0
    const poll = async () => {
      if (!active) return
      setPolling(true)
      const result = await entitlementRefetch()
      const status = result.data?.status
      if (status === 'active' || status === 'grace') {
        setPolling(false)
        if (calculationId) {
          router.replace({ pathname: '/(modals)/processing-paid', params: { calculation_id: calculationId } })
        } else {
          router.back()
        }
        return
      }
      attempts += 1
      if (attempts < 45) {
        timeoutId = setTimeout(poll, 2000)
        return
      }
      setPolling(false)
      Alert.alert(t('paywall.paymentPendingTitle'), t('paywall.paymentPendingBody'))
    }
    poll()
    return () => {
      active = false
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
    }
  }, [accessToken, calculationId, entitlementRefetch, router, statusParam, t])

  useEffect(() => {
    if (!accessToken) return
    if (!hasActiveEntitlement) return
    if (calculationId) {
      router.replace({ pathname: '/(modals)/processing-paid', params: { calculation_id: calculationId } })
      return
    }
    router.replace('/(tabs)/insights')
  }, [accessToken, calculationId, hasActiveEntitlement, router])

  const handleClose = () => {
    if (isWeb && typeof window !== 'undefined' && window.history.length > 1 && !statusParam) {
      router.back()
      return
    }
    const fallback = hasActiveEntitlement ? '/(tabs)/insights' : '/(tabs)/assets'
    router.replace(fallback)
  }

  const openLegalUrl = async (url: string) => {
    try {
      await Linking.openURL(url)
    } catch {
      Alert.alert(t('paywall.legalOpenFailTitle'), t('paywall.legalOpenFailBody'))
    }
  }

  return (
    <Screen scroll>
      <Text style={{ fontSize: 26, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
        {t('paywall.title')}
      </Text>
      <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {t('paywall.subtitle')}
      </Text>

      <View style={{ marginTop: theme.spacing.md, gap: theme.spacing.sm }}>
        {['paywall.bullet1', 'paywall.bullet2', 'paywall.bullet3'].map((key) => (
          <View key={key} style={{ flexDirection: 'row', gap: theme.spacing.sm, alignItems: 'center' }}>
            <Ionicons name="checkmark-circle" size={18} color={theme.colors.accent} />
            <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t(key as any)}</Text>
          </View>
        ))}
      </View>

      <View style={{ marginTop: theme.spacing.md }}>
        {(plans.data ?? []).map((plan) => {
          const isAnnual = plan.interval === 'year' || plan.plan_id === 'annual'
          const productId = isIOS ? plan.product_ids?.apple : isAndroid ? plan.product_ids?.google : undefined
          const storePrice = productId ? iapProducts[productId]?.price : undefined
          const priceLabel = storePrice ?? formatCurrency(plan.price, plan.currency)

          return (
            <Card
              key={plan.plan_id}
              style={{
                marginBottom: theme.spacing.sm,
                borderColor: isAnnual ? theme.colors.accent : undefined,
                backgroundColor: isAnnual ? theme.colors.surfaceElevated : undefined,
              }}
            >
              <View
                style={{
                  flexDirection: 'row',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  gap: theme.spacing.sm,
                }}
              >
                <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>{plan.name}</Text>
                {isAnnual ? <Badge label="80% off" tone="medium" /> : null}
              </View>
              <View style={{ flexDirection: 'row', alignItems: 'baseline', marginTop: theme.spacing.xs }}>
                <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.display, fontSize: 22 }}>
                  {priceLabel}
                </Text>
                <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body, marginLeft: 6 }}>
                  / {plan.interval}
                </Text>
              </View>
              <Button
                title={t('paywall.planCta', { name: plan.name })}
                style={{ marginTop: theme.spacing.sm }}
                onPress={() => {
                  if (isWeb) {
                    handleStripePurchase(plan.plan_id).catch((err) => {
                      Alert.alert(
                        t('paywall.paymentFailTitle'),
                        err instanceof Error ? err.message : t('paywall.paymentFailBody')
                      )
                    })
                    return
                  }
                  const productId = isIOS ? plan.product_ids?.apple : plan.product_ids?.google
                  handleIapPurchase(plan.plan_id, productId).catch((err) => {
                    Alert.alert(
                      t('paywall.paymentFailTitle'),
                      err instanceof Error ? err.message : t('paywall.paymentFailBody')
                    )
                  })
                }}
                loading={purchasingPlan === plan.plan_id}
                disabled={!!purchasingPlan}
              />
            </Card>
          )
        })}
      </View>

      <View style={{ marginTop: theme.spacing.md, gap: theme.spacing.xs }}>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body, fontSize: 12, lineHeight: 18 }}>
          {t('paywall.legalDisclosure')}
        </Text>
        <View style={{ flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap' }}>
          <Pressable onPress={() => openLegalUrl(termsUrl)}>
            <Text
              style={{
                color: theme.colors.accent,
                fontFamily: theme.fonts.bodyBold,
                fontSize: 12,
                textDecorationLine: 'underline',
              }}
            >
              {t('paywall.legalTerms')}
            </Text>
          </Pressable>
          <Text style={{ marginHorizontal: 6, color: theme.colors.muted, fontFamily: theme.fonts.body, fontSize: 12 }}>
            {t('paywall.legalLinkSeparator')}
          </Text>
          <Pressable onPress={() => openLegalUrl(privacyUrl)}>
            <Text
              style={{
                color: theme.colors.accent,
                fontFamily: theme.fonts.bodyBold,
                fontSize: 12,
                textDecorationLine: 'underline',
              }}
            >
              {t('paywall.legalPrivacy')}
            </Text>
          </Pressable>
        </View>
      </View>

      {polling ? (
        <Text style={{ marginTop: theme.spacing.md, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {t('paywall.paymentPending')}
        </Text>
      ) : null}

      {(isIOS || isAndroid) && (
        <Button
          title={t('common.restorePurchases')}
          variant="ghost"
          style={{ marginTop: theme.spacing.md }}
          onPress={handleRestorePurchases}
          disabled={restoring || !!purchasingPlan}
          loading={restoring}
        />
      )}

      <Button title={t('common.close')} variant="ghost" style={{ marginTop: theme.spacing.lg }} onPress={handleClose} />
    </Screen>
  )
}
