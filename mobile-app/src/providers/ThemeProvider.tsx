import React, { createContext, useContext } from 'react'

import { colors, fonts, radius, spacing } from '../utils/theme'

const ThemeContext = createContext({ colors, radius, spacing, fonts })

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return <ThemeContext.Provider value={{ colors, radius, spacing, fonts }}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  return useContext(ThemeContext)
}
