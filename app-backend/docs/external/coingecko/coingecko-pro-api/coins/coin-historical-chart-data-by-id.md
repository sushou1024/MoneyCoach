# Coin Historical Chart Data by ID

> This endpoint allows you to **get the historical chart data of a coin including time in UNIX, price, market cap and 24hr volume based on particular coin ID**

<Tip>
  ### Tips

  * You may obtain the coin ID (API ID) via several ways:
    * refers to respective coin page and find 'API ID'.
    * refers to [`/coins/list`](/reference/coins-list) endpoint.
    * refers to google sheets [here](https://docs.google.com/spreadsheets/d/1wTTuxXt8n9q7C4NDXqQpI3wpKu1_5bGVmP9Xz0XGSyU/edit?usp=sharing).
  * You may use tools like [epoch converter ](https://www.epochconverter.com) to convert human readable date to UNIX timestamp.
</Tip>

<Note>
  ### Note

  * You may leave the interval params as empty for automatic granularity:
    * 1 day from current time = **5-minutely** data
    * 2 - 90 days from current time = **hourly** data
    * above 90 days from current time = **daily** data (00:00 UTC)
  * For **non-Enterprise plan subscribers** who would like to get hourly data, please leave the interval params empty for auto granularity.
  * The **5-minutely** and **hourly** interval params are also exclusively available to **Enterprise plan subscribers,** bypassing auto-granularity:
    * `interval=5m`: 5-minutely historical data (responses include information from the past 10 days, up until now).
    * `interval=hourly`: hourly historical data (responses include information from the past 100 days, up until now).
  * Cache / Update Frequency:
    * Every 30 seconds for all the API plans (for last data point).
    * The last completed UTC day (00:00) data is available **10 minutes after midnight** on the next UTC day (00:10).
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /coins/{id}/market_chart
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
  /coins/{id}/market_chart:
    get:
      tags:
        - Coins
      summary: Coin Historical Chart Data by ID
      description: >-
        This endpoint allows you to **get the historical chart data of a coin
        including time in UNIX, price, market cap and 24hr volume based on
        particular coin ID**
      operationId: coins-id-market-chart
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
        - name: vs_currency
          in: query
          description: |-
            target currency of market data 
             *refers to [`/simple/supported_vs_currencies`](/reference/simple-supported-currencies).
          required: true
          schema:
            type: string
            example: usd
            default: usd
        - name: days
          in: query
          description: |-
            data up to number of days ago 
             You may use any integer or `max` for number of days
          required: true
          schema:
            type: string
            default: '1'
          examples:
            value-1:
              value: '1'
            value-2:
              value: max
        - name: interval
          in: query
          description: data interval, leave empty for auto granularity
          required: false
          schema:
            type: string
            enum:
              - 5m
              - hourly
              - daily
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
          description: >-
            Get historical market data include price, market cap, and 24hr
            volume (granularity auto)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CoinsMarketChart'
components:
  schemas:
    CoinsMarketChart:
      type: object
      properties:
        prices:
          type: array
          items:
            type: array
            items:
              type: number
        market_caps:
          type: array
          items:
            type: array
            items:
              type: number
        total_volumes:
          type: array
          items:
            type: array
            items:
              type: number
      example:
        prices:
          - - 1711843200000
            - 69702.3087473573
          - - 1711929600000
            - 71246.9514406015
          - - 1711983682000
            - 68887.7495158568
        market_caps:
          - - 1711843200000
            - 1370247487960.09
          - - 1711929600000
            - 1401370211582.37
          - - 1711983682000
            - 1355701979725.16
        total_volumes:
          - - 1711843200000
            - 16408802301.8374
          - - 1711929600000
            - 19723005998.215
          - - 1711983682000
            - 30137418199.6431
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