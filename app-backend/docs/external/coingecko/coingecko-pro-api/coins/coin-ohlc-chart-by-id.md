# Coin OHLC Chart by ID

> This endpoint allows you to **get the OHLC chart (Open, High, Low, Close) of a coin based on particular coin ID**

<Tip>
  ### Tips

  * You may obtain the coin ID (API ID) via several ways:
    * refers to respective coin page and find 'API ID'.
    * refers to [`/coins/list`](/reference/coins-list) endpoint.
    * refers to Google Sheets [here](https://docs.google.com/spreadsheets/d/1wTTuxXt8n9q7C4NDXqQpI3wpKu1_5bGVmP9Xz0XGSyU/edit?usp=sharing).
  * For historical chart data with better granularity, you may consider using [/coins/\{id}/market\_chart](/reference/coins-id-market-chart) endpoint.
</Tip>

<Note>
  ### Note

  * The timestamp displayed in the payload (response) indicates the end (or close) time of the OHLC data.
  * Data granularity (candle's body) is automatic:
    * 1 - 2 days: 30 minutes
    * 3 - 30 days: 4 hours
    * 31 days and beyond: 4 days
  * Cache / Update Frequency:
    * Every 15 minutes for all the API plans
    * The last completed UTC day (00:00) is available 35 minutes after midnight on the next UTC day (00:35).
  * Exclusive **daily** and **hourly** candle interval parameter for all paid plan subscribers (`interval = daily`, `interval=hourly`)
    * '**daily**' interval is available for **1 / 7 / 14 / 30 / 90 / 180** days only.
    * '**hourly**' interval is available for  **1 / 7 / 14 / 30 / 90** days only.
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /coins/{id}/ohlc
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
  /coins/{id}/ohlc:
    get:
      tags:
        - Coins
      summary: Coin OHLC Chart by ID
      description: >-
        This endpoint allows you to **get the OHLC chart (Open, High, Low,
        Close) of a coin based on particular coin ID**
      operationId: coins-id-ohlc
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
            target currency of price data 
             *refers to [`/simple/supported_vs_currencies`](/reference/simple-supported-currencies).
          required: true
          schema:
            type: string
            example: usd
            default: usd
        - name: days
          in: query
          description: 'data up to number of days ago '
          required: true
          schema:
            type: string
            enum:
              - '1'
              - '7'
              - '14'
              - '30'
              - '90'
              - '180'
              - '365'
              - max
        - name: interval
          in: query
          description: data interval, leave empty for auto granularity
          required: false
          schema:
            type: string
            enum:
              - daily
              - hourly
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
          description: Get coin's OHLC
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CoinsOHLC'
components:
  schemas:
    CoinsOHLC:
      type: array
      items:
        type: array
        items:
          type: number
      example:
        - - 1709395200000
          - 61942
          - 62211
          - 61721
          - 61845
        - - 1709409600000
          - 61828
          - 62139
          - 61726
          - 62139
        - - 1709424000000
          - 62171
          - 62210
          - 61821
          - 62068
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