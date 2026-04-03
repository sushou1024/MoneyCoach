https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyPriceperformancestatsLatest

# Price Performance Stats v2

GET https://pro-api.coinmarketcap.com/v2/cryptocurrency/price-performance-stats/latest

Returns price performance statistics for one or more cryptocurrencies including launch price ROI and all-time high / all-time low. Stats are returned for an all_time period by default. UTC yesterday and a number of rolling time periods may be requested using the time_period parameter. Utilize the convert parameter to translate values into multiple fiats or cryptocurrencies using historical rates.

Cache / Update frequency: Every 60 seconds.

## Query Parameters
 id	
string
One or more comma-separated cryptocurrency CoinMarketCap IDs. Example: 1,2

 slug	
string
Alternatively pass a comma-separated list of cryptocurrency slugs. Example: "bitcoin,ethereum"

 symbol	
string
Alternatively pass one or more comma-separated cryptocurrency symbols. Example: "BTC,ETH". At least one "id" or "slug" or "symbol" is required for this request.

 time_period	
string
Default: "all_time"
Specify one or more comma-delimited time periods to return stats for. all_time is the default. Pass all_time,yesterday,24h,7d,30d,90d,365d to return all supported time periods. All rolling periods have a rolling close time of the current request time. For example 24h would have a close time of now and an open time of 24 hours before now. Please note: yesterday is a UTC period and currently does not currently support high and low timestamps.

 convert	
string
Optionally calculate quotes in up to 120 currencies at once by passing a comma-separated list of cryptocurrency or fiat currency symbols. Each additional convert option beyond the first requires an additional call credit. A list of supported fiat options can be found here. Each conversion is returned in its own "quote" object.

 convert_id	
string
Optionally calculate quotes by CoinMarketCap ID instead of symbol. This option is identical to convert outside of ID format. Ex: convert_id=1,2781 would replace convert=BTC,USD in your query. This parameter cannot be used when convert is used.

 skip_invalid	
boolean
Default: true
Pass true to relax request validation rules. When requesting records on multiple cryptocurrencies an error is returned if no match is found for 1 or more requested cryptocurrencies. If set to true, invalid lookups will be skipped allowing valid cryptocurrencies to still be returned.

## Response

{
  "data": {
    "1": {
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "slug": "bitcoin",
      "last_updated": "2019-08-22T01:51:32.000Z",
      "periods": {
        "all_time": {
          "open_timestamp": "2013-04-28T00:00:00.000Z",
          "high_timestamp": "2017-12-17T12:19:14.000Z",
          "low_timestamp": "2013-07-05T18:56:01.000Z",
          "close_timestamp": "2019-08-22T01:52:18.613Z",
          "quote": {
            "USD": {
              "open": 135.3000030517578,
              "open_timestamp": "2013-04-28T00:00:00.000Z",
              "high": 20088.99609375,
              "high_timestamp": "2017-12-17T12:19:14.000Z",
              "low": 65.5260009765625,
              "low_timestamp": "2013-07-05T18:56:01.000Z",
              "close": 65.5260009765625,
              "close_timestamp": "2019-08-22T01:52:18.618Z",
              "percent_change": 7223.718930042746,
              "price_change": 9773.691932798241
            }
          }
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