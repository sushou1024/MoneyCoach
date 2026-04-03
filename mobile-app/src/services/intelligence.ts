import { apiRequest } from './api'
import { AssetBrief, MarketRegime } from '../types/api'

export async function fetchMarketRegime(token: string) {
  return apiRequest<MarketRegime>('/v1/intelligence/regime', { token })
}

export async function fetchAssetBrief(token: string, assetKey: string) {
  return apiRequest<AssetBrief>(`/v1/intelligence/assets/${encodeURIComponent(assetKey)}`, { token })
}
