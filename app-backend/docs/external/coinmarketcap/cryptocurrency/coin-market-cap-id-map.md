https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyMap

# CoinMarketCap ID Map

GET https://pro-api.coinmarketcap.com/v1/cryptocurrency/map

Returns a mapping of all cryptocurrencies to unique CoinMarketCap ids. Per our Best Practices we recommend utilizing CMC ID instead of cryptocurrency symbols to securely identify cryptocurrencies with our other endpoints and in your own application logic. Each cryptocurrency returned includes typical identifiers such as name, symbol, and token_address for flexible mapping to id.

By default this endpoint returns cryptocurrencies that have actively tracked markets on supported exchanges. You may receive a map of all inactive cryptocurrencies by passing listing_status=inactive. You may also receive a map of registered cryptocurrency projects that are listed but do not yet meet methodology requirements to have tracked markets via listing_status=untracked. Please review our methodology documentation for additional details on listing states.

Cryptocurrencies returned include first_historical_data and last_historical_data timestamps to conveniently reference historical date ranges available to query with historical time-series data endpoints. You may also use the aux parameter to only include properties you require to slim down the payload if calling this endpoint frequently.

Cache / Update frequency: Mapping data is updated only as needed, every 30 seconds.

## Query Parameters

 listing_status	
string
Default: "active"
Only active cryptocurrencies are returned by default. Pass inactive to get a list of cryptocurrencies that are no longer active. Pass untracked to get a list of cryptocurrencies that are listed but do not yet meet methodology requirements to have tracked markets available. You may pass one or more comma-separated values.

 start	
integer >= 1
Default: 1
Optionally offset the start (1-based index) of the paginated list of items to return.

 limit	
integer [ 1 .. 5000 ]
Optionally specify the number of results to return. Use this parameter and the "start" parameter to determine your own pagination size.

 sort	
string
Default: "id"
"cmc_rank""id"
What field to sort the list of cryptocurrencies by.

 symbol	
string
Optionally pass a comma-separated list of cryptocurrency symbols to return CoinMarketCap IDs for. If this option is passed, other options will be ignored.

 aux	
string
Default: "platform,first_historical_data,last_historical_data,is_active"
Optionally specify a comma-separated list of supplemental data fields to return. Pass platform,first_historical_data,last_historical_data,is_active,status to include all auxiliary fields.

## Response

{
  "data": [
    {
      "id": 1,
      "rank": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "slug": "bitcoin",
      "is_active": 1,
      "first_historical_data": "2013-04-28T18:47:21.000Z",
      "last_historical_data": "2020-05-05T20:44:01.000Z",
      "platform": null
    },
    {
      "id": 1839,
      "rank": 3,
      "name": "Binance Coin",
      "symbol": "BNB",
      "slug": "binance-coin",
      "is_active": 1,
      "first_historical_data": "2017-07-25T04:30:05.000Z",
      "last_historical_data": "2020-05-05T20:44:02.000Z",
      "platform": {
        "id": 1027,
        "name": "Ethereum",
        "symbol": "ETH",
        "slug": "ethereum",
        "token_address": "0xB8c77482e45F1F44dE1745F52C74426C631bDD52"
      }
    },
    {
      "id": 825,
      "rank": 5,
      "name": "Tether",
      "symbol": "USDT",
      "slug": "tether",
      "is_active": 1,
      "first_historical_data": "2015-02-25T13:34:26.000Z",
      "last_historical_data": "2020-05-05T20:44:01.000Z",
      "platform": {
        "id": 1027,
        "name": "Ethereum",
        "symbol": "ETH",
        "slug": "ethereum",
        "token_address": "0xdac17f958d2ee523a2206206994597c13d831ec7"
      }
    }
  ],
  "status": {
    "timestamp": "2018-06-02T22:51:28.209Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 10,
    "credit_count": 1
  }
}