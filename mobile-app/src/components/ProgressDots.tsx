import React from 'react'
import { StyleSheet, View } from 'react-native'

import { useTheme } from '../providers/ThemeProvider'

export function ProgressDots({ total, activeIndex }: { total: number; activeIndex: number }) {
  const theme = useTheme()
  return (
    <View style={styles.row}>
      {Array.from({ length: total }).map((_, index) => (
        <View
          key={`dot-${index}`}
          style={[
            styles.dot,
            {
              backgroundColor: index === activeIndex ? theme.colors.accent : theme.colors.surfaceElevated,
              borderColor: index === activeIndex ? theme.colors.accent : theme.colors.border,
            },
          ]}
        />
      ))}
    </View>
  )
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 8,
  },
  dot: {
    width: 8,
    height: 8,
    borderRadius: 99,
    borderWidth: 1,
  },
})
