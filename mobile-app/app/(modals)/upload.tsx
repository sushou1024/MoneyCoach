import { Ionicons } from '@expo/vector-icons'
import { File } from 'expo-file-system'
import * as ImagePicker from 'expo-image-picker'
import { useRouter } from 'expo-router'
import React, { useState } from 'react'
import { Alert, FlatList, Image, Platform, Pressable, Text, View } from 'react-native'

import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { ProgressBar } from '../../src/components/ProgressBar'
import { Screen } from '../../src/components/Screen'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { computeClientChecksum, completeUploadBatch, createUploadBatch, uploadImage } from '../../src/services/uploads'
import { useUploadDraftStore } from '../../src/stores/uploadDraft'

function guessMime(uri: string) {
  if (uri.endsWith('.png')) return 'image/png'
  if (uri.endsWith('.jpg') || uri.endsWith('.jpeg')) return 'image/jpeg'
  return 'image/png'
}

function filenameFromUri(uri: string) {
  const parts = uri.split('/')
  return parts[parts.length - 1] ?? `image_${Date.now()}.png`
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

export default function UploadScreen() {
  const theme = useTheme()
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken } = useAuth()
  const { setDraft } = useUploadDraftStore()
  const [images, setImages] = useState<ImagePicker.ImagePickerAsset[]>([])
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState({ current: 0, total: 0 })

  const handleCancel = () => {
    router.replace('/(tabs)/assets')
  }

  const pickImages = async () => {
    try {
      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: 'images',
        allowsMultipleSelection: true,
        quality: 0.8,
      })
      if (!result.canceled) {
        const merged = [...images, ...result.assets].reduce<ImagePicker.ImagePickerAsset[]>((acc, item) => {
          if (!acc.find((existing) => existing.uri === item.uri)) acc.push(item)
          return acc
        }, [])
        if (merged.length > 15) {
          Alert.alert(t('upload.limitTitle'), t('upload.limitBody'))
        }
        setImages(merged.slice(0, 15))
      }
    } catch {
      Alert.alert(t('upload.permissionTitle'), t('upload.permissionBody'))
    }
  }

  const removeImage = (uri: string) => {
    setImages((prev) => prev.filter((img) => img.uri !== uri))
  }

  const handleUpload = async () => {
    if (!accessToken || images.length === 0) return
    setUploading(true)
    setProgress({ current: 0, total: images.length })
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

    const imageMeta = await Promise.all(
      images.map(async (image) => {
        const sizeBytes = await resolveImageSizeBytes(image)
        return {
          file_name: image.fileName ?? filenameFromUri(image.uri),
          mime_type: image.mimeType ?? guessMime(image.uri),
          size_bytes: sizeBytes,
          uri: image.uri,
        }
      })
    )

    const createResp = await createUploadBatch(accessToken, 'holdings', imageMeta, timezone)
    if (createResp.error ?? !createResp.data) {
      setUploading(false)
      Alert.alert(t('upload.failTitle'), createResp.error?.message ?? t('upload.failBody'))
      return
    }

    const uploads = createResp.data.image_uploads
    setDraft(
      createResp.data.upload_batch_id,
      uploads.map((upload, index) => ({ imageId: upload.image_id, uri: imageMeta[index].uri }))
    )
    for (let idx = 0; idx < uploads.length; idx += 1) {
      const upload = uploads[idx]
      const file = imageMeta[idx]
      let status = 0
      try {
        status = await uploadImage(upload.upload_url, upload.headers ?? {}, file.uri)
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
      setProgress((prev) => ({ ...prev, current: prev.current + 1 }))
    }

    const checksum = await computeClientChecksum(
      uploads.map((upload, index) => ({
        image_id: upload.image_id,
        uri: imageMeta[index].uri,
        size_bytes: imageMeta[index].size_bytes,
      }))
    )

    const imageIds = uploads.map((item) => item.image_id)
    const completeResp = await completeUploadBatch(accessToken, createResp.data.upload_batch_id, imageIds, checksum)
    setUploading(false)
    if (completeResp.error) {
      Alert.alert(t('upload.failTitle'), completeResp.error.message)
      return
    }

    router.replace({ pathname: '/(modals)/processing-ocr', params: { id: createResp.data.upload_batch_id } })
  }

  return (
    <Screen scroll>
      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
        <Text style={{ fontSize: 24, color: theme.colors.ink, fontFamily: theme.fonts.display }}>
          {t('upload.title')}
        </Text>
        <Button title={t('common.cancel')} variant="ghost" onPress={handleCancel} />
      </View>
      <Text style={{ marginTop: theme.spacing.sm, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
        {t('upload.subtitle')}
      </Text>

      <Button title={t('upload.select')} style={{ marginTop: theme.spacing.md }} onPress={pickImages} />

      <FlatList
        data={images}
        keyExtractor={(item) => item.uri}
        numColumns={3}
        scrollEnabled={false}
        columnWrapperStyle={{ gap: theme.spacing.sm, marginTop: theme.spacing.md }}
        renderItem={({ item }) => (
          <View style={{ flex: 1 / 3 }}>
            <Card style={{ padding: 0, overflow: 'hidden' }}>
              <Image source={{ uri: item.uri }} style={{ width: '100%', height: 96 }} />
              <Pressable
                onPress={() => removeImage(item.uri)}
                style={{
                  position: 'absolute',
                  top: 6,
                  right: 6,
                  backgroundColor: theme.colors.overlay,
                  borderRadius: 12,
                  padding: 2,
                }}
              >
                <Ionicons name="close" size={14} color={theme.colors.ink} />
              </Pressable>
            </Card>
          </View>
        )}
      />

      <View style={{ marginTop: theme.spacing.md }}>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('upload.privacy')}</Text>
      </View>

      {uploading ? (
        <View style={{ marginTop: theme.spacing.md }}>
          <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>
            {t('upload.progress', { current: progress.current, total: progress.total })}
          </Text>
          <View style={{ marginTop: theme.spacing.sm }}>
            <ProgressBar progress={progress.total ? progress.current / progress.total : 0} />
          </View>
        </View>
      ) : null}

      <Button
        title={uploading ? t('upload.uploading') : t('upload.uploadCta')}
        style={{ marginTop: theme.spacing.lg }}
        onPress={handleUpload}
        disabled={uploading || images.length === 0}
      />
    </Screen>
  )
}
