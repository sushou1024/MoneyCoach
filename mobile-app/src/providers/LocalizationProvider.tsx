import AsyncStorage from '@react-native-async-storage/async-storage'
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

import { SupportedLocale, getCurrentLocale, resolveLocale, setCurrentLocale, translate } from '../utils/i18n'

const STORAGE_KEY = 'mc_locale_override'

type LocalizationContextValue = {
  locale: SupportedLocale
  setLocale: (locale: SupportedLocale) => Promise<void>
  t: (key: Parameters<typeof translate>[1], params?: Record<string, string | number>) => string
}

const LocalizationContext = createContext<LocalizationContextValue | null>(null)

export function LocalizationProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<SupportedLocale>(getCurrentLocale())

  useEffect(() => {
    const hydrate = async () => {
      const stored = await AsyncStorage.getItem(STORAGE_KEY)
      if (stored) {
        const resolved = resolveLocale(stored)
        setLocaleState(resolved)
        setCurrentLocale(resolved)
      }
    }
    hydrate()
  }, [])

  const setLocale = useCallback(async (next: SupportedLocale) => {
    setLocaleState(next)
    setCurrentLocale(next)
    await AsyncStorage.setItem(STORAGE_KEY, next)
  }, [])

  const t = useCallback(
    (key: Parameters<typeof translate>[1], params?: Record<string, string | number>) => translate(locale, key, params),
    [locale]
  )

  const value = useMemo(
    () => ({
      locale,
      setLocale,
      t,
    }),
    [locale, setLocale, t]
  )

  return <LocalizationContext.Provider value={value}>{children}</LocalizationContext.Provider>
}

export function useLocalization() {
  const ctx = useContext(LocalizationContext)
  if (!ctx) {
    throw new Error('LocalizationProvider missing')
  }
  return ctx
}
