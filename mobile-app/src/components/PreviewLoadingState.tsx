import React, { useEffect, useState } from 'react'
import { Text, View } from 'react-native'

import { Card } from './Card'
import { Screen } from './Screen'
import { ShimmerBlock } from './ShimmerBlock'
import { useTheme } from '../providers/ThemeProvider'

type PreviewLoadingStateProps = {
  title: string
  steps?: string[]
}

export function PreviewLoadingState({ title, steps }: PreviewLoadingStateProps) {
  const theme = useTheme()
  const [stepIndex, setStepIndex] = useState(0)
  const stepCount = steps?.length ?? 0
  const stepLabel = stepCount > 0 ? steps?.[stepIndex] : null

  useEffect(() => {
    if (!stepCount) return
    setStepIndex(0)
    const interval = setInterval(() => {
      setStepIndex((prev) => (prev + 1) % stepCount)
    }, 2000)
    return () => clearInterval(interval)
  }, [stepCount])

  return (
    <Screen>
      <Text style={{ fontSize: 22, color: theme.colors.ink, fontFamily: theme.fonts.display }}>{title}</Text>
      {stepLabel ? (
        <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
          {stepLabel}
        </Text>
      ) : null}
      <View style={{ marginTop: theme.spacing.lg }}>
        <Card>
          <ShimmerBlock width="52%" height={10} radius={6} />
          <ShimmerBlock width="70%" height={22} radius={8} style={{ marginTop: theme.spacing.sm }} />
          <View style={{ flexDirection: 'row', gap: theme.spacing.md, marginTop: theme.spacing.md }}>
            <ShimmerBlock width={84} height={84} radius={42} />
            <ShimmerBlock width={84} height={84} radius={42} />
          </View>
        </Card>
        <Card style={{ marginTop: theme.spacing.md }}>
          <ShimmerBlock width="40%" height={10} radius={6} />
          <View style={{ marginTop: theme.spacing.md, gap: theme.spacing.md }}>
            {Array.from({ length: 3 }).map((_, index) => (
              <View key={`risk-${index}`} style={{ flexDirection: 'row', alignItems: 'center' }}>
                <ShimmerBlock width={28} height={28} radius={14} />
                <View style={{ flex: 1, marginLeft: theme.spacing.sm }}>
                  <ShimmerBlock width="60%" height={10} radius={6} />
                  <ShimmerBlock width="85%" height={8} radius={6} style={{ marginTop: theme.spacing.xs }} />
                </View>
              </View>
            ))}
          </View>
        </Card>
      </View>
    </Screen>
  )
}
