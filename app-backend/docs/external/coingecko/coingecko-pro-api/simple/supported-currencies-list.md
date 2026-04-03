# Supported Currencies List

> This endpoint allows you to **query all the supported currencies on CoinGecko**

<Tip>
  ### Tips

  * You may use this endpoint to query the list of currencies for other endpoints that contain params like `vs_currencies`.
</Tip>

<Note>
  ### Note

  * Cache/Update Frequency: every 30 seconds for Pro API (Analyst, Lite, Pro, Enterprise).
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /simple/supported_vs_currencies
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
  /simple/supported_vs_currencies:
    get:
      tags:
        - Simple
      summary: Supported Currencies List
      description: >-
        This endpoint allows you to **query all the supported currencies on
        CoinGecko**
      operationId: simple-supported-currencies
      responses:
        '200':
          description: list of supported currencies
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CurrencyList'
components:
  schemas:
    CurrencyList:
      type: array
      items:
        type: string
      example:
        - btc
        - eth
        - ltc
        - bch
        - bnb
        - eos
        - xrp
        - xlm
        - link
        - dot
        - yfi
        - sol
        - usd
        - aed
        - ars
        - aud
        - bdt
        - bhd
        - bmd
        - brl
        - cad
        - chf
        - clp
        - cny
        - czk
        - dkk
        - eur
        - gbp
        - gel
        - hkd
        - huf
        - idr
        - ils
        - inr
        - jpy
        - krw
        - kwd
        - lkr
        - mmk
        - mxn
        - myr
        - ngn
        - nok
        - nzd
        - php
        - pkr
        - pln
        - rub
        - sar
        - sek
        - sgd
        - thb
        - try
        - twd
        - uah
        - vef
        - vnd
        - zar
        - xdr
        - xag
        - xau
        - bits
        - sats
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