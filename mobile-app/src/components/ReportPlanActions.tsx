import { Ionicons } from '@expo/vector-icons'
import React, { useMemo } from 'react'
import { Animated, Pressable, StyleSheet, Text, View } from 'react-native'

import { useLocalization } from '../providers/LocalizationProvider'
import { useTheme } from '../providers/ThemeProvider'

type ReportPlanActionsProps = {
  onAutoExecute: () => void
  onViewStrategy: () => void
  onNotify: () => void
  notifyEnabled: boolean
  notifyEnabledAnim: Animated.Value
}

export function ReportPlanActions({
  onAutoExecute,
  onViewStrategy,
  onNotify,
  notifyEnabled,
  notifyEnabledAnim,
}: ReportPlanActionsProps) {
  const theme = useTheme()
  const { t } = useLocalization()
  const styles = useMemo(() => createStyles(theme), [theme])

  return (
    <View style={styles.container}>
      <Pressable
        accessibilityRole="button"
        onPress={onAutoExecute}
        style={({ pressed }) => [styles.autoExecute, pressed && styles.autoExecutePressed]}
      >
        <Text style={styles.autoExecuteText}>{t('report.autoExecute')}</Text>
        <Ionicons name="arrow-forward" size={16} color={theme.colors.ink} style={styles.autoExecuteIcon} />
      </Pressable>
      <View style={styles.chipRow}>
        <Pressable
          accessibilityRole="button"
          onPress={onViewStrategy}
          style={({ pressed }) => [styles.chip, pressed && styles.chipPressed]}
        >
          <Text style={styles.chipText}>{t('report.viewStrategy')}</Text>
        </Pressable>
        <Pressable
          accessibilityRole="button"
          onPress={onNotify}
          disabled={notifyEnabled}
          style={({ pressed }) => [
            styles.chip,
            notifyEnabled ? styles.chipEnabled : null,
            pressed && !notifyEnabled ? styles.chipPressed : null,
          ]}
        >
          <View style={styles.notifyContent}>
            {notifyEnabled ? (
              <Animated.View
                style={{
                  opacity: notifyEnabledAnim,
                  transform: [
                    {
                      scale: notifyEnabledAnim.interpolate({ inputRange: [0, 1], outputRange: [0.92, 1] }),
                    },
                  ],
                }}
              >
                <Ionicons name="checkmark-circle" size={14} color={theme.colors.accent} />
              </Animated.View>
            ) : null}
            <Text style={[styles.chipText, notifyEnabled ? styles.chipTextEnabled : null]}>
              {notifyEnabled ? t('report.notifyEnabledCta') : t('report.notifyCta')}
            </Text>
          </View>
        </Pressable>
      </View>
    </View>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    container: {
      marginTop: theme.spacing.sm,
    },
    autoExecute: {
      flexDirection: 'row',
      alignItems: 'center',
      paddingVertical: 6,
      paddingHorizontal: 2,
    },
    autoExecutePressed: {
      opacity: 0.6,
    },
    autoExecuteText: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 14,
    },
    autoExecuteIcon: {
      marginLeft: 6,
    },
    chipRow: {
      flexDirection: 'row',
      flexWrap: 'wrap',
      alignItems: 'center',
      marginTop: theme.spacing.xs,
      gap: theme.spacing.xs,
    },
    chip: {
      paddingVertical: 6,
      paddingHorizontal: 10,
      borderRadius: theme.radius.sm,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
    },
    chipEnabled: {
      borderColor: theme.colors.accentSoft,
      backgroundColor: theme.colors.accentSoft,
      opacity: 0.9,
    },
    chipPressed: {
      opacity: 0.7,
    },
    chipText: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 13,
    },
    chipTextEnabled: {
      color: theme.colors.accent,
    },
    notifyContent: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: 6,
    },
  })
