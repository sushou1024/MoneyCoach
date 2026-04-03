import React, { useEffect, useMemo, useRef, useState } from 'react'
import {
  Animated,
  Easing,
  PanResponder,
  ScrollView,
  StyleProp,
  Text,
  View,
  ViewStyle,
  useWindowDimensions,
} from 'react-native'

import { Button } from './Button'
import { ProgressDots } from './ProgressDots'
import { Screen } from './Screen'
import { useTheme } from '../providers/ThemeProvider'
import { useOnboardingStore } from '../stores/onboarding'
import {
  ONBOARDING_SWIPE_ACTIVATION_DX,
  applyOnboardingSwipeResistance,
  resolveOnboardingSwipeAction,
} from '../utils/onboardingSwipe'

type OnboardingStepProps = {
  children: React.ReactNode
  nextLabel: string
  onNext: () => void
  onBack?: () => void
  nextDisabled?: boolean
  progressTotal?: number
  progressIndex?: number
  footerNote?: string
  animateIn?: boolean
  contentScroll?: boolean
  swipeEnabled?: boolean
  nextButtonVariant?: 'primary' | 'secondary' | 'ghost'
  nextButtonStyle?: StyleProp<ViewStyle>
}

export function OnboardingStep({
  children,
  nextLabel,
  onNext,
  onBack,
  nextDisabled,
  progressTotal,
  progressIndex,
  footerNote,
  animateIn = true,
  contentScroll = false,
  swipeEnabled = false,
  nextButtonVariant,
  nextButtonStyle,
}: OnboardingStepProps) {
  const theme = useTheme()
  const { width } = useWindowDimensions()
  const { transitionDirection, setTransitionDirection } = useOnboardingStore((state) => ({
    transitionDirection: state.transitionDirection,
    setTransitionDirection: state.setTransitionDirection,
  }))
  const safeWidth = Math.max(width, 1)
  const translateX = useRef(new Animated.Value(animateIn ? safeWidth : 0)).current
  const [transitioning, setTransitioning] = useState(false)
  const canSwipeBack = swipeEnabled && !!onBack
  const canSwipeForward = swipeEnabled && !nextDisabled

  const animateExit = (direction: 'forward' | 'backward', callback: () => void) => {
    if (transitioning) return
    const target = direction === 'forward' ? -safeWidth : safeWidth
    setTransitionDirection(direction)
    setTransitioning(true)
    Animated.timing(translateX, {
      toValue: target,
      duration: 220,
      easing: Easing.in(Easing.cubic),
      useNativeDriver: true,
    }).start(({ finished }) => {
      if (!finished) {
        setTransitioning(false)
        return
      }
      callback()
      requestAnimationFrame(() => {
        translateX.setValue(0)
        setTransitioning(false)
      })
    })
  }

  const snapBack = () => {
    Animated.timing(translateX, {
      toValue: 0,
      duration: 180,
      easing: Easing.out(Easing.cubic),
      useNativeDriver: true,
    }).start()
  }

  useEffect(() => {
    if (!animateIn) {
      translateX.setValue(0)
      return
    }
    translateX.setValue(transitionDirection === 'backward' ? -safeWidth : safeWidth)
    Animated.timing(translateX, {
      toValue: 0,
      duration: 260,
      easing: Easing.out(Easing.cubic),
      useNativeDriver: true,
    }).start()
  }, [animateIn, safeWidth, transitionDirection, translateX])

  const handleNext = () => {
    if (transitioning || nextDisabled) return
    animateExit('forward', onNext)
  }

  const panResponder = useMemo(
    () =>
      PanResponder.create({
        onStartShouldSetPanResponder: () => false,
        onStartShouldSetPanResponderCapture: () => false,
        onMoveShouldSetPanResponder: (_, gestureState) => {
          if (!swipeEnabled || transitioning) {
            return false
          }
          const { dx, dy } = gestureState
          return Math.abs(dx) > ONBOARDING_SWIPE_ACTIVATION_DX && Math.abs(dx) > Math.abs(dy) * 1.2
        },
        onMoveShouldSetPanResponderCapture: (_, gestureState) => {
          if (!swipeEnabled || transitioning) {
            return false
          }
          const { dx, dy } = gestureState
          return Math.abs(dx) > ONBOARDING_SWIPE_ACTIVATION_DX && Math.abs(dx) > Math.abs(dy) * 1.2
        },
        onPanResponderMove: (_, gestureState) => {
          const offset = applyOnboardingSwipeResistance({
            dx: gestureState.dx,
            canGoBack: canSwipeBack,
            canGoForward: canSwipeForward,
          })
          translateX.setValue(offset)
        },
        onPanResponderRelease: (_, gestureState) => {
          const action = resolveOnboardingSwipeAction({
            dx: gestureState.dx,
            vx: gestureState.vx,
            canGoBack: canSwipeBack,
            canGoForward: canSwipeForward,
          })

          if (action === 'next') {
            animateExit('forward', onNext)
            return
          }
          if (action === 'back' && onBack) {
            animateExit('backward', onBack)
            return
          }

          snapBack()
        },
        onPanResponderTerminate: snapBack,
        onPanResponderTerminationRequest: () => true,
      }),
    [animateExit, canSwipeBack, canSwipeForward, onBack, onNext, swipeEnabled, transitioning, translateX]
  )

  const contentNode = contentScroll ? (
    <ScrollView
      contentContainerStyle={{ flexGrow: 1, paddingBottom: theme.spacing.md }}
      showsVerticalScrollIndicator={false}
    >
      {children}
    </ScrollView>
  ) : (
    children
  )

  return (
    <Screen>
      <View style={{ flex: 1 }}>
        <Animated.View
          style={{ flex: 1, transform: [{ translateX }] }}
          {...(swipeEnabled ? panResponder.panHandlers : {})}
        >
          {contentNode}
        </Animated.View>
        <View style={{ marginTop: theme.spacing.md }}>
          {footerNote ? (
            <Text style={{ color: theme.colors.muted, marginBottom: theme.spacing.md, fontFamily: theme.fonts.body }}>
              {footerNote}
            </Text>
          ) : null}
          {progressTotal && progressIndex !== undefined ? (
            <ProgressDots total={progressTotal} activeIndex={progressIndex} />
          ) : null}
          <Button
            title={nextLabel}
            variant={nextButtonVariant}
            style={[{ marginTop: theme.spacing.lg }, nextButtonStyle]}
            onPress={handleNext}
            disabled={nextDisabled || transitioning}
          />
        </View>
      </View>
    </Screen>
  )
}
