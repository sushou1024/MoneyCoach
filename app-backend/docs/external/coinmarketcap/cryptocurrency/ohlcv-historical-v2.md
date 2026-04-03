https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyOhlcvHistorical

# Market Pairs Latest v2

GET https://pro-api.coinmarketcap.com/v2/cryptocurrency/ohlcv/historical

Lists all active market pairs that CoinMarketCap tracks for a given cryptocurrency or fiat currency. All markets with this currency as the pair base or pair quote will be returned. The latest price and volume information is returned for each market. Use the "convert" option to return market values in multiple fiat and cryptocurrency conversions in the same call.

Cache / Update frequency: Every 1 minute.

## Query Parameters
 id	
string
A cryptocurrency or fiat currency by CoinMarketCap ID to list market pairs for. Example: "1"

 slug	
string
Alternatively pass a cryptocurrency by slug. Example: "bitcoin"

 symbol	
string
Alternatively pass a cryptocurrency by symbol. Fiat currencies are not supported by this field. Example: "BTC". A single cryptocurrency "id", "slug", or "symbol" is required.

 start	
integer >= 1
Default: 1
Optionally offset the start (1-based index) of the paginated list of items to return.

 limit	
integer [ 1 .. 5000 ]
Default: 100
Optionally specify the number of results to return. Use this parameter and the "start" parameter to determine your own pagination size.

 sort_dir	
string
Default: "desc"
Valid values: "asc""desc"
Optionally specify the sort direction of markets returned.

 sort	
string
Default: "volume_24h_strict"
Valid values: "volume_24h_strict""cmc_rank""cmc_rank_advanced""effective_liquidity""market_score""market_reputation"
Optionally specify the sort order of markets returned. By default we return a strict sort on 24 hour reported volume. Pass cmc_rank to return a CMC methodology based sort where markets with excluded volumes are returned last.

 aux	
string
Default: "num_market_pairs,category,fee_type"
Optionally specify a comma-separated list of supplemental data fields to return. Pass num_market_pairs,category,fee_type,market_url,currency_name,currency_slug,price_quote,notice,cmc_rank,effective_liquidity,market_score,market_reputation to include all auxiliary fields.

 matched_id	
string
Optionally include one or more fiat or cryptocurrency IDs to filter market pairs by. For example ?id=1&matched_id=2781 would only return BTC markets that matched: "BTC/USD" or "USD/BTC". This parameter cannot be used when matched_symbol is used.

 matched_symbol	
string
Optionally include one or more fiat or cryptocurrency symbols to filter market pairs by. For example ?symbol=BTC&matched_symbol=USD would only return BTC markets that matched: "BTC/USD" or "USD/BTC". This parameter cannot be used when matched_id is used.

 category	
string
Default: "all"
Valid values: "all""spot""derivatives""otc""perpetual"
The category of trading this market falls under. Spot markets are the most common but options include derivatives and OTC.

 fee_type	
string
Default: "all"
Valid values: "all""percentage""no-fees""transactional-mining""unknown"
The fee type the exchange enforces for this market.

 convert	
string
Optionally calculate market quotes in up to 120 currencies at once by passing a comma-separated list of cryptocurrency or fiat currency symbols. Each additional convert option beyond the first requires an additional call credit. A list of supported fiat options can be found here. Each conversion is returned in its own "quote" object.

 convert_id	
string
Optionally calculate market quotes by CoinMarketCap ID instead of symbol. This option is identical to convert outside of ID format. Ex: convert_id=1,2781 would replace convert=BTC,USD in your query. This parameter cannot be used when convert is used.

## Response

{
  "data": {
    "id": 1,
    "name": "Bitcoin",
    "symbol": "BTC",
    "quotes": [
      {
        "time_open": "2019-01-02T00:00:00.000Z",
        "time_close": "2019-01-02T23:59:59.999Z",
        "time_high": "2019-01-02T03:53:00.000Z",
        "time_low": "2019-01-02T02:43:00.000Z",
        "quote": {
          "USD": {
            "open": 3849.21640853,
            "high": 3947.9812729,
            "low": 3817.40949569,
            "close": 3943.40933686,
            "volume": 5244856835.70851,
            "market_cap": 68849856731.6738,
            "timestamp": "2019-01-02T23:59:59.999Z"
          }
        }
      },
      {
        "time_open": "2019-01-03T00:00:00.000Z",
        "time_close": "2019-01-03T23:59:59.999Z",
        "time_high": "2019-01-02T03:53:00.000Z",
        "time_low": "2019-01-02T02:43:00.000Z",
        "quote": {
          "USD": {
            "open": 3931.04863841,
            "high": 3935.68513083,
            "low": 3826.22287069,
            "close": 3836.74131867,
            "volume": 4530215218.84018,
            "market_cap": 66994920902.7202,
            "timestamp": "2019-01-03T23:59:59.999Z"
          }
        }
      }
    ]
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