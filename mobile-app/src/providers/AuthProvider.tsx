import { useQueryClient } from '@tanstack/react-query'
import * as SecureStore from 'expo-secure-store'
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

import { setRefreshTokenHandler } from '../services/api'
import { logout, refreshSession } from '../services/auth'

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
      const [accessToken, refreshToken, userId] = await Promise.all([
        SecureStore.getItemAsync(ACCESS_KEY),
        SecureStore.getItemAsync(REFRESH_KEY),
        SecureStore.getItemAsync(USER_ID_KEY),
      ])
      setState({ accessToken, refreshToken, userId, isLoading: false })
    }
    bootstrap()
  }, [])

  const setSession = async (accessToken: string, refreshToken: string, userId: string) => {
    await Promise.all([
      SecureStore.setItemAsync(ACCESS_KEY, accessToken),
      SecureStore.setItemAsync(REFRESH_KEY, refreshToken),
      SecureStore.setItemAsync(USER_ID_KEY, userId),
    ])
    setState({ accessToken, refreshToken, userId, isLoading: false })
  }

  const clearSession = async () => {
    if (state.refreshToken) {
      await logout(state.refreshToken)
    }
    await Promise.all([
      SecureStore.deleteItemAsync(ACCESS_KEY),
      SecureStore.deleteItemAsync(REFRESH_KEY),
      SecureStore.deleteItemAsync(USER_ID_KEY),
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
    await SecureStore.setItemAsync(ACCESS_KEY, nextAccessToken)
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
