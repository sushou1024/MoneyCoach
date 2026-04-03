import { Ionicons } from '@expo/vector-icons'
import { useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { Alert, Pressable, StyleSheet, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Input } from '../../src/components/Input'
import { Screen } from '../../src/components/Screen'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { loginWithEmail, registerWithEmail, startEmailRegistration } from '../../src/services/auth'
import { useOnboardingStore } from '../../src/stores/onboarding'
import { finalizeAuthFlow } from '../../src/utils/authFlow'

type AuthMode = 'signin' | 'create'

export default function SC07a() {
  const theme = useTheme()
  const router = useRouter()
  const { setSession } = useAuth()
  const onboarding = useOnboardingStore()
  const { t } = useLocalization()
  const styles = useMemo(() => createStyles(theme), [theme])

  const [mode, setMode] = useState<AuthMode>('create')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [verificationCode, setVerificationCode] = useState('')
  const [devCode, setDevCode] = useState<string | null>(null)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const allowDevOTP = __DEV__ || process.env.EXPO_PUBLIC_SHOW_DEV_OTP === '1'

  const normalizedEmail = email.trim().toLowerCase()
  const emailValid = normalizedEmail.includes('@') && normalizedEmail.includes('.')
  const passwordValid = mode === 'signin' ? password.trim().length > 0 : password.length >= 8
  const confirmValid = mode === 'signin' ? true : confirmPassword.length > 0 && password === confirmPassword
  const verificationValid = mode === 'signin' ? true : verificationCode.trim().length > 0
  const canSubmit = emailValid && passwordValid && confirmValid && verificationValid && !submitting
  const showMismatchHint = mode === 'create' && confirmPassword.length > 0 && password !== confirmPassword

  const handleSendCode = async () => {
    if (!emailValid || sendingCode) return
    setSendingCode(true)
    const response = await startEmailRegistration(normalizedEmail)
    setSendingCode(false)
    if (response.error) {
      Alert.alert(t('onboarding.sc07a.sendCodeFailTitle'), response.error.message)
      return
    }
    const returnedCode = response.data?.code
    if (returnedCode && allowDevOTP) {
      setDevCode(returnedCode)
      setVerificationCode(returnedCode)
    } else {
      setDevCode(null)
    }
    Alert.alert(t('onboarding.sc07a.codeSentTitle'), t('onboarding.sc07a.codeSentBody'))
  }

  const handleSubmit = async () => {
    if (!canSubmit) return
    setSubmitting(true)
    const response =
      mode === 'create'
        ? await registerWithEmail(normalizedEmail, password, verificationCode.trim())
        : await loginWithEmail(normalizedEmail, password)
    setSubmitting(false)

    if (response.error ?? !response.data) {
      if (mode === 'create') {
        Alert.alert(
          t('onboarding.sc07a.createFailTitle'),
          response.error?.message ?? t('onboarding.sc07a.createFailBody')
        )
      } else {
        Alert.alert(
          t('onboarding.sc07a.signInFailTitle'),
          response.error?.message ?? t('onboarding.sc07a.signInFailBody')
        )
      }
      return
    }

    const nextRoute = await finalizeAuthFlow({
      accessToken: response.data.access_token,
      refreshToken: response.data.refresh_token,
      userId: response.data.user_id,
      setSession,
      onboarding,
    })
    router.replace(nextRoute)
  }

  return (
    <Screen decorativeBackground={false} scroll>
      <View style={styles.container}>
        <View style={styles.header}>
          <Text style={styles.title}>{t('onboarding.sc07a.title')}</Text>
          <Text style={styles.subtitle}>{t('onboarding.sc07a.subtitle')}</Text>
        </View>

        <View style={styles.modeSelector}>
          <Pressable
            onPress={() => setMode('create')}
            style={[styles.modeButton, mode === 'create' ? styles.modeButtonActive : null]}
            accessibilityRole="button"
          >
            <Text style={[styles.modeText, mode === 'create' ? styles.modeTextActive : null]}>
              {t('onboarding.sc07a.modeCreate')}
            </Text>
          </Pressable>
          <Pressable
            onPress={() => setMode('signin')}
            style={[styles.modeButton, mode === 'signin' ? styles.modeButtonActive : null]}
            accessibilityRole="button"
          >
            <Text style={[styles.modeText, mode === 'signin' ? styles.modeTextActive : null]}>
              {t('onboarding.sc07a.modeSignIn')}
            </Text>
          </Pressable>
        </View>

        <View style={styles.card}>
          <View style={styles.fieldGroup}>
            <View style={styles.labelRow}>
              <Ionicons name="mail-outline" size={16} color={theme.colors.accent} />
              <Text style={styles.fieldLabel}>{t('onboarding.sc07a.emailLabel')}</Text>
            </View>
            <Input
              value={email}
              onChangeText={setEmail}
              placeholder={t('onboarding.sc07a.emailPlaceholder')}
              keyboardType="email-address"
              autoCapitalize="none"
              autoCorrect={false}
              style={styles.input}
            />
            {!emailValid && email.length > 0 ? (
              <Text style={styles.hint}>{t('onboarding.sc07a.emailHint')}</Text>
            ) : null}
          </View>

          <View style={styles.fieldGroup}>
            <View style={styles.passwordLabelRow}>
              <View style={styles.labelRow}>
                <Ionicons name="lock-closed-outline" size={16} color={theme.colors.accent} />
                <Text style={styles.fieldLabel}>{t('onboarding.sc07a.passwordLabel')}</Text>
              </View>
              <Pressable onPress={() => setShowPassword((value) => !value)} accessibilityRole="button">
                <Text style={styles.toggleText}>
                  {showPassword ? t('onboarding.sc07a.hidePassword') : t('onboarding.sc07a.showPassword')}
                </Text>
              </Pressable>
            </View>
            <Input
              value={password}
              onChangeText={setPassword}
              placeholder={t('onboarding.sc07a.passwordPlaceholder')}
              secureTextEntry={!showPassword}
              autoCapitalize="none"
              autoCorrect={false}
              style={styles.input}
            />
            {mode === 'create' ? <Text style={styles.hint}>{t('onboarding.sc07a.passwordHint')}</Text> : null}
          </View>

          {mode === 'create' ? (
            <View style={styles.fieldGroup}>
              <View style={styles.passwordLabelRow}>
                <View style={styles.labelRow}>
                  <Ionicons name="checkmark-circle-outline" size={16} color={theme.colors.accent} />
                  <Text style={styles.fieldLabel}>{t('onboarding.sc07a.confirmPasswordLabel')}</Text>
                </View>
                <Pressable onPress={() => setShowConfirmPassword((value) => !value)} accessibilityRole="button">
                  <Text style={styles.toggleText}>
                    {showConfirmPassword ? t('onboarding.sc07a.hidePassword') : t('onboarding.sc07a.showPassword')}
                  </Text>
                </Pressable>
              </View>
              <Input
                value={confirmPassword}
                onChangeText={setConfirmPassword}
                placeholder={t('onboarding.sc07a.confirmPasswordPlaceholder')}
                secureTextEntry={!showConfirmPassword}
                autoCapitalize="none"
                autoCorrect={false}
                style={styles.input}
              />
              {showMismatchHint ? (
                <Text style={styles.errorHint}>{t('onboarding.sc07a.confirmPasswordHint')}</Text>
              ) : null}
            </View>
          ) : null}

          {mode === 'create' ? (
            <View style={styles.fieldGroup}>
              <View style={styles.passwordLabelRow}>
                <View style={styles.labelRow}>
                  <Ionicons name="shield-checkmark-outline" size={16} color={theme.colors.accent} />
                  <Text style={styles.fieldLabel}>{t('onboarding.sc07a.verificationCodeLabel')}</Text>
                </View>
                <Pressable
                  onPress={handleSendCode}
                  style={[styles.sendCodeButton, !emailValid || sendingCode ? styles.sendCodeButtonDisabled : null]}
                  accessibilityRole="button"
                >
                  <Text style={styles.sendCodeText}>
                    {sendingCode ? t('onboarding.sc07a.sendingCodeCta') : t('onboarding.sc07a.sendCodeCta')}
                  </Text>
                </Pressable>
              </View>
              <Input
                value={verificationCode}
                onChangeText={setVerificationCode}
                placeholder={t('onboarding.sc07a.verificationCodePlaceholder')}
                keyboardType="number-pad"
                style={styles.input}
              />
              <Text style={styles.hint}>{t('onboarding.sc07a.verificationCodeHint')}</Text>
              {devCode ? <Text style={styles.devCode}>{t('onboarding.sc07a.devCode', { code: devCode })}</Text> : null}
            </View>
          ) : null}

          <Button
            title={mode === 'create' ? t('onboarding.sc07a.createCta') : t('onboarding.sc07a.signInCta')}
            onPress={handleSubmit}
            loading={submitting}
            disabled={!canSubmit}
            style={styles.submitButton}
          />
        </View>

        <Pressable onPress={() => router.back()} style={styles.backLink} accessibilityRole="button">
          <Text style={styles.backLinkText}>{t('onboarding.sc07a.backToMethods')}</Text>
        </Pressable>
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
    modeSelector: {
      flexDirection: 'row',
      gap: theme.spacing.sm,
      backgroundColor: theme.colors.surface,
      borderRadius: theme.radius.lg,
      borderWidth: 1,
      borderColor: theme.colors.border,
      padding: theme.spacing.xs,
    },
    modeButton: {
      flex: 1,
      borderRadius: theme.radius.md,
      paddingVertical: theme.spacing.sm,
      alignItems: 'center',
      justifyContent: 'center',
    },
    modeButtonActive: {
      backgroundColor: theme.colors.surfaceElevated,
      borderWidth: 1,
      borderColor: theme.colors.accent,
    },
    modeText: {
      fontFamily: theme.fonts.bodyMedium,
      color: theme.colors.muted,
      fontSize: 14,
    },
    modeTextActive: {
      color: theme.colors.ink,
    },
    card: {
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: theme.radius.lg,
      backgroundColor: theme.colors.surfaceElevated,
      paddingHorizontal: theme.spacing.lg,
      paddingVertical: theme.spacing.lg,
      gap: theme.spacing.md,
      shadowColor: '#0b0c10',
      shadowOpacity: 0.05,
      shadowRadius: 12,
      shadowOffset: { width: 0, height: 8 },
      elevation: 1,
    },
    fieldGroup: {
      gap: theme.spacing.xs,
    },
    labelRow: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.xs,
    },
    passwordLabelRow: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing.sm,
    },
    fieldLabel: {
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
      color: theme.colors.ink,
    },
    input: {
      backgroundColor: theme.colors.surface,
    },
    hint: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    errorHint: {
      color: theme.colors.danger,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    toggleText: {
      color: theme.colors.accent,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 13,
    },
    sendCodeButton: {
      borderWidth: 1,
      borderColor: theme.colors.accent,
      borderRadius: theme.radius.sm,
      paddingHorizontal: theme.spacing.sm,
      paddingVertical: 6,
      backgroundColor: theme.colors.surface,
    },
    sendCodeButtonDisabled: {
      borderColor: theme.colors.border,
      opacity: 0.6,
    },
    sendCodeText: {
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 12,
      color: theme.colors.accent,
    },
    devCode: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: 12,
    },
    submitButton: {
      marginTop: theme.spacing.xs,
      shadowOpacity: 0,
      shadowRadius: 0,
      shadowOffset: { width: 0, height: 0 },
      elevation: 0,
    },
    backLink: {
      marginTop: theme.spacing.xs,
      alignItems: 'center',
      paddingVertical: theme.spacing.sm,
    },
    backLinkText: {
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
      color: theme.colors.accent,
    },
  })
}
