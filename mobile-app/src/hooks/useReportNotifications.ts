import { QueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useRef, useState } from 'react'
import { Alert, Animated, Easing } from 'react-native'

import { useProfile } from './useProfile'
import { useLocalization } from '../providers/LocalizationProvider'
import { getPushPermission, PushPermissionStatus, registerForPush } from '../services/notifications'
import { updateProfile } from '../services/profile'

type UseReportNotificationsParams = {
  accessToken: string | null
  userId?: string
  queryClient: QueryClient
}

export function useReportNotifications({ accessToken, userId, queryClient }: UseReportNotificationsParams) {
  const profile = useProfile()
  const { t } = useLocalization()
  const [showNotifyModal, setShowNotifyModal] = useState(false)
  const [notifyPlanId, setNotifyPlanId] = useState<string | null>(null)
  const [pushPermissionStatus, setPushPermissionStatus] = useState<PushPermissionStatus>('unknown')
  const [notifyInlineState, setNotifyInlineState] = useState<'success' | 'denied' | null>(null)
  const [isEnablingNotifications, setIsEnablingNotifications] = useState(false)
  const notifyEnabledAnim = useRef(new Animated.Value(0)).current
  const notifyInlineAnim = useRef(new Animated.Value(0)).current
  const notifyInlineTimeout = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isMountedRef = useRef(true)

  useEffect(() => {
    isMountedRef.current = true
    return () => {
      isMountedRef.current = false
      if (notifyInlineTimeout.current) {
        clearTimeout(notifyInlineTimeout.current)
        notifyInlineTimeout.current = null
      }
    }
  }, [])

  useEffect(() => {
    const syncPermission = async () => {
      const permission = await getPushPermission()
      if (!isMountedRef.current) return
      setPushPermissionStatus(permission.status)
      notifyEnabledAnim.setValue(permission.status === 'granted' ? 1 : 0)
    }
    syncPermission()
  }, [notifyEnabledAnim])

  const showNotifyInline = useCallback(
    (state: 'success' | 'denied') => {
      setNotifyInlineState(state)
      notifyInlineAnim.setValue(0)
      Animated.timing(notifyInlineAnim, {
        toValue: 1,
        duration: 160,
        easing: Easing.out(Easing.cubic),
        useNativeDriver: true,
      }).start()

      if (notifyInlineTimeout.current) {
        clearTimeout(notifyInlineTimeout.current)
      }
      notifyInlineTimeout.current = setTimeout(() => {
        Animated.timing(notifyInlineAnim, {
          toValue: 0,
          duration: 220,
          easing: Easing.in(Easing.cubic),
          useNativeDriver: true,
        }).start(({ finished }) => {
          if (!finished || !isMountedRef.current) return
          setNotifyInlineState(null)
        })
      }, 2200)
    },
    [notifyInlineAnim]
  )

  const enableNotifications = useCallback(async () => {
    if (!accessToken) return
    setIsEnablingNotifications(true)
    try {
      const resp = await registerForPush(accessToken)
      setPushPermissionStatus(resp.permission?.status ?? 'unknown')
      if (!resp.registered) {
        if (resp.permission?.status && resp.permission.status !== 'granted') {
          setShowNotifyModal(false)
          showNotifyInline('denied')
          return
        }
        Alert.alert(t('report.notifyFailTitle'), resp.error ?? t('report.notifyFailBody'))
        return
      }

      const currentPrefs = profile.data?.notification_prefs ?? {}
      const nextPrefs = {
        ...currentPrefs,
        portfolio_alerts: true,
        action_alerts: true,
        market_alpha: currentPrefs.market_alpha ?? false,
      }
      const updateResp = await updateProfile(accessToken, { notification_prefs: nextPrefs })
      if (updateResp.error) {
        Alert.alert(t('report.notifyFailTitle'), updateResp.error.message)
        return
      }
      if (userId) {
        await queryClient.invalidateQueries({ queryKey: ['profile', userId] })
      }

      Animated.timing(notifyEnabledAnim, {
        toValue: 1,
        duration: 160,
        easing: Easing.out(Easing.cubic),
        useNativeDriver: true,
      }).start()
      setPushPermissionStatus('granted')
      setShowNotifyModal(false)
      showNotifyInline('success')
    } finally {
      if (isMountedRef.current) {
        setIsEnablingNotifications(false)
      }
    }
  }, [accessToken, notifyEnabledAnim, profile.data?.notification_prefs, queryClient, showNotifyInline, t, userId])

  const handleNotify = useCallback((planId: string) => {
    setNotifyPlanId(planId)
    setNotifyInlineState(null)
    setShowNotifyModal(true)
  }, [])

  const closeNotifyModal = useCallback(() => {
    if (isEnablingNotifications) return
    setShowNotifyModal(false)
  }, [isEnablingNotifications])

  return {
    showNotifyModal,
    notifyPlanId,
    notifyInlineState,
    notifyEnabledAnim,
    notifyInlineAnim,
    notifyEnabled: pushPermissionStatus === 'granted',
    isEnablingNotifications,
    handleNotify,
    enableNotifications,
    closeNotifyModal,
    setShowNotifyModal,
  }
}
