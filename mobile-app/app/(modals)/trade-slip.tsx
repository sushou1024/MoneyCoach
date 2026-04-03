import { File } from 'expo-file-system'
import * as ImagePicker from 'expo-image-picker'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useState } from 'react'
import { Alert, Image, Platform, Text } from 'react-native'

import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Screen } from '../../src/components/Screen'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { executeInsight } from '../../src/services/insights'
import {
  computeClientChecksum,
  completeUploadBatch,
  createUploadBatch,
  fetchUploadBatch,
  uploadImage,
} from '../../src/services/uploads'
import { formatDateTime } from '../../src/utils/format'

function guessMime(uri: string) {
  if (uri.endsWith('.png')) return 'image/png'
  if (uri.endsWith('.jpg') || uri.endsWith('.jpeg')) return 'image/jpeg'
  return 'image/png'
}

async function resolveImageSizeBytes(asset: ImagePicker.ImagePickerAsset) {
  if (Platform.OS === 'web') {
    if (typeof asset.fileSize === 'number') {
      return asset.fileSize
    }
    try {
      const response = await fetch(asset.uri)
      if (!response.ok) {
        return asset.fileSize ?? 0
      }
      const blob = await response.blob()
      return blob.size
    } catch {
      return asset.fileSize ?? 0
    }
  }
  try {
    const info = new File(asset.uri).info()
    if (info.exists && typeof info.size === 'number') {
      return info.size
    }
  } catch {
    return asset.fileSize ?? 0
  }
  return asset.fileSize ?? 0
}

export default function TradeSlipScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken } = useAuth()
  const params = useLocalSearchParams<{ insight_id?: string; return_to?: string }>()
  const insightId = Array.isArray(params.insight_id) ? params.insight_id[0] : (params.insight_id ?? '')
  const returnTo = Array.isArray(params.return_to) ? params.return_to[0] : params.return_to
  const [image, setImage] = useState<ImagePicker.ImagePickerAsset | null>(null)
  const [uploading, setUploading] = useState(false)
  const [result, setResult] = useState<{ count: number; warnings: string[]; updatedAt: string } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const successRoute = returnTo || '/(tabs)/insights'

  const pickImage = async () => {
    try {
      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: 'images',
        allowsMultipleSelection: false,
        quality: 0.8,
      })
      if (!result.canceled) {
        setImage(result.assets[0])
      }
    } catch {
      Alert.alert(t('upload.permissionTitle'), t('upload.permissionBody'))
    }
  }

  const handleUpload = async () => {
    if (!accessToken || !image) return
    setUploading(true)
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
    const sizeBytes = await resolveImageSizeBytes(image)
    const meta = {
      file_name: image.fileName ?? 'trade-slip.png',
      mime_type: image.mimeType ?? guessMime(image.uri),
      size_bytes: sizeBytes,
      uri: image.uri,
    }

    const createResp = await createUploadBatch(accessToken, 'trade_slip', [meta], timezone)
    if (createResp.error ?? !createResp.data) {
      setUploading(false)
      Alert.alert(t('upload.failTitle'), createResp.error?.message ?? t('upload.failBody'))
      return
    }

    const batch = createResp.data
    const upload = batch.image_uploads[0]
    let status = 0
    try {
      status = await uploadImage(upload.upload_url, upload.headers ?? {}, meta.uri)
    } catch {
      setUploading(false)
      Alert.alert(t('upload.failTitle'), t('upload.uploadBody'))
      return
    }
    if (status < 200 || status >= 300) {
      setUploading(false)
      Alert.alert(t('upload.failTitle'), t('upload.uploadBody'))
      return
    }

    const checksum = await computeClientChecksum([
      { image_id: upload.image_id, uri: meta.uri, size_bytes: meta.size_bytes },
    ])

    const completeResp = await completeUploadBatch(accessToken, batch.upload_batch_id, [upload.image_id], checksum)
    if (completeResp.error) {
      setUploading(false)
      Alert.alert(t('upload.failTitle'), completeResp.error.message)
      return
    }

    const poll = async () => {
      const resp = await fetchUploadBatch(accessToken, batch.upload_batch_id)
      if (resp.data && 'portfolio_snapshot_id' in resp.data && resp.data.status === 'completed') {
        setUploading(false)
        const transactionCount = resp.data.transaction_ids?.length ?? 0
        const warnings = resp.data.warnings ?? []
        if (insightId && resp.data.transaction_ids?.length) {
          const executeResp = await executeInsight(accessToken, insightId, {
            method: 'trade_slip',
            transaction_ids: resp.data.transaction_ids,
          })
          if (executeResp.error) {
            Alert.alert(t('quickUpdate.failTitle'), executeResp.error.message)
          }
        }
        setResult({
          count: transactionCount,
          warnings,
          updatedAt: new Date().toISOString(),
        })
        return
      }
      if (resp.data && 'status' in resp.data && resp.data.status === 'failed') {
        setUploading(false)
        setError(resp.data.error_code ?? t('tradeSlip.failBody'))
        return
      }
      setTimeout(poll, 1500)
    }
    poll()
  }

  return (
    <Screen scroll>
      <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
        {t('tradeSlip.title')}
      </Text>
      <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {t('tradeSlip.subtitle')}
      </Text>

      {error ? (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={{ color: theme.colors.danger, fontFamily: theme.fonts.bodyBold }}>
            {t('tradeSlip.failTitle')}
          </Text>
          <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.sm, fontFamily: theme.fonts.body }}>
            {error}
          </Text>
        </Card>
      ) : null}

      {result ? (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
            {t('tradeSlip.successTitle')}
          </Text>
          <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.sm, fontFamily: theme.fonts.body }}>
            {t('tradeSlip.successBody', { count: result.count })}
          </Text>
          <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.xs, fontFamily: theme.fonts.body }}>
            {t('tradeSlip.updatedAt', { time: formatDateTime(result.updatedAt) })}
          </Text>
          {result.warnings.length ? (
            <Text style={{ color: theme.colors.warning, marginTop: theme.spacing.sm, fontFamily: theme.fonts.body }}>
              {result.warnings.join('\n')}
            </Text>
          ) : null}
          <Button
            title={t('common.close')}
            style={{ marginTop: theme.spacing.md }}
            onPress={() => router.replace(successRoute)}
          />
        </Card>
      ) : null}

      {image ? (
        <Card style={{ marginTop: theme.spacing.md, padding: 0, overflow: 'hidden' }}>
          <Image source={{ uri: image.uri }} style={{ width: '100%', height: 220 }} />
        </Card>
      ) : null}

      <Button title={t('tradeSlip.select')} style={{ marginTop: theme.spacing.md }} onPress={pickImage} />
      <Button
        title={uploading ? t('upload.uploading') : t('tradeSlip.uploadCta')}
        style={{ marginTop: theme.spacing.md }}
        onPress={handleUpload}
        disabled={uploading || !image || !!result}
      />
      {!result ? (
        <Button
          title={t('common.close')}
          variant="ghost"
          style={{ marginTop: theme.spacing.md }}
          onPress={() => router.back()}
        />
      ) : null}
    </Screen>
  )
}
