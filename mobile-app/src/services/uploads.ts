import * as Crypto from 'expo-crypto'
import { File } from 'expo-file-system'
import { Platform } from 'react-native'

import { apiRequest } from './api'
import { UploadBatchComplete, UploadBatchCreateResponse, UploadBatchNeedsReview } from '../types/api'

export interface UploadImageMeta {
  file_name: string
  mime_type: string
  size_bytes: number
  uri: string
}

export interface UploadBatchReviewPayload {
  platform_overrides: { image_id: string; platform_guess: string }[]
  resolutions: {
    symbol_raw: string
    asset_type: string
    symbol: string
    asset_key: string
    exchange_mic: string
  }[]
  edits: {
    asset_id: string
    action?: string
    symbol?: string
    asset_type?: string
    amount?: number
    value_from_screenshot?: number
    display_currency?: string
    avg_price?: number
    manual_value_display?: number
  }[]
  duplicate_overrides: { image_id: string; include: boolean }[]
}

export async function createUploadBatch(
  token: string,
  purpose: 'holdings' | 'trade_slip',
  images: UploadImageMeta[],
  deviceTimezone: string
) {
  return apiRequest<UploadBatchCreateResponse>('/v1/upload-batches', {
    method: 'POST',
    token,
    body: {
      purpose,
      image_count: images.length,
      images: images.map((img) => ({
        file_name: img.file_name,
        mime_type: img.mime_type,
        size_bytes: img.size_bytes,
      })),
      device_timezone: deviceTimezone,
    },
  })
}

async function readUriBlob(uri: string) {
  const response = await fetch(uri)
  if (!response.ok) {
    throw new Error('Failed to read image data')
  }
  return response.blob()
}

export async function uploadImage(uploadUrl: string, headers: Record<string, string>, uri: string) {
  if (Platform.OS === 'web') {
    const blob = await readUriBlob(uri)
    const response = await fetch(uploadUrl, {
      method: 'PUT',
      headers,
      body: blob,
    })
    return response.status
  }
  const file = new File(uri)
  const bytes = await file.bytes()
  const body =
    bytes.byteOffset === 0 && bytes.byteLength === bytes.buffer.byteLength
      ? bytes.buffer
      : bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength)
  const response = await fetch(uploadUrl, {
    method: 'PUT',
    headers,
    body,
  })
  return response.status
}

function bufferToHex(buffer: ArrayBuffer) {
  const bytes = new Uint8Array(buffer)
  let hex = ''
  for (const byte of bytes) {
    hex += byte.toString(16).padStart(2, '0')
  }
  return hex
}

export async function computeFileChecksum(uri: string) {
  let bytes: Uint8Array<ArrayBuffer>
  if (Platform.OS === 'web') {
    const response = await fetch(uri)
    if (!response.ok) {
      throw new Error('Failed to read image data')
    }
    const buffer = await response.arrayBuffer()
    bytes = new Uint8Array(buffer)
  } else {
    const file = new File(uri)
    bytes = await file.bytes()
  }
  const digest = await Crypto.digest(Crypto.CryptoDigestAlgorithm.SHA256, bytes)
  return bufferToHex(digest)
}

export async function computeClientChecksum(images: { image_id: string; uri: string; size_bytes: number }[]) {
  const hashes = await Promise.all(images.map((image) => computeFileChecksum(image.uri)))
  const manifest = images.map((image, index) => `${image.image_id}:${hashes[index]}:${image.size_bytes}`).join('\n')
  const manifestHash = await Crypto.digestStringAsync(Crypto.CryptoDigestAlgorithm.SHA256, manifest, {
    encoding: Crypto.CryptoEncoding.HEX,
  })
  return `sha256:${manifestHash}`
}

export async function completeUploadBatch(token: string, batchId: string, imageIds: string[], clientChecksum: string) {
  return apiRequest<{ status: string; poll_after_ms?: number }>(`/v1/upload-batches/${batchId}/complete`, {
    method: 'POST',
    token,
    body: {
      image_ids: imageIds,
      client_checksum: clientChecksum,
    },
  })
}

export async function reviewUploadBatch(token: string, batchId: string, payload: UploadBatchReviewPayload) {
  return apiRequest<{ status: string; poll_after_ms?: number }>(`/v1/upload-batches/${batchId}/review`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function fetchUploadBatch(token: string, batchId: string) {
  return apiRequest<UploadBatchNeedsReview | UploadBatchComplete | { status: string; error_code?: string }>(
    `/v1/upload-batches/${batchId}`,
    { token }
  )
}
