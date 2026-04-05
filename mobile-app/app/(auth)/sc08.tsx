import { Ionicons } from '@expo/vector-icons'
import { useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { Pressable, StyleSheet, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Screen } from '../../src/components/Screen'
import { useEntitlement } from '../../src/hooks/useEntitlement'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { registerForPush } from '../../src/services/notifications'
import { useResponsiveScale } from '../../src/utils/responsive'

export default function SC08() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken } = useAuth()
  const entitlement = useEntitlement()
  const [loading, setLoading] = useState(false)
  const { compact, font } = useResponsiveScale()
  const titleSize = font(30, 26)
  const subtitleSize = font(15, 13)
  const styles = useMemo(
    () => createStyles(theme, { titleSize, subtitleSize, compact }),
    [theme, titleSize, subtitleSize, compact]
  )

  const navigateForward = () => {
    const status = entitlement.data?.status
    if (status === 'active' || status === 'grace') {
      router.replace('/(tabs)/insights')
    } else {
      router.replace('/(tabs)/assets')
    }
  }

  const handleEnable = async () => {
    setLoading(true)
    try {
      await registerForPush(accessToken ?? '')
    } catch {
      // Permission denied or error - proceed anyway
    } finally {
      setLoading(false)
      navigateForward()
    }
  }

  const handleSkip = () => {
    navigateForward()
  }

  return (
    <Screen decorativeBackground={false}>
      <View style={styles.container}>
        <View style={styles.content}>
          <View style={styles.iconContainer}>
            <Ionicons name="notifications-outline" size={64} color={theme.colors.accent} />
          </View>

          <View style={styles.header}>
            <Text style={styles.title}>{t('onboarding.sc08.title')}</Text>
            <Text style={styles.subtitle}>{t('onboarding.sc08.subtitle')}</Text>
          </View>

          <View style={styles.benefitList}>
            {(['pulse-outline', 'trending-up-outline', 'shield-checkmark-outline'] as const).map(
              (icon, index) => {
                const labels = [
                  'Daily portfolio briefings',
                  'Market movement alerts',
                  'Strategy signals & insights',
                ]
                return (
                  <View key={icon} style={styles.benefitRow}>
                    <Ionicons name={icon} size={20} color={theme.colors.accent} />
                    <Text style={styles.benefitText}>{labels[index]}</Text>
                  </View>
                )
              }
            )}
          </View>
        </View>

        <View style={styles.actions}>
          <Button
            title={t('onboarding.sc08.enable')}
            variant="primary"
            onPress={handleEnable}
            loading={loading}
            style={styles.enableButton}
          />
          <Pressable onPress={handleSkip} style={styles.skipLink} accessibilityRole="button">
            <Text style={styles.skipText}>{t('onboarding.sc08.skip')}</Text>
          </Pressable>
        </View>
      </View>
    </Screen>
  )
}

function createStyles(
  theme: ReturnType<typeof useTheme>,
  sizes: { titleSize: number; subtitleSize: number; compact: boolean }
) {
  return StyleSheet.create({
    container: {
      flex: 1,
      justifyContent: 'space-between',
    },
    content: {
      flex: 1,
      justifyContent: 'center',
      alignItems: 'center',
      gap: theme.spacing.xl,
    },
    iconContainer: {
      width: 120,
      height: 120,
      borderRadius: 60,
      backgroundColor: theme.colors.accentSoft,
      alignItems: 'center',
      justifyContent: 'center',
    },
    header: {
      gap: theme.spacing.sm,
      alignItems: 'center',
      paddingHorizontal: theme.spacing.md,
    },
    title: {
      fontSize: sizes.titleSize,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
      letterSpacing: -0.3,
      textAlign: 'center',
    },
    subtitle: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      fontSize: sizes.subtitleSize,
      lineHeight: Math.round(sizes.subtitleSize * 1.45),
      textAlign: 'center',
    },
    benefitList: {
      gap: theme.spacing.md,
      paddingHorizontal: theme.spacing.lg,
      alignSelf: 'stretch',
    },
    benefitRow: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.sm,
      paddingVertical: theme.spacing.xs,
      paddingHorizontal: theme.spacing.md,
      borderRadius: theme.radius.lg,
      backgroundColor: theme.colors.surface,
      borderWidth: 1,
      borderColor: theme.colors.border,
    },
    benefitText: {
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
      color: theme.colors.ink,
    },
    actions: {
      gap: theme.spacing.sm,
      paddingTop: theme.spacing.lg,
    },
    enableButton: {
      borderRadius: 999,
      minHeight: sizes.compact ? 52 : 56,
      shadowOpacity: 0.16,
      shadowRadius: 18,
      shadowOffset: { width: 0, height: 10 },
      elevation: 2,
    },
    skipLink: {
      alignItems: 'center',
      paddingVertical: theme.spacing.sm,
    },
    skipText: {
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 14,
      color: theme.colors.muted,
    },
  })
}
