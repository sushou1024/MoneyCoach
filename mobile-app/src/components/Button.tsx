import React from 'react'
import { ActivityIndicator, Pressable, StyleProp, StyleSheet, Text, TextStyle, View, ViewStyle } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

type ButtonVariant = 'primary' | 'secondary' | 'ghost'

interface ButtonProps {
  title: string
  onPress?: () => void
  variant?: ButtonVariant
  disabled?: boolean
  loading?: boolean
  style?: StyleProp<ViewStyle>
  icon?: React.ReactNode
  contentStyle?: StyleProp<ViewStyle>
  textStyle?: StyleProp<TextStyle>
  labelCentered?: boolean
}

export function Button({
  title,
  onPress,
  variant = 'primary',
  disabled,
  loading,
  style,
  icon,
  contentStyle,
  textStyle,
  labelCentered = false,
}: ButtonProps) {
  const theme = useTheme()
  const styles = makeStyles(theme, variant, disabled)
  const isDisabled = Boolean(disabled) || Boolean(loading)
  const shouldCenterLabel = Boolean(labelCentered && icon)

  return (
    <Pressable
      style={({ pressed }) => [styles.button, pressed && styles.pressed, style]}
      onPress={onPress}
      disabled={isDisabled}
      accessibilityRole="button"
    >
      {loading ? (
        <ActivityIndicator color={styles.text.color} />
      ) : (
        <View style={[styles.content, contentStyle, shouldCenterLabel && styles.contentCentered]}>
          {icon ? <View style={[styles.icon, shouldCenterLabel && styles.iconAbsolute]}>{icon}</View> : null}
          <Text style={[styles.text, textStyle, shouldCenterLabel && styles.textCentered]}>{title}</Text>
        </View>
      )}
    </Pressable>
  )
}

const makeStyles = (theme: ReturnType<typeof useTheme>, variant: ButtonVariant, disabled?: boolean) =>
  StyleSheet.create({
    button: {
      backgroundColor:
        variant === 'primary'
          ? theme.colors.accent
          : variant === 'secondary'
            ? theme.colors.surfaceElevated
            : 'transparent',
      paddingVertical: 14,
      paddingHorizontal: 20,
      minHeight: 50,
      borderRadius: theme.radius.lg,
      borderWidth: variant === 'primary' ? 0 : 1,
      borderColor: variant === 'primary' ? 'transparent' : theme.colors.border,
      opacity: disabled ? 0.5 : 1,
      alignItems: 'center',
      justifyContent: 'center',
      shadowColor: variant === 'primary' ? theme.colors.accent : 'transparent',
      shadowOpacity: variant === 'primary' ? 0.12 : 0,
      shadowRadius: 14,
      shadowOffset: { width: 0, height: 8 },
      elevation: variant === 'primary' ? 1 : 0,
    },
    pressed: {
      transform: [{ scale: 0.99 }],
      opacity: 0.92,
    },
    text: {
      color:
        variant === 'primary'
          ? theme.colors.surfaceElevated
          : variant === 'ghost'
            ? theme.colors.accent
            : theme.colors.ink,
      fontSize: 15,
      fontFamily: theme.fonts.bodyBold,
      letterSpacing: 0.2,
      textAlign: 'center',
    },
    content: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'center',
      gap: theme.spacing.xs,
      width: '100%',
      flex: 1,
    },
    contentCentered: {
      position: 'relative',
      justifyContent: 'center',
    },
    icon: {
      marginRight: 4,
    },
    iconAbsolute: {
      position: 'absolute',
      left: 0,
    },
    textCentered: {
      width: '100%',
      textAlign: 'center',
    },
  })
