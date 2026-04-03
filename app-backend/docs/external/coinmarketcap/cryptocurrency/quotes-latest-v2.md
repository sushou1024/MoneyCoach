https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyQuotesLatest

# Quotes Latest v2

GET https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest

Returns the latest market quote for 1 or more cryptocurrencies. Use the "convert" option to return market values in multiple fiat and cryptocurrency conversions in the same call.

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

 convert	
string
Optionally calculate market quotes in up to 120 currencies at once by passing a comma-separated list of cryptocurrency or fiat currency symbols. Each additional convert option beyond the first requires an additional call credit. A list of supported fiat options can be found here. Each conversion is returned in its own "quote" object.

 convert_id	
string
Optionally calculate market quotes by CoinMarketCap ID instead of symbol. This option is identical to convert outside of ID format. Ex: convert_id=1,2781 would replace convert=BTC,USD in your query. This parameter cannot be used when convert is used.

 aux	
string
Default: "num_market_pairs,cmc_rank,date_added,tags,platform,max_supply,circulating_supply,total_supply,is_active,is_fiat"
Optionally specify a comma-separated list of supplemental data fields to return. Pass num_market_pairs,cmc_rank,date_added,tags,platform,max_supply,circulating_supply,total_supply,market_cap_by_total_supply,volume_24h_reported,volume_7d,volume_7d_reported,volume_30d,volume_30d_reported,is_active,is_fiat to include all auxiliary fields.

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
      "is_active": 1,
      "is_fiat": 0,
      "circulating_supply": 17199862,
      "total_supply": 17199862,
      "max_supply": 21000000,
      "date_added": "2013-04-28T00:00:00.000Z",
      "num_market_pairs": 331,
      "cmc_rank": 1,
      "last_updated": "2018-08-09T21:56:28.000Z",
      "tags": [
        "mineable"
      ],
      "platform": null,
      "self_reported_circulating_supply": null,
      "self_reported_market_cap": null,
      "minted_market_cap": 1802955697670.94,
      "quote": {
        "USD": {
          "price": 6602.60701122,
          "volume_24h": 4314444687.5194,
          "volume_change_24h": -0.152774,
          "percent_change_1h": 0.988615,
          "percent_change_24h": 4.37185,
          "percent_change_7d": -12.1352,
          "percent_change_30d": -12.1352,
          "market_cap": 852164659250.2758,
          "market_cap_dominance": 51,
          "fully_diluted_market_cap": 952835089431.14,
          "last_updated": "2018-08-09T21:56:28.000Z"
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