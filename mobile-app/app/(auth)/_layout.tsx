import { Stack } from 'expo-router'
import React from 'react'

export default function AuthLayout() {
  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="sc00" options={{ animation: 'none' }} />
      <Stack.Screen name="sc01" options={{ animation: 'none' }} />
      <Stack.Screen name="sc01b" options={{ animation: 'none' }} />
      <Stack.Screen name="sc02" options={{ animation: 'none' }} />
      <Stack.Screen name="sc03" options={{ animation: 'none' }} />
      <Stack.Screen name="sc04" options={{ animation: 'none' }} />
      <Stack.Screen name="sc05" options={{ animation: 'none' }} />
      <Stack.Screen name="sc06" options={{ animation: 'none' }} />
      <Stack.Screen name="sc07" options={{ animation: 'none' }} />
      <Stack.Screen name="sc07a" options={{ animation: 'none' }} />
      <Stack.Screen name="sc08" options={{ animation: 'none' }} />
    </Stack>
  )
}
