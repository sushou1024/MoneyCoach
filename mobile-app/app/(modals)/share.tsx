import { useQuery } from '@tanstack/react-query'
import * as MediaLibrary from 'expo-media-library'
import { useRouter } from 'expo-router'
import * as Sharing from 'expo-sharing'
import React, { useMemo, useRef, useState } from 'react'
import { Alert, Platform, Share, StyleSheet, Text, View } from 'react-native'
import ViewShot from 'react-native-view-shot'
import type { CaptureOptions } from 'react-native-view-shot'

import { Button } from '../../src/components/Button'
import { Screen } from '../../src/components/Screen'
import { SpeedometerGauge } from '../../src/components/SpeedometerGauge'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchPortfolioSnapshots } from '../../src/services/portfolio'
import { fetchReport } from '../../src/services/reports'
import { isPaidReport } from '../../src/utils/report'

export default function ShareScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken, userId } = useAuth()
  const viewShotRef = useRef<ViewShot>(null)
  const webCaptureRef = useRef<View>(null)
  const isWeb = Platform.OS === 'web'
  const [webStatus, setWebStatus] = useState<{ message: string; tone: 'muted' | 'danger' } | null>(null)
  const styles = useMemo(() => createStyles(theme), [theme])
  const viewShotOptions = useMemo<CaptureOptions>(
    () => ({
      format: 'png',
      quality: 0.96,
      result: isWeb ? 'data-uri' : 'tmpfile',
    }),
    [isWeb]
  )

  const reportsQuery = useQuery({
    queryKey: ['reports', userId],
    queryFn: async () => {
      if (!accessToken) return null
      const reportsResp = await fetchPortfolioSnapshots(accessToken)
      if (reportsResp.error) {
        throw new Error(reportsResp.error.message)
      }
      return reportsResp.data?.items ?? []
    },
    enabled: !!accessToken,
  })

  const readyReport = useMemo(() => {
    const items = reportsQuery.data ?? []
    return items.find((item) => item.status === 'ready') ?? null
  }, [reportsQuery.data])

  const reportQuery = useQuery({
    queryKey: ['share-report', userId, readyReport?.calculation_id],
    queryFn: async () => {
      if (!accessToken || !readyReport?.calculation_id) return null
      const reportResp = await fetchReport(accessToken, readyReport.calculation_id)
      if (reportResp.error) return null
      return reportResp.data ?? null
    },
    enabled: !!accessToken && !!readyReport?.calculation_id,
  })

  const score = readyReport?.health_score
  const reportData = reportQuery.data
  const verdict = isPaidReport(reportData)
    ? (reportData.the_verdict?.constructive_comment ?? reportData.risk_summary ?? t('share.verdictFallback'))
    : t('share.verdictFallback')

  const setWebMessage = (message: string, tone: 'muted' | 'danger' = 'muted') => {
    if (!isWeb) return
    setWebStatus({ message, tone })
  }

  const captureWebCard = async () => {
    try {
      const node = webCaptureRef.current as unknown as HTMLElement | null
      if (!node) {
        setWebMessage(t('share.captureFailBody'), 'danger')
        return null
      }
      if (typeof document !== 'undefined' && document.fonts?.ready) {
        try {
          await document.fonts.ready
        } catch {
          // Ignore font loading errors and attempt capture anyway.
        }
      }
      await new Promise((resolve) => requestAnimationFrame(resolve))
      const rect = node.getBoundingClientRect()
      const computedBackground = window.getComputedStyle(node).backgroundColor
      const backgroundColor =
        computedBackground && computedBackground !== 'rgba(0, 0, 0, 0)' && computedBackground !== 'transparent'
          ? computedBackground
          : undefined
      try {
        const htmlToImage = await import('html-to-image')
        const dataUrl = await htmlToImage.toPng(node, {
          cacheBust: true,
          pixelRatio: window.devicePixelRatio || 2,
          width: Math.ceil(rect.width),
          height: Math.ceil(rect.height),
          backgroundColor,
        })
        return dataUrl
      } catch {
        // Fall back to html2canvas for older browsers or if DOM cloning fails.
      }
      const html2canvasModule = await import('html2canvas')
      const html2canvas = html2canvasModule.default ?? html2canvasModule
      const canvas = await html2canvas(node, {
        backgroundColor: backgroundColor ?? null,
        scale: window.devicePixelRatio || 1,
        width: Math.ceil(rect.width),
        height: Math.ceil(rect.height),
        useCORS: true,
        logging: false,
      })
      return canvas.toDataURL('image/png', 0.96)
    } catch {
      setWebMessage(t('share.captureFailBody'), 'danger')
      return null
    }
  }

  const captureCard = async () => {
    if (isWeb) {
      return captureWebCard()
    }
    const uri = await viewShotRef.current?.capture?.()
    if (!uri) {
      Alert.alert(t('share.captureFailTitle'), t('share.captureFailBody'))
      return null
    }
    return uri
  }

  const dataUriToBlob = (dataUri: string) => {
    const [header, data] = dataUri.split(',')
    if (!data) return null
    const match = /data:(.*);base64/.exec(header)
    const mime = match?.[1] ?? 'image/png'
    const binary = atob(data)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i)
    }
    return new Blob([bytes], { type: mime })
  }

  const downloadWebImage = async (dataUri: string) => {
    try {
      const blob = dataUriToBlob(dataUri)
      if (!blob || typeof document === 'undefined') {
        setWebMessage(t('share.captureFailBody'), 'danger')
        return false
      }
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = 'moneycoach-health-score.png'
      document.body.appendChild(link)
      link.click()
      link.remove()
      setTimeout(() => URL.revokeObjectURL(url), 1000)
      return true
    } catch {
      setWebMessage(t('share.captureFailBody'), 'danger')
      return false
    }
  }

  const shareWebImage = async (dataUri: string) => {
    const webNavigator =
      typeof navigator === 'undefined'
        ? null
        : (navigator as { share?: (data: any) => Promise<void>; canShare?: (data: any) => boolean })
    if (!webNavigator?.share) return false
    const message = t('share.message', { score: score ?? '--' })
    try {
      const blob = dataUriToBlob(dataUri)
      if (blob && typeof File !== 'undefined') {
        const file = new File([blob], 'moneycoach-health-score.png', { type: blob.type || 'image/png' })
        if (webNavigator.canShare?.({ files: [file] })) {
          await webNavigator.share({ files: [file], title: t('share.title'), text: message })
          return true
        }
      }
      await webNavigator.share({ title: t('share.title'), text: message })
      return true
    } catch {
      return false
    }
  }

  const handleSave = async () => {
    if (isWeb) {
      setWebStatus(null)
      const uri = await captureCard()
      if (!uri) return
      const saved = await downloadWebImage(uri)
      if (saved) {
        setWebMessage(t('share.webDownloadReady'))
      }
      return
    }
    const permission = await MediaLibrary.requestPermissionsAsync()
    if (!permission.granted) {
      Alert.alert(t('share.permissionTitle'), t('share.permissionBody'))
      return
    }
    const uri = await captureCard()
    if (!uri) return
    await MediaLibrary.saveToLibraryAsync(uri)
    Alert.alert(t('share.saveSuccessTitle'), t('share.saveSuccessBody'))
  }

  const handleShare = async () => {
    if (isWeb) {
      setWebStatus(null)
      const uri = await captureCard()
      if (!uri) return
      const shared = await shareWebImage(uri)
      if (!shared) {
        const downloaded = await downloadWebImage(uri)
        if (downloaded) {
          setWebMessage(t('share.webShareFallback'))
        }
      }
      return
    }
    const uri = await captureCard()
    if (!uri) return
    if (await Sharing.isAvailableAsync()) {
      await Sharing.shareAsync(uri)
      return
    }
    await Share.share({ url: uri, message: t('share.message', { score: score ?? '--' }) })
  }

  return (
    <Screen scroll>
      <View style={styles.header}>
        <Text style={styles.title}>{t('share.title')}</Text>
        <Text style={styles.subtitle}>{t('share.subtitle')}</Text>
      </View>
      {isWeb ? (
        <View style={styles.captureWrap}>
          <View ref={webCaptureRef} style={styles.shareCard} collapsable={false}>
            <View style={styles.cardGlowTop} />
            <View style={styles.cardGlowBottom} />
            <View style={styles.cardRail} />
            <View style={styles.cardContent}>
              <View style={styles.cardHeader}>
                <View style={styles.brandRow}>
                  <View style={styles.brandDot} />
                  <Text style={styles.brandText}>{t('share.watermark')}</Text>
                </View>
                <View style={styles.scorePill}>
                  <Text style={styles.scorePillText}>{t('share.scoreLabel')}</Text>
                </View>
              </View>
              <View style={styles.scoreWrap}>
                <View style={styles.scoreRingWrap}>
                  <SpeedometerGauge size={150} value={score ?? undefined} label={t('share.scoreLabel')} />
                </View>
              </View>
              <View style={styles.verdictBox}>
                <Text style={styles.verdictText}>{verdict}</Text>
              </View>
              <View style={styles.footerRow}>
                <View style={styles.footerRule} />
                <Text style={styles.footerText}>{t('share.watermark')}</Text>
              </View>
            </View>
          </View>
        </View>
      ) : (
        <ViewShot ref={viewShotRef} options={viewShotOptions} style={styles.captureWrap}>
          <View style={styles.shareCard} collapsable={false}>
            <View style={styles.cardGlowTop} />
            <View style={styles.cardGlowBottom} />
            <View style={styles.cardRail} />
            <View style={styles.cardContent}>
              <View style={styles.cardHeader}>
                <View style={styles.brandRow}>
                  <View style={styles.brandDot} />
                  <Text style={styles.brandText}>{t('share.watermark')}</Text>
                </View>
                <View style={styles.scorePill}>
                  <Text style={styles.scorePillText}>{t('share.scoreLabel')}</Text>
                </View>
              </View>
              <View style={styles.scoreWrap}>
                <View style={styles.scoreRingWrap}>
                  <SpeedometerGauge size={150} value={score ?? undefined} label={t('share.scoreLabel')} />
                </View>
              </View>
              <View style={styles.verdictBox}>
                <Text style={styles.verdictText}>{verdict}</Text>
              </View>
              <View style={styles.footerRow}>
                <View style={styles.footerRule} />
                <Text style={styles.footerText}>{t('share.watermark')}</Text>
              </View>
            </View>
          </View>
        </ViewShot>
      )}
      <View style={styles.actions}>
        <Button title={t('share.saveCta')} onPress={handleSave} />
        <Button
          title={t('share.shareCta')}
          variant="secondary"
          style={{ marginTop: theme.spacing.sm }}
          onPress={handleShare}
        />
        <Button
          title={t('common.close')}
          variant="ghost"
          style={{ marginTop: theme.spacing.sm }}
          onPress={() => router.back()}
        />
        {isWeb && webStatus ? (
          <Text
            style={[
              styles.webStatus,
              { color: webStatus.tone === 'danger' ? theme.colors.danger : theme.colors.muted },
            ]}
          >
            {webStatus.message}
          </Text>
        ) : null}
      </View>
    </Screen>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    header: {
      alignItems: 'center',
    },
    title: {
      fontSize: 24,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    subtitle: {
      marginTop: theme.spacing.xs,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      textAlign: 'center',
    },
    captureWrap: {
      marginTop: theme.spacing.lg,
      width: '100%',
      maxWidth: 420,
      alignSelf: 'center',
    },
    shareCard: {
      borderRadius: theme.radius.xl,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surface,
      overflow: 'hidden',
      minHeight: 320,
    },
    cardGlowTop: {
      position: 'absolute',
      width: 220,
      height: 220,
      borderRadius: 999,
      right: -80,
      top: -120,
      backgroundColor: theme.colors.accentSoft,
      opacity: 0.4,
    },
    cardGlowBottom: {
      position: 'absolute',
      width: 240,
      height: 240,
      borderRadius: 999,
      left: -120,
      bottom: -140,
      backgroundColor: theme.colors.glow,
      opacity: 0.25,
    },
    cardRail: {
      position: 'absolute',
      left: 0,
      top: 0,
      bottom: 0,
      width: 5,
      backgroundColor: theme.colors.accentSoft,
    },
    cardContent: {
      padding: theme.spacing.lg,
      gap: theme.spacing.md,
    },
    cardHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    brandRow: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.xs,
    },
    brandDot: {
      width: 6,
      height: 6,
      borderRadius: 999,
      backgroundColor: theme.colors.accent,
    },
    brandText: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 12,
      letterSpacing: 1.6,
      textTransform: 'uppercase',
    },
    scorePill: {
      paddingVertical: 4,
      paddingHorizontal: theme.spacing.sm,
      borderRadius: 999,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
    },
    scorePillText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 11,
      letterSpacing: 1.1,
      textTransform: 'uppercase',
    },
    scoreWrap: {
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: 170,
    },
    scoreRingWrap: {
      alignItems: 'center',
    },
    verdictBox: {
      paddingVertical: theme.spacing.sm,
      paddingHorizontal: theme.spacing.md,
      borderRadius: theme.radius.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
    },
    verdictText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      textAlign: 'center',
    },
    footerRow: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.sm,
    },
    footerRule: {
      flex: 1,
      height: 1,
      backgroundColor: theme.colors.border,
    },
    footerText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 11,
      letterSpacing: 1.4,
      textTransform: 'uppercase',
    },
    actions: {
      marginTop: theme.spacing.lg,
    },
    webStatus: {
      marginTop: theme.spacing.sm,
      textAlign: 'center',
      fontFamily: theme.fonts.body,
    },
  })
