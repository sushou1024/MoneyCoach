import { apiRequest } from './api'

export type OHLCVPoint = [number, number, number, number, number, number]

export async function fetchOHLCV(
  token: string,
  params: {
    asset_key?: string
    asset_type?: 'crypto' | 'stock'
    symbol?: string
    interval: '4h' | '1d'
    start?: string
    end?: string
  }
) {
  const query = new URLSearchParams()
  if (params.asset_key) query.set('asset_key', params.asset_key)
  if (params.asset_type) query.set('asset_type', params.asset_type)
  if (params.symbol) query.set('symbol', params.symbol)
  query.set('interval', params.interval)
  if (params.start) query.set('start', params.start)
  if (params.end) query.set('end', params.end)
  return apiRequest<{ series: OHLCVPoint[]; quote_currency: string }>(`/v1/market-data/ohlcv?${query.toString()}`, {
    token,
  })
}
