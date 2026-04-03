import { Stack } from 'expo-router'
import React from 'react'

export default function ModalLayout() {
  return <Stack screenOptions={{ headerShown: false, presentation: 'modal' }} />
}
