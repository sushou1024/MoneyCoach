# Coins List with Market Data

> This endpoint allows you to **query all the supported coins with price, market cap, volume and market related data**

<Tip>
  ### Tips

  * You can retrieve specific coins using their unique `ids`, `names`, or `symbols` instead of returning the whole list.
  * To filter results based on the coin's category, use the `category` param (refer to [`/coins/categories/list`](/reference/coins-categories-list) for available categories).
  * Use the `per_page` and `page` params to manage the number of results you receive and navigate through the data.
</Tip>

<Note>
  ### Note

  * When multiple lookup params are provided, the following priority order is applied: `category` (highest) > `ids` > `names` > `symbols` (lowest).
  * When searching by `name`, you need to URL-encode any spaces (e.g. "Binance Coin" becomes "Binance%20Coin").
  * The `include_tokens=all` param is exclusively for use with the `symbols` lookup and is limited to maximum of 50 symbols per request.
  * Wildcard searches are not supported for lookup params (`ids`, `names`, `symbols`).
  * Cache/Update Frequency: every 45 seconds for Pro API (Analyst, Lite, Pro, Enterprise).
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /coins/markets
openapi: 3.0.0
info:
  title: CoinGecko Pro API
  version: 3.0.0
servers:
  - url: https://pro-api.coingecko.com/api/v3
security:
  - apiKeyAuth: []
  - apiKeyQueryParam: []
paths:
  /coins/markets:
    get:
      tags:
        - Coins
      summary: Coins List with Market Data
      description: >-
        This endpoint allows you to **query all the supported coins with price,
        market cap, volume and market related data**
      operationId: coins-markets
      parameters:
        - name: vs_currency
          in: query
          description: |-
            target currency of coins and market data 
             *refers to [`/simple/supported_vs_currencies`](/reference/simple-supported-currencies).
          required: true
          schema:
            type: string
            example: usd
            default: usd
        - name: ids
          in: query
          description: |-
            coins' IDs, comma-separated if querying more than 1 coin. 
             *refers to [`/coins/list`](/reference/coins-list).
          required: false
          schema:
            type: string
            default: bitcoin
          examples:
            one value:
              value: bitcoin
            multiple values:
              value: bitcoin,ethereum
        - name: names
          in: query
          description: coins' names, comma-separated if querying more than 1 coin.
          required: false
          schema:
            type: string
            default: Bitcoin
          examples:
            one value:
              value: Bitcoin
            multiple values:
              value: Bitcoin,Ethereum
        - name: symbols
          in: query
          description: coins' symbols, comma-separated if querying more than 1 coin.
          required: false
          schema:
            type: string
            default: btc
          examples:
            one value:
              value: btc
            multiple values:
              value: btc,eth
        - name: include_tokens
          in: query
          description: |-
            for `symbols` lookups, specify `all` to include all matching tokens 
             Default `top` returns top-ranked tokens (by market cap or volume)
          required: false
          schema:
            type: string
            enum:
              - top
              - all
        - name: category
          in: query
          description: |-
            filter based on coins' category 
             *refers to [`/coins/categories/list`](/reference/coins-categories-list).
          required: false
          schema:
            type: string
            example: layer-1
            default: layer-1
        - name: order
          in: query
          description: 'sort result by field, default: market_cap_desc'
          required: false
          schema:
            type: string
            enum:
              - market_cap_asc
              - market_cap_desc
              - volume_asc
              - volume_desc
              - id_asc
              - id_desc
        - name: per_page
          in: query
          description: |-
            total results per page, default: 100 
             Valid values: 1...250
          required: false
          schema:
            type: number
        - name: page
          in: query
          description: 'page through results, default: 1'
          required: false
          schema:
            type: number
        - name: sparkline
          in: query
          description: 'include sparkline 7 days data, default: false'
          required: false
          schema:
            type: boolean
        - name: price_change_percentage
          in: query
          description: >-
            include price change percentage timeframe, comma-separated if query
            more than 1 timeframe 
             Valid values: 1h, 24h, 7d, 14d, 30d, 200d, 1y
          required: false
          schema:
            type: string
            default: 1h
          examples:
            one value:
              value: 1h
            multiple values:
              value: 1h,24h,7d
        - name: locale
          in: query
          description: 'language background, default: en'
          required: false
          schema:
            type: string
            enum:
              - ar
              - bg
              - cs
              - da
              - de
              - el
              - en
              - es
              - fi
              - fr
              - he
              - hi
              - hr
              - hu
              - id
              - it
              - ja
              - ko
              - lt
              - nl
              - 'no'
              - pl
              - pt
              - ro
              - ru
              - sk
              - sl
              - sv
              - th
              - tr
              - uk
              - vi
              - zh
              - zh-tw
        - name: precision
          in: query
          description: decimal place for currency price value
          required: false
          schema:
            type: string
            enum:
              - full
              - '0'
              - '1'
              - '2'
              - '3'
              - '4'
              - '5'
              - '6'
              - '7'
              - '8'
              - '9'
              - '10'
              - '11'
              - '12'
              - '13'
              - '14'
              - '15'
              - '16'
              - '17'
              - '18'
      responses:
        '200':
          description: List all coins with market data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CoinsMarkets'
components:
  schemas:
    CoinsMarkets:
      type: array
      items:
        $ref: '#/components/schemas/CoinsMarketsItem'
      example:
        - id: bitcoin
          symbol: btc
          name: Bitcoin
          image: >-
            <https://assets.coingecko.com/coins/images/1/large/bitcoin.png?1696501400>
          current_price: 70187
          market_cap: 1381651251183
          market_cap_rank: 1
          fully_diluted_valuation: 1474623675796
          total_volume: 20154184933
          high_24h: 70215
          low_24h: 68060
          price_change_24h: 2126.88
          price_change_percentage_24h: 3.12502
          market_cap_change_24h: 44287678051
          market_cap_change_percentage_24h: 3.31157
          circulating_supply: 19675987
          total_supply: 21000000
          max_supply: 21000000
          ath: 73738
          ath_change_percentage: -4.77063
          ath_date: '2024-03-14T07:10:36.635Z'
          atl: 67.81
          atl_change_percentage: 103455.83335
          atl_date: '2013-07-06T00:00:00.000Z'
          roi: null
          last_updated: '2024-04-07T16:49:31.736Z'
    CoinsMarketsItem:
      type: object
      properties:
        id:
          type: string
          description: coin ID
        symbol:
          type: string
          description: coin symbol
        name:
          type: string
          description: coin name
        image:
          type: string
          description: coin image url
        current_price:
          type: number
          nullable: true
          description: coin current price in currency
        market_cap:
          type: number
          nullable: true
          description: coin market cap in currency
        market_cap_rank:
          type: number
          nullable: true
          description: coin rank by market cap
        fully_diluted_valuation:
          type: number
          nullable: true
          description: coin fully diluted valuation (fdv) in currency
        total_volume:
          type: number
          nullable: true
          description: coin total trading volume in currency
        high_24h:
          type: number
          nullable: true
          description: coin 24hr price high in currency
        low_24h:
          type: number
          nullable: true
          description: coin 24hr price low in currency
        price_change_24h:
          type: number
          nullable: true
          description: coin 24hr price change in currency
        price_change_percentage_24h:
          type: number
          nullable: true
          description: coin 24hr price change in percentage
        market_cap_change_24h:
          type: number
          nullable: true
          description: coin 24hr market cap change in currency
        market_cap_change_percentage_24h:
          type: number
          nullable: true
          description: coin 24hr market cap change in percentage
        circulating_supply:
          type: number
          nullable: true
          description: coin circulating supply
        total_supply:
          type: number
          nullable: true
          description: coin total supply
        max_supply:
          type: number
          nullable: true
          description: coin max supply
        ath:
          type: number
          nullable: true
          description: coin all time high (ATH) in currency
        ath_change_percentage:
          type: number
          nullable: true
          description: coin all time high (ATH) change in percentage
        ath_date:
          type: string
          format: date-time
          nullable: true
          description: coin all time high (ATH) date
        atl:
          type: number
          nullable: true
          description: coin all time low (atl) in currency
        atl_change_percentage:
          type: number
          nullable: true
          description: coin all time low (atl) change in percentage
        atl_date:
          type: string
          format: date-time
          nullable: true
          description: coin all time low (atl) date
        roi:
          type: object
          nullable: true
          allOf:
            - $ref: '#/components/schemas/RoiData'
          description: return on investment data
        last_updated:
          type: string
          format: date-time
          description: coin last updated timestamp
    RoiData:
      type: object
      properties:
        times:
          type: number
          description: ROI multiplier
        currency:
          type: string
          description: ROI currency
        percentage:
          type: number
          description: ROI percentage
      required:
        - times
        - currency
        - percentage
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: x-cg-pro-api-key
    apiKeyQueryParam:
      type: apiKey
      in: query
      name: x_cg_pro_api_key

````

---

> To find navigation and other pages in this documentation, fetch the llms.txt file at: https://docs.coingecko.com/llms.txt