import { useQuery } from '@tanstack/react-query'
import { useLocalSearchParams, useRouter } from 'expo-router'
import React, { useMemo, useState } from 'react'
import { Alert, FlatList, Image, Pressable, StyleSheet, Switch, Text, View } from 'react-native'

import { AssetLogo } from '../../src/components/AssetLogo'
import { Badge } from '../../src/components/Badge'
import { Button } from '../../src/components/Button'
import { Card } from '../../src/components/Card'
import { Input } from '../../src/components/Input'
import { Screen } from '../../src/components/Screen'
import { useEntitlement } from '../../src/hooks/useEntitlement'
import { useAuth } from '../../src/providers/AuthProvider'
import { useLocalization } from '../../src/providers/LocalizationProvider'
import { useTheme } from '../../src/providers/ThemeProvider'
import { fetchUploadBatch, reviewUploadBatch } from '../../src/services/uploads'
import type { UploadBatchReviewPayload } from '../../src/services/uploads'
import { useUploadDraftStore } from '../../src/stores/uploadDraft'
import { OCRAmbiguity, OCRAsset } from '../../src/types/api'
import { formatCurrency, formatNumber } from '../../src/utils/format'

type Candidate = OCRAmbiguity['candidates'][number]

type AssetEditState = {
  amount?: string
  avg_price?: string
  manual_value_display?: string
}

const STABLECOINS = new Set(['USDT', 'USDC', 'DAI', 'TUSD', 'BUSD', 'FDUSD', 'USDP', 'FRAX'])
const hkSymbolPattern = /^0*(\d{1,5})\.?HK$/i

function candidateLabel(candidate: Candidate) {
  const exchange = candidate.exchange_mic ? ` · ${candidate.exchange_mic}` : ''
  const name = candidate.name ? ` (${candidate.name})` : ''
  return `${candidate.symbol}${exchange} [${candidate.asset_type}]${name}`
}

function normalizeHKSymbol(symbol: string) {
  const trimmed = symbol.trim()
  const match = trimmed.match(hkSymbolPattern)
  if (!match) return trimmed
  return `${match[1]}.HK`
}

function getAssetIdentity(asset: OCRAsset) {
  const raw = asset.symbol ?? asset.symbol_raw
  const normalized = raw ? normalizeHKSymbol(raw) : asset.symbol_raw
  const name = asset.name?.trim()
  const exchangeMic = (asset.exchange_mic ?? '').toUpperCase()
  const isHongKongStock =
    asset.asset_type === 'stock' && (exchangeMic === 'XHKG' || normalized.toUpperCase().endsWith('.HK'))
  const showSymbolChip = Boolean(name && normalized) && !isHongKongStock
  const title = name && isHongKongStock ? `${name} (${normalized})` : name || normalized || asset.symbol_raw
  return {
    name,
    symbol: normalized,
    title,
    showSymbolChip,
  }
}

function parseNumber(value: string) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) return null
  return parsed
}

function errorCopy(code?: string) {
  switch (code) {
    case 'INVALID_IMAGE':
      return 'review.error.invalid'
    case 'UNSUPPORTED_ASSET_VIEW':
      return 'review.error.unsupported'
    case 'EXTRACTION_FAILURE':
      return 'review.error.extraction'
    default:
      return 'review.error.generic'
  }
}

export default function ReviewScreen() {
  const theme = useTheme()
  const styles = useMemo(() => createStyles(theme), [theme])
  const router = useRouter()
  const { t } = useLocalization()
  const { accessToken, userId } = useAuth()
  const entitlement = useEntitlement()
  const params = useLocalSearchParams<{ id?: string }>()
  const batchId = params.id ?? ''
  const { drafts, clearDraft } = useUploadDraftStore()
  const draftImages = drafts[batchId]?.images ?? []

  const [removedIds, setRemovedIds] = useState<string[]>([])
  const [assetEdits, setAssetEdits] = useState<Record<string, AssetEditState>>({})
  const [resolutions, setResolutions] = useState<Record<string, Candidate>>({})
  const [duplicateOverrides, setDuplicateOverrides] = useState<Record<string, boolean>>({})

  const query = useQuery({
    queryKey: ['upload-batch', userId, batchId],
    queryFn: async () => {
      if (!accessToken || !batchId) return null
      const resp = await fetchUploadBatch(accessToken, batchId)
      if (resp.error) {
        throw new Error(resp.error.message)
      }
      return resp.data ?? null
    },
    enabled: !!accessToken && !!batchId,
    retry: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    staleTime: 60 * 1000,
    refetchInterval: false,
  })

  const assets = useMemo(() => {
    if (!query.data || !('ocr_assets' in query.data)) return []
    return query.data.ocr_assets
  }, [query.data])
  const baseCurrency = useMemo(() => {
    if (!query.data || !('base_currency' in query.data)) return 'USD'
    return query.data.base_currency ?? 'USD'
  }, [query.data])
  const baseFxRateToUSD = useMemo(() => {
    if (!query.data || !('base_fx_rate_to_usd' in query.data)) return 1
    const rate = query.data.base_fx_rate_to_usd ?? 1
    return rate > 0 ? rate : 1
  }, [query.data])

  const ambiguityMap = useMemo(() => {
    if (!query.data || !('ambiguities' in query.data)) return {}
    return query.data.ambiguities.reduce<Record<string, OCRAmbiguity>>((acc, item) => {
      acc[item.symbol_raw] = item
      return acc
    }, {})
  }, [query.data])

  const unpricedAssets = useMemo(() => assets.filter((asset) => asset.value_usd_priced_draft == null), [assets])
  const hasUnpriced = unpricedAssets.length > 0
  const draftValueUSDByAsset = useMemo(() => {
    const values: Record<string, number | null> = {}
    for (const asset of assets) {
      if (removedIds.includes(asset.asset_id)) {
        values[asset.asset_id] = null
        continue
      }
      const edit = assetEdits[asset.asset_id]
      let rawUSD = asset.value_usd_priced_draft
      if (rawUSD == null && asset.value_display_draft != null) {
        rawUSD = asset.value_display_draft * baseFxRateToUSD
      }
      if (edit?.manual_value_display) {
        const parsed = parseNumber(edit.manual_value_display)
        if (parsed !== null) {
          rawUSD = parsed * baseFxRateToUSD
        }
      } else if (edit?.amount) {
        const parsed = parseNumber(edit.amount)
        if (parsed !== null && rawUSD != null && asset.amount > 0) {
          rawUSD = (rawUSD / asset.amount) * parsed
        }
      }
      values[asset.asset_id] = rawUSD ?? null
    }
    return values
  }, [assets, assetEdits, baseFxRateToUSD, removedIds])

  const draftTotalUSD = useMemo(() => {
    return Object.values(draftValueUSDByAsset).reduce<number>((sum, value) => sum + (value ?? 0), 0)
  }, [draftValueUSDByAsset])

  const lowValueThreshold = useMemo(() => {
    if (draftTotalUSD <= 0) return 0
    return draftTotalUSD * 0.01
  }, [draftTotalUSD])

  const showAvgPricePrompt = useMemo(() => {
    const eligible = assets.filter((asset) => {
      const value = asset.value_usd_priced_draft
      const assetType = asset.asset_type ?? ''
      const symbol = asset.symbol ?? asset.symbol_raw
      if (!value || value <= 0) return false
      if (assetType !== 'crypto' && assetType !== 'stock') return false
      if (STABLECOINS.has(symbol)) return false
      return true
    })
    if (!eligible.length) return false
    const total = eligible.reduce((sum, asset) => sum + (asset.value_usd_priced_draft ?? 0), 0)
    if (total <= 0) return false
    const topHoldings = eligible
      .sort((a, b) => (b.value_usd_priced_draft ?? 0) - (a.value_usd_priced_draft ?? 0))
      .slice(0, 3)
      .filter((asset) => (asset.value_usd_priced_draft ?? 0) >= total * 0.1)
    return topHoldings.some((asset) => {
      const edited = assetEdits[asset.asset_id]?.avg_price
      return !(edited && Number(edited) > 0) && !asset.avg_price
    })
  }, [assetEdits, assets])

  const toggleRemove = (assetId: string) => {
    setRemovedIds((prev) => (prev.includes(assetId) ? prev.filter((id) => id !== assetId) : [...prev, assetId]))
  }

  const updateEdit = (assetId: string, key: keyof AssetEditState, value: string) => {
    setAssetEdits((prev) => ({
      ...prev,
      [assetId]: {
        ...prev[assetId],
        [key]: value,
      },
    }))
  }

  const handleResolve = (symbolRaw: string, candidate: Candidate) => {
    setResolutions((prev) => ({ ...prev, [symbolRaw]: candidate }))
  }

  const handleDuplicateOverride = (imageId: string, value: boolean) => {
    setDuplicateOverrides((prev) => ({ ...prev, [imageId]: value }))
  }

  const buildEdits = (items: OCRAsset[]) => {
    const edits: UploadBatchReviewPayload['edits'] = []
    for (const asset of items) {
      if (removedIds.includes(asset.asset_id)) {
        edits.push({ asset_id: asset.asset_id, action: 'remove' })
        continue
      }
      const editState = assetEdits[asset.asset_id]
      if (!editState) continue
      const payload: UploadBatchReviewPayload['edits'][number] = { asset_id: asset.asset_id }
      if (editState.amount) {
        const parsed = parseNumber(editState.amount)
        if (parsed !== null && parsed > 0 && parsed !== asset.amount) {
          payload.amount = parsed
        }
      }
      if (editState.avg_price) {
        const parsed = parseNumber(editState.avg_price)
        if (parsed !== null && parsed >= 0) {
          payload.avg_price = parsed
        }
      }
      if (editState.manual_value_display) {
        const parsed = parseNumber(editState.manual_value_display)
        if (parsed !== null && parsed >= 0) {
          payload.manual_value_display = parsed
        }
      }
      if (Object.keys(payload).length > 1) {
        edits.push(payload)
      }
    }
    return edits
  }

  const buildResolutions = () => {
    return Object.entries(resolutions).map(([symbolRaw, candidate]) => ({
      symbol_raw: symbolRaw,
      asset_type: candidate.asset_type,
      symbol: candidate.symbol,
      asset_key: candidate.asset_key,
      exchange_mic: candidate.exchange_mic ?? '',
    }))
  }

  const buildDuplicateOverrides = () => {
    return Object.entries(duplicateOverrides).map(([imageId, include]) => ({
      image_id: imageId,
      include,
    }))
  }

  const handleConfirm = async () => {
    if (!accessToken || !batchId) return
    const remainingSymbols = new Set(
      assets.filter((asset) => !removedIds.includes(asset.asset_id)).map((asset) => asset.symbol_raw)
    )
    const unresolved = Object.keys(ambiguityMap).filter(
      (symbolRaw) => remainingSymbols.has(symbolRaw) && !resolutions[symbolRaw]
    )
    if (unresolved.length > 0) {
      Alert.alert(t('review.resolveTitle'), t('review.resolveBody', { symbols: unresolved.join(', ') }))
      return
    }
    const resp = await reviewUploadBatch(accessToken, batchId, {
      platform_overrides: [],
      resolutions: buildResolutions(),
      edits: buildEdits(assets),
      duplicate_overrides: buildDuplicateOverrides(),
    })
    if (resp.error) {
      Alert.alert(t('review.failTitle'), resp.error.message)
      return
    }
    clearDraft(batchId)
    const paid = ['active', 'grace'].includes(entitlement.data?.status ?? '')
    router.replace({
      pathname: paid ? '/(modals)/processing-paid' : '/(modals)/preview',
      params: paid ? { id: batchId } : { batch_id: batchId },
    })
  }

  if (query.isLoading || !query.data) {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('review.loading')}</Text>
      </Screen>
    )
  }

  if ('status' in query.data && query.data.status === 'failed') {
    const key = errorCopy(query.data.error_code)
    return (
      <Screen>
        <Card>
          <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>{t('review.errorTitle')}</Text>
          <Text style={{ color: theme.colors.muted, marginTop: theme.spacing.sm, fontFamily: theme.fonts.body }}>
            {t(key as any)}
          </Text>
          <Button
            title={t('review.retryCta')}
            style={{ marginTop: theme.spacing.md }}
            onPress={() => router.replace('/(modals)/upload')}
          />
        </Card>
      </Screen>
    )
  }

  if ('status' in query.data && query.data.status !== 'needs_review') {
    return (
      <Screen>
        <Text style={{ color: theme.colors.muted, fontFamily: theme.fonts.body }}>{t('review.processing')}</Text>
        <Button
          title={t('common.retry')}
          style={{ marginTop: theme.spacing.md }}
          onPress={() => router.replace({ pathname: '/(modals)/processing-ocr', params: { id: batchId } })}
        />
      </Screen>
    )
  }

  const images = 'images' in query.data ? query.data.images : []

  const renderAmbiguity = (symbolRaw: string, ambiguity: OCRAmbiguity) => {
    const selected = resolutions[symbolRaw]
    return (
      <View style={styles.ambiguityBlock}>
        <Text style={styles.ambiguityTitle}>{t('review.resolveLabel', { symbol: symbolRaw })}</Text>
        <View style={styles.ambiguityList}>
          {ambiguity.candidates.map((candidate) => {
            const isSelected =
              selected &&
              selected.asset_key === candidate.asset_key &&
              (selected.exchange_mic ?? '') === (candidate.exchange_mic ?? '')
            return (
              <Pressable
                key={`${candidate.asset_key}-${candidate.exchange_mic ?? 'na'}`}
                onPress={() => handleResolve(symbolRaw, candidate)}
                style={[styles.candidateCard, isSelected && styles.candidateCardSelected]}
              >
                <Text style={styles.candidateText}>{candidateLabel(candidate)}</Text>
              </Pressable>
            )
          })}
        </View>
        <Text style={styles.ambiguityNote}>{t('review.rememberNote')}</Text>
      </View>
    )
  }

  return (
    <Screen scroll>
      <Text style={styles.title}>{t('review.title')}</Text>
      <Text style={styles.subtitle}>{t('review.subtitle')}</Text>

      {showAvgPricePrompt ? (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>
            {t('review.avgPricePromptTitle')}
          </Text>
          <Text style={{ marginTop: theme.spacing.xs, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
            {t('review.avgPricePromptBody')}
          </Text>
        </Card>
      ) : null}

      {hasUnpriced ? (
        <Card style={{ marginTop: theme.spacing.md }}>
          <Text style={{ color: theme.colors.ink, fontFamily: theme.fonts.bodyBold }}>{t('review.unpricedTitle')}</Text>
          <Text style={{ marginTop: theme.spacing.xs, color: theme.colors.muted, fontFamily: theme.fonts.body }}>
            {t('review.unpricedBody', { currency: baseCurrency })}
          </Text>
        </Card>
      ) : null}

      <FlatList
        data={images}
        keyExtractor={(item) => item.image_id}
        horizontal
        showsHorizontalScrollIndicator={false}
        style={styles.scanStrip}
        renderItem={({ item }) => {
          const draftImage = draftImages.find((image) => image.imageId === item.image_id)
          const warnings = item.warnings ?? []
          return (
            <Card style={styles.scanCard}>
              {draftImage ? (
                <Image source={{ uri: draftImage.uri }} style={styles.scanImage} />
              ) : (
                <View style={styles.scanPlaceholder}>
                  <Text style={styles.scanPlaceholderText}>{item.image_id.slice(0, 6)}</Text>
                </View>
              )}
              {warnings.length ? <Text style={styles.scanWarning}>{warnings.join('\n')}</Text> : null}
              {warnings.length ? (
                <View style={styles.scanToggleRow}>
                  <Text style={styles.scanToggleLabel}>{t('review.duplicateToggle')}</Text>
                  <Switch
                    value={!!duplicateOverrides[item.image_id]}
                    onValueChange={(value) => handleDuplicateOverride(item.image_id, value)}
                  />
                </View>
              ) : null}
            </Card>
          )
        }}
      />

      <View style={styles.sectionHeader}>
        <Text style={styles.sectionTitle}>{t('review.detectedTitle')}</Text>
      </View>

      <FlatList
        data={assets}
        keyExtractor={(item) => item.asset_id}
        scrollEnabled={false}
        style={styles.assetList}
        renderItem={({ item }) => {
          const removed = removedIds.includes(item.asset_id)
          const edit = assetEdits[item.asset_id] ?? {}
          const ambiguity = removed ? null : ambiguityMap[item.symbol_raw]
          const isUnpriced = item.value_usd_priced_draft === null || item.value_usd_priced_draft === undefined
          const manualValue = edit.manual_value_display ?? item.manual_value_display ?? item.manual_value_usd
          const hasManual = manualValue !== null && manualValue !== undefined && manualValue !== ''
          const avgPriceDisplay =
            edit.avg_price && edit.avg_price.trim() ? parseNumber(edit.avg_price.trim()) : item.avg_price_display
          const avgPriceValue = avgPriceDisplay ?? item.avg_price
          const avgPriceCurrency = avgPriceDisplay !== null && avgPriceDisplay !== undefined ? baseCurrency : 'USD'
          const draftValueDisplay = item.value_display_draft
          const draftValue = draftValueDisplay ?? item.value_usd_priced_draft
          const draftCurrency = draftValueDisplay !== null && draftValueDisplay !== undefined ? baseCurrency : 'USD'
          const hasAvgPrice = avgPriceValue !== null && avgPriceValue !== undefined && avgPriceValue > 0
          const showDraftValue = !isUnpriced && draftValue !== null && draftValue !== undefined
          const draftValueUSD = draftValueUSDByAsset[item.asset_id]
          const { name: assetName, symbol: assetSymbol, title: assetTitle, showSymbolChip } = getAssetIdentity(item)
          const logoLabel = assetName || assetSymbol || item.symbol_raw
          const isIgnored =
            !removed &&
            lowValueThreshold > 0 &&
            draftValueUSD !== null &&
            draftValueUSD !== undefined &&
            draftValueUSD > 0
              ? draftValueUSD < lowValueThreshold
              : false
          return (
            <Card style={[styles.assetCard, removed && styles.assetCardRemoved]}>
              <View style={styles.assetHeader}>
                <AssetLogo uri={item.logo_url} label={logoLabel} size={44} />
                <View style={styles.assetHeaderText}>
                  <View style={styles.assetTitleRow}>
                    <Text style={styles.assetTitle}>{assetTitle}</Text>
                    {showSymbolChip ? (
                      <View style={styles.symbolChip}>
                        <Text style={styles.symbolChipText}>{assetSymbol}</Text>
                      </View>
                    ) : null}
                  </View>
                  <View style={styles.assetDetailRow}>
                    <Text style={styles.assetDetailText}>
                      {t('review.amountLabel', { amount: formatNumber(item.amount) })}
                    </Text>
                    {hasAvgPrice ? (
                      <Text style={styles.assetDetailText}>
                        {t('review.avgCostLabel', { value: formatCurrency(avgPriceValue, avgPriceCurrency) })}
                      </Text>
                    ) : null}
                  </View>
                </View>
              </View>

              <View style={styles.assetMetaRow}>
                {showDraftValue ? (
                  <Text style={styles.draftValueText}>
                    {t('review.draftLabel', { value: formatCurrency(draftValue ?? 0, draftCurrency) })}
                  </Text>
                ) : null}
                <View style={styles.badgeRowInline}>
                  {isUnpriced ? <Badge label={t('review.unpricedLabel')} tone="medium" /> : null}
                  {isIgnored ? (
                    <View style={styles.ignoredChip}>
                      <Text style={styles.ignoredChipText}>{t('review.ignoredLabel')}</Text>
                    </View>
                  ) : null}
                </View>
              </View>

              {hasManual ? <Text style={styles.noteText}>{t('review.manualIncludedNote')}</Text> : null}

              {ambiguity && renderAmbiguity(item.symbol_raw, ambiguity)}

              <View style={styles.fieldBlock}>
                <View style={[styles.fieldRow, styles.fieldRowFirst]}>
                  <View style={styles.field}>
                    <Text style={styles.fieldLabel}>{t('review.field.amount')}</Text>
                    <Input
                      value={edit.amount ?? String(item.amount)}
                      onChangeText={(value) => updateEdit(item.asset_id, 'amount', value)}
                      keyboardType="decimal-pad"
                      editable={!removed}
                    />
                  </View>
                  <View style={styles.field}>
                    <Text style={styles.fieldLabel}>{t('review.field.avgPrice', { currency: baseCurrency })}</Text>
                    <Input
                      value={
                        edit.avg_price ??
                        (item.avg_price_display
                          ? String(item.avg_price_display)
                          : item.avg_price
                            ? String(item.avg_price)
                            : '')
                      }
                      onChangeText={(value) => updateEdit(item.asset_id, 'avg_price', value)}
                      keyboardType="decimal-pad"
                      editable={!removed}
                    />
                  </View>
                </View>
                {isUnpriced ? (
                  <View style={styles.fieldRow}>
                    <View style={styles.field}>
                      <Text style={styles.fieldLabel}>{t('review.field.manualValue', { currency: baseCurrency })}</Text>
                      <Input
                        value={
                          edit.manual_value_display ??
                          (item.manual_value_display
                            ? String(item.manual_value_display)
                            : item.manual_value_usd
                              ? String(item.manual_value_usd)
                              : '')
                        }
                        onChangeText={(value) => updateEdit(item.asset_id, 'manual_value_display', value)}
                        keyboardType="decimal-pad"
                        editable={!removed}
                      />
                    </View>
                  </View>
                ) : null}
              </View>

              <View style={styles.actionRow}>
                <Pressable
                  onPress={() => toggleRemove(item.asset_id)}
                  style={[styles.removeButton, removed && styles.removeButtonActive]}
                >
                  <Text style={[styles.removeText, removed && styles.removeTextActive]}>
                    {removed ? t('review.undoRemove') : t('review.remove')}
                  </Text>
                </Pressable>
              </View>
            </Card>
          )
        }}
      />

      <Button title={t('review.confirm')} style={{ marginTop: theme.spacing.lg }} onPress={handleConfirm} />
    </Screen>
  )
}

const createStyles = (theme: ReturnType<typeof useTheme>) =>
  StyleSheet.create({
    title: {
      fontSize: 24,
      color: theme.colors.ink,
      fontFamily: theme.fonts.display,
    },
    subtitle: {
      marginTop: theme.spacing.sm,
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    scanStrip: {
      marginTop: theme.spacing.md,
    },
    scanCard: {
      marginRight: theme.spacing.sm,
      width: 160,
    },
    scanImage: {
      width: '100%',
      height: 90,
      borderRadius: theme.radius.sm,
    },
    scanPlaceholder: {
      width: '100%',
      height: 90,
      borderRadius: theme.radius.sm,
      backgroundColor: theme.colors.surfaceElevated,
      alignItems: 'center',
      justifyContent: 'center',
    },
    scanPlaceholderText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    scanWarning: {
      marginTop: theme.spacing.xs,
      color: theme.colors.danger,
      fontFamily: theme.fonts.body,
    },
    scanToggleRow: {
      flexDirection: 'row',
      alignItems: 'center',
      marginTop: theme.spacing.xs,
    },
    scanToggleLabel: {
      color: theme.colors.muted,
      flex: 1,
      fontFamily: theme.fonts.body,
    },
    sectionHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: 'space-between',
      marginTop: theme.spacing.lg,
    },
    sectionTitle: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      fontSize: 16,
    },
    assetList: {
      marginTop: theme.spacing.md,
    },
    assetCard: {
      marginBottom: theme.spacing.sm,
      paddingVertical: theme.spacing.md,
    },
    assetCardRemoved: {
      opacity: 0.55,
    },
    assetHeader: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: theme.spacing.md,
    },
    assetHeaderText: {
      flex: 1,
    },
    assetTitleRow: {
      flexDirection: 'row',
      alignItems: 'center',
      flexWrap: 'wrap',
      gap: theme.spacing.xs,
    },
    assetTitle: {
      fontSize: 16,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
      flexShrink: 1,
    },
    symbolChip: {
      paddingHorizontal: 8,
      paddingVertical: 2,
      borderRadius: 999,
      backgroundColor: theme.colors.surfaceElevated,
      borderWidth: 1,
      borderColor: theme.colors.border,
    },
    symbolChipText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 11,
      letterSpacing: 0.2,
    },
    assetDetailRow: {
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: theme.spacing.sm,
      marginTop: theme.spacing.xs,
    },
    assetDetailText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    assetMetaRow: {
      marginTop: theme.spacing.sm,
      flexDirection: 'row',
      alignItems: 'center',
      flexWrap: 'wrap',
      gap: theme.spacing.sm,
    },
    badgeRowInline: {
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: theme.spacing.xs,
      alignItems: 'center',
    },
    draftValueText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
    },
    ignoredChip: {
      paddingHorizontal: 8,
      paddingVertical: 3,
      borderRadius: 999,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: 'transparent',
    },
    ignoredChipText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyMedium,
      fontSize: 11,
      letterSpacing: 0.2,
    },
    noteText: {
      color: theme.colors.muted,
      marginTop: theme.spacing.xs,
      fontFamily: theme.fonts.body,
    },
    ambiguityBlock: {
      marginTop: theme.spacing.md,
    },
    ambiguityTitle: {
      fontSize: 14,
      color: theme.colors.ink,
      fontFamily: theme.fonts.bodyBold,
    },
    ambiguityList: {
      marginTop: theme.spacing.sm,
    },
    candidateCard: {
      padding: theme.spacing.sm,
      borderWidth: 1,
      borderColor: theme.colors.border,
      borderRadius: theme.radius.sm,
      marginBottom: theme.spacing.sm,
      backgroundColor: theme.colors.surface,
    },
    candidateCardSelected: {
      borderColor: theme.colors.accent,
      backgroundColor: theme.colors.accentSoft,
    },
    candidateText: {
      color: theme.colors.ink,
      fontFamily: theme.fonts.body,
    },
    ambiguityNote: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
    },
    fieldBlock: {
      marginTop: theme.spacing.md,
    },
    fieldRow: {
      flexDirection: 'row',
      gap: theme.spacing.sm,
      marginTop: theme.spacing.sm,
    },
    fieldRowFirst: {
      marginTop: 0,
    },
    field: {
      flex: 1,
    },
    fieldLabel: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.body,
      marginBottom: theme.spacing.xs,
    },
    actionRow: {
      marginTop: theme.spacing.md,
      flexDirection: 'row',
      justifyContent: 'flex-end',
    },
    removeButton: {
      paddingVertical: 6,
      paddingHorizontal: theme.spacing.sm,
      borderRadius: theme.radius.md,
      borderWidth: 1,
      borderColor: theme.colors.border,
      backgroundColor: theme.colors.surfaceElevated,
    },
    removeButtonActive: {
      borderColor: theme.colors.accent,
      backgroundColor: theme.colors.accentSoft,
    },
    removeText: {
      color: theme.colors.muted,
      fontFamily: theme.fonts.bodyBold,
    },
    removeTextActive: {
      color: theme.colors.accent,
    },
  })
