import { Redirect } from 'expo-router'
import React from 'react'

import { LoadingScreen } from '../src/components/LoadingScreen'
import { useEntitlement } from '../src/hooks/useEntitlement'
import { useAuth } from '../src/providers/AuthProvider'
import { useLocalization } from '../src/providers/LocalizationProvider'

export default function Index() {
  const { accessToken, isLoading } = useAuth()
  const entitlement = useEntitlement()
  const { t } = useLocalization()

  if (isLoading || (accessToken && entitlement.isLoading)) {
    return <LoadingScreen label={t('app.loadingDashboard')} />
  }

  if (!accessToken) {
    return <Redirect href="/(auth)/sc00" />
  }

  if (entitlement.data && (entitlement.data.status === 'active' || entitlement.data.status === 'grace')) {
    return <Redirect href="/(tabs)/insights" />
  }

  return <Redirect href="/(tabs)/assets" />
}
