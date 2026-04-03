import { Ionicons } from '@expo/vector-icons'
import * as AppleAuthentication from 'expo-apple-authentication'
import * as AuthSession from 'expo-auth-session'
import * as Google from 'expo-auth-session/providers/google'
import Constants from 'expo-constants'
import { useRouter } from 'expo-router'
import * as WebBrowser from 'expo-web-browser'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Alert, Platform, StyleSheet, Text, View } from 'react-native'
import Svg, { Path } from 'react-native-svg'

import { Button } from '../../src/components/Button'
import { Screen } from '../../src/components/Screen'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { oauthLogin } from '../../src/services/auth'
import {
  isGoogleSignInCancelledError,
  isGoogleSignInConfigError,
  signInWithGoogle,
} from '../../src/services/googleSignIn'
import { useOnboardingStore } from '../../src/stores/onboarding'
import { finalizeAuthFlow } from '../../src/utils/authFlow'
import { isExpoGo } from '../../src/utils/expoEnvironment'
import { TranslationKey } from '../../src/utils/i18n'

WebBrowser.maybeCompleteAuthSession()

type TranslateFn = (key: TranslationKey, params?: Record<string, string | number>) => string

type GoogleRuntimeConfig = {
  iosClientId: string
  webClientId: string
  expoClientId: string
  isExpoGoEnv: boolean
  proxyRedirectUri: string
}

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

const localizeValues = (values: string[], map: Record<string, TranslationKey>, t: TranslateFn) =>
  values.map((value) => {
    const key = map[value]
    return key ? t(key) : value
  })

const localizeSingleValue = (value: string, map: Record<string, TranslationKey>, t: TranslateFn) => {
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

const missingGoogleVarsForCurrentPlatform = (options: {
  iosClientId: string
  webClientId: string
  expoClientId: string
  isExpoGoEnv: boolean
  proxyProjectName: string
  useNativeAndroidGoogleSignIn: boolean
}) => {
  const missing: string[] = []
  if (Platform.OS === 'android' && options.useNativeAndroidGoogleSignIn) {
    if (!options.webClientId) {
      missing.push('EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID')
    }
    return missing
  }
  if (Platform.OS === 'ios' && !options.iosClientId) {
    missing.push('EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID')
  }
  if (Platform.OS === 'web' && !options.webClientId) {
    missing.push('EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID')
  }
  if (options.isExpoGoEnv && !options.expoClientId) {
    missing.push('EXPO_PUBLIC_GOOGLE_EXPO_CLIENT_ID')
  }
  if (options.isExpoGoEnv && !options.proxyProjectName) {
    missing.push('EXPO_PROJECT_FULL_NAME')
  }
  return missing
}

function GoogleLogoIcon({ size = 18 }: { size?: number }) {
  return (
    <Svg width={size} height={size} viewBox="0 0 48 48" accessibilityRole="image">
      <Path
        fill="#FFC107"
        d="M43.6 20.5H42V20H24v8h11.3C33.5 32.6 29.2 36 24 36c-6.6 0-12-5.4-12-12s5.4-12 12-12c3.1 0 5.9 1.1 8.1 3.2l5.7-5.7C34.6 6.5 29.6 4 24 4 12.9 4 4 12.9 4 24s8.9 20 20 20 20-8.9 20-20c0-1.3-.1-2.7-.4-3.5z"
      />
      <Path
        fill="#FF3D00"
        d="M6.3 14.7l6.6 4.8C14.9 16.2 19.1 13 24 13c3.1 0 5.9 1.1 8.1 3.2l5.7-5.7C34.6 6.5 29.6 4 24 4 16.3 4 9.6 8.4 6.3 14.7z"
      />
      <Path
        fill="#4CAF50"
        d="M24 44c5.1 0 9.8-2 13.4-5.3l-6.2-5.1C29.1 35.9 26.6 37 24 37c-5.2 0-9.5-3.4-11-8.1l-6.5 5C9.7 39.5 16.3 44 24 44z"
      />
      <Path
        fill="#1976D2"
        d="M43.6 20.5H42V20H24v8h11.3c-1.1 3.2-3.3 5.8-6.1 7.5l6.2 5.1C39.8 36.8 44 30.9 44 24c0-1.3-.1-2.7-.4-3.5z"
      />
    </Svg>
  )
}

function BrowserGoogleMethodButton({
  t,
  styles,
  config,
  onComplete,
}: {
  t: TranslateFn
  styles: ReturnType<typeof createStyles>
  config: GoogleRuntimeConfig
  onComplete: (idToken: string) => Promise<void>
}) {
  const [loading, setLoading] = useState(false)
  const [awaitingExchange, setAwaitingExchange] = useState(false)
  const expectedAuthCodeRef = useRef<string | null>(null)
  const handledIdTokenRef = useRef<string | null>(null)
  const exchangeTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const [request, response, promptAsync] = Google.useAuthRequest({
    ...(config.isExpoGoEnv ? { clientId: config.expoClientId || undefined } : {}),
    ...(Platform.OS === 'web'
      ? { webClientId: config.webClientId || undefined }
      : { iosClientId: config.iosClientId || undefined }),
    ...(config.isExpoGoEnv && config.proxyRedirectUri ? { redirectUri: config.proxyRedirectUri } : {}),
    responseType: Platform.OS === 'web' ? AuthSession.ResponseType.IdToken : undefined,
    scopes: ['profile', 'email'],
  })

  const clearExchangeTimeout = useCallback(() => {
    if (exchangeTimeoutRef.current) {
      clearTimeout(exchangeTimeoutRef.current)
      exchangeTimeoutRef.current = null
    }
  }, [])

  const failGoogle = useCallback(() => {
    Alert.alert(t('onboarding.sc07.googleFailTitle'), t('onboarding.sc07.googleFailBody'))
  }, [t])

  const submitIdToken = useCallback(
    async (idToken: string) => {
      try {
        await onComplete(idToken)
      } catch {
        failGoogle()
      } finally {
        setLoading(false)
      }
    },
    [failGoogle, onComplete]
  )

  useEffect(() => {
    return () => {
      clearExchangeTimeout()
    }
  }, [clearExchangeTimeout])

  useEffect(() => {
    if (!awaitingExchange || !response || response.type !== 'success') {
      return
    }
    const expectedCode = expectedAuthCodeRef.current
    if (!expectedCode || response.params?.code !== expectedCode) {
      return
    }
    const idToken = response.authentication?.idToken ?? response.params?.id_token
    if (!idToken) {
      return
    }
    if (handledIdTokenRef.current === idToken) {
      return
    }
    handledIdTokenRef.current = idToken
    expectedAuthCodeRef.current = null
    setAwaitingExchange(false)
    clearExchangeTimeout()
    submitIdToken(idToken).catch(() => undefined)
  }, [awaitingExchange, clearExchangeTimeout, response, submitIdToken])

  const handleGoogle = async () => {
    if (!request) {
      failGoogle()
      return
    }

    setLoading(true)
    setAwaitingExchange(false)
    expectedAuthCodeRef.current = null
    handledIdTokenRef.current = null
    clearExchangeTimeout()

    try {
      let result
      if (config.isExpoGoEnv && config.proxyRedirectUri) {
        const authUrl = await request.makeAuthUrlAsync(Google.discovery)
        const startUrl = `${config.proxyRedirectUri}/start?authUrl=${encodeURIComponent(authUrl)}&returnUrl=${encodeURIComponent(config.proxyRedirectUri)}`
        result = await promptAsync({ url: startUrl })
      } else {
        result = await promptAsync()
      }

      if (result.type !== 'success') {
        setLoading(false)
        return
      }

      const idToken = result.authentication?.idToken ?? result.params?.id_token
      if (idToken) {
        handledIdTokenRef.current = idToken
        await submitIdToken(idToken)
        return
      }

      const authCode = result.params?.code
      if (!authCode) {
        setLoading(false)
        failGoogle()
        return
      }

      expectedAuthCodeRef.current = authCode
      setAwaitingExchange(true)
      exchangeTimeoutRef.current = setTimeout(() => {
        if (handledIdTokenRef.current) {
          return
        }
        expectedAuthCodeRef.current = null
        setAwaitingExchange(false)
        setLoading(false)
        failGoogle()
      }, 10000)
    } catch {
      setAwaitingExchange(false)
      setLoading(false)
      failGoogle()
    }
  }

  return (
    <Button
      title={t('onboarding.sc07.google')}
      variant="secondary"
      onPress={handleGoogle}
      loading={loading}
      disabled={!request}
      icon={<GoogleLogoIcon size={18} />}
      style={styles.methodButton}
      contentStyle={styles.methodContent}
      textStyle={styles.methodText}
      labelCentered
    />
  )
}

function NativeAndroidGoogleMethodButton({
  t,
  styles,
  webClientId,
  onComplete,
}: {
  t: TranslateFn
  styles: ReturnType<typeof createStyles>
  webClientId: string
  onComplete: (idToken: string) => Promise<void>
}) {
  const [loading, setLoading] = useState(false)

  const failGoogle = useCallback(() => {
    Alert.alert(t('onboarding.sc07.googleFailTitle'), t('onboarding.sc07.googleFailBody'))
  }, [t])

  const handleGoogle = async () => {
    setLoading(true)
    try {
      const { idToken } = await signInWithGoogle({ webClientId })
      await onComplete(idToken)
    } catch (error) {
      if (isGoogleSignInCancelledError(error)) {
        return
      }
      if (isGoogleSignInConfigError(error)) {
        Alert.alert(
          t('onboarding.sc07.googleMissingTitle'),
          t('onboarding.sc07.googleMissingBody', { vars: error.missingVars.join(', ') })
        )
        return
      }
      failGoogle()
    } finally {
      setLoading(false)
    }
  }

  return (
    <Button
      title={t('onboarding.sc07.google')}
      variant="secondary"
      onPress={handleGoogle}
      loading={loading}
      icon={<GoogleLogoIcon size={18} />}
      style={styles.methodButton}
      contentStyle={styles.methodContent}
      textStyle={styles.methodText}
      labelCentered
    />
  )
}

export default function SC07() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { setSession } = useAuth()
  const onboarding = useOnboardingStore()
  const { markets, experience, style, painPoints, riskPreference } = onboarding
  const [appleLoading, setAppleLoading] = useState(false)
  const [appleAvailable, setAppleAvailable] = useState(false)
  const styles = useMemo(() => createStyles(theme), [theme])

  const isWeb = Platform.OS === 'web'
  const isExpoGoEnv = !isWeb && isExpoGo()
  const useNativeAndroidGoogleSignIn = Platform.OS === 'android' && !isExpoGoEnv

  const iosClientId = process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID ?? ''
  const expoWebClientId = process.env.EXPO_PUBLIC_GOOGLE_EXPO_CLIENT_ID ?? ''
  const webClientId = process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID ?? ''

  const proxyProjectName = useMemo(() => {
    const owner = Constants.expoConfig?.owner
    const slug = Constants.expoConfig?.slug
    if (owner && slug) {
      return `@${owner}/${slug}`
    }
    return Constants.expoConfig?.originalFullName ?? ''
  }, [])

  const proxyRedirectUri = useMemo(() => {
    if (!isExpoGoEnv || !proxyProjectName) {
      return ''
    }
    return `https://auth.expo.io/${proxyProjectName}`
  }, [isExpoGoEnv, proxyProjectName])

  const googleMissingVars = useMemo(
    () =>
      missingGoogleVarsForCurrentPlatform({
        iosClientId,
        webClientId,
        expoClientId: expoWebClientId,
        isExpoGoEnv,
        proxyProjectName,
        useNativeAndroidGoogleSignIn,
      }),
    [expoWebClientId, iosClientId, isExpoGoEnv, proxyProjectName, useNativeAndroidGoogleSignIn, webClientId]
  )

  const googleRuntimeConfig = useMemo<GoogleRuntimeConfig>(
    () => ({
      iosClientId,
      webClientId,
      expoClientId: expoWebClientId,
      isExpoGoEnv,
      proxyRedirectUri,
    }),
    [expoWebClientId, iosClientId, isExpoGoEnv, proxyRedirectUri, webClientId]
  )

  useEffect(() => {
    if (Platform.OS !== 'ios') return
    let active = true
    AppleAuthentication.isAvailableAsync()
      .then((available) => {
        if (active) setAppleAvailable(available)
      })
      .catch(() => {
        if (active) setAppleAvailable(false)
      })
    return () => {
      active = false
    }
  }, [])

  const completeGoogleLogin = useCallback(
    async (idToken: string) => {
      const resp = await oauthLogin('google', idToken)
      if (resp.error ?? !resp.data) {
        Alert.alert(t('onboarding.sc07.googleFailTitle'), resp.error?.message ?? t('onboarding.sc07.googleFailBody'))
        return
      }
      const nextRoute = await finalizeAuthFlow({
        accessToken: resp.data.access_token,
        refreshToken: resp.data.refresh_token,
        userId: resp.data.user_id,
        setSession,
        onboarding,
      })
      router.replace(nextRoute)
    },
    [onboarding, router, setSession, t]
  )

  const profileRows = useMemo(() => {
    const missingLabel = t('onboarding.sc07.selectionMissing')
    const marketsValues = localizeValues(markets, marketLabelKeyByValue, t)
    const focusValues = localizeValues(painPoints, painPointLabelKeyByValue, t)
    const experienceValue = localizeSingleValue(experience, experienceLabelKeyByValue, t)
    const styleValue = localizeSingleValue(style, styleLabelKeyByValue, t)
    const riskValue = localizeSingleValue(riskPreference, riskLabelKeyByValue, t)

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
  }, [experience, markets, painPoints, riskPreference, style, t])

  const handleGoogleMissingConfig = () => {
    Alert.alert(
      t('onboarding.sc07.googleMissingTitle'),
      t('onboarding.sc07.googleMissingBody', { vars: googleMissingVars.join(', ') })
    )
  }

  const handleApple = async () => {
    if (Platform.OS !== 'ios' || !appleAvailable) {
      Alert.alert(t('onboarding.sc07.appleUnavailableTitle'), t('onboarding.sc07.appleUnavailableBody'))
      return
    }
    setAppleLoading(true)
    try {
      const credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
          AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
        ],
      })
      const idToken = credential.identityToken
      if (!idToken) {
        Alert.alert(t('onboarding.sc07.appleFailTitle'), t('onboarding.sc07.appleFailBody'))
        return
      }
      const resp = await oauthLogin('apple', idToken)
      if (resp.error ?? !resp.data) {
        Alert.alert(t('onboarding.sc07.appleFailTitle'), resp.error?.message ?? t('onboarding.sc07.appleFailBody'))
        return
      }
      const nextRoute = await finalizeAuthFlow({
        accessToken: resp.data.access_token,
        refreshToken: resp.data.refresh_token,
        userId: resp.data.user_id,
        setSession,
        onboarding,
      })
      router.replace(nextRoute)
    } catch (err) {
      const appleError = err as { code?: string } | null
      if (appleError?.code === 'ERR_REQUEST_CANCELED') {
        return
      }
      Alert.alert(t('onboarding.sc07.appleFailTitle'), t('onboarding.sc07.appleFailBody'))
    } finally {
      setAppleLoading(false)
    }
  }

  return (
    <Screen decorativeBackground={false} scroll>
      <View style={styles.container}>
        <View style={styles.header}>
          <Text style={styles.title}>{t('onboarding.sc07.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc07.subtitle')}</Text>
        </View>

        <View style={styles.profileBlock}>
          <View style={styles.profileList}>
            {profileRows.map((row) => (
              <View key={row.key} style={styles.profileCard}>
                <Text style={styles.profileLabel}>{row.label}</Text>
                <Text style={styles.profileValue} numberOfLines={2}>
                  {row.value}
                </Text>
              </View>
            ))}
          </View>
        </View>

        <View style={styles.methodsBlock}>
          <Text style={styles.methodsTitle}>{t('onboarding.sc07.methodsTitle')}</Text>
          <View style={styles.methodsStack}>
            {googleMissingVars.length === 0 ? (
              useNativeAndroidGoogleSignIn ? (
                <NativeAndroidGoogleMethodButton
                  t={t}
                  styles={styles}
                  webClientId={webClientId}
                  onComplete={completeGoogleLogin}
                />
              ) : (
                <BrowserGoogleMethodButton
                  t={t}
                  styles={styles}
                  config={googleRuntimeConfig}
                  onComplete={completeGoogleLogin}
                />
              )
            ) : (
              <Button
                title={t('onboarding.sc07.google')}
                variant="secondary"
                onPress={handleGoogleMissingConfig}
                icon={<GoogleLogoIcon size={18} />}
                style={styles.methodButton}
                contentStyle={styles.methodContent}
                textStyle={styles.methodText}
                labelCentered
              />
            )}

            {Platform.OS === 'ios' ? (
              <Button
                title={t('onboarding.sc07.apple')}
                variant="secondary"
                onPress={handleApple}
                loading={appleLoading}
                icon={<Ionicons name="logo-apple" size={18} color={theme.colors.ink} />}
                style={styles.methodButton}
                contentStyle={styles.methodContent}
                textStyle={styles.methodText}
                labelCentered
              />
            ) : null}

            <Button
              title={t('onboarding.sc07.email')}
              variant="ghost"
              onPress={() => router.push('/(auth)/sc07a')}
              icon={<Ionicons name="mail-outline" size={18} color={theme.colors.accent} />}
              style={[styles.methodButton, styles.methodButtonOutline]}
              contentStyle={styles.methodContent}
              textStyle={styles.methodTextAccent}
              labelCentered
            />
          </View>
          <Text style={styles.methodsHelper}>{t('onboarding.sc07.methodsHelper')}</Text>
        </View>
      </View>
    </Screen>
  )
}

function createStyles(theme: ReturnType<typeof useTheme>) {
  return StyleSheet.create({
    container: {
      flexGrow: 1,
      gap: theme.spacing.lg,
    },
    header: {
      gap: theme.spacing.sm,
      paddingRight: theme.spacing.sm,
    },
    title: {
      fontSize: 30,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      letterSpacing: -0.3,
    },
    subtitle: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 15,
      lineHeight: 21,
    },
    profileBlock: {
      gap: 0,
    },
    profileList: {
      gap: theme.spacing.sm,
    },
    profileCard: {
      flexDirection: 'row',
      alignItems: 'center',
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.sm,
      minHeight: 48,
      gap: theme.spacing.sm,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surface,
      shadowColor: '#0b0c10',
      shadowOpacity: 0.04,
      shadowRadius: 10,
      shadowOffset: { width: 0, height: 6 },
      elevation: 1,
    },
    profileLabel: {
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 11,
      letterSpacing: 0.4,
      textTransform: 'uppercase',
      color: theme.colors.muted,
    },
    profileValue: {
      flex: 1,
      textAlign: 'right',
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
      lineHeight: 18,
      color: theme.colors.ink,
    },
    methodsBlock: {
      gap: theme.spacing.sm,
    },
    methodsTitle: {
      fontFamily: theme.fonts.bodyBold,
      fontSize: 15,
      color: theme.colors.ink,
    },
    methodsStack: {
      gap: theme.spacing.sm,
    },
    methodButton: {
      marginTop: 0,
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: theme.radius.lg,
      backgroundColor: theme.colors.surfaceElevated,
      paddingHorizontal: theme.spacing.md,
      paddingVertical: theme.spacing.md,
      minHeight: 56,
      alignItems: 'stretch',
    },
    methodButtonOutline: {
      backgroundColor: theme.colors.surface,
      borderColor: theme.colors.accent,
    },
    methodContent: {
      width: '100%',
      justifyContent: 'center',
      alignItems: 'center',
    },
    methodText: {
      color: theme.colors.ink,
    },
    methodTextAccent: {
      color: theme.colors.accent,
    },
    methodsHelper: {
      fontFamily: theme.fonts.body,
      fontSize: 13,
      color: theme.colors.muted,
    },
  })
}
