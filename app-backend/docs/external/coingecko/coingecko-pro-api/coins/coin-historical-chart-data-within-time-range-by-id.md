# Coin Historical Chart Data within Time Range by ID

> This endpoint allows you to **get the historical chart data of a coin within certain time range in UNIX along with price, market cap and 24hr volume based on particular coin ID**

<Tip>
  ### Tips

  * You may obtain the coin ID (API ID) via several ways:
    * refers to respective coin page and find 'API ID'.
    * refers to [`/coins/list`](/reference/coins-list) endpoint.
    * refers to google sheets [here](https://docs.google.com/spreadsheets/d/1wTTuxXt8n9q7C4NDXqQpI3wpKu1_5bGVmP9Xz0XGSyU/edit?usp=sharing).
  * Supports ISO date strings (`YYYY-MM-DD` or\
    `YYYY-MM-DDTHH:MM`, recommended for best compatibility) or UNIX timestamps.
</Tip>

<Note>
  ### Note

  * You may leave the interval params as empty for automatic granularity:
    * 1 day from current time = **5-minutely** data
    * 1 day from any time (except current time) = **hourly** data
    * 2 - 90 days from any time = **hourly** data
    * above 90 days from any time = **daily** data (00:00 UTC)
  * For **non-Enterprise plan subscribers** who would like to get hourly data, please leave the interval params empty for auto granularity.
  * The **5-minutely** and **hourly** interval params are also exclusively available to **Enterprise plan subscribers**, bypassing auto-granularity:
    * `interval=5m`: 5-minutely historical data, supports up to **any 10 days** date range per request.
    * `interval=hourly`: hourly historical data, supports up to **any 100 days** date range per request.
  * Data availability:
    * `interval=5m`: Available from 9 February 2018 onwards.
    * `interval=hourly`: Available from 30 Jan 2018 onwards.
  * Cache / Update Frequency:\
    Based on days range (all the API plans)
    * 1 day = 30 seconds cache
    * 2 -90 days = 30 minutes cache
    * 90 days = 12 hours cache
    * The last completed UTC day (00:00) is available 35 minutes after midnight on the next UTC day (00:35). The cache will always expire at 00:40 UTC.
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /coins/{id}/market_chart/range
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
  /coins/{id}/market_chart/range:
    get:
      tags:
        - Coins
      summary: Coin Historical Chart Data within Time Range by ID
      description: >-
        This endpoint allows you to **get the historical chart data of a coin
        within certain time range in UNIX along with price, market cap and 24hr
        volume based on particular coin ID**
      operationId: coins-id-market-chart-range
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
        - name: from
          in: query
          description: >-
            starting date in ISO date string (`YYYY-MM-DD` or
            `YYYY-MM-DDTHH:MM`) or UNIX timestamp. 
             **use ISO date string for best compatibility**
          required: true
          schema:
            type: string
            default: '2024-01-01'
          examples:
            iso_date:
              value: '2024-01-01'
            iso_datetime:
              value: 2024-01-01T00:00
            unix_timestamp:
              value: '1609459200'
        - name: to
          in: query
          description: >-
            ending date in ISO date string (`YYYY-MM-DD` or `YYYY-MM-DDTHH:MM`)
            or UNIX timestamp. 
             **use ISO date string for best compatibility**
          required: true
          schema:
            type: string
            default: '2024-12-31'
          examples:
            iso_date:
              value: '2024-12-31'
            iso_datetime:
              value: 2024-12-31T23:59
            unix_timestamp:
              value: '1640995200'
        - name: interval
          in: query
          description: 'data interval, leave empty for auto granularity '
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
                $ref: '#/components/schemas/CoinsMarketChartRange'
components:
  schemas:
    CoinsMarketChartRange:
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
          - - 1704067241331
            - 42261.0406175669
          - - 1704070847420
            - 42493.2764087546
          - - 1704074443652
            - 42654.0731066594
        market_caps:
          - - 1704067241331
            - 827596236151.196
          - - 1704070847420
            - 831531023621.411
          - - 1704074443652
            - 835499399014.932
        total_volumes:
          - - 1704067241331
            - 14305769170.9498
          - - 1704070847420
            - 14130205376.1709
          - - 1704074443652
            - 13697382902.2424
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