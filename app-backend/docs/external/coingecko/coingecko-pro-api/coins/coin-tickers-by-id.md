# Coin Tickers by ID

> This endpoint allows you to **query the coin tickers on both centralized exchange (CEX) and decentralized exchange (DEX) based on a particular coin ID**

<Tip>
  ### Tips

  * You may obtain the coin ID (API ID) via several ways:
    * refers to respective coin page and find 'API ID'.
    * refers to [`/coins/list`](/reference/coins-list) endpoint.
    * refers to google sheets [here](https://docs.google.com/spreadsheets/d/1wTTuxXt8n9q7C4NDXqQpI3wpKu1_5bGVmP9Xz0XGSyU/edit?usp=sharing).
  * You may specify the `exchange_ids` if you want to retrieve tickers for specific exchange only.
  * You may include values such as  `page` to specify which page of responses you would like to show.
  * You may also flag to include more data such as exchange logo and depth.
</Tip>

<Note>
  ### Note

  * The tickers are paginated to 100 items.
  * When `dex_pair_format=symbol`, the DEX pair `base` and `target` are displayed in symbol format (e.g. `WETH`, `USDC`) instead of as contract addresses.
  * When order is sorted by `volume`, ***converted\_volume*** will be used instead of ***volume***.
  * Cache / Update Frequency:  every 2 minutes for all the API plans.
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /coins/{id}/tickers
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
  /coins/{id}/tickers:
    get:
      tags:
        - Coins
      summary: Coin Tickers by ID
      description: >-
        This endpoint allows you to **query the coin tickers on both centralized
        exchange (CEX) and decentralized exchange (DEX) based on a particular
        coin ID**
      operationId: coins-id-tickers
      parameters:
        - name: id
          in: path
          description: |-
            coin ID 
             *refers to [`/coins/list`](/reference/coins-list).
          required: true
          schema:
            type: string
            example: bitcoin
            default: bitcoin
        - name: exchange_ids
          in: query
          description: |-
            exchange ID 
             *refers to [`/exchanges/list`](/reference/exchanges-list).
          required: false
          schema:
            type: string
            example: binance
            default: binance
        - name: include_exchange_logo
          in: query
          description: 'include exchange logo, default: false'
          required: false
          schema:
            type: boolean
        - name: page
          in: query
          description: page through results
          required: false
          schema:
            type: number
        - name: order
          in: query
          description: 'use this to sort the order of responses, default: trust_score_desc'
          required: false
          schema:
            type: string
            enum:
              - trust_score_desc
              - trust_score_asc
              - volume_desc
              - volume_asc
        - name: depth
          in: query
          description: >-
            include 2% orderbook depth, ie. `cost_to_move_up_usd` and
            `cost_to_move_down_usd` 
             Default: false
          required: false
          schema:
            type: boolean
        - name: dex_pair_format
          in: query
          description: >-
            set to `symbol` to display DEX pair base and target as symbols,
            default: `contract_address`
          required: false
          schema:
            type: string
            enum:
              - contract_address
              - symbol
      responses:
        '200':
          description: Get coin tickers
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CoinsTickers'
components:
  schemas:
    CoinsTickers:
      type: object
      properties:
        name:
          type: string
          description: coin name
        tickers:
          type: array
          description: list of tickers
          items:
            type: object
            properties:
              base:
                type: string
                description: coin ticker base currency
              target:
                type: string
                description: coin ticker target currency
              market:
                type: object
                description: coin ticker exchange
                properties:
                  name:
                    type: string
                    description: exchange name
                  identifier:
                    type: string
                    description: exchange identifier
                  has_trading_incentive:
                    type: boolean
                    description: exchange trading incentive
                  logo:
                    type: string
                    description: exchange image url
                required:
                  - name
                  - identifier
                  - has_trading_incentive
              last:
                type: number
                description: coin ticker last price
              volume:
                type: number
                description: coin ticker volume
              cost_to_move_up_usd:
                type: number
                description: coin ticker cost to move up in usd
              cost_to_move_down_usd:
                type: number
                description: coin ticker cost to move down in usd
              converted_last:
                type: object
                description: coin ticker converted last price
                properties:
                  btc:
                    type: number
                  eth:
                    type: number
                  usd:
                    type: number
              converted_volume:
                type: object
                description: coin ticker converted volume
                properties:
                  btc:
                    type: number
                  eth:
                    type: number
                  usd:
                    type: number
              trust_score:
                type: string
                description: coin ticker trust score
              bid_ask_spread_percentage:
                type: number
                description: coin ticker bid ask spread percentage
              timestamp:
                type: string
                description: coin ticker timestamp
              last_traded_at:
                type: string
                description: coin ticker last traded timestamp
              last_fetch_at:
                type: string
                description: coin ticker last fetch timestamp
              is_anomaly:
                type: boolean
                description: coin ticker anomaly
              is_stale:
                type: boolean
                description: coin ticker stale
              trade_url:
                type: string
                description: coin ticker trade url
              token_info_url:
                type: string
                description: coin ticker token info url
                nullable: true
              coin_id:
                type: string
                description: coin ticker base currency coin ID
              target_coin_id:
                type: string
                description: coin ticker target currency coin ID
      example:
        name: Bitcoin
        tickers:
          - base: BTC
            target: USDT
            market:
              name: Binance
              identifier: binance
              has_trading_incentive: false
              logo: >-
                https://assets.coingecko.com/markets/images/52/small/binance.jpg?1706864274
            last: 69476
            volume: 20242.03975
            cost_to_move_up_usd: 19320706.3958517
            cost_to_move_down_usd: 16360235.3694131
            converted_last:
              btc: 1.000205
              eth: 20.291404
              usd: 69498
            converted_volume:
              btc: 20249
              eth: 410802
              usd: 1406996874
            trust_score: green
            bid_ask_spread_percentage: 0.010014
            timestamp: '2024-04-08T04:02:01+00:00'
            last_traded_at: '2024-04-08T04:02:01+00:00'
            last_fetch_at: '2024-04-08T04:03:00+00:00'
            is_anomaly: false
            is_stale: false
            trade_url: https://www.binance.com/en/trade/BTC_USDT?ref=37754157
            token_info_url: null
            coin_id: bitcoin
            target_coin_id: tether
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