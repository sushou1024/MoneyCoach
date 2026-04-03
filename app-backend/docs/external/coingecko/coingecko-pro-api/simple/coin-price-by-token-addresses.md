# Coin Price by Token Addresses

> This endpoint allows you to **query one or more token prices using their token contract addresses**

<Tip>
  ### Tips

  * You may obtain the asset platform and contract address via several ways:
    * refers to respective coin page and find 'contract address'.
    * refers to [`/coins/list`](/reference/coins-list) endpoint (`include platform = true`).
  * You may flag to include more data such as market cap, 24hr volume, 24hr change, last updated time etc.
</Tip>

<Note>
  ### Note

  * The endpoint returns the global average price of the coin that is aggregated across all active exchanges on CoinGecko.
  * You may cross-check the price on [CoinGecko](https://www.coingecko.com) and learn more about our price methodology [here](https://www.coingecko.com/en/methodology).
  * Cache/Update Frequency: every 20 seconds for Pro API (Analyst, Lite, Pro, Enterprise).
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /simple/token_price/{id}
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
  /simple/token_price/{id}:
    get:
      tags:
        - Simple
      summary: Coin Price by Token Addresses
      description: >-
        This endpoint allows you to **query one or more token prices using their
        token contract addresses**
      operationId: simple-token-price
      parameters:
        - name: id
          in: path
          description: |-
            asset platform's ID 
             *refers to [`/asset_platforms`](/reference/asset-platforms-list).
          required: true
          schema:
            type: string
            example: ethereum
            default: ethereum
        - name: contract_addresses
          in: query
          description: >-
            the contract addresses of tokens, comma-separated if querying more
            than 1 token's contract address
          required: true
          schema:
            type: string
            default: '0x2260fac5e5542a773aa44fbcfedf7c193bc2c599'
          examples:
            one value:
              value: '0x2260fac5e5542a773aa44fbcfedf7c193bc2c599'
            multiple values:
              value: >-
                0x2260fac5e5542a773aa44fbcfedf7c193bc2c599,0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2
        - name: vs_currencies
          in: query
          description: >-
            target currency of coins, comma-separated if querying more than 1
            currency. 
             *refers to [`/simple/supported_vs_currencies`](/reference/simple-supported-currencies).
          required: true
          schema:
            type: string
            default: usd
          examples:
            one value:
              value: usd
            multiple values:
              value: usd,eur
        - name: include_market_cap
          in: query
          description: 'include market capitalization, default: false'
          required: false
          schema:
            type: boolean
        - name: include_24hr_vol
          in: query
          description: 'include 24hr volume, default: false'
          required: false
          schema:
            type: boolean
        - name: include_24hr_change
          in: query
          description: |-
            include 24hr change 
             default: false
          required: false
          schema:
            type: boolean
        - name: include_last_updated_at
          in: query
          description: 'include last updated price time in UNIX , default: false'
          required: false
          schema:
            type: boolean
        - name: precision
          in: query
          description: 'decimal place for currency price value '
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
          description: price(s) of cryptocurrency
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SimpleTokenPrice'
components:
  schemas:
    SimpleTokenPrice:
      type: object
      properties:
        usd:
          type: number
          description: price in USD
        usd_market_cap:
          type: number
          description: market cap in USD
        usd_24h_vol:
          type: number
          description: 24hr volume in USD
        usd_24h_change:
          type: number
          description: 24hr change in USD
        last_updated_at:
          type: number
          description: last updated timestamp
      example:
        '0x2260fac5e5542a773aa44fbcfedf7c193bc2c599':
          usd: 67187.3358936566
          usd_market_cap: 1317802988326.25
          usd_24h_vol: 31260929299.5248
          usd_24h_change: 3.63727894677354
          last_updated_at: 1711356300
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