import AsyncStorage from '@react-native-async-storage/async-storage'
import { useQueryClient } from '@tanstack/react-query'
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

import { setRefreshTokenHandler } from '../services/api'
import { logout, oauthLogin, refreshSession } from '../services/auth'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  userId: string | null
  isLoading: boolean
}

interface AuthContextValue extends AuthState {
  setSession: (accessToken: string, refreshToken: string, userId: string) => Promise<void>
  clearSession: () => Promise<void>
  refreshAccessToken: () => Promise<string | null>
}

const AuthContext = createContext<AuthContextValue | null>(null)

const ACCESS_KEY = 'mc_access_token'
const REFRESH_KEY = 'mc_refresh_token'
const USER_ID_KEY = 'mc_user_id'
const GOOGLE_STATE_KEY = 'mc_google_oauth_state'

const parseAuthParams = () => {
  if (typeof window === 'undefined') return null
  const url = window.location.href
  const params = new URLSearchParams()
  const queryIndex = url.indexOf('?')
  const hashIndex = url.indexOf('#')
  if (queryIndex !== -1) {
    const queryPart = url.slice(queryIndex + 1, hashIndex === -1 ? undefined : hashIndex)
    new URLSearchParams(queryPart).forEach((value, key) => {
      params.set(key, value)
    })
  }
  if (hashIndex !== -1) {
    const hashPart = url.slice(hashIndex + 1)
    new URLSearchParams(hashPart).forEach((value, key) => {
      params.set(key, value)
    })
  }
  if (params.size === 0) {
    return null
  }
  return params
}

const clearAuthParamsFromUrl = () => {
  if (typeof window === 'undefined') return
  const cleanUrl = `${window.location.origin}${window.location.pathname}${window.location.search}`
  window.history.replaceState({}, document.title, cleanUrl)
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient()
  const [state, setState] = useState<AuthState>({
    accessToken: null,
    refreshToken: null,
    userId: null,
    isLoading: true,
  })

  useEffect(() => {
    const bootstrap = async () => {
      const params = parseAuthParams()
      if (params?.get('id_token') && params.get('state')) {
        const expectedState = window.localStorage.getItem(GOOGLE_STATE_KEY)
        window.localStorage.removeItem(GOOGLE_STATE_KEY)
        clearAuthParamsFromUrl()
        if (expectedState && expectedState === params.get('state')) {
          const resp = await oauthLogin('google', params.get('id_token') ?? '')
          if (!resp.error && resp.data) {
            await Promise.all([
              AsyncStorage.setItem(ACCESS_KEY, resp.data.access_token),
              AsyncStorage.setItem(REFRESH_KEY, resp.data.refresh_token),
              AsyncStorage.setItem(USER_ID_KEY, resp.data.user_id),
            ])
            setState({
              accessToken: resp.data.access_token,
              refreshToken: resp.data.refresh_token,
              userId: resp.data.user_id,
              isLoading: false,
            })
            return
          }
        }
      }
      const [accessToken, refreshToken, userId] = await Promise.all([
        AsyncStorage.getItem(ACCESS_KEY),
        AsyncStorage.getItem(REFRESH_KEY),
        AsyncStorage.getItem(USER_ID_KEY),
      ])
      setState({ accessToken, refreshToken, userId, isLoading: false })
    }
    bootstrap()
  }, [])

  const setSession = async (accessToken: string, refreshToken: string, userId: string) => {
    await Promise.all([
      AsyncStorage.setItem(ACCESS_KEY, accessToken),
      AsyncStorage.setItem(REFRESH_KEY, refreshToken),
      AsyncStorage.setItem(USER_ID_KEY, userId),
    ])
    setState({ accessToken, refreshToken, userId, isLoading: false })
  }

  const clearSession = async () => {
    if (state.refreshToken) {
      await logout(state.refreshToken)
    }
    await Promise.all([
      AsyncStorage.removeItem(ACCESS_KEY),
      AsyncStorage.removeItem(REFRESH_KEY),
      AsyncStorage.removeItem(USER_ID_KEY),
    ])
    queryClient.clear()
    setState({ accessToken: null, refreshToken: null, userId: null, isLoading: false })
  }

  const refreshAccessToken = useCallback(async () => {
    if (!state.refreshToken) return null
    const resp = await refreshSession(state.refreshToken)
    if (resp.error) {
      return null
    }
    const nextAccessToken = resp.data?.access_token ?? null
    if (!nextAccessToken) {
      return null
    }
    await AsyncStorage.setItem(ACCESS_KEY, nextAccessToken)
    setState((prev) => ({ ...prev, accessToken: nextAccessToken }))
    return nextAccessToken
  }, [state.refreshToken])

  useEffect(() => {
    setRefreshTokenHandler(() => refreshAccessToken())
    return () => setRefreshTokenHandler(null)
  }, [refreshAccessToken])

  const value = useMemo(
    () => ({
      ...state,
      setSession,
      clearSession,
      refreshAccessToken,
    }),
    [state]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('AuthProvider missing')
  }
  return ctx
}
