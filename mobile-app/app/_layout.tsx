import {
  useFonts,
  SpaceGrotesk_400Regular,
  SpaceGrotesk_500Medium,
  SpaceGrotesk_600SemiBold,
} from '@expo-google-fonts/space-grotesk'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Stack, useRouter, useSegments } from 'expo-router'
import { StatusBar } from 'expo-status-bar'
import * as WebBrowser from 'expo-web-browser'
import React, { useEffect, useRef } from 'react'
import { Platform, StyleSheet, View, useWindowDimensions } from 'react-native'
import { SafeAreaProvider } from 'react-native-safe-area-context'

import { LoadingScreen } from '../src/components/LoadingScreen'
import { AuthProvider, useAuth } from '../src/providers/AuthProvider'
import { IAPProvider } from '../src/providers/IAPProvider'
import { LocalizationProvider } from '../src/providers/LocalizationProvider'
import { ThemeProvider, useTheme } from '../src/providers/ThemeProvider'
import { useOnboardingStore } from '../src/stores/onboarding'

const queryClient = new QueryClient()

// Ensure OAuth popups on web can complete even when they return to a non-auth route.
WebBrowser.maybeCompleteAuthSession()

function AuthGate() {
  const { accessToken, isLoading } = useAuth()
  const segments = useSegments()
  const router = useRouter()
  const onboardingMode = useOnboardingStore((state) => state.mode)
  const resetOnboarding = useOnboardingStore((state) => state.reset)
  const wasInAuthRef = useRef(false)

  useEffect(() => {
    if (isLoading) return
    const inAuth = segments[0] === '(auth)'
    const wasInAuth = wasInAuthRef.current
    if (accessToken && !inAuth && onboardingMode === 'retake' && wasInAuth) {
      resetOnboarding()
    }
    if (!accessToken && !inAuth) {
      router.replace('/(auth)/sc00')
      wasInAuthRef.current = inAuth
      return
    }
    if (accessToken && inAuth && onboardingMode !== 'retake' && (segments as string[])[1] !== 'sc08') {
      router.replace('/')
    }
    wasInAuthRef.current = inAuth
  }, [accessToken, isLoading, onboardingMode, resetOnboarding, router, segments])

  return null
}

const IPHONE_17_PRO_ASPECT = 9 / 19.5
const WEB_FRAME_SCALE = 0.98

function AppFrame({ children }: { children: React.ReactNode }) {
  const theme = useTheme()
  const { width, height } = useWindowDimensions()

  if (Platform.OS !== 'web') {
    return <View style={{ flex: 1 }}>{children}</View>
  }

  const widthByHeight = height * IPHONE_17_PRO_ASPECT
  let frameWidth = width
  let frameHeight = width / IPHONE_17_PRO_ASPECT
  if (widthByHeight <= width) {
    frameWidth = widthByHeight
    frameHeight = height
  }
  frameWidth *= WEB_FRAME_SCALE
  frameHeight *= WEB_FRAME_SCALE

  return (
    <View style={[styles.webRoot, { backgroundColor: theme.colors.background }]}>
      <View
        style={[styles.webFrame, { width: frameWidth, height: frameHeight, backgroundColor: theme.colors.background }]}
      >
        {children}
      </View>
    </View>
  )
}

export default function RootLayout() {
  const [fontsLoaded] = useFonts({
    SpaceGrotesk_400Regular,
    SpaceGrotesk_500Medium,
    SpaceGrotesk_600SemiBold,
  })

  if (!fontsLoaded) {
    return (
      <SafeAreaProvider>
        <ThemeProvider>
          <StatusBar style="dark" />
          <AppFrame>
            <LoadingScreen label="Preparing your dashboard..." />
          </AppFrame>
        </ThemeProvider>
      </SafeAreaProvider>
    )
  }

  return (
    <SafeAreaProvider>
      <LocalizationProvider>
        <QueryClientProvider client={queryClient}>
          <AuthProvider>
            <IAPProvider>
              <ThemeProvider>
                <StatusBar style="dark" />
                <AuthGate />
                <AppFrame>
                  <Stack screenOptions={{ headerShown: false }}>
                    <Stack.Screen name="(auth)" />
                    <Stack.Screen name="(tabs)" />
                    <Stack.Screen name="(modals)" options={{ presentation: 'modal' }} />
                  </Stack>
                </AppFrame>
              </ThemeProvider>
            </IAPProvider>
          </AuthProvider>
        </QueryClientProvider>
      </LocalizationProvider>
    </SafeAreaProvider>
  )
}

const styles = StyleSheet.create({
  webRoot: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  webFrame: {
    overflow: 'hidden',
    borderRadius: 32,
  },
})
