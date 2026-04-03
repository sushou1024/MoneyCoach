https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyOhlcvLatest

# OHLCV Latest v2

GET https://pro-api.coinmarketcap.com/v2/cryptocurrency/ohlcv/latest

Returns the latest OHLCV (Open, High, Low, Close, Volume) market values for one or more cryptocurrencies for the current UTC day. Since the current UTC day is still active these values are updated frequently. You can find the final calculated OHLCV values for the last completed UTC day along with all historic days using /cryptocurrency/ohlcv/historical.

Cache / Update frequency: Every 10 minutes. Additional OHLCV intervals and 1 minute updates will be available in the future.

## Query Parameters
 id	
string
One or more comma-separated cryptocurrency CoinMarketCap IDs. Example: 1,2

 symbol	
string
Alternatively pass one or more comma-separated cryptocurrency symbols. Example: "BTC,ETH". At least one "id" or "symbol" is required.

 convert	
string
Optionally calculate market quotes in up to 120 currencies at once by passing a comma-separated list of cryptocurrency or fiat currency symbols. Each additional convert option beyond the first requires an additional call credit. A list of supported fiat options can be found here. Each conversion is returned in its own "quote" object.

 convert_id	
string
Optionally calculate market quotes by CoinMarketCap ID instead of symbol. This option is identical to convert outside of ID format. Ex: convert_id=1,2781 would replace convert=BTC,USD in your query. This parameter cannot be used when convert is used.

 skip_invalid	
boolean
Default: true
Pass true to relax request validation rules. When requesting records on multiple cryptocurrencies an error is returned if any invalid cryptocurrencies are requested or a cryptocurrency does not have matching records in the requested timeframe. If set to true, invalid lookups will be skipped allowing valid cryptocurrencies to still be returned.

## Response

{
  "data": {
    "1": {
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "last_updated": "2018-09-10T18:54:00.000Z",
      "time_open": "2018-09-10T00:00:00.000Z",
      "time_close": null,
      "time_high": "2018-09-10T00:00:00.000Z",
      "time_low": "2018-09-10T00:00:00.000Z",
      "quote": {
        "USD": {
          "open": 6301.57,
          "high": 6374.98,
          "low": 6292.76,
          "close": 6308.76,
          "volume": 3786450000,
          "last_updated": "2018-09-10T18:54:00.000Z"
        }
      }
    }
  },
  "status": {
    "timestamp": "2026-01-01T19:49:23.887Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 10,
    "credit_count": 1,
    "notice": ""
  }
}